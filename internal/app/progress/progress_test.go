package progress_test

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/progress"
)

// dir is the project root the verb sees. Every port is a double: progress is
// read-only, and a test that cannot write is the strongest way to say so.
const dir = "/proyecto"

// store is the clone's state, inside the project it describes and outside the
// worktree: the real one is <proyecto>/.git/odachi.
var store = filepath.Join(dir, ".git", "odachi")

// fakeFiles is the disk port in memory. It records writes so a test can prove
// there are none.
type fakeFiles struct {
	files    map[string][]byte
	writes   []string
	existErr error
}

func newFakeFiles(seed map[string]string) *fakeFiles {
	f := &fakeFiles{files: map[string][]byte{}}
	for path, content := range seed {
		f.files[fullPath(path)] = []byte(content)
	}
	return f
}

func fullPath(rel string) string {
	return filepath.Join(dir, filepath.FromSlash(rel))
}

func (f *fakeFiles) Exists(path string) (bool, error) {
	if f.existErr != nil {
		return false, f.existErr
	}
	_, ok := f.files[path]
	return ok, nil
}

func (f *fakeFiles) Read(path string) ([]byte, error) {
	data, ok := f.files[path]
	if !ok {
		return nil, errMissing(path)
	}
	return data, nil
}

func (f *fakeFiles) List(target string) ([]string, error) {
	prefix := strings.TrimSuffix(filepath.ToSlash(target), "/") + "/"
	var names []string
	for path := range f.files {
		slashed := filepath.ToSlash(path)
		if !strings.HasPrefix(slashed, prefix) {
			continue
		}
		rest := strings.TrimPrefix(slashed, prefix)
		if strings.Contains(rest, "/") {
			continue
		}
		names = append(names, rest)
	}
	sortStrings(names)
	return names, nil
}

func (f *fakeFiles) Write(path string, data []byte) error {
	f.writes = append(f.writes, path)
	f.files[path] = data
	return nil
}

// fakeGit answers the one fact progress needs from the repository.
type fakeGit struct {
	tree string
	err  error

	// onBranch, trees and treeCommits answer the ledger questions: a test seeds
	// only the ones it exercises.
	onBranch    map[string]bool
	trees       map[string]string
	treeCommits map[string]string
}

func (g *fakeGit) TreeHash(string) (string, error) { return g.tree, g.err }

func (g *fakeGit) TreeOf(_, commit string) (string, error) {
	tree, ok := g.trees[commit]
	if !ok {
		return "", errors.New("unknown commit")
	}
	return tree, nil
}

func (g *fakeGit) Has(_, commit string) (bool, error) { return g.onBranch[commit], g.err }

func (g *fakeGit) CommitWithTree(_, tree string) (string, error) {
	return g.treeCommits[tree], g.err
}

// missing is the error a read of a file that is not there returns.
type missing string

func (m missing) Error() string { return "missing file: " + string(m) }

func errMissing(path string) error { return missing(path) }

// sortStrings is a tiny insertion sort: the fixture never holds many names and
// the order has to be stable for the report to be comparable.
func sortStrings(names []string) {
	for i := 1; i < len(names); i++ {
		for j := i; j > 0 && names[j] < names[j-1]; j-- {
			names[j], names[j-1] = names[j-1], names[j]
		}
	}
}

// taskDoc is one feature with a done step - with its commit - and a step still
// owed.
const taskDoc = `# reserva

## Requirements (verbatim, from the user)
- R1 · "reservar turnos"

## Done when
- reservo un turno

## Steps
- [x] T1 · el turno se guarda   R1 · abc1234
- [ ] T2 · el turno se puede cancelar   R1

## History
L1 · "reservar turnos"
`

// project is one run's wiring.
type project struct {
	files *fakeFiles
	git   *fakeGit
}

func newProject() *project {
	return &project{
		files: newFakeFiles(map[string]string{"odd/tasks/0007-reserva.md": taskDoc}),
		git: &fakeGit{
			tree:     "4a1c9f7d2e3b5a6c8d9e0f1a2b3c4d5e6f7a8b9c",
			onBranch: map[string]bool{"abc1234": true, "def5678": true},
		},
	}
}

func (p *project) run(opts progress.Options) progress.Result {
	opts.Dir = dir
	opts.Store = store
	return progress.Run(opts, p.files, p.git)
}

// proofFor is a proof file of the shape the tool writes: the tree it measured,
// the command, and its output.
func proofFor(tree string) string {
	return "# tree " + tree + "\n# command go test ./...\n# output\nok\n"
}

// TestProgressSaysWhetherThisTreeHasItsOwnProof: a proof is evidence about one
// tree. One from another tree is not evidence about this one, and the report
// says which of the three it is: this tree, another, or none.
func TestProgressSaysWhetherThisTreeHasItsOwnProof(t *testing.T) {
	t.Run("this tree", func(t *testing.T) {
		p := newProject()
		p.files.files[fullPath("proof/0007-reserva.txt")] = []byte(proofFor(p.git.tree))

		lines := strings.Join(p.run(progress.Options{}).Lines, "\n")

		if !strings.Contains(lines, "proof: proof/0007-reserva.txt · this tree") {
			t.Errorf("the report does not name the proof of this tree:\n%s", lines)
		}
	})

	t.Run("another tree", func(t *testing.T) {
		p := newProject()
		p.files.files[fullPath("proof/0007-reserva.txt")] = []byte(proofFor("9f2b7c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b"))

		lines := strings.Join(p.run(progress.Options{}).Lines, "\n")

		if !strings.Contains(lines, "another tree") {
			t.Errorf("a proof about other bytes is reported as this tree's:\n%s", lines)
		}
	})

	t.Run("none", func(t *testing.T) {
		lines := strings.Join(newProject().run(progress.Options{}).Lines, "\n")

		if !strings.Contains(lines, "proof: none yet") {
			t.Errorf("a project with no proof does not say so:\n%s", lines)
		}
	})
}

// owing is the gate's answer for a tree that owes a verdict: do it, and name
// this base.
func owing(base string) func() (bool, string, error) {
	return func() (bool, string, error) { return true, base, nil }
}

// nothingOwed is the gate's answer for a tree it does not ask about: an open
// feature, or documents written after the close.
func nothingOwed() (bool, string, error) { return false, "", nil }

// TestProgressSaysWhetherTheReviewIsOwed: the gate asks once per feature, at the
// close, and only when the project has the review on. The report says which of
// those moments this tree is in, and whether a verdict already covers it.
func TestProgressSaysWhetherTheReviewIsOwed(t *testing.T) {
	closed := strings.Replace(taskDoc, "- [ ] T2 · el turno se puede cancelar   R1", "- [x] T2 · el turno se puede cancelar   R1 · def5678", 1)

	t.Run("open feature", func(t *testing.T) {
		p := newProject()

		lines := strings.Join(p.run(progress.Options{ReviewMode: true, Review: nothingOwed}).Lines, "\n")

		if !strings.Contains(lines, "review: not due yet") {
			t.Errorf("an open feature asks for the review early:\n%s", lines)
		}
	})

	t.Run("closed feature owes one", func(t *testing.T) {
		p := newProject()
		p.files.files[fullPath("odd/tasks/0007-reserva.md")] = []byte(closed)

		lines := strings.Join(p.run(progress.Options{ReviewMode: true, Review: owing("abc1234")}).Lines, "\n")

		if !strings.Contains(lines, "review: owed") {
			t.Errorf("a closed feature owes the verdict and the report hides it:\n%s", lines)
		}
	})

	t.Run("closed feature with a verdict", func(t *testing.T) {
		p := newProject()
		p.files.files[fullPath("odd/tasks/0007-reserva.md")] = []byte(closed)
		p.files.files[filepath.Join(store, "reviews", p.git.tree+".json")] = []byte(`{"tree":"` + p.git.tree + `"}`)

		lines := strings.Join(p.run(progress.Options{ReviewMode: true, Review: nothingOwed}).Lines, "\n")

		if !strings.Contains(lines, "review: on file") {
			t.Errorf("a tree with a verdict is reported as owing one:\n%s", lines)
		}
	})

	t.Run("review off", func(t *testing.T) {
		p := newProject()
		p.files.files[fullPath("odd/tasks/0007-reserva.md")] = []byte(closed)

		lines := strings.Join(p.run(progress.Options{}).Lines, "\n")

		if !strings.Contains(lines, "review: off") {
			t.Errorf("a project with the review off is asked for a verdict anyway:\n%s", lines)
		}
	})
}

// TestProgressSaysWhenTheGateCannotAnswer: the rule is a dependency, and a
// dependency can fail. A failed question is reported as such: turning it into
// "not due" would tell the human there is nothing to do while the gate cannot say.
func TestProgressSaysWhenTheGateCannotAnswer(t *testing.T) {
	p := newProject()
	failing := func() (bool, string, error) { return false, "", errors.New("git roto") }

	lines := strings.Join(p.run(progress.Options{ReviewMode: true, Review: failing}).Lines, "\n")

	if !strings.Contains(lines, "review: cannot be read") {
		t.Errorf("a failed rule is not reported:\n%s", lines)
	}
	if strings.Contains(lines, "review: owed") || strings.Contains(lines, "not due") {
		t.Errorf("a failed rule was answered as if it had said something:\n%s", lines)
	}
}

// TestProgressSaysWhenTheStoreCannotBeRead: an I/O failure on the verdict store
// is not the absence of a verdict. Reporting "owed" would send the human to pay
// for a review that may already exist, from the one verb whose job is accuracy.
func TestProgressSaysWhenTheStoreCannotBeRead(t *testing.T) {
	p := newProject()
	p.files.existErr = errors.New("permission denied reading .odachi/reviews")

	lines := strings.Join(p.run(progress.Options{ReviewMode: true, Review: owing("abc1234")}).Lines, "\n")

	if !strings.Contains(lines, "review: cannot be read") {
		t.Errorf("the report does not say the store could not be read:\n%s", lines)
	}
	if strings.Contains(lines, "review: owed") {
		t.Errorf("an unreadable store is reported as a verdict owed:\n%s", lines)
	}
}

// TestProgressAsksTheGateInsteadOfDeriving: the report does not re-derive the
// gate's rule from what it can see; it asks. A closed feature whose only change
// since the close is documents owes nothing - the exemption v0.11.18 already
// gives the gate - and the report says that instead of pointing at a review
// nobody owes.
func TestProgressAsksTheGateInsteadOfDeriving(t *testing.T) {
	closed := strings.Replace(taskDoc, "- [ ] T2 · el turno se puede cancelar   R1", "- [x] T2 · el turno se puede cancelar   R1 · def5678", 1)
	p := newProject()
	p.files.files[fullPath("odd/tasks/0007-reserva.md")] = []byte(closed)

	asked := false
	answer := func() (bool, string, error) {
		asked = true
		return false, "", nil
	}

	lines := strings.Join(p.run(progress.Options{ReviewMode: true, Review: answer}).Lines, "\n")

	if !asked {
		t.Error("the report answered the question without asking the gate")
	}
	if strings.Contains(lines, "review: owed") {
		t.Errorf("a tree the gate does not ask about is reported as owing a verdict:\n%s", lines)
	}
	if !strings.Contains(lines, "the documents after the close") {
		t.Errorf("the report does not say why nothing is owed:\n%s", lines)
	}
}

// TestProgressDoesNotDependOnTheReviewForItsCode: a report is not a gate. A
// feature that owes a review still exits 0, because the human asked what is
// missing, not whether it may commit.
func TestProgressDoesNotDependOnTheReviewForItsCode(t *testing.T) {
	p := newProject()
	p.files.files[fullPath("odd/tasks/0007-reserva.md")] = []byte(strings.Replace(taskDoc, "- [ ] T2 · el turno se puede cancelar   R1", "- [x] T2 · el turno se puede cancelar   R1 · def5678", 1))

	if result := p.run(progress.Options{ReviewMode: true}); result.Code != 0 {
		t.Errorf("code = %d, want 0: the report names what is missing, it does not gate it", result.Code)
	}
}

// TestProgressSaysSoWhenThereIsNoFeature: an empty report would look like a
// project with nothing to do. The line names the command that creates the first
// feature instead.
func TestProgressSaysSoWhenThereIsNoFeature(t *testing.T) {
	p := &project{
		files: newFakeFiles(map[string]string{"brief.md": "# algo\n"}),
		git:   &fakeGit{tree: "4a1c9f7d2e3b5a6c8d9e0f1a2b3c4d5e6f7a8b9c"},
	}

	result := p.run(progress.Options{})

	if result.Code != 0 {
		t.Errorf("code = %d, want 0: a project with no feature is not an error", result.Code)
	}
	if !strings.Contains(result.Message, "odachi start") {
		t.Errorf("message = %q, want the command that creates the first feature", result.Message)
	}
	if len(result.Lines) != 1 {
		t.Errorf("lines = %v, want one line: an empty report would look like work to do", result.Lines)
	}
}

// TestProgressWritesNothing: the verb that reports the state cannot be the verb
// that changes it. Every path is exercised - feature, proof, verdict - and the
// disk port records zero writes.
func TestProgressWritesNothing(t *testing.T) {
	p := newProject()
	p.files.files[fullPath("proof/0007-reserva.txt")] = []byte(proofFor(p.git.tree))
	p.files.files[filepath.Join(store, "reviews", p.git.tree+".json")] = []byte(`{"tree":"` + p.git.tree + `"}`)
	p.files.writes = nil

	p.run(progress.Options{ReviewMode: true})

	if len(p.files.writes) != 0 {
		t.Errorf("progress wrote %v, want nothing: it only reads", p.files.writes)
	}
}

// TestTheStateNamesAreTheContract: another program reads these strings, and a
// test that compares a name against its own constant pins nothing - rename the
// constant's value and the suite stays green while the contract changes. Each
// literal appears here once, so a rename is a failure.
func TestTheStateNamesAreTheContract(t *testing.T) {
	names := map[string]string{
		"this_tree":       progress.ProofThisTree,
		"another_tree":    progress.ProofAnotherTree,
		"none":            progress.ProofNone,
		"unmeasurable":    progress.ProofUnmeasurable,
		"off":             progress.ReviewOff,
		"not_due":         progress.ReviewNotDue,
		"owed":            progress.ReviewOwed,
		"waived":          progress.ReviewWaived,
		"on_file":         progress.ReviewOnFile,
		"feature_open":    progress.WhyOpen,
		"after_the_close": progress.WhyAfterClose,
	}
	for want, got := range names {
		if got != want {
			t.Errorf("a state name changed: the contract says %q, the code says %q", want, got)
		}
	}

	// Two states share the word: a proof and a review that cannot be read.
	if progress.ProofUnreadable != "unreadable" {
		t.Errorf("ProofUnreadable = %q, want %q", progress.ProofUnreadable, "unreadable")
	}
	if progress.ReviewUnreadable != "unreadable" {
		t.Errorf("ReviewUnreadable = %q, want %q", progress.ReviewUnreadable, "unreadable")
	}
}

// TestTheReportCarriesTheOpenQuestions: the count is not the fact, the words are.
// The state publishes each open question with the id the history gives it and the
// sentence the human is being asked, so a renderer can show what blocks.
func TestTheReportCarriesTheOpenQuestions(t *testing.T) {
	p := newProject()
	p.files.files[fullPath("odd/tasks/0007-reserva.md")] = []byte(strings.Replace(
		taskDoc, "L1 · ", "L2 · PREGUNTA P1 · \"¿Se puede cancelar?\"\nL1 · ", 1))

	result := p.run(progress.Options{ReviewMode: true, Review: nothingOwed})

	if result.State == nil {
		t.Fatal("no state")
	}
	want := []progress.Question{{ID: "P1", Text: "¿Se puede cancelar?"}}
	if !reflect.DeepEqual(result.State.OpenQuestions, want) {
		t.Errorf("open questions = %+v, want %+v", result.State.OpenQuestions, want)
	}
	// The human reads the same words: the block comes right under the headline,
	// before the plan it stops.
	if got := result.Lines[1]; got != `P1 · "¿Se puede cancelar?"` {
		t.Errorf("the line under the headline is %q, want the question", got)
	}
	if lines := strings.Join(result.Lines, "\n"); !strings.Contains(lines, "1 open question") {
		t.Errorf("the lines do not say the count:\n%s", lines)
	}
}

// TestAReportWithNoQuestionsSaysNothingExtra: a feature with nothing to answer
// prints what it printed before -the headline, the steps, the proof and the
// review- and no empty section for a list that has no rows.
func TestAReportWithNoQuestionsSaysNothingExtra(t *testing.T) {
	p := newProject()

	result := p.run(progress.Options{ReviewMode: true, Review: nothingOwed})

	lines := result.Lines
	if want := 1 + 2 + 2; len(lines) != want {
		t.Fatalf("the report has %d lines, want %d:\n%s", len(lines), want, strings.Join(lines, "\n"))
	}
	if !strings.HasPrefix(lines[1], "  T1 · ") {
		t.Errorf("the first line under the headline is %q, want the first step", lines[1])
	}
}

// TestAFeatureWithNoQuestionsPublishesAnEmptyList: the field is a list, always.
// A nil would make every reader tell "no questions" from "the field is missing",
// and the JSON would say null where the renderer expects to loop.
func TestAFeatureWithNoQuestionsPublishesAnEmptyList(t *testing.T) {
	p := newProject()

	result := p.run(progress.Options{ReviewMode: true, Review: nothingOwed})

	if result.State == nil {
		t.Fatal("no state")
	}
	if result.State.OpenQuestions == nil {
		t.Fatal("open questions = nil, want an empty list")
	}
	if len(result.State.OpenQuestions) != 0 {
		t.Errorf("open questions = %+v, want none", result.State.OpenQuestions)
	}
	if data, err := json.Marshal(result.State.OpenQuestions); err != nil || string(data) != "[]" {
		t.Errorf("the list marshals as %s (err %v), want []", data, err)
	}
}

// TestTheReportIsDataBeforeItIsLines: the report is one shape rendered twice. The
// state carries the same facts the lines carry, so the machine reader and the
// human cannot be told different things.
func TestTheReportIsDataBeforeItIsLines(t *testing.T) {
	p := newProject()
	p.files.files[fullPath("proof/0007-reserva.txt")] = []byte(proofFor(p.git.tree))

	result := p.run(progress.Options{ReviewMode: true, Review: owing("abc1234")})

	if result.State == nil {
		t.Fatal("the report has no state: a machine can only read the lines")
	}
	state := result.State
	if state.Number != 7 || state.Slug != "reserva" {
		t.Errorf("feature = %04d-%s, want 0007-reserva", state.Number, state.Slug)
	}
	if len(state.OpenQuestions) != 0 {
		t.Errorf("open questions = %+v, want none", state.OpenQuestions)
	}
	if len(state.Steps) != 2 {
		t.Fatalf("steps = %v, want two", state.Steps)
	}
	if !state.Steps[0].Done || state.Steps[0].Commit != "abc1234" {
		t.Errorf("first step = %+v, want done at abc1234", state.Steps[0])
	}
	if state.Steps[1].Done || state.Steps[1].Commit != "" {
		t.Errorf("second step = %+v, want pending", state.Steps[1])
	}
	if state.Proof.State != progress.ProofThisTree || state.Proof.Path != "proof/0007-reserva.txt" {
		t.Errorf("proof = %+v, want this tree's", state.Proof)
	}
	if state.Review.State != progress.ReviewOwed || state.Review.Base != "abc1234" {
		t.Errorf("review = %+v, want owed at abc1234", state.Review)
	}

	if lines := strings.Join(result.Lines, "\n"); !strings.Contains(lines, "1 of 2") {
		t.Errorf("the lines do not say what the state says:\n%s", lines)
	}
}

// TestProgressNamesTheFeatureAndItsSteps: the report answers one question - where
// does the open feature stand - and it answers it with what the task doc already
// says: the name, how many steps are done, and the commit each one landed in.
func TestProgressNamesTheFeatureAndItsSteps(t *testing.T) {
	p := newProject()

	result := p.run(progress.Options{})

	lines := strings.Join(result.Lines, "\n")
	for _, want := range []string{
		"0007-reserva",
		"1 of 2",
		"abc1234",
		"el turno se guarda",
		"el turno se puede cancelar",
	} {
		if !strings.Contains(lines, want) {
			t.Errorf("the report does not say %q:\n%s", want, lines)
		}
	}
	if result.Code != 0 {
		t.Errorf("code = %d, want 0: a report is not a gate", result.Code)
	}
}

// TestProgressNamesTheCommitAStepLandedIn: the task doc records the tree a step
// produced, and a rebase keeps the tree and rewrites the commit. The report names
// the commit that carries it; an older mark that landed with another sha is found
// through its bytes; and bytes that never landed as a commit say so instead of
// naming one that does not exist.
func TestProgressNamesTheCommitAStepLandedIn(t *testing.T) {
	const tree = "7bc78db4200c32c2d4fb706b78a5b40c485170ac"

	t.Run("a tree mark", func(t *testing.T) {
		p := newProject()
		p.git.treeCommits = map[string]string{tree: "82377db0000000000000000000000000000000000"}
		doc := strings.Replace(taskDoc, "R1 · abc1234", "R1 · tree:"+tree, 1)
		p.files.files[fullPath("odd/tasks/0007-reserva.md")] = []byte(doc)

		result := p.run(progress.Options{})

		step := result.State.Steps[0]
		if step.Commit != "82377db" || step.Note != "" {
			t.Errorf("step = %+v, want the commit that carries the tree", step)
		}
		if lines := strings.Join(result.Lines, "\n"); !strings.Contains(lines, "82377db") {
			t.Errorf("the report does not name the landed commit:\n%s", lines)
		}
	})

	t.Run("a commit that landed with another sha", func(t *testing.T) {
		p := newProject()
		p.git.onBranch = map[string]bool{}
		p.git.trees = map[string]string{"682f1d4": tree}
		p.git.treeCommits = map[string]string{tree: "82377db0000000000000000000000000000000000"}
		doc := strings.Replace(taskDoc, "R1 · abc1234", "R1 · 682f1d4", 1)
		p.files.files[fullPath("odd/tasks/0007-reserva.md")] = []byte(doc)

		result := p.run(progress.Options{})

		if step := result.State.Steps[0]; step.Commit != "82377db" || step.Note != "" {
			t.Errorf("step = %+v, want the commit found through its bytes", step)
		}
	})

	t.Run("bytes that never landed", func(t *testing.T) {
		p := newProject()
		p.git.onBranch = map[string]bool{}
		p.git.trees = map[string]string{"682f1d4": tree}
		doc := strings.Replace(taskDoc, "R1 · abc1234", "R1 · 682f1d4", 1)
		p.files.files[fullPath("odd/tasks/0007-reserva.md")] = []byte(doc)

		result := p.run(progress.Options{})

		step := result.State.Steps[0]
		if step.Commit != "" || step.Note != "not on this branch" {
			t.Errorf("step = %+v, want no commit and the reason", step)
		}
		if lines := strings.Join(result.Lines, "\n"); !strings.Contains(lines, "not on this branch") {
			t.Errorf("the report names a commit that does not exist:\n%s", lines)
		}
	})
}

// TestAReportSaysTheReviewWasWaived: a closed feature that waived its review is
// not "not due yet" and not "on file": it is the waiver, named, so the panel and
// the human read the decision instead of a silence.
func TestAReportSaysTheReviewWasWaived(t *testing.T) {
	p := newProject()
	doc := strings.Replace(taskDoc, "- [ ] T2 · el turno se puede cancelar   R1", "- [x] T2 · el turno se puede cancelar   R1 · abc1234", 1)
	doc = strings.Replace(doc, "# reserva\n", "# reserva\nReview: waived\n", 1)
	p.files.files[fullPath("odd/tasks/0007-reserva.md")] = []byte(doc)

	result := p.run(progress.Options{ReviewMode: true, Review: nothingOwed})

	if result.State == nil {
		t.Fatal("no state")
	}
	if got := result.State.Review.State; got != progress.ReviewWaived {
		t.Fatalf("review = %q, want %q", got, progress.ReviewWaived)
	}
	if lines := strings.Join(result.Lines, "\n"); !strings.Contains(lines, "review: waived by the human") {
		t.Errorf("the lines do not name the waiver:\n%s", lines)
	}
}
