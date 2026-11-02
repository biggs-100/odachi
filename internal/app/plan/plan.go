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
	// Route is who runs the new step: inline, delegated, or empty when nobody
	// declared it. Agent names the other agent, and only exists with delegated.
	Route, Agent string
	// Commit is the commit that closed the step; it only makes sense with Done.
	Commit string
	// Git answers the two trees the done form compares. It is optional: without
	// it a step closes with no mark, which is what the first day and a test
	// without a repository need.
	Git Git
	// Feature is the number of the task document to write to. Zero means the
	// highest, which is the feature the human is working on now.
	Feature int
}

// Git answers the done form's two questions: are the bytes the step produced the
// commit that carries them, and which tree does a commit carry? A mark is only
// honest when the worktree is that commit, and a commit is remembered by its tree.
type Git interface {
	HeadTree(root string) (string, error)
	WorktreeTree(root string) (string, error)
	TreeOf(root, commit string) (string, error)
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
const usage = `odachi plan "<text>" [--req R1] [--route inline | --route delegated --agent <name>] · odachi plan --done T2 [--commit abc1234]`

// Run performs one `odachi plan`.
func Run(opts Options, files Files) Result {
	if opts.Dir == "" {
		opts.Dir = "."
	}

	form, problem := validate(opts)
	if problem != "" {
		return blocked(problem)
	}

	current, err := tasks.Feature(files, opts.Dir, opts.Feature)
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
	route := strings.TrimSpace(opts.Route)
	agent := strings.TrimSpace(opts.Agent)

	switch {
	case done == "" && commit != "":
		return "", `--commit only goes with a step you are marking: run odachi plan --done T2 --commit <sha>`
	case done != "" && (route != "" || agent != ""):
		return "", `--route and --agent describe a step you open, and this run closes one: run odachi plan "<text>" --route ` + routeAndAgent(route, agent)
	case text != "" && done != "":
		return "", "use one form of odachi plan at a time: " + usage
	case done != "" && requirement != "":
		return "", `--req only goes with the text of a new step: run odachi plan "<text>" --req ` + requirement
	case done != "":
		return formDone, ""
	case text == "":
		return "", `missing the step: write odachi plan "<text>" to add one or odachi plan --done T2 to mark one done`
	}
	return formAdd, ""
}

// routeAndAgent writes the pair an error repeats, so a message that refuses a
// call still names the call that works.
func routeAndAgent(route, agent string) string {
	if route == "" {
		route = "inline"
	}
	if agent == "" {
		return route
	}
	return route + " --agent " + agent
}

// add appends the next step. The number and the link are decided by the domain,
// and the write is canonical by construction.
func add(current tasks.Document, files Files, opts Options) Result {
	grown, err := current.Doc.AddStepBy(opts.Text, opts.Requirement, taskdoc.Route{Kind: opts.Route, Agent: opts.Agent})
	if err != nil {
		return blocked(oneLine(err.Error()) + " (en " + current.Rel + ")")
	}

	if err := tasks.SaveWithBackup(files, current, grown); err != nil {
		return blocked(err.Error())
	}

	step := grown.Steps[len(grown.Steps)-1]
	return Result{
		Code:    CodeOK,
		Message: fmt.Sprintf("step %s added to %s", step.ID, current.Rel),
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
			Message: fmt.Sprintf("step %s was already done", id),
			Task:    current.Rel,
		}
	}

	// A commit recorded on an open step is not thrown away when the caller does
	// not pass one: plan records, it never erases.
	mark := step.Commit
	if given := strings.TrimSpace(opts.Commit); given != "" {
		// --commit names a commit; the document records its bytes, because a rebase
		// rewrites the commit and keeps the tree.
		mark = canonicalMark(opts, given)
	}
	// No commit was given: the step is remembered by the bytes it produced, when
	// those bytes are a commit. The sha of a local commit does not survive a
	// rebase; a tree does, and it is the identity a proof and a verdict already use.
	var note string
	if mark == "" {
		if hasDoneAfter(current.Doc, id) {
			// Marking an older step after a later one is done: HEAD is not this
			// step's commit, and guessing would attach the wrong bytes.
			note = "pass --commit: a later step is already done"
		} else {
			mark, note = committedTree(opts)
		}
	}

	grown, err := current.Doc.Record(id, mark)
	if err != nil {
		return blocked(oneLine(err.Error()) + " (en " + current.Rel + ")")
	}
	if err := tasks.SaveWithBackup(files, current, grown); err != nil {
		return blocked(err.Error())
	}

	message := fmt.Sprintf("step %s done", id)
	if mark != "" {
		message += " · " + shortMark(mark)
	}
	if note != "" {
		message += " · " + note
	}
	return Result{
		Code:    CodeOK,
		Message: message,
		Task:    current.Rel,
		Files:   []string{current.Rel},
	}
}

// canonicalMark names the bytes of a commit the repository can read: the same
// identity the default records, so an explicit --commit survives a rebase too. A
// commit the repository cannot read keeps the spelling it came with, because that
// is all there is.
func canonicalMark(opts Options, commit string) string {
	if opts.Git == nil {
		return commit
	}
	tree, err := opts.Git.TreeOf(opts.Dir, commit)
	if err != nil || strings.TrimSpace(tree) == "" {
		return commit
	}
	return "tree:" + tree
}

// hasDoneAfter reports whether a step later than id is already done. Marking an
// older step then would read HEAD, which is a later commit, and record its bytes
// as this step's: a wrong mark is worse than none.
func hasDoneAfter(doc taskdoc.TaskDoc, id string) bool {
	seen := false
	for _, step := range doc.Steps {
		if seen && step.Done {
			return true
		}
		if strings.EqualFold(step.ID, id) {
			seen = true
		}
	}
	return false
}

// committedTree is the mark a step can honestly carry: the tree of HEAD, and only
// when the worktree is that commit. Marking a step before committing it is legal
// and records nothing, as it always did; a tree describing uncommitted bytes would
// be a claim the repository cannot back.
func committedTree(opts Options) (string, string) {
	if opts.Git == nil {
		return "", ""
	}
	head, err := opts.Git.HeadTree(opts.Dir)
	if err != nil || strings.TrimSpace(head) == "" {
		return "", ""
	}
	worktree, err := opts.Git.WorktreeTree(opts.Dir)
	if err != nil {
		return "", ""
	}
	if worktree != head {
		return "", "the tree was not recorded: the worktree is not the commit"
	}
	return "tree:" + head, ""
}

// shortMark is what the message prints: a tree or a commit in the seven characters
// a human reads, while the document keeps the whole value.
func shortMark(mark string) string {
	if tree, ok := strings.CutPrefix(mark, "tree:"); ok && len(tree) > 7 {
		return "tree:" + tree[:7]
	}
	return mark
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
		return fmt.Sprintf("unknown step: %s; this task doc has no open steps (in %s); add one with odachi plan \"<text>\"",
			id, current.Rel)
	}
	return fmt.Sprintf("unknown step: %s; the open steps are %s (in %s)", id, strings.Join(open, ", "), current.Rel)
}

// blocked is a Result with exit code 1.
func blocked(message string) Result {
	return Result{Code: CodeBlocked, Message: oneLine(message)}
}

// oneLine keeps every message to a single line, whatever came in from a flag.
func oneLine(message string) string {
	return strings.Join(strings.Fields(message), " ")
}
