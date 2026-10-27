// Package start implements `odachi start`: the interview that turns a human's
// answers into the four artifacts (brief, profile, fence and task document).
//
// It is pure orchestration. It declares two ports — Interview (the human) and
// Files (the disk) — and never touches os: the adapters in cmd/odachi do the
// I/O and the domain packages own the artifact formats.
//
// Exit codes are the CLI contract: 0 artifacts written, 1 blocked, 2 a human
// answer is missing. Every message is English, one line, and names the next
// action when the run is blocked.
package start

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/biggs-100/odachi/internal/app/lock"
	"github.com/biggs-100/odachi/internal/domain/brief"
	"github.com/biggs-100/odachi/internal/domain/profile"
)

// Exit codes (Contract III).
const (
	// CodeOK means the artifacts were written.
	CodeOK = 0
	// CodeBlocked means an invalid flag or an existing brief.md without --force.
	CodeBlocked = 1
	// CodeNeedsAnswer means a human answer is missing and cannot be asked here.
	CodeNeedsAnswer = 2
)

// The artifact paths, relative to the project root.
const (
	briefFile   = "brief.md"
	taskDir     = "odd/tasks"
	profileFile = ".odachi/profile.toml"
	fenceFile   = ".odachi/fence.toml"
)

// Question is one interview question. Why is the one-line explanation the tutor
// shows with the question, never as a separate message.
type Question struct {
	Text    string
	Why     string
	Options []string
}

// Interview is the human port. Ask returns the raw answer, Say shows one line.
type Interview interface {
	Ask(q Question) (string, error)
	Say(line string) error
}

// Files is the disk port. Write replaces a file atomically and creates its
// parent directory; Backup copies an existing file aside before a rewrite. The
// application calls Backup itself, so "never lose the human's file" is
// observable at this seam instead of hidden inside the adapter.
type Files interface {
	Exists(path string) (bool, error)
	Read(path string) ([]byte, error)
	List(dir string) ([]string, error)
	Backup(path string) (string, error)
	Write(path string, data []byte) error
}

// Options is one `odachi start` invocation. Every answer flag is optional: what
// is not answered on the command line is asked, unless the profile forbids it.
type Options struct {
	Dir string

	Outcome   string
	Audience  string
	Done      []string
	Exclude   []string
	Sensitive string
	Deadline  string
	Learn     []string

	BriefPath string
	Force     bool
	Machine   bool

	// Base is the commit this feature starts from, as the caller read it from
	// the repository. The task doc records it so the closing gate can name the
	// review of the feature instead of the commit that closes it. Empty in a
	// repository with no commit yet.
	Base string

	// WaiveReview writes the human's decision that this feature lands without a
	// review verdict. It is the one ceremony a small change may skip without
	// breaking a hard rule: the proof and the fence still apply, and the waiver
	// is recorded in the task doc so the gate names it instead of hiding it.
	WaiveReview bool

	// Profile is the already-resolved profile loaded by the caller from
	// .odachi/profile.toml. Nil means "infer it from what we can see".
	Profile *profile.Profile

	// Agents is the stored client selection, already resolved by the caller
	// from .odachi/agents.toml. Empty means both, the default of a project
	// that never chose.
	Agents string

	// Store is the absolute root of the clone's state, resolved by the caller
	// from the repository's git common directory. It is where the write lock
	// lives. Empty means the project is not a git repository: there is no store
	// to lock, and the run proceeds without the lock instead of inventing one.
	Store string
}

// Result is what the CLI prints and exits with.
type Result struct {
	Code    int
	Message string
	// Lines is the block a successful run prints, one item per line. It is
	// empty for a run with nothing extra to say, which prints Message.
	Lines   []string
	Profile string
	Task    string
	Files   []string
}

// verbName is the lock owner: what the next run reads when this one is still
// writing.
const verbName = "start"

// Run performs one `odachi start`.
func Run(opts Options, interview Interview, files Files, guard lock.Lock) Result {
	r := &runner{opts: opts, interview: interview, files: files}
	if r.opts.Dir == "" {
		r.opts.Dir = "."
	}

	// The lock comes before the first read that decides anything. These four
	// artifacts are written one by one, so two interleaved runs would leave every
	// file valid and the set incoherent; the defer is what releases it when the
	// run stops on its own rules, not only when it succeeds.
	//
	// A run outside a git repository has no store to hold a lock in. That is the
	// one case where this verb writes more than one file unlocked, and it is
	// named here instead of hidden: there is nowhere else to put the file.
	held := lock.Held{}
	if r.opts.Store != "" {
		held = lock.Take(guard, r.opts.Store, verbName)
	}
	if held.Refusal != "" {
		return blocked(held.Refusal)
	}
	defer held.Release()

	return withNotice(r.run(), held.Notice)
}

// run does the work once the lock is held.
func (r *runner) run() Result {
	fromFlags, err := answersFromFlags(r.opts)
	if err != nil {
		return blocked(err.Error())
	}
	r.answers = fromFlags

	if r.opts.BriefPath != "" {
		return r.runFromBrief()
	}
	return r.runFromInterview()
}

// withNotice puts the lock's announcement in front of the run's own answer: a
// run that took over an abandoned lock says so, on its own line when it
// succeeded and at the front of the message when it did not.
func withNotice(result Result, notice string) Result {
	if notice == "" {
		return result
	}

	if result.Code == CodeOK {
		result.Lines = append([]string{notice}, result.Message)
	}
	if result.Message == "" {
		result.Message = notice
		return result
	}
	result.Message = notice + " " + result.Message
	return result
}

// runner holds the mutable state of one run.
type runner struct {
	opts      Options
	interview Interview
	files     Files
	answers   answers
	profile   profile.Profile
	written   []string
}

// runFromInterview asks what the flags did not answer, then writes the four
// artifacts.
func (r *runner) runFromInterview() Result {
	if r.briefExists() && !r.opts.Force {
		return blocked(fmt.Sprintf(
			"there is already %s and I do not want to overwrite it without permission; run odachi start --force again (or delete %s)",
			briefFile, briefFile))
	}

	resolved, problem := r.resolveProfile()
	if problem != nil {
		return *problem
	}
	r.profile = resolved

	if problem := r.collect(); problem != nil {
		return *problem
	}
	return r.writeArtifacts(artifactInput{
		title:      taskSlug(r.answers.outcome),
		nonGoals:   r.answers.exclude,
		writeBrief: true,
	})
}

// runFromBrief uses an existing brief instead of asking, and never rewrites it:
// the human already wrote that file.
func (r *runner) runFromBrief() Result {
	if r.hasAnswerFlags() {
		return blocked("do not mix --brief with --outcome, --audience, --done, --exclude, --sensitive, --deadline or --learn; use one of the two (odachi start --brief <file>)")
	}

	loaded, problem := r.loadBrief()
	if problem != nil {
		return *problem
	}

	resolved, problem := r.resolveProfile()
	if problem != nil {
		return *problem
	}
	r.profile = resolved
	r.answers = answersFromBrief(loaded)

	title := strings.TrimSpace(loaded.Name)
	if len(slugWords(title)) == 0 {
		title = taskSlug(r.answers.outcome)
	}

	return r.writeArtifacts(artifactInput{
		title:      title,
		nonGoals:   loaded.NonGoals,
		writeBrief: false,
	})
}

// resolveProfile loads the profile the caller supplied or infers it from what
// the tool can observe. A profile changes ceremony, never rigor, so the only
// thing read here is how much to ask.
func (r *runner) resolveProfile() (profile.Profile, *Result) {
	if r.opts.Profile != nil {
		if err := profile.Validate(*r.opts.Profile); err != nil {
			return profile.Profile{}, failure(blocked("the profile is not valid: " + oneLine(err.Error()) + "; fix .odachi/profile.toml and run odachi start again"))
		}
		return *r.opts.Profile, nil
	}

	hasBrief, err := r.files.Exists(filepath.Join(r.opts.Dir, briefFile))
	if err != nil {
		return profile.Profile{}, failure(blocked("could not read the project directory: " + oneLine(err.Error()) + "; check the permissions and run odachi start again"))
	}

	name := profile.Infer(profile.Facts{HasBrief: hasBrief, Machine: r.opts.Machine})
	resolved, err := profile.Resolve(profile.Builtin(), name)
	if err != nil {
		return profile.Profile{}, failure(blocked("could not resolve the profile " + name + ": " + oneLine(err.Error())))
	}
	return resolved, nil
}

// collect asks for every answer that is still missing, following the pacing of
// the profile. It never asks when the run cannot reach a human.
func (r *runner) collect() *Result {
	if len(r.answers.missing()) == 0 {
		return nil
	}

	pacing := r.profile.Questions
	if pacing == "" {
		pacing = "one-by-one"
	}
	if pacing == "all" || r.interview == nil {
		return failure(r.needsAnswer())
	}

	for _, spec := range questionSpecs() {
		if r.answers.has(spec.field) {
			continue
		}
		value, problem := r.ask(spec)
		if problem != nil {
			return problem
		}
		r.set(spec.field, value)
	}

	// A senior sees all seven answers as one block before anything is written.
	if pacing == "batched" {
		r.showAnswers()
	}
	return nil
}

// ask runs the validation loop for one question. The write path is lenient: a
// human who types something unexpected is told what fits and asked again.
func (r *runner) ask(spec questionSpec) (string, *Result) {
	question := Question{Text: spec.text, Why: spec.why, Options: spec.options}

	for {
		raw, err := r.interview.Ask(question)
		if err != nil {
			// A terminal that died mid-question and a run with no terminal are
			// the same state: the human is not there. The actionable answer is
			// the flags that unblock the run, not the adapter's read error.
			return "", failure(r.needsAnswer())
		}
		value := strings.TrimSpace(raw)

		switch spec.field {
		case fieldOutcome:
			if len(slugWords(value)) == 0 {
				if problem := r.say(retryOutcome); problem != nil {
					return "", problem
				}
				continue
			}
			return value, nil
		case fieldAudience:
			if normalized, ok := normalizeAudience(value); ok {
				return normalized, nil
			}
			if problem := r.say("did not understand " + quote(value) + ": write 1 (phone), 2 (computer) or 3 (both)"); problem != nil {
				return "", problem
			}
		case fieldDone:
			if value == "" {
				if problem := r.say("I need at least one way to know it worked; write it on one line"); problem != nil {
					return "", problem
				}
				continue
			}
			return value, nil
		case fieldSensitive:
			if normalized, ok := normalizeSensitive(value); ok {
				return normalized, nil
			}
			if problem := r.say("did not understand " + quote(value) + ": write 1 (none), 2 (names), 3 (health) or 4 (money)"); problem != nil {
				return "", problem
			}
		case fieldDeadline:
			if normalized, ok := normalizeDeadline(value); ok {
				return normalized, nil
			}
			if problem := r.say("did not understand " + quote(value) + ": write 1 (today), 2 (this week) or 3 (no rush)"); problem != nil {
				return "", problem
			}
		default: // fieldExclude, fieldLearn: free text, empty is allowed
			return value, nil
		}
	}
}

// say shows one line to the human. A closed output is a missing human answer,
// not a crash.
func (r *runner) say(line string) *Result {
	if err := r.interview.Say(line); err != nil {
		return failure(Result{Code: CodeNeedsAnswer, Message: oneLine(
			"could not show the question: " + err.Error() + "; run odachi start again in a terminal")})
	}
	return nil
}

// set stores a raw answer in the field it belongs to.
func (r *runner) set(f field, value string) {
	switch f {
	case fieldOutcome:
		r.answers.outcome = value
	case fieldAudience:
		r.answers.audience = value
	case fieldDone:
		r.answers.done = appendUnique(r.answers.done, value)
	case fieldExclude:
		for _, item := range splitList(value) {
			if len(slugWords(item)) > 0 {
				r.answers.exclude = appendUnique(r.answers.exclude, item)
			}
		}
	case fieldSensitive:
		r.answers.sensitive = value
	case fieldDeadline:
		r.answers.deadline = value
	case fieldLearn:
		for _, item := range splitList(value) {
			r.answers.learn = appendUnique(r.answers.learn, item)
		}
	}
}

// showAnswers prints the seven answers together, before anything is written.
func (r *runner) showAnswers() {
	lines := []string{
		"this is how your answers look; if they are right, I write the files",
		"1 · what you want to exist: " + r.answers.outcome,
		"2 · who uses it: " + audienceLabel(r.answers.audience),
		"3 · how you know it worked: " + strings.Join(r.answers.done, "; "),
		"4 · what it does not include: " + orNone(r.answers.exclude),
		"5 · sensitive data: " + sensitiveLabel(r.answers.sensitive),
		"6 · time: " + deadlineLabel(r.answers.deadline),
		"7 · what you want to learn: " + orNone(r.answers.learn),
	}
	for _, line := range lines {
		if problem := r.say(line); problem != nil {
			return
		}
	}
}

// needsAnswer reports every flag that would unblock a run with no human.
func (r *runner) needsAnswer() Result {
	flags := readableFlags(r.answers.missing())
	return Result{Code: CodeNeedsAnswer, Message: oneLine(
		"answers are missing and this run cannot ask: " +
			strings.Join(flags, ", ") +
			"; run odachi start again with those flags (or pass --brief <file>)")}
}

// briefExists reports whether brief.md is already there.
func (r *runner) briefExists() bool {
	exists, err := r.files.Exists(filepath.Join(r.opts.Dir, briefFile))
	return err == nil && exists
}

// hasAnswerFlags reports whether the caller also supplied answers on the
// command line, which would otherwise be silently ignored next to --brief.
func (r *runner) hasAnswerFlags() bool {
	return r.opts.Outcome != "" ||
		r.opts.Audience != "" ||
		len(r.opts.Done) > 0 ||
		len(r.opts.Exclude) > 0 ||
		r.opts.Sensitive != "" ||
		r.opts.Deadline != "" ||
		len(r.opts.Learn) > 0
}

// loadBrief reads and parses the brief named by --brief.
func (r *runner) loadBrief() (brief.Brief, *Result) {
	path := r.opts.BriefPath
	if !filepath.IsAbs(path) {
		path = filepath.Join(r.opts.Dir, path)
	}

	data, err := r.files.Read(path)
	if err != nil {
		return brief.Brief{}, failure(blocked(fmt.Sprintf(
			"could not read the brief %s: %s; check --brief and run odachi start again",
			r.opts.BriefPath, oneLine(err.Error()))))
	}

	loaded := brief.Parse(string(data))
	if len(loaded.Outcome) == 0 {
		return brief.Brief{}, failure(blocked(fmt.Sprintf(
			`the brief %s has no "## What must exist"; add it and run odachi start --brief %s again`,
			r.opts.BriefPath, r.opts.BriefPath)))
	}
	return loaded, nil
}

// failure boxes a blocked/needs-answer result for the internal call chain.
func failure(result Result) *Result {
	return &result
}

// blocked is a Result with exit code 1.
func blocked(message string) Result {
	return Result{Code: CodeBlocked, Message: oneLine(message)}
}

// oneLine keeps every message to a single line, whatever an adapter said.
func oneLine(message string) string {
	return strings.Join(strings.Fields(message), " ")
}
