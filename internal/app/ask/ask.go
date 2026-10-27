// Package ask implements `odachi ask`: the three ways the agent records
// something in the current task document's history.
//
//   - a question only the human can answer  -> PREGUNTA P<n>
//   - the human's answer                    -> VOS P<n>
//   - a decision the agent took alone       -> OFICIO ... (porque ...)
//
// The history is append-only, so a run either adds one line or writes nothing.
// Every message is English, one line, and names the flag that unblocks it.
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

// Options is one `odachi ask` invocation. Exactly one of Question, Answer,
// Craft or Change is set: one call records one thing.
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
	// Change is a requirement the human changed, in their words, and
	// Requirement names the R it changes: the steps that cite it reopen.
	Change      string
	Requirement string
	// Feature is the number of the task document to write to. Zero means the
	// highest, which is the feature the human is working on now.
	Feature int
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

// The four forms of the command, once the flags are validated.
const (
	formQuestion = "question"
	formAnswer   = "answer"
	formCraft    = "craft"
	formChange   = "change"
)

// usage is the one-line reminder of the four shapes, repeated in errors because
// a blocked message must name the command that unblocks it.
const usage = `odachi ask "<question>" · odachi ask --answer "<answer>" [--for P1] · odachi ask --craft "<decision>" --because "<reason>" · odachi ask --change "<words>" --req R1`

// Run performs one `odachi ask`.
func Run(opts Options, files Files) Result {
	if opts.Dir == "" {
		opts.Dir = "."
	}

	form, problem := validate(opts)
	if problem != nil {
		return *problem
	}

	current, err := tasks.Feature(files, opts.Dir, opts.Feature)
	if err != nil {
		return blocked(oneLine(err.Error()))
	}

	switch form {
	case formQuestion:
		return record(current, files, current.Doc.AppendHistory(questionLine(current.Doc, opts.Question)),
			fmt.Sprintf("opened question %s in %s; answer it with odachi ask --answer \"<answer>\" --for %s",
				nextQuestionID(current.Doc), current.Rel, nextQuestionID(current.Doc)))
	case formCraft:
		line := fmt.Sprintf("L%d · OFICIO · %s (porque %s)",
			current.Doc.NextLine(), quoted(opts.Craft), oneLine(opts.Because))
		return record(current, files, current.Doc.AppendHistory(line),
			"recorded the craft decision in "+current.Rel)
	case formChange:
		return change(current, files, opts)
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
	change := strings.TrimSpace(opts.Change)
	requirement := strings.TrimSpace(opts.Requirement)

	given := 0
	for _, value := range []string{question, answer, craft, change} {
		if value != "" {
			given++
		}
	}

	switch {
	case given > 1:
		return "", failure(blocked("use one form of odachi ask at a time: " + usage))
	case craft != "" && because == "":
		return "", failure(blocked(`missing --because: a decision without a reason teaches nothing; run odachi ask --craft "<decision>" --because "<reason>"`))
	case craft == "" && because != "":
		return "", failure(blocked(`--because only goes with --craft: run odachi ask --craft "<decision>" --because "<reason>"`))
	case change != "" && requirement == "":
		return "", failure(blocked(`missing --req: a change names the requirement it changes, and that is what reopens its steps; run odachi ask --change "<words>" --req R1`))
	case change == "" && requirement != "":
		return "", failure(blocked(`--req only goes with --change: run odachi ask --change "<words>" --req ` + requirement))
	case which != "" && answer == "":
		return "", failure(blocked(`--for only goes with --answer: run odachi ask --answer "<answer>" --for ` + which))
	case given == 0 && opts.Question == "":
		return "", failure(Result{Code: CodeNeedsAnswer, Message: oneLine(
			"missing what to record: " + usage)})
	case given == 0:
		return "", failure(blocked(`the question cannot be empty: it is what the human decides; write it in your words (odachi ask "<question>")`))
	case change != "":
		return formChange, nil
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
		return blocked(fmt.Sprintf("%s has no open questions; open one with odachi ask \"<question>\"", current.Rel))
	}

	id := strings.ToUpper(strings.TrimSpace(opts.For))
	if id == "" {
		if len(open) > 1 {
			return blocked(fmt.Sprintf("%s has %d open questions; tell me which one you answered with %s",
				current.Rel, len(open), forChoices(open)))
		}
		id = open[0].ID
	}

	closed, err := current.Doc.CloseQuestion(id, opts.Answer)
	if err != nil {
		return blocked(oneLine(err.Error()) + " (en " + current.Rel + ")")
	}
	return record(current, files, closed, fmt.Sprintf("closed question %s in %s", id, current.Rel))
}

// change records a requirement the human changed. Their words go to the history
// as given, and every done step that cites the requirement reopens: a step that
// claimed it cannot stay closed over a requirement that moved.
func change(current tasks.Document, files Files, opts Options) Result {
	requirement := strings.ToUpper(strings.TrimSpace(opts.Requirement))
	if !current.Doc.HasRequirement(requirement) {
		return blocked(fmt.Sprintf("%s has no requirement %s; the requirements are %s",
			current.Rel, requirement, requirementChoices(current.Doc)))
	}

	line := fmt.Sprintf("L%d · CAMBIO %s · %s", current.Doc.NextLine(), requirement, quoted(opts.Change))
	grown, reopened := current.Doc.AppendHistory(line).Reopen(requirement)

	message := "recorded the change in " + current.Rel
	if len(reopened) == 0 {
		message += "; no done step cites " + requirement + ", nothing reopened"
	} else {
		message += "; reopened " + strings.Join(reopened, ", ")
	}
	return record(current, files, grown, message)
}

// requirementChoices names the requirements the document declares, so a typo in
// --req is answered with the ones that exist.
func requirementChoices(doc taskdoc.TaskDoc) string {
	ids := make([]string, 0, len(doc.Requirements))
	for _, requirement := range doc.Requirements {
		ids = append(ids, requirement.ID)
	}
	return strings.Join(ids, ", ")
}

// record rewrites the task document with the grown version and reports where it
// happened. The write goes through the shared guard in tasks: it refuses to
// rewrite a document it would damage, because the log is the memory of the
// work and losing it is worse than not recording.
func record(current tasks.Document, files Files, grown taskdoc.TaskDoc, message string) Result {
	if err := tasks.Save(files, current, grown); err != nil {
		return blocked(err.Error())
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
	return "--for " + ids[0] + " (the open ones are " + strings.Join(ids, ", ") + ")"
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
