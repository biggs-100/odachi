package review_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/review"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
)

// oneAnswer is what a fake reviewer answers for one draw: no findings, and the
// declaration every complete answer carries.
const oneAnswer = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

// warningDraw is one draw's answer: a warning at that line, in the words that draw
// used. Two draws of the same judge name the same defect differently, which is why
// the union needs an identity that is not the sentence.
func warningDraw(line int, why string) string {
	return fmt.Sprintf(`{"inspection":{"status":"completed","paths":["a.go"]},"findings":[{"path":"a.go","line":%d,"severity":"WARNING","lens":"reliability","causal":true,"why":%q}]}`, line, why)
}

// filedVerdict reads the verdict the run filed for the current tree.
func filedVerdict(t *testing.T, p *project) domainreview.Verdict {
	t.Helper()
	data, ok := p.files.files[storePath(currentTree)]
	if !ok {
		t.Fatal("the run filed no verdict for the tree")
	}
	parsed, err := domainreview.Parse(data)
	if err != nil {
		t.Fatalf("the filed verdict does not parse: %v", err)
	}
	return parsed
}

// filedFindings reads the rows the run filed for the current tree.
func filedFindings(t *testing.T, p *project) []domainreview.Finding {
	t.Helper()
	return filedVerdict(t, p).Findings
}

// TestReviewRefusesADrawCountItCannotPay: --runs is a count of answers, and a count
// outside the range is refused before the store is read or a reviewer is started.
// Eleven draws of four lenses is an experiment, and an experiment is designed
// somewhere else - the ceiling is the one `review probe` already names.
func TestReviewRefusesADrawCountItCannotPay(t *testing.T) {
	for _, runs := range []int{-1, review.MaxRuns + 1} {
		p := newProject()
		p.host.model = "host/default"
		p.shell.output = oneAnswer

		result := p.run(review.Options{Exec: "revisor --model {model}", Runs: runs})

		if result.Code != review.CodeBlocked {
			t.Fatalf("--runs %d = code %d, want blocked (%s)", runs, result.Code, result.Message)
		}
		if !strings.Contains(result.Message, strconv.Itoa(review.MaxRuns)) {
			t.Errorf("--runs %d does not name the ceiling %d: %s", runs, review.MaxRuns, result.Message)
		}
		if len(p.shell.commands) != 0 {
			t.Errorf("--runs %d started a reviewer: %v", runs, p.shell.commands)
		}
		if written := p.files.written(); len(written) != 0 {
			t.Errorf("--runs %d wrote %v, want nothing", runs, written)
		}
	}
}

// TestReviewWithoutAnAnswerHasNoDrawToCount: the bare form prints the question and
// asks nobody, so a count of answers means nothing there. The refusal names the
// flag that gives it a meaning, the way --model does.
func TestReviewWithoutAnAnswerHasNoDrawToCount(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"

	result := p.run(review.Options{Runs: 3})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want blocked (%s)", result.Code, result.Message)
	}
	for _, want := range []string{"--runs 3", "--exec"} {
		if !strings.Contains(result.Message, want) {
			t.Errorf("the refusal does not name %q: %s", want, result.Message)
		}
	}
}

// TestAZeroDrawCountIsTheCallerThatNeverSaid: the Go zero value is the run that
// chose no count, and it means one - the same rule an absent field follows in a
// stored verdict. Otherwise a stored verdict would depend on which caller asked.
func TestAZeroDrawCountIsTheCallerThatNeverSaid(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.output = oneAnswer

	result := p.run(review.Options{Exec: "revisor --model {model}"})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if len(p.shell.commands) != 1 {
		t.Errorf("reviewer calls = %v, want one", p.shell.commands)
	}
}

// TestEveryDrawIsAnAnswerAndTheUnionIsTheVerdict: --runs is a count of answers, and
// the verdict is what all of them said. A defect one draw saw is still a defect: the
// other draws being quiet is one sample each, not an acquittal.
func TestEveryDrawIsAnAnswerAndTheUnionIsTheVerdict(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.script = []fakeAnswer{
		{output: warningDraw(3, "the first draw saw it")},
		{output: warningDraw(4, "the second draw saw it")},
		{output: warningDraw(5, "the third draw saw it")},
	}

	result := p.run(review.Options{Exec: "revisor --model {model}", Runs: 3})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if len(p.shell.commands) != 3 {
		t.Fatalf("reviewer calls = %d, want one per draw", len(p.shell.commands))
	}
	if got := len(filedFindings(t, p)); got != 3 {
		t.Errorf("rows = %d, want the three the draws named", got)
	}
}

// TestTwoDrawsThatNameOneDefectPayOneRow: the identity of a defect is its file, its
// bytes, its lens, its place and its severity - never the sentence that describes
// it. Two draws of one judge name it differently and it is still one row, with the
// words the first payment bought.
func TestTwoDrawsThatNameOneDefectPayOneRow(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.script = []fakeAnswer{
		{output: warningDraw(3, "the first draw said it this way")},
		{output: warningDraw(3, "the second draw said it another way")},
	}

	if result := p.run(review.Options{Exec: "revisor --model {model}", Runs: 2}); result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	findings := filedFindings(t, p)
	if len(findings) != 1 {
		t.Fatalf("rows = %d, want one: %v", len(findings), findings)
	}
	if findings[0].Why != "the first draw said it this way" {
		t.Errorf("the row kept %q, want the words of the draw that paid first", findings[0].Why)
	}
}

// TestTheVerdictSaysHowManyDrawsPaidForEachLens: the count is in the file, so a
// reader of an approval can tell one answer from three without asking the person
// who paid. One landing is one answer.
func TestTheVerdictSaysHowManyDrawsPaidForEachLens(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.output = oneAnswer

	if result := p.run(review.Options{Exec: "revisor --model {model}", Runs: 3}); result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	if got := filedVerdict(t, p).DrawsFor("reliability"); got != 3 {
		t.Errorf("draws for reliability = %d, want the three answers that landed", got)
	}
}

// TestShowPrintsTheDrawsBesideTheLenses: --show is where a person reads what a tree
// holds, so the count has to be there. A verdict approved by three answers must not
// read like one approved by one.
func TestShowPrintsTheDrawsBesideTheLenses(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.output = oneAnswer
	if result := p.run(review.Options{Exec: "revisor --model {model}", Runs: 3}); result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	shown := strings.Join(p.run(review.Options{Show: true}).Lines, "\n")

	if !strings.Contains(shown, "reliability (3 draws)") {
		t.Errorf("--show does not print the count beside the lens:\n%s", shown)
	}
}

// TestAVerdictStoredBeforeTheFieldReadsAsOneDraw: every verdict already in a store
// was paid by one answer, and the field that says so arrived later. Reading silence
// as zero would turn the whole store into reviews nobody paid for.
func TestAVerdictStoredBeforeTheFieldReadsAsOneDraw(t *testing.T) {
	p := newProject()
	p.files.files[storePath(currentTree)] = domainreview.Render(warningVerdict(t, currentTree))

	shown := strings.Join(p.run(review.Options{Show: true}).Lines, "\n")

	if !strings.Contains(shown, "reliability (1 draw)") {
		t.Errorf("--show does not read an absent count as one draw:\n%s", shown)
	}
}

// drawsVerdict is a verdict this tree already holds: paid by that many answers, and
// naming a warning at each line it is given. It is what a top-up finds on disk.
func drawsVerdict(t *testing.T, draws int, lines ...int) domainreview.Verdict {
	t.Helper()
	verdict := domainreview.Verdict{
		Reviewer: "revisor",
		Lenses:   []string{"reliability"},
		Draws:    map[string]int{"reliability": draws},
	}
	for _, line := range lines {
		verdict.Findings = append(verdict.Findings, domainreview.Finding{
			Path:     "a.go",
			Line:     line,
			Severity: domainreview.Warning,
			Causal:   true,
			Why:      "the draw before this one saw it",
			Lens:     "reliability",
		})
	}
	sealed, err := domainreview.Seal(verdict, currentTree, "HEAD", now)
	if err != nil {
		t.Fatalf("sealing the verdict this tree already holds: %v", err)
	}
	return sealed
}

// TestALaterRunPaysOnlyTheDrawsThatAreMissing: a lens paid by one answer still owes
// the two a run of three asks for, and the run pays those two -not three, and not
// nothing. The count is what makes "the lens already paid for" a fact instead of a
// refusal.
func TestALaterRunPaysOnlyTheDrawsThatAreMissing(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.files.files[storePath(currentTree)] = domainreview.Render(drawsVerdict(t, 1))
	p.shell.output = oneAnswer

	result := p.run(review.Options{Exec: "revisor --model {model}", Runs: 3})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if len(p.shell.commands) != 2 {
		t.Errorf("reviewer calls = %d, want only the two draws missing", len(p.shell.commands))
	}
	if got := filedVerdict(t, p).DrawsFor("reliability"); got != 3 {
		t.Errorf("draws = %d, want the three the file now counts", got)
	}
}

// TestATopUpAddsItsRowsToTheOnesAlreadyPaid: the union grows across runs. The rows a
// first run paid for stand, the new answer adds what it found, and the file is one
// verdict with a count.
func TestATopUpAddsItsRowsToTheOnesAlreadyPaid(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.files.files[storePath(currentTree)] = domainreview.Render(drawsVerdict(t, 1, 3))
	p.shell.output = warningDraw(5, "the second draw saw it")

	if result := p.run(review.Options{Exec: "revisor --model {model}", Runs: 2}); result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	findings := filedFindings(t, p)
	if len(findings) != 2 {
		t.Fatalf("rows = %d, want the one that stood and the one this draw added: %v", len(findings), findings)
	}
	if got := filedVerdict(t, p).DrawsFor("reliability"); got != 2 {
		t.Errorf("draws = %d, want 2", got)
	}
}

// TestARunThatAsksForNoNewDrawIsRefusedNamingTheCount: asking for three on a lens
// that already has three pays nothing, and saying so with the count is what lets a
// person ask for a fourth instead of guessing why the run refused.
func TestARunThatAsksForNoNewDrawIsRefusedNamingTheCount(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.files.files[storePath(currentTree)] = domainreview.Render(drawsVerdict(t, 3))
	p.shell.output = oneAnswer

	result := p.run(review.Options{Exec: "revisor --model {model}", Runs: 3})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want blocked (%s)", result.Code, result.Message)
	}
	if len(p.shell.commands) != 0 {
		t.Errorf("reviewer calls = %v, want none", p.shell.commands)
	}
	for _, want := range []string{"3 draws", "--runs 4"} {
		if !strings.Contains(strings.Join(append([]string{result.Message}, result.Lines...), "\n"), want) {
			t.Errorf("the refusal does not name %q: %s %v", want, result.Message, result.Lines)
		}
	}
}

// TestADrawThatDoesNotAnswerLeavesTheOnesThatDid: the answers that landed are paid
// for, and a reviewer that fails on the third does not throw them away. The verdict
// says what it holds, the run names the draw it lost, and the command that pays the
// rest is a top-up -the one thing that makes the loss recoverable.
func TestADrawThatDoesNotAnswerLeavesTheOnesThatDid(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.script = []fakeAnswer{
		{output: warningDraw(3, "the first draw saw it")},
		{output: warningDraw(4, "the second draw saw it")},
		{output: "boom: the model did not answer", exitCode: 1},
	}

	result := p.run(review.Options{Exec: "revisor --model {model}", Runs: 3})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want blocked: the verb did not do what it was asked (%s)", result.Code, result.Message)
	}
	held := filedVerdict(t, p)
	if got := held.DrawsFor("reliability"); got != 2 {
		t.Errorf("draws = %d, want the two answers that landed", got)
	}
	if len(held.Findings) != 2 {
		t.Errorf("rows = %d, want the two defects the draws named", len(held.Findings))
	}
	said := strings.Join(append([]string{result.Message}, result.Lines...), "\n")
	for _, want := range []string{"draw 3 of 3", "holds 2 draws", "--runs 3"} {
		if !strings.Contains(said, want) {
			t.Errorf("the run does not say %q:\n%s", want, said)
		}
	}
}

// TestADrawThatFailsWithNothingLandedIsTheRefusalItAlwaysWas: no answer exists, so
// there is no count to report and nothing to keep. The run fails like it always did
// and invents no verdict.
func TestADrawThatFailsWithNothingLandedIsTheRefusalItAlwaysWas(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.script = []fakeAnswer{{output: "boom: the model did not answer", exitCode: 1}}

	result := p.run(review.Options{Exec: "revisor --model {model}", Runs: 3})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want blocked (%s)", result.Code, result.Message)
	}
	if _, stored := p.files.files[storePath(currentTree)]; stored {
		t.Error("a verdict was stored with no answer")
	}
	if strings.Contains(strings.Join(result.Lines, "\n"), "did not answer, and") {
		t.Errorf("the run talks about a count with nothing landed:\n%s", strings.Join(result.Lines, "\n"))
	}
}
