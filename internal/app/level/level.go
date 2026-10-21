// Package level implements `odachi level`: the facts about how far the human
// has come, the graduation proposal when it is earned, and the one command that
// accepts it.
//
// It never rewrites the profile on its own. `level --to <profile>` is the human
// accepting the proposal, and that is the only path that writes.
package level

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/biggs-100/odachi/internal/app/start"
	"github.com/biggs-100/odachi/internal/app/tasks"
	"github.com/biggs-100/odachi/internal/domain/profile"
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
const profileFile = ".odachi/profile.toml"

// Options is one `odachi level` invocation. Current is the profile in force,
// already resolved by the caller from .odachi/profile.toml or from inference.
type Options struct {
	Dir string

	// Current is the profile the project is running under.
	Current profile.Profile
	// To is the profile the human accepted with --to. Empty means "show me".
	To string
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

// evidence is what the project shows: the three numbers the tour of the tool
// reports and the two the graduation is made of.
type evidence struct{ features, answers, craft, open int }

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
		name = "(sin perfil)"
	}

	line := fmt.Sprintf("perfil: %s · features con pasos: %d · preguntas contestadas: %d · decisiones de oficio: %d · preguntas abiertas: %d",
		name, e.features, e.answers, e.craft, e.open)
	lines := []string{line}

	earned := profile.Evidence{Features: e.features, VisionAnswers: e.answers, CraftDecisions: e.craft}
	switch next, ok := profile.Next(name, earned); {
	case ok:
		lines = append(lines, fmt.Sprintf(
			"ganaste el perfil %s: %d features con pasos hechos y %d preguntas contestadas; para aceptarlo corré odachi level --to %s",
			next, e.features, e.answers, next))
	case name != "novato":
		// Only a novato graduates: there is nothing to propose to someone who
		// is already past the first step of the ladder.
		lines = append(lines, fmt.Sprintf("ya estás en %s: no hay nada que proponerte", name))
	default:
		lines = append(lines, "todavía no está ganado el perfil senior: seguí sumando features con pasos hechos y preguntas contestadas")
	}

	return Result{Code: CodeOK, Message: line, Lines: lines, Profile: name}
}

// apply rewrites .odachi/profile.toml with the profile the human accepted,
// backing up the previous file first and reporting every field that changed.
func apply(opts Options, target string, files Files) Result {
	// The built-in chain plus the profile the human already runs: a project that
	// named its own profile can still graduate from it.
	known := profile.Builtin()
	if name := strings.TrimSpace(opts.Current.Name); name != "" {
		known[name] = opts.Current
	}

	to, err := profile.Resolve(known, target)
	if err != nil {
		return blocked(oneLine(err.Error()) + "; volvé a correr odachi level --to <perfil> con uno de esos nombres")
	}

	_, changes := profile.Apply(opts.Current, to)
	if len(changes) == 0 {
		return Result{Code: CodeOK, Message: fmt.Sprintf("el perfil ya era %s; no escribí nada", to.Name), Profile: to.Name}
	}

	path := filepath.Join(opts.Dir, profileFile)
	exists, err := files.Exists(path)
	if err != nil {
		return blocked(fmt.Sprintf("no pude revisar %s: %s; revisá los permisos y volvé a correr odachi level --to %s",
			profileFile, oneLine(err.Error()), to.Name))
	}

	backup := ""
	if exists {
		// The human's file is copied aside before it is replaced, always.
		backup, err = files.Backup(path)
		if err != nil {
			return blocked(fmt.Sprintf("no pude respaldar %s antes de reescribirlo: %s; revisá los permisos y volvé a correr odachi level --to %s",
				profileFile, oneLine(err.Error()), to.Name))
		}
	}

	if err := files.Write(path, []byte(start.RenderProfile(to))); err != nil {
		return blocked(fmt.Sprintf("no pude escribir %s: %s; revisá los permisos y volvé a correr odachi level --to %s",
			profileFile, oneLine(err.Error()), to.Name))
	}

	message := fmt.Sprintf("listo: apliqué el perfil %s en %s", to.Name, profileFile)
	if backup != "" {
		message += "; el anterior quedó en " + relative(opts.Dir, backup)
	}

	lines := make([]string, 0, len(changes)+1)
	lines = append(lines, message)
	for _, change := range changes {
		lines = append(lines, fmt.Sprintf("%s: %s -> %s", change.Field, shown(change.From), shown(change.To)))
	}

	return Result{Code: CodeOK, Message: message, Lines: lines, Profile: to.Name, Files: []string{profileFile}}
}

// shown renders an empty field value the way the human sees it in the file.
func shown(value string) string {
	if value == "" {
		return "(vacío)"
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
