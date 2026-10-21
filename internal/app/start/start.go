// Package start implements `odachi start`: the interview that turns a human's
// answers into the four artifacts (brief, profile, fence and task document).
//
// It is pure orchestration. It declares two ports — Interview (the human) and
// Files (the disk) — and never touches os: the adapters in cmd/odachi do the
// I/O and the domain packages own the artifact formats.
//
// Exit codes are the CLI contract: 0 artifacts written, 1 blocked, 2 a human
// answer is missing. Every message is Spanish, one line, and names the next
// action when the run is blocked.
package start

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/biggz/odachi/internal/domain/brief"
	"github.com/biggz/odachi/internal/domain/profile"
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

	// Profile is the already-resolved profile loaded by the caller from
	// .odachi/profile.toml. Nil means "infer it from what we can see".
	Profile *profile.Profile
}

// Result is what the CLI prints and exits with.
type Result struct {
	Code    int
	Message string
	Profile string
	Task    string
	Files   []string
}

// Run performs one `odachi start`.
func Run(opts Options, interview Interview, files Files) Result {
	r := &runner{opts: opts, interview: interview, files: files}
	if r.opts.Dir == "" {
		r.opts.Dir = "."
	}

	fromFlags, err := answersFromFlags(opts)
	if err != nil {
		return blocked(err.Error())
	}
	r.answers = fromFlags

	if opts.BriefPath != "" {
		return r.runFromBrief()
	}
	return r.runFromInterview()
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
			"ya existe %s y no quiero pisarlo sin permiso; volvé a correr odachi start --force (o borrá %s)",
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
		return blocked("no mezcles --brief con --outcome, --audience, --done, --exclude, --sensitive, --deadline ni --learn; usá uno de los dos (odachi start --brief <archivo>)")
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
			return profile.Profile{}, failure(blocked("el perfil no es válido: " + oneLine(err.Error()) + "; corregí .odachi/profile.toml y volvé a correr odachi start"))
		}
		return *r.opts.Profile, nil
	}

	hasBrief, err := r.files.Exists(filepath.Join(r.opts.Dir, briefFile))
	if err != nil {
		return profile.Profile{}, failure(blocked("no pude leer el directorio del proyecto: " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi start"))
	}
	hasRepo, err := r.files.Exists(filepath.Join(r.opts.Dir, ".git"))
	if err != nil {
		return profile.Profile{}, failure(blocked("no pude leer el directorio del proyecto: " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi start"))
	}

	name := profile.Infer(profile.Facts{HasBrief: hasBrief, HasRepo: hasRepo, Machine: r.opts.Machine})
	resolved, err := profile.Resolve(profile.Builtin(), name)
	if err != nil {
		return profile.Profile{}, failure(blocked("no pude resolver el perfil " + name + ": " + oneLine(err.Error())))
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
			return "", failure(Result{Code: CodeNeedsAnswer, Message: oneLine(
				"no pude leer la respuesta: " + err.Error() + "; volvé a correr odachi start y respondé las preguntas")})
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
			if problem := r.say("no entendí " + quote(value) + ": escribí 1 (celular), 2 (compu) o 3 (los dos)"); problem != nil {
				return "", problem
			}
		case fieldDone:
			if value == "" {
				if problem := r.say("necesito al menos una forma de saber que funcionó; escribila en una línea"); problem != nil {
					return "", problem
				}
				continue
			}
			return value, nil
		case fieldSensitive:
			if normalized, ok := normalizeSensitive(value); ok {
				return normalized, nil
			}
			if problem := r.say("no entendí " + quote(value) + ": escribí 1 (ninguno), 2 (nombres), 3 (salud) o 4 (plata)"); problem != nil {
				return "", problem
			}
		case fieldDeadline:
			if normalized, ok := normalizeDeadline(value); ok {
				return normalized, nil
			}
			if problem := r.say("no entendí " + quote(value) + ": escribí 1 (hoy), 2 (esta semana) o 3 (sin apuro)"); problem != nil {
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
			"no pude mostrar la pregunta: " + err.Error() + "; volvé a correr odachi start en una terminal")})
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
		"así quedaron tus respuestas; si está bien, escribo los archivos",
		"1 · qué querés que exista: " + r.answers.outcome,
		"2 · quién lo usa: " + audienceLabel(r.answers.audience),
		"3 · cómo sabés que funcionó: " + strings.Join(r.answers.done, "; "),
		"4 · qué no incluye: " + orNone(r.answers.exclude),
		"5 · datos sensibles: " + sensitiveLabel(r.answers.sensitive),
		"6 · tiempo: " + deadlineLabel(r.answers.deadline),
		"7 · qué querés aprender: " + orNone(r.answers.learn),
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
		"faltan respuestas y esta corrida no puede preguntar: " +
			strings.Join(flags, ", ") +
			"; volvé a correr odachi start con esos flags (o pasá --brief <archivo>)")}
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
			"no pude leer el brief %s: %s; revisá --brief y volvé a correr odachi start",
			r.opts.BriefPath, oneLine(err.Error()))))
	}

	loaded := brief.Parse(string(data))
	if len(loaded.Outcome) == 0 {
		return brief.Brief{}, failure(blocked(fmt.Sprintf(
			`el brief %s no tiene "## Qué tiene que existir"; agregalo y volvé a correr odachi start --brief %s`,
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
