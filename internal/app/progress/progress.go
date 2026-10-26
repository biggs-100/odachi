// Package progress reports where the open feature stands: the steps the task doc
// records, the commit each done step landed in, the proof of this tree and the
// review it still owes. It only reads: a verb that reports the state must never be
// a verb that changes it.
//
// The report is computed once as State and the human lines are rendered from it.
// One shape, two readers: a renderer that parses the lines would drift from the
// data the moment either changes.
package progress

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/biggs-100/odachi/internal/app/tasks"
	"github.com/biggs-100/odachi/internal/domain/proof"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// CodeOK is the only code this verb returns: a report is not a gate, and a
// script that reads it asks the report, not the exit code.
const CodeOK = 0

// The states a proof can be in, as a machine reads them.
const (
	ProofThisTree     = "this_tree"
	ProofAnotherTree  = "another_tree"
	ProofNone         = "none"
	ProofUnreadable   = "unreadable"
	ProofUnmeasurable = "unmeasurable"
	ReviewOff         = "off"
	ReviewNotDue      = "not_due"
	ReviewOwed        = "owed"
	ReviewOnFile      = "on_file"
	ReviewUnreadable  = "unreadable"

	// Why the review is not due yet: an open feature, or the documents written
	// after the close.
	WhyOpen       = "feature_open"
	WhyAfterClose = "after_the_close"
)

// Options is what the verb needs from the invocation. The directory is the
// project root; ReviewMode is the project's switch, read by the caller the same
// way check reads it, so the report never claims a gate the project turned off.
type Options struct {
	Dir        string
	ReviewMode bool

	// Feature is the number of the task document this report describes. Zero
	// means the highest, which is the feature the human is working on now.
	Feature int

	// Review answers the question the gate answers: does this tree owe a
	// verdict, and which base must that review name? The caller wires the gate's
	// own rule in. A report that re-derived it would disagree with the gate the
	// day either one changed, and this verb exists to say what is missing.
	Review func() (owed bool, base string, err error)
}

// Files is the disk port: enough to find and read the task doc. Every Files
// adapter that can read a project satisfies it.
type Files interface {
	Exists(path string) (bool, error)
	Read(path string) ([]byte, error)
	List(dir string) ([]string, error)
}

// Git is the repository port: enough to name the tree this report describes.
type Git interface {
	TreeHash(root string) (string, error)
}

// Step is one task-doc step as data.
type Step struct {
	ID     string `json:"id"`
	Text   string `json:"text"`
	Done   bool   `json:"done"`
	Commit string `json:"commit,omitempty"`
}

// Question is one open question as data: the id the history gave it and the words
// the human is being asked.
type Question struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// Proof is the proof line as data: which file, in which of the states.
type Proof struct {
	State string `json:"state"`
	Path  string `json:"path,omitempty"`
}

// Review is the review line as data: the moment the feature is in, why nothing
// is due when nothing is, the verdict file when one covers this tree, and the base
// a review must name when it is owed.
type Review struct {
	State string `json:"state"`
	Why   string `json:"why,omitempty"`
	Path  string `json:"path,omitempty"`
	Base  string `json:"base,omitempty"`
}

// State is the report as data. The names are the contract an external renderer
// depends on, so they are pinned by a test.
type State struct {
	Number        int        `json:"number"`
	Slug          string     `json:"slug"`
	Steps         []Step     `json:"steps"`
	OpenQuestions []Question `json:"open_questions"`
	Proof         Proof      `json:"proof"`
	Review        Review     `json:"review"`
}

// Result is the report: the lines the human reads, in order, and the state they
// were rendered from. State is nil when there is no feature yet: the message
// says what to run.
type Result struct {
	Code    int
	Message string
	Lines   []string
	State   *State
}

// Run reads the open feature and reports where it stands.
func Run(opts Options, files Files, git Git) Result {
	doc, problem := currentFeature(files, opts.Dir, opts.Feature)
	if problem != "" {
		return Result{Code: CodeOK, Message: problem, Lines: []string{problem}}
	}

	// One read of the tree: the proof is evidence about one tree, and the verdict
	// is filed under one tree, so both questions are asked about the same bytes.
	tree, treeErr := git.TreeHash(opts.Dir)

	state := State{
		Number:        doc.Number,
		Slug:          doc.Slug,
		Steps:         steps(doc.Doc.Steps),
		OpenQuestions: questions(doc.Doc.OpenQuestions()),
		Proof:         proofState(files, opts.Dir, doc, tree, treeErr),
		Review:        reviewState(opts, files, doc, tree, treeErr),
	}

	lines := render(state)
	return Result{
		Code:    CodeOK,
		Message: lines[0],
		Lines:   lines,
		State:   &state,
	}
}

// currentFeature is the highest-numbered task doc: the feature the human is
// working on now. A project with no feature is not an error to hide behind an
// empty report: the line says what to run.
func currentFeature(files Files, dir string, number int) (tasks.Document, string) {
	doc, err := tasks.Feature(files, dir, number)
	if err != nil {
		return tasks.Document{}, oneLine(err.Error())
	}
	return doc, ""
}

// proofState says whether this tree has its own proof. A proof is evidence about
// the bytes it measured: one from another tree is not evidence about this one,
// and the report never rounds that up.
func proofState(files Files, dir string, doc tasks.Document, tree string, treeErr error) Proof {
	rel := proofRel(doc)
	data, err := files.Read(fullPath(dir, rel))
	if err != nil {
		return Proof{State: ProofNone}
	}
	parsed, err := proof.Parse(string(data))
	if err != nil {
		return Proof{State: ProofUnreadable, Path: rel}
	}
	if treeErr != nil {
		return Proof{State: ProofUnmeasurable, Path: rel}
	}
	if parsed.Matches(tree) {
		return Proof{State: ProofThisTree, Path: rel}
	}
	return Proof{State: ProofAnotherTree, Path: rel}
}

// reviewState says which moment the feature is in. The rule belongs to the gate:
// this function asks it instead of re-deriving it from the steps, so the report
// and the gate cannot disagree about a tree.
func reviewState(opts Options, files Files, doc tasks.Document, tree string, treeErr error) Review {
	if !opts.ReviewMode {
		return Review{State: ReviewOff}
	}
	if treeErr != nil {
		return Review{State: ReviewUnreadable}
	}
	verdict, err := domainreview.Path(tree)
	if err != nil {
		return Review{State: ReviewUnreadable}
	}
	exists, err := files.Exists(fullPath(opts.Dir, verdict))
	if err != nil {
		// A store that cannot be read is not a store with no verdict: claiming a
		// review is owed would send the human to pay for one that may be there.
		return Review{State: ReviewUnreadable}
	}
	if exists {
		return Review{State: ReviewOnFile, Path: verdict}
	}

	if opts.Review == nil {
		return Review{State: ReviewUnreadable}
	}
	owed, base, err := opts.Review()
	if err != nil {
		return Review{State: ReviewUnreadable}
	}
	if !owed {
		if done, total := doc.Doc.Progress(); total > 0 && done == total {
			return Review{State: ReviewNotDue, Why: WhyAfterClose}
		}
		return Review{State: ReviewNotDue, Why: WhyOpen}
	}
	return Review{State: ReviewOwed, Base: strings.TrimSpace(base)}
}

// render is the human report, derived from the state. Every line here is a
// projection of a field: the machine reader and the human read the same facts.
func render(state State) []string {
	lines := []string{fmt.Sprintf("%04d-%s · %d of %d steps · %s",
		state.Number, state.Slug, doneSteps(state.Steps), len(state.Steps), questionCount(len(state.OpenQuestions)))}

	// The block sits between the headline and the plan: a question is what stops
	// the work, and the human reads it before the steps they cannot finish.
	for _, question := range state.OpenQuestions {
		lines = append(lines, fmt.Sprintf("%s · %q", question.ID, question.Text))
	}

	for _, step := range state.Steps {
		landed := "pending"
		if step.Done {
			landed = step.Commit
			if landed == "" {
				landed = "done"
			}
		}
		lines = append(lines, fmt.Sprintf("  %s · %s · %s", step.ID, landed, step.Text))
	}
	return append(lines, proofLine(state.Proof), reviewLine(state.Review))
}

// proofLine renders the proof state, in the voice the rest of the tool uses.
func proofLine(p Proof) string {
	switch p.State {
	case ProofThisTree:
		return "proof: " + p.Path + " · this tree"
	case ProofAnotherTree:
		return "proof: " + p.Path + " · another tree"
	case ProofUnreadable:
		return "proof: " + p.Path + " · cannot be read"
	case ProofUnmeasurable:
		return "proof: " + p.Path + " · this tree cannot be measured"
	default:
		return "proof: none yet"
	}
}

// reviewLine renders the review state, including why nothing is owed.
func reviewLine(r Review) string {
	switch r.State {
	case ReviewOff:
		return "review: off · the project does not ask for one"
	case ReviewOnFile:
		return "review: on file · " + r.Path
	case ReviewOwed:
		if r.Base == "" {
			return "review: owed · the feature is closed"
		}
		return "review: owed · odachi review --base " + r.Base
	case ReviewUnreadable:
		return "review: cannot be read · this tree cannot be measured"
	default:
		if r.Why == WhyAfterClose {
			return "review: not due yet · the documents after the close do not reopen it"
		}
		return "review: not due yet · the feature is open"
	}
}

// steps turns the task doc's steps into data, in the order the plan wrote them.
func steps(recorded []taskdoc.Step) []Step {
	out := make([]Step, 0, len(recorded))
	for _, step := range recorded {
		out = append(out, Step{
			ID:     step.ID,
			Text:   step.Text,
			Done:   step.Done,
			Commit: step.Commit,
		})
	}
	return out
}

// doneSteps is how many steps the feature finished.
func doneSteps(steps []Step) int {
	done := 0
	for _, step := range steps {
		if step.Done {
			done++
		}
	}
	return done
}

// proofRel is where this feature's proof lives: the number and slug the task doc
// already carries, so the report never guesses. It is the same name check writes.
func proofRel(doc tasks.Document) string {
	return fmt.Sprintf("proof/%04d-%s.txt", doc.Number, doc.Slug)
}

// fullPath resolves a project-relative path the way every adapter does.
func fullPath(dir, rel string) string {
	return filepath.Join(dir, filepath.FromSlash(rel))
}

// questions turns the document's open questions into data, in the order the
// history records them. The list is never nil: a renderer loops over it, and "no
// questions" must not look like a missing field.
func questions(open []taskdoc.Question) []Question {
	out := make([]Question, 0, len(open))
	for _, question := range open {
		out = append(out, Question{ID: question.ID, Text: question.Text})
	}
	return out
}

// questionCount says how many decisions the feature still owes, in the voice the
// rest of the tool uses.
func questionCount(n int) string {
	if n == 0 {
		return "0 open questions"
	}
	if n == 1 {
		return "1 open question"
	}
	return fmt.Sprintf("%d open questions", n)
}

// oneLine keeps every message to a single line, whatever an adapter said.
func oneLine(message string) string {
	return strings.Join(strings.Fields(message), " ")
}
