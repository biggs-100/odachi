// Package ask implements `odachi ask`: the three ways the agent records
// something in the current task document's history.
//
//   - a question only the human can answer  -> PREGUNTA P<n>
//   - the human's answer                    -> VOS P<n>
//   - a decision the agent took alone       -> OFICIO ... (porque ...)
//
// The history is append-only, so a run either adds one line or writes nothing.
// Every message is Spanish, one line, and names the flag that unblocks it.
package ask

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/biggs-100/odachi/internal/app/tasks"
	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// Exit codes (Contract IV).
const (
	// CodeOK means the record was written.
	CodeOK = 0
	// CodeBlocked means there is nothing to write to, or the flags contradict.
	CodeBlocked = 1
	// CodeNeedsAnswer means a human answer is missing and cannot be inferred.
	CodeNeedsAnswer = 2
)

// Options is one `odachi ask` invocation. Exactly one of Question, Answer or
// Craft is set: one call records one thing.
type Options struct {
	Dir string

	// Question is the text of a new open question.
	Question string
	// Answer closes an open question; For names which one when several are
	// open.
	Answer string
	For    string
	// Craft is a decision the agent took alone, and Because is why.
	Craft   string
	Because string
}

// Files is the disk port: enough to find, read and rewrite a task document.
type Files interface {
	List(dir string) ([]string, error)
	Read(path string) ([]byte, error)
	Write(path string, data []byte) error
}

// Result is what the CLI prints and exits with. It has the same shape as
// start.Result so the CLI reports every verb the same way; Lines carries a
// multi-line block when the verb has one.
type Result struct {
	Code    int
	Message string
	Lines   []string
	Task    string
	Files   []string
}

// The three forms of the command, once the flags are validated.
const (
	formQuestion = "question"
	formAnswer   = "answer"
	formCraft    = "craft"
)

// usage is the one-line reminder of the three shapes, repeated in errors
// because a blocked message must name the command that unblocks it.
const usage = `odachi ask "<pregunta>" · odachi ask --answer "<respuesta>" [--for P1] · odachi ask --craft "<decisión>" --because "<razón>"`

// Run performs one `odachi ask`.
func Run(opts Options, files Files) Result {
	if opts.Dir == "" {
		opts.Dir = "."
	}

	form, problem := validate(opts)
	if problem != nil {
		return *problem
	}

	current, err := tasks.Current(files, opts.Dir)
	if err != nil {
		return blocked(oneLine(err.Error()))
	}

	switch form {
	case formQuestion:
		return record(current, files, current.Doc.AppendHistory(questionLine(current.Doc, opts.Question)),
			fmt.Sprintf("abrí la pregunta %s en %s; contestala con odachi ask --answer \"<respuesta>\" --for %s",
				nextQuestionID(current.Doc), current.Rel, nextQuestionID(current.Doc)))
	case formCraft:
		line := fmt.Sprintf("L%d · OFICIO · %s (porque %s)",
			current.Doc.NextLine(), quoted(opts.Craft), oneLine(opts.Because))
		return record(current, files, current.Doc.AppendHistory(line),
			"anoté la decisión de oficio en "+current.Rel)
	default:
		return answer(current, files, opts)
	}
}

// validate picks the form and rejects the combinations that record nothing. A
// run with no argument at all needs a human, so it is exit 2 like `odachi
// start` with no answers; a contradictory call is exit 1.
func validate(opts Options) (string, *Result) {
	question := strings.TrimSpace(opts.Question)
	answer := strings.TrimSpace(opts.Answer)
	craft := strings.TrimSpace(opts.Craft)
	because := strings.TrimSpace(opts.Because)
	which := strings.TrimSpace(opts.For)

	given := 0
	for _, value := range []string{question, answer, craft} {
		if value != "" {
			given++
		}
	}

	switch {
	case given > 1:
		return "", failure(blocked("usá una sola forma de odachi ask por vez: " + usage))
	case craft != "" && because == "":
		return "", failure(blocked(`falta --because: una decisión sin razón no enseña nada; corré odachi ask --craft "<decisión>" --because "<razón>"`))
	case craft == "" && because != "":
		return "", failure(blocked(`--because solo acompaña a --craft: corré odachi ask --craft "<decisión>" --because "<razón>"`))
	case which != "" && answer == "":
		return "", failure(blocked(`--for solo acompaña a --answer: corré odachi ask --answer "<respuesta>" --for ` + which))
	case given == 0 && opts.Question == "":
		return "", failure(Result{Code: CodeNeedsAnswer, Message: oneLine(
			"falta lo que querés anotar: " + usage)})
	case given == 0:
		return "", failure(blocked(`la pregunta no puede quedar vacía: es lo que el humano decide; escribila en tus palabras (odachi ask "<pregunta>")`))
	case craft != "":
		return formCraft, nil
	case answer != "":
		return formAnswer, nil
	default:
		return formQuestion, nil
	}
}

// answer closes the open question the human just answered. With one open
// question it needs no --for; with two or more it never guesses.
func answer(current tasks.Document, files Files, opts Options) Result {
	open := current.Doc.OpenQuestions()
	if len(open) == 0 {
		return blocked(fmt.Sprintf("%s no tiene preguntas abiertas; abrí una con odachi ask \"<pregunta>\"", current.Rel))
	}

	id := strings.ToUpper(strings.TrimSpace(opts.For))
	if id == "" {
		if len(open) > 1 {
			return blocked(fmt.Sprintf("%s tiene %d preguntas abiertas; decime cuál contestaste con %s",
				current.Rel, len(open), forChoices(open)))
		}
		id = open[0].ID
	}

	closed, err := current.Doc.CloseQuestion(id, opts.Answer)
	if err != nil {
		return blocked(oneLine(err.Error()) + " (en " + current.Rel + ")")
	}
	return record(current, files, closed, fmt.Sprintf("cerré la pregunta %s en %s", id, current.Rel))
}

// record rewrites the task document with the grown history and reports where it
// happened. It refuses to rewrite a document it would damage: the log is the
// memory of the work, and losing it is worse than not recording.
func record(current tasks.Document, files Files, grown taskdoc.TaskDoc, message string) Result {
	rendered := grown.Render()

	if dropped := firstDroppedLine(current.Raw, rendered); dropped != "" {
		return blocked(fmt.Sprintf(
			"no anoté nada en %s: reescribirlo perdería la línea %q, que no está en el formato canónico; pasá esas notas a ## Historial y volvé a intentar",
			current.Rel, dropped))
	}
	if err := files.Write(current.Path, []byte(rendered)); err != nil {
		return blocked(fmt.Sprintf("no pude escribir %s: %s; revisá los permisos y volvé a intentar",
			current.Rel, oneLine(err.Error())))
	}

	return Result{Code: CodeOK, Message: message, Task: current.Rel, Files: []string{current.Rel}}
}

// questionLine is the PREGUNTA shape of Contract IV.
func questionLine(doc taskdoc.TaskDoc, text string) string {
	return fmt.Sprintf("L%d · PREGUNTA %s · %s", doc.NextLine(), nextQuestionID(doc), quoted(text))
}

// nextQuestionID is the P number this document has not used yet.
func nextQuestionID(doc taskdoc.TaskDoc) string {
	return "P" + strconv.Itoa(doc.NextQuestion())
}

// forChoices names the open questions the human can pick with --for.
func forChoices(open []taskdoc.Question) string {
	ids := make([]string, 0, len(open))
	for _, question := range open {
		ids = append(ids, question.ID)
	}
	return "--for " + ids[0] + " (las abiertas son " + strings.Join(ids, ", ") + ")"
}

// firstDroppedLine returns the first non-empty line of before that would not
// survive rendering, or "" when the rewrite loses nothing. Rendering is
// canonical: anything the agent or the human wrote outside the canonical
// sections is not carried over, so the caller must not throw it away silently.
func firstDroppedLine(before, after string) string {
	present := map[string]bool{}
	for _, line := range strings.Split(after, "\n") {
		present[strings.TrimSpace(line)] = true
	}
	for _, line := range strings.Split(before, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if !present[trimmed] {
			return trimmed
		}
	}
	return ""
}

// quoted renders text inside the canonical quotes, on one line.
func quoted(text string) string {
	return `"` + oneLine(text) + `"`
}

// failure boxes a result for the internal call chain.
func failure(result Result) *Result {
	return &result
}

// blocked is a Result with exit code 1.
func blocked(message string) Result {
	return Result{Code: CodeBlocked, Message: oneLine(message)}
}

// oneLine keeps every message to a single line, whatever came in from a flag.
func oneLine(message string) string {
	return strings.Join(strings.Fields(message), " ")
}
