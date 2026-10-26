// Package progress reports where the open feature stands: the steps the task doc
// records and the commit each done step landed in. It only reads: a verb that
// reports the state must never be a verb that changes it.
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

// Options is what the verb needs from the invocation. The directory is the
// project root; ReviewMode is the project's switch, read by the caller the same
// way check reads it, so the report never claims a gate the project turned off.
type Options struct {
	Dir        string
	ReviewMode bool

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

// Result is the report: the lines the human reads, in order.
type Result struct {
	Code    int
	Message string
	Lines   []string
}

// Run reads the open feature and prints where it stands.
func Run(opts Options, files Files, git Git) Result {
	doc, problem := currentFeature(files, opts.Dir)
	if problem != "" {
		return Result{Code: CodeOK, Message: problem, Lines: []string{problem}}
	}

	// One read of the tree: the proof is evidence about one tree, and the verdict
	// is filed under one tree, so both questions are asked about the same bytes.
	tree, treeErr := git.TreeHash(opts.Dir)

	lines := []string{headline(doc)}
	lines = append(lines, stepLines(doc.Doc.Steps)...)
	lines = append(lines, proofLine(files, opts.Dir, doc, tree, treeErr))
	lines = append(lines, reviewLine(opts, files, doc, tree, treeErr))
	return Result{
		Code:    CodeOK,
		Message: lines[0],
		Lines:   lines,
	}
}

// proofLine says whether this tree has its own proof. A proof is evidence about
// the bytes it measured: one from another tree is not evidence about this one,
// and the report never rounds that up.
func proofLine(files Files, dir string, doc tasks.Document, tree string, treeErr error) string {
	rel := proofRel(doc)
	data, err := files.Read(fullPath(dir, rel))
	if err != nil {
		return "proof: none yet"
	}
	parsed, err := proof.Parse(string(data))
	if err != nil {
		return "proof: " + rel + " · cannot be read"
	}
	if treeErr != nil {
		return "proof: " + rel + " · this tree cannot be measured"
	}
	if parsed.Matches(tree) {
		return "proof: " + rel + " · this tree"
	}
	return "proof: " + rel + " · another tree"
}

// reviewLine says which moment the feature is in. The rule belongs to the gate:
// this function asks it instead of re-deriving it from the steps, so the report
// and the gate cannot disagree about a tree.
func reviewLine(opts Options, files Files, doc tasks.Document, tree string, treeErr error) string {
	if !opts.ReviewMode {
		return "review: off · the project does not ask for one"
	}
	if treeErr != nil {
		return "review: cannot be read · this tree cannot be measured"
	}
	verdict, err := domainreview.Path(tree)
	if err != nil {
		return "review: cannot be read · " + oneLine(err.Error())
	}
	if exists, err := files.Exists(fullPath(opts.Dir, verdict)); err == nil && exists {
		return "review: on file · " + verdict
	}

	if opts.Review == nil {
		return "review: cannot be read · the gate's rule was not wired"
	}
	owed, base, err := opts.Review()
	if err != nil {
		return "review: cannot be read · " + oneLine(err.Error())
	}
	if !owed {
		if done, total := doc.Doc.Progress(); total > 0 && done == total {
			return "review: not due yet · the documents after the close do not reopen it"
		}
		return "review: not due yet · the feature is open"
	}
	base = strings.TrimSpace(base)
	if base == "" {
		return "review: owed · the feature is closed"
	}
	return "review: owed · odachi review --base " + base
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

// currentFeature is the highest-numbered task doc: the feature the human is
// working on now. A project with no feature is not an error to hide behind an
// empty report: the line says what to run.
func currentFeature(files Files, dir string) (tasks.Document, string) {
	doc, err := tasks.Current(files, dir)
	if err != nil {
		return tasks.Document{}, oneLine(err.Error())
	}
	return doc, ""
}

// headline is the one line that names the feature: its number and slug, how many
// steps are done, and how many questions are still open.
func headline(doc tasks.Document) string {
	done, total := countSteps(doc.Doc.Steps)
	return fmt.Sprintf("%04d-%s · %d of %d steps · %s",
		doc.Number, doc.Slug, done, total, questions(len(doc.Doc.OpenQuestions())))
}

// stepLines is one line per step, in the order the plan wrote them: the id, the
// commit it landed in (or "pending"), and the behavior it delivers.
func stepLines(steps []taskdoc.Step) []string {
	lines := make([]string, 0, len(steps))
	for _, step := range steps {
		landed := "pending"
		if step.Done {
			landed = step.Commit
			if strings.TrimSpace(landed) == "" {
				landed = "done"
			}
		}
		lines = append(lines, fmt.Sprintf("  %s · %s · %s", step.ID, landed, step.Text))
	}
	return lines
}

// countSteps is how many steps the feature finished, of how many it planned.
func countSteps(steps []taskdoc.Step) (done, total int) {
	for _, step := range steps {
		if step.Done {
			done++
		}
	}
	return done, len(steps)
}

// questions says how many decisions the feature still owes, in the voice the rest
// of the tool uses.
func questions(n int) string {
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
