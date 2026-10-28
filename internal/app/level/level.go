// Package level implements `odachi level`: the facts about how far the human
// has come, the graduation proposal when it is earned, and the one command that
// accepts it.
//
// It never rewrites the profile on its own. `level --to <profile>` is the human
// accepting the proposal, and that is the only path that writes. It writes two
// things for one decision: the ceremony in odd/profile.toml and the voice of
// the agent in the last section of the AGENTS.md block. Every other byte of the
// block is preserved, so a human's lines and the rules survive.
package level

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/biggs-100/odachi/internal/app/start"
	"github.com/biggs-100/odachi/internal/app/tasks"
	"github.com/biggs-100/odachi/internal/domain/layout"
	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/profile"
	"github.com/biggs-100/odachi/internal/domain/rules"
	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// Exit codes (Contract IV).
const (
	// CodeOK means the facts were shown, or the profile was applied.
	CodeOK = 0
	// CodeBlocked means an unknown profile or a file that could not be written.
	CodeBlocked = 1
)

// profileFile is the human's profile, relative to the project root.
const profileFile = layout.Profile

// agentsFile is the file the rules block —and the voice, its last section— lives
// in. The renderer owns the name; this package only writes one section of it.
const agentsFile = rules.File

// Options is one `odachi level` invocation. Current is the profile in force,
// already resolved by the caller from odd/profile.toml or from inference.
type Options struct {
	Dir string

	// Current is the profile the project is running under.
	Current profile.Profile
	// To is the profile the human accepted with --to. Empty means "show me".
	To string
	// Agents is the stored client selection, already resolved by the caller
	// from the clone's store. Empty means both, the default of a project
	// that never chose.
	Agents string
}

// Files is the disk port: enough to read every feature, back up the human's
// profile and replace it.
type Files interface {
	Exists(path string) (bool, error)
	Read(path string) ([]byte, error)
	List(dir string) ([]string, error)
	Backup(path string) (string, error)
	Write(path string, data []byte) error
}

// Result is what the CLI prints and exits with. It has the same shape as
// start.Result; Lines carries the block a human reads.
type Result struct {
	Code    int
	Message string
	Lines   []string
	Profile string
	Task    string
	Files   []string
}

// Run performs one `odachi level`.
func Run(opts Options, files Files) Result {
	if opts.Dir == "" {
		opts.Dir = "."
	}

	if target := strings.TrimSpace(opts.To); target != "" {
		return apply(opts, target, files)
	}

	docs, err := tasks.All(files, opts.Dir)
	if err != nil {
		return blocked(oneLine(err.Error()))
	}
	return facts(opts.Current, count(docs))
}

// evidence is what the project shows: the four numbers the tour of the tool
// reports and the two the graduation is made of.
type evidence struct{ features, answers, changes, craft, open int }

// count reads every task document once. A feature counts when at least one of
// its steps is done: a document nobody advanced is a plan, not a feature.
func count(docs []tasks.Document) evidence {
	var e evidence
	for _, doc := range docs {
		if done, _ := doc.Doc.Progress(); done > 0 {
			e.features++
		}
		for _, event := range doc.Doc.Events() {
			switch event.Kind {
			case taskdoc.KindAnswer:
				e.answers++
			case taskdoc.KindChange:
				e.changes++
			case taskdoc.KindCraft:
				e.craft++
			}
		}
		e.open += len(doc.Doc.OpenQuestions())
	}
	return e
}

// facts shows how far the human has come, and offers the graduation when it is
// earned. It proposes; it never takes.
func facts(current profile.Profile, e evidence) Result {
	name := current.Name
	if name == "" {
		name = "(no profile)"
	}

	line := fmt.Sprintf("profile: %s · features with steps: %d · answered questions: %d · requirement changes: %d · craft decisions: %d · open questions: %d",
		name, e.features, e.answers, e.changes, e.craft, e.open)
	lines := []string{line}

	// A change is a vision decision: the human moved the requirement, which is
	// the strongest thing they can decide. It sums with the answers; a craft
	// decision still does not.
	vision := e.answers + e.changes
	earned := profile.Evidence{Features: e.features, VisionAnswers: vision, CraftDecisions: e.craft}
	switch next, ok := profile.Next(name, earned); {
	case ok:
		lines = append(lines, fmt.Sprintf(
			"you earned the %s profile: %d features with done steps and %d vision decisions (answered questions and requirement changes); to accept it run odachi level --to %s",
			next, e.features, vision, next))
	case name != "novato":
		// Only a novato graduates: there is nothing to propose to someone who
		// is already past the first step of the ladder.
		lines = append(lines, fmt.Sprintf("you are already at %s: there is nothing to propose", name))
	default:
		lines = append(lines, "the senior profile is not earned yet: keep adding features with done steps and vision decisions (answered questions or requirement changes)")
	}

	return Result{Code: CodeOK, Message: line, Lines: lines, Profile: name}
}

// apply writes the profile the human accepted and the voice that goes with it,
// backing up each file before it is replaced and reporting every field that
// changed.
//
// The two writes are one command because they are one choice: the profile decides
// the ceremony and the profile's name picks the voice, so a human who accepted
// senior must not keep reading an agent that talks to a novice. Each half is
// idempotent on its own — a profile that already is the target is not rewritten,
// and a voice section that already says the right thing is not touched — so the
// second run of the same command writes nothing at all.
func apply(opts Options, target string, files Files) Result {
	// The built-in chain plus the profile the human already runs: a project that
	// named its own profile can still graduate from it.
	known := profile.Builtin()
	if name := strings.TrimSpace(opts.Current.Name); name != "" {
		known[name] = opts.Current
	}

	to, err := profile.Resolve(known, target)
	if err != nil {
		return blocked(oneLine(err.Error()) + "; run odachi level --to <profile> again with one of those names")
	}

	lines := make([]string, 0, 4)
	written := make([]string, 0, 2)

	_, changes := profile.Apply(opts.Current, to)
	applied := len(changes) > 0
	if applied {
		path := filepath.Join(opts.Dir, profileFile)
		exists, err := files.Exists(path)
		if err != nil {
			return blocked(fmt.Sprintf("could not check %s: %s; check the permissions and run odachi level --to %s again",
				profileFile, oneLine(err.Error()), to.Name))
		}

		backup := ""
		if exists {
			// The human's file is copied aside before it is replaced, always.
			backup, err = files.Backup(path)
			if err != nil {
				return blocked(fmt.Sprintf("could not back up %s before rewriting it: %s; check the permissions and run odachi level --to %s again",
					profileFile, oneLine(err.Error()), to.Name))
			}
		}

		if err := files.Write(path, []byte(start.RenderProfile(to))); err != nil {
			return blocked(fmt.Sprintf("could not write %s: %s; check the permissions and run odachi level --to %s again",
				profileFile, oneLine(err.Error()), to.Name))
		}
		written = append(written, profileFile)
		if backup != "" {
			lines = append(lines, "the previous profile is at "+relative(opts.Dir, backup))
		}
		for _, change := range changes {
			lines = append(lines, fmt.Sprintf("%s: %s -> %s", change.Field, shown(change.From), shown(change.To)))
		}
	}

	// The voice: the last section of the rules block, the only part of it this
	// command owns, and the native agent that mirrors the block.
	voiceName, voiceErr := persona.VoiceName(to.Name)
	voiced := false
	personaFiles := []string{}
	switch {
	case voiceErr != nil:
		// A profile name odachi has no voice for. The ceremony is still the
		// human's decision, so the block is left alone and the line says which
		// names have a voice.
		lines = append(lines, "the block voice is unchanged: "+oneLine(voiceErr.Error())+"; use novato, senior or ci in "+profileFile)
	default:
		line, wrote, problem := rewriteVoice(opts.Dir, voiceName, files)
		if problem != nil {
			// The profile may already be written: say how much of the job is
			// done instead of pretending nothing happened.
			problem.Message = "applied the " + to.Name + " profile in " + profileFile + "; " + problem.Message
			problem.Lines = lines
			problem.Files = written
			return *problem
		}
		if wrote {
			voiced = true
			written = append(written, agentsFile)
		}
		if line != "" {
			lines = append(lines, line)
		}

		// The native agent mirrors the block the human already has, so the
		// selector and AGENTS.md cannot disagree after a profile change. A
		// missing file is init's to install, not level's: this command refreshes
		// what is there.
		refreshed, personaLine, problem := refreshPersonas(opts.Dir, files, opts.Agents)
		if problem != nil {
			problem.Message = "applied the " + to.Name + " profile in " + profileFile + "; " + problem.Message
			problem.Lines = lines
			problem.Files = append(append([]string{}, written...), refreshed...)
			return *problem
		}
		personaFiles = refreshed
		if personaLine != "" {
			lines = append(lines, personaLine)
		}
		written = append(written, refreshed...)
	}

	// One line, and it says what happened to every surface of the same choice.
	// A run that changed nothing says so instead of claiming work it did not do,
	// which is also what keeps the second run of the same command silent.
	message := fmt.Sprintf("the profile was already %s", to.Name)
	switch {
	case voiceErr != nil:
		message += "; I wrote nothing"
	case applied && voiced:
		message = fmt.Sprintf("done: applied the %s profile in %s and the block voice is %s", to.Name, profileFile, voiceName)
	case applied:
		message = fmt.Sprintf("done: applied the %s profile in %s; the block voice was already %s", to.Name, profileFile, voiceName)
	case voiced:
		message = fmt.Sprintf("done: the profile was already %s; I updated the block voice to %s", to.Name, voiceName)
	case len(personaFiles) > 0:
		message = fmt.Sprintf("done: the profile was already %s and the block voice is also %s; I updated the native agent", to.Name, voiceName)
	default:
		message = fmt.Sprintf("the profile was already %s and the block voice is also %s; I wrote nothing", to.Name, voiceName)
	}
	if len(personaFiles) > 0 && (applied || voiced) {
		message += "; I updated the native agent (" + strings.Join(personaFiles, ", ") + ")"
	}

	// The message goes first, because the printer prints Lines and not the
	// message when there are any: a detail without the sentence that explains it
	// is how a human reads "AGENTS.md: unchanged" and concludes nothing ran.
	if len(lines) > 0 {
		lines = append([]string{message}, lines...)
	}

	return Result{Code: CodeOK, Message: message, Lines: lines, Profile: to.Name, Files: written}
}

// rewriteVoice refreshes the voice section of the rules block in place: every
// other byte of AGENTS.md —the rules, the verbs, whatever the human wrote above
// and below the block— is preserved, the way the .gitignore block preserves its
// file. A block written before the voice moved into it gains the section.
//
// A project with no block is not an error here: the profile is still applied and
// the line names the command that writes the block. problem is for a file odachi
// could not read or write, where nothing was changed.
func rewriteVoice(dir, voiceName string, files Files) (line string, wrote bool, problem *Result) {
	path := filepath.Join(dir, agentsFile)
	missing := "AGENTS.md: there is no odachi block: the voice is not in the project; run odachi init"
	noFile := "AGENTS.md: there is no AGENTS.md: the voice is not in the project; run odachi init"

	exists, err := files.Exists(path)
	if err != nil {
		failure := blocked(fmt.Sprintf("could not check %s: %s; check the permissions and run odachi level --to %s again",
			agentsFile, oneLine(err.Error()), voiceName))
		return "", false, &failure
	}
	if !exists {
		return noFile, false, nil
	}

	data, err := files.Read(path)
	if err != nil {
		failure := blocked(fmt.Sprintf("could not read %s: %s; check the permissions and run odachi level --to %s again",
			agentsFile, oneLine(err.Error()), voiceName))
		return "", false, &failure
	}

	current := string(data)
	if !rules.HasBlock(current) {
		return missing, false, nil
	}

	updated, changed, err := rules.WithVoice(current, voiceName)
	if err != nil {
		failure := blocked(fmt.Sprintf("could not build the %s voice: %s; run odachi level --to %s again",
			agentsFile, oneLine(err.Error()), voiceName))
		return "", false, &failure
	}
	if !changed {
		// Nothing changed, so nothing to report: the message says the voice was
		// already the right one, and a second run writes no line either.
		return "", false, nil
	}

	backup, err := files.Backup(path)
	if err != nil {
		failure := blocked(fmt.Sprintf("could not back up %s before rewriting it: %s; check the permissions and run odachi level --to %s again",
			agentsFile, oneLine(err.Error()), voiceName))
		return "", false, &failure
	}
	if err := files.Write(path, []byte(updated)); err != nil {
		failure := blocked(fmt.Sprintf("could not write %s: %s; check the permissions and run odachi level --to %s again",
			agentsFile, oneLine(err.Error()), voiceName))
		return "", false, &failure
	}

	return "AGENTS.md: voice " + voiceName + " in the block (the previous one is at " + relative(dir, backup) + ")", true, nil
}

// refreshPersonas keeps the native agent in step with the block the human
// already has. It mirrors the block's own bytes, not a freshly rendered copy, so
// a human who edited the rules by hand gets the edit in the selector too.
//
// Only files that already exist are refreshed: installing them is `odachi init`'s
// job, and a level run on a project without them still applies the profile. A
// project without a block has nothing to mirror and is left alone.
func refreshPersonas(dir string, files Files, selection string) (changed []string, line string, problem *Result) {
	agentsPath := filepath.Join(dir, agentsFile)
	exists, err := files.Exists(agentsPath)
	if err != nil {
		failure := blocked(fmt.Sprintf("could not check %s: %s; check the permissions and run odachi level --to <profile> again",
			agentsFile, oneLine(err.Error())))
		return nil, "", &failure
	}
	if !exists {
		return nil, "", nil
	}

	data, err := files.Read(agentsPath)
	if err != nil {
		failure := blocked(fmt.Sprintf("could not read %s: %s; check the permissions and run odachi level --to <profile> again",
			agentsFile, oneLine(err.Error())))
		return nil, "", &failure
	}
	body, ok := rules.BodyIn(string(data))
	if !ok {
		return nil, "", nil
	}

	selected, err := persona.FilesForSelection(body, selection)
	if err != nil {
		failure := blocked(fmt.Sprintf("could not resolve the client selection: %s; run odachi init --agents both and run odachi level --to <profile> again",
			oneLine(err.Error())))
		return changed, "", &failure
	}
	for _, file := range selected {
		path := filepath.Join(dir, filepath.FromSlash(file.Path))

		exists, err := files.Exists(path)
		if err != nil {
			failure := blocked(fmt.Sprintf("could not check %s: %s; check the permissions and run odachi level --to <profile> again",
				file.Path, oneLine(err.Error())))
			return changed, "", &failure
		}
		if !exists {
			continue
		}

		current, err := files.Read(path)
		if err != nil {
			failure := blocked(fmt.Sprintf("could not read %s: %s; check the permissions and run odachi level --to <profile> again",
				file.Path, oneLine(err.Error())))
			return changed, "", &failure
		}
		if string(current) == file.Content {
			continue
		}

		if _, err := files.Backup(path); err != nil {
			failure := blocked(fmt.Sprintf("could not back up %s before rewriting it: %s; check the permissions and run odachi level --to <profile> again",
				file.Path, oneLine(err.Error())))
			return changed, "", &failure
		}
		if err := files.Write(path, []byte(file.Content)); err != nil {
			failure := blocked(fmt.Sprintf("could not write %s: %s; check the permissions and run odachi level --to <profile> again",
				file.Path, oneLine(err.Error())))
			return changed, "", &failure
		}
		changed = append(changed, file.Path)
	}

	if len(changed) == 0 {
		return nil, "", nil
	}
	return changed, "the native agent is up to date: " + strings.Join(changed, ", "), nil
}

// shown renders an empty field value the way the human sees it in the file.
func shown(value string) string {
	if value == "" {
		return "(empty)"
	}
	return value
}

// relative reports a path the way the human typed it: inside the project, not
// an absolute path from a temporary directory.
func relative(dir, path string) string {
	if rel, err := filepath.Rel(dir, path); err == nil {
		return filepath.ToSlash(rel)
	}
	return filepath.ToSlash(path)
}

// blocked is a Result with exit code 1.
func blocked(message string) Result {
	return Result{Code: CodeBlocked, Message: oneLine(message)}
}

// oneLine keeps every message to a single line, whatever an adapter said.
func oneLine(message string) string {
	return strings.Join(strings.Fields(message), " ")
}
