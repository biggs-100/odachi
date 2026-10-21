// Package plan implements `odachi plan`: adding the next step to the current
// task document and closing a step that is done.
//
// The agent used to add steps by editing Markdown by hand. A hand-written
// section that the canonical renderer does not know made every later `ask` fail
// and the project got stuck. `plan` closes that hole: the step is added to the
// parsed document and the whole file is written back through the canonical
// renderer, using the same write guard as `ask`.
package plan

import (
	"fmt"
	"strings"

	"github.com/biggs-100/odachi/internal/app/tasks"
	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// Exit codes (Contract IX). Unlike `ask`, plan never needs a human answer: it
// either writes or explains what is wrong.
const (
	// CodeOK means the document was written, or the step was already done.
	CodeOK = 0
	// CodeBlocked means the flags contradict, the text is invalid, or there is
	// nothing to plan into.
	CodeBlocked = 1
)

// Options is one `odachi plan` invocation. Either Text (add a step) or Done
// (close a step) is set, never both.
type Options struct {
	Dir string

	// Text is the new step's text.
	Text string
	// Requirement links the new step to an R the document declares.
	Requirement string
	// Done is the T id to mark as done.
	Done string
	// Commit is the commit that closed the step; it only makes sense with Done.
	Commit string
}

// Files is the disk port: enough to find the document, read it, back it up and
// replace it.
type Files interface {
	List(dir string) ([]string, error)
	Read(path string) ([]byte, error)
	Write(path string, data []byte) error
	Backup(path string) (string, error)
}

// Result is what the CLI prints and exits with. It has the same shape as
// start.Result so the CLI reports every verb the same way.
type Result struct {
	Code    int
	Message string
	Task    string
	Files   []string
}

// The two forms of the command, once the flags are validated.
const (
	formAdd  = "add"
	formDone = "done"
)

// usage is the one-line reminder of the two shapes, repeated in errors because
// a blocked message must name the command that unblocks it.
const usage = `odachi plan "<texto>" [--req R1] · odachi plan --done T2 [--commit abc1234]`

// Run performs one `odachi plan`.
func Run(opts Options, files Files) Result {
	if opts.Dir == "" {
		opts.Dir = "."
	}

	form, problem := validate(opts)
	if problem != "" {
		return blocked(problem)
	}

	current, err := tasks.Current(files, opts.Dir)
	if err != nil {
		return blocked(oneLine(err.Error()))
	}

	if form == formDone {
		return markDone(current, files, opts)
	}
	return add(current, files, opts)
}

// validate picks the form and rejects the combinations that plan nothing. A
// contradictory call is blocked before the document is even read.
func validate(opts Options) (string, string) {
	text := strings.TrimSpace(opts.Text)
	done := strings.TrimSpace(opts.Done)
	requirement := strings.TrimSpace(opts.Requirement)
	commit := strings.TrimSpace(opts.Commit)

	switch {
	case done == "" && commit != "":
		return "", `--commit solo acompaña a un paso que estás marcando: corré odachi plan --done T2 --commit <sha>`
	case text != "" && done != "":
		return "", "usá una sola forma de odachi plan por vez: " + usage
	case done != "" && requirement != "":
		return "", `--req solo acompaña al texto de un paso nuevo: corré odachi plan "<texto>" --req ` + requirement
	case done != "":
		return formDone, ""
	case text == "":
		return "", `falta el paso: escribí odachi plan "<texto>" para agregar uno o odachi plan --done T2 para marcar uno hecho`
	}
	return formAdd, ""
}

// add appends the next step. The number and the link are decided by the domain,
// and the write is canonical by construction.
func add(current tasks.Document, files Files, opts Options) Result {
	grown, err := current.Doc.AddStep(opts.Text, opts.Requirement)
	if err != nil {
		return blocked(oneLine(err.Error()) + " (en " + current.Rel + ")")
	}

	if err := tasks.SaveWithBackup(files, current, grown); err != nil {
		return blocked(err.Error())
	}

	step := grown.Steps[len(grown.Steps)-1]
	return Result{
		Code:    CodeOK,
		Message: fmt.Sprintf("paso %s agregado a %s", step.ID, current.Rel),
		Task:    current.Rel,
		Files:   []string{current.Rel},
	}
}

// markDone closes a step and records the commit that closed it. Closing a
// closed step is a no-op that writes nothing: the document is the memory of the
// work, and re-marking it would rewrite that memory for no reason.
func markDone(current tasks.Document, files Files, opts Options) Result {
	id := strings.ToUpper(strings.TrimSpace(opts.Done))

	step, ok := findStep(current.Doc, id)
	if !ok {
		return blocked(unknownStep(current, id))
	}
	// The document's own spelling wins: a hand-written `t2` stays `t2` when the
	// tool marks it, because Record matches the id it has.
	id = step.ID
	if step.Done {
		return Result{
			Code:    CodeOK,
			Message: fmt.Sprintf("paso %s ya estaba hecho", id),
			Task:    current.Rel,
		}
	}

	// A commit recorded on an open step is not thrown away when the caller does
	// not pass one: plan records, it never erases.
	commit := step.Commit
	if given := strings.TrimSpace(opts.Commit); given != "" {
		commit = given
	}

	grown, err := current.Doc.Record(id, commit)
	if err != nil {
		return blocked(oneLine(err.Error()) + " (en " + current.Rel + ")")
	}
	if err := tasks.SaveWithBackup(files, current, grown); err != nil {
		return blocked(err.Error())
	}

	message := fmt.Sprintf("paso %s hecho", id)
	if commit != "" {
		message += " · " + commit
	}
	return Result{
		Code:    CodeOK,
		Message: message,
		Task:    current.Rel,
		Files:   []string{current.Rel},
	}
}

// findStep returns the step with the given id, ignoring case.
func findStep(doc taskdoc.TaskDoc, id string) (taskdoc.Step, bool) {
	for _, step := range doc.Steps {
		if strings.EqualFold(step.ID, id) {
			return step, true
		}
	}
	return taskdoc.Step{}, false
}

// unknownStep explains an id the document does not have, naming the steps the
// agent can still close.
func unknownStep(current tasks.Document, id string) string {
	open := make([]string, 0, len(current.Doc.Steps))
	for _, step := range current.Doc.Steps {
		if !step.Done {
			open = append(open, step.ID)
		}
	}

	if len(open) == 0 {
		return fmt.Sprintf("paso desconocido: %s; este task doc no tiene pasos abiertos (en %s); agregá uno con odachi plan \"<texto>\"",
			id, current.Rel)
	}
	return fmt.Sprintf("paso desconocido: %s; los pasos abiertos son %s (en %s)", id, strings.Join(open, ", "), current.Rel)
}

// blocked is a Result with exit code 1.
func blocked(message string) Result {
	return Result{Code: CodeBlocked, Message: oneLine(message)}
}

// oneLine keeps every message to a single line, whatever came in from a flag.
func oneLine(message string) string {
	return strings.Join(strings.Fields(message), " ")
}
