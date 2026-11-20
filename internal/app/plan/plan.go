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
	"github.com/biggs-100/odachi/internal/domain/authorship"
	"github.com/biggs-100/odachi/internal/domain/gate"
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
	// Skills are the craft a delegated step owes, by name: the plan records them
	// and `odachi delegate` resolves them to the exact paths the worker reads.
	Skills []string
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

// Git answers the questions both forms ask. Closing a step needs the two trees to
// compare -whether the bytes the step produced are the commit that carries them,
// and which tree a commit carries- because a mark is only honest when the worktree
// is that commit. Opening one needs the diff of the feature so far, which is how
// the plan weighs a feature that has already written more than one task may carry.
type Git interface {
	HeadTree(root string) (string, error)
	WorktreeTree(root string) (string, error)
	TreeOf(root, commit string) (string, error)
	// Diff is the text of the change between base and the tree that stands: the
	// same bytes the plan weighs before a step opens and the close names after
	// it, so no two doors can disagree about what this feature already authored
	// (0334, 0339).
	Diff(root, base string) (string, error)
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
const usage = `odachi plan "<text>" [--req R1] [--route inline | --route delegated --agent <name> --skill <name>] · odachi plan --done T2 [--commit abc1234]`

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
	case done != "" && (route != "" || agent != "" || len(opts.Skills) > 0):
		return "", `--route, --agent and --skill describe a step you open, and this run closes one: run odachi plan "<text>" --route inline, or --route delegated --agent <name> --skill <name>`
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

// add appends the next step. The number and the link are decided by the domain,
// and the write is canonical by construction.
func add(current tasks.Document, files Files, opts Options) Result {
	// A light feature is one line, and its document says so. A second step would
	// leave a document that still claims the lane while carrying a plan: the way
	// out is a brief, not a quieter lie (0166).
	if current.Doc.Light {
		return blocked("this is a light feature: a second step is the signal to leave the lane; " +
			"write a brief and run odachi start again")
	}

	// The feature is weighed before the step exists. The close measures the same
	// candidate once the work is paid; this is the one moment the number can stop
	// more of it from arriving (0334).
	refusal, note := sliceVerdict(current, opts)
	if refusal != "" {
		return blocked(refusal)
	}

	grown, err := current.Doc.AddStepBySkills(opts.Text, opts.Requirement, taskdoc.Route{Kind: opts.Route, Agent: opts.Agent}, opts.Skills)
	if err != nil {
		return blocked(oneLine(err.Error()) + " (en " + current.Rel + ")")
	}

	if err := tasks.SaveWithBackup(files, current, grown); err != nil {
		return blocked(err.Error())
	}

	step := grown.Steps[len(grown.Steps)-1]
	message := fmt.Sprintf("step %s added to %s", step.ID, current.Rel)
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

// Why a count was not taken. The first two are not failures -a document that
// records no base has nothing to be measured against, and a caller that wired no
// repository has no trees to compare- so the plan stays silent for them, which is
// the sentence it always printed. Anything else is a repository that could not
// answer, and both moments name it (0332, 0334, 0339).
const (
	noBase = "this task doc records no base"
	noRepo = "no repository is wired"
)

// measurement is the one reading of what a feature has authored: the lines
// between the base the document records and the tree that stands, and that base.
// Why is empty when Lines is a number; when it is not, Lines carries no value and
// Why is the clause that explains the absence.
//
// This package asks the question twice -the plan weighs a feature before another
// step opens, the close records it once the work is paid- and both read measure,
// which counts with the same authorship.Lines the gate counts with. A second
// count would let the moments speak about different bytes (0331, 0334, 0339).
type measurement struct {
	lines int
	base  string
	why   string
}

// took reports whether the reading produced a number.
func (m measurement) took() bool { return m.why == "" }

// measure reads the feature's cost, or says why it could not be read.
func measure(current tasks.Document, opts Options) measurement {
	if opts.Git == nil {
		return measurement{why: noRepo}
	}
	base := strings.TrimSpace(current.Doc.Base)
	if base == "" {
		return measurement{why: noBase}
	}

	m := measurement{base: base}
	diff, err := opts.Git.Diff(opts.Dir, base)
	if err != nil {
		m.why = oneLine(err.Error())
		return m
	}
	m.lines = authorship.Lines(diff)
	return m
}

// sliceVerdict answers whether this feature may still grow a step. A refusal
// carries the number it measured, the limit one task carries and the door that
// starts the rest as a feature of its own; a note says the number could not be
// taken, because a measurement nobody could read is not a candidate under it, and
// a repository a caller did not wire -or a document with no base- has nothing to
// weigh against (0334). The reading itself is measure's, so the close cannot
// disagree with this door about the same feature (0339).
func sliceVerdict(current tasks.Document, opts Options) (refusal, note string) {
	m := measure(current, opts)
	switch {
	case m.why == noRepo || m.why == noBase:
		// Nothing to weigh, and no note: the step opens with the sentence it
		// always printed.
		return "", ""
	case !m.took():
		return "", "the slice was not measured: " + m.why
	case m.lines > gate.SliceLines:
		return fmt.Sprintf("the candidate is %d authored lines and one feature authors at most %d: plan the rest with odachi start",
			m.lines, gate.SliceLines), ""
	}
	return "", ""
}

// markDone closes a step and records the commit that closed it. Closing a
// closed step is a no-op that writes nothing: the document is the memory of the
// work, and re-marking it would rewrite that memory for no reason.
//
// It also names what the feature has authored since its base. The number cannot
// stop this work any more -it is paid- so the close reports it and the limit the
// feature's own close applies, instead of refusing what is already written (0339).
func markDone(current tasks.Document, files Files, opts Options) Result {
	id := strings.ToUpper(strings.TrimSpace(opts.Done))

	step, ok := findStep(current.Doc, id)
	if !ok {
		return blocked(unknownStep(current, id))
	}
	// The document's own spelling wins: a hand-written `t2` stays `t2` when the
	// tool marks it, because Record matches the id it has.
	id = step.ID
	given := strings.TrimSpace(opts.Commit)

	// A step already done is the memory of work finished, and re-running the verb
	// does not rewrite it. One exception: a step that closed with no mark can
	// still be given the bytes it produced, because a record nobody can repair is
	// a hole in the memory (0164).
	if step.Done {
		switch {
		case step.Commit != "":
			return Result{
				Code:    CodeOK,
				Message: fmt.Sprintf("step %s was already done", id),
				Task:    current.Rel,
			}
		case given == "":
			return Result{
				Code: CodeOK,
				Message: fmt.Sprintf("step %s was already done with no mark; pass --commit <sha> to record the tree it produced",
					id),
				Task: current.Rel,
			}
		}

		mark := canonicalMark(opts, given)
		grown, err := current.Doc.Record(id, mark)
		if err != nil {
			return blocked(oneLine(err.Error()) + " (en " + current.Rel + ")")
		}
		if err := tasks.SaveWithBackup(files, current, grown); err != nil {
			return blocked(err.Error())
		}
		return Result{
			Code:    CodeOK,
			Message: fmt.Sprintf("step %s already done · recorded %s", id, shortMark(mark)),
			Task:    current.Rel,
			Files:   []string{current.Rel},
		}
	}

	// A commit recorded on an open step is not thrown away when the caller does
	// not pass one: plan records, it never erases.
	mark := step.Commit
	if given != "" {
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

	// The number is read before the document is written: the mark and the backup
	// are the close's own paperwork, not authored work, and counting them would
	// make the line a human reads depend on the act of printing it (0339).
	cost := measure(current, opts)

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
	message += " · " + closingSentence(cost)
	return Result{
		Code:    CodeOK,
		Message: message,
		Task:    current.Rel,
		Files:   []string{current.Rel},
	}
}

// closingSentence is what a close adds to the one line a person reads: the
// authored lines accumulated since the feature's base, that base, and the limit
// the feature's own close applies to them. A step owes no verifier, so the number
// is a record and never a refusal: past the limit the sentence only says what
// follows, in the gate's words, and the step still closes (0339).
//
// A count nobody could take is named as such, with why and with no number: a
// figure invented at the close would be a claim the diff never made.
func closingSentence(m measurement) string {
	if !m.took() {
		return "the count was not taken: " + m.why
	}
	if m.lines > gate.SliceLines {
		return fmt.Sprintf("%d authored lines since base %s, past the %d a close that owes a verifier may carry: this step owes no verifier and closes anyway, and the rest of the work belongs to a feature of its own",
			m.lines, shortBase(m.base), gate.SliceLines)
	}
	return fmt.Sprintf("%d authored lines since base %s, of the %d a close that owes a verifier may carry",
		m.lines, shortBase(m.base), gate.SliceLines)
}

// shortBase names the base in the seven characters a human reads, the way the
// mark is already shortened; a base shorter than that keeps its spelling.
func shortBase(base string) string {
	if len(base) > 7 {
		return base[:7]
	}
	return base
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
		return "", "the tree was not recorded: the worktree is not the commit; pass --commit <sha> to record it"
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
