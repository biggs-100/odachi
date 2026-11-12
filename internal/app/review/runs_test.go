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

// filedFindings reads the verdict the run filed for the current tree.
func filedFindings(t *testing.T, p *project) []domainreview.Finding {
	t.Helper()
	data, ok := p.files.files[storePath(currentTree)]
	if !ok {
		t.Fatal("the run filed no verdict for the tree")
	}
	parsed, err := domainreview.Parse(data)
	if err != nil {
		t.Fatalf("the filed verdict does not parse: %v", err)
	}
	return parsed.Findings
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
