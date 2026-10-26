package review_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/review"
)

// carriedPrompt is a candidate whose store carries the rows under test. The
// prompt is the only thing asserted: what the store holds is another test's
// business.
func carriedPrompt(findings []review.Finding, passes []review.Pass) string {
	in := promptInput()
	in.Known = findings
	in.KnownPasses = passes
	return review.Prompt(in)
}

// count is how many times one row is printed. A reviewer reads the section to
// know what not to re-report; the second copy teaches it nothing.
func count(text, row string) int {
	return strings.Count(text, row)
}

// TestCarriedSectionPrintsEachRowOnce: a row travels from verdict to verdict as
// it is inherited, so the store accumulates copies of the same finding. The
// prompt says once what is already carried: the same sentence twice is a cost
// paid on every review of every tree below it, not a warning.
func TestCarriedSectionPrintsEachRowOnce(t *testing.T) {
	finding := review.Finding{
		Path: "a.go", Line: 3, Lens: review.Reliability,
		Severity: review.Warning, Why: "pierde el turno", Blob: "b",
	}
	clean := review.Pass{Path: "a.go", Lens: review.Reliability, Blob: "b"}

	got := carriedPrompt(
		[]review.Finding{finding, finding, finding},
		[]review.Pass{clean, clean},
	)

	if n := count(got, "- a.go:3 (reliability, WARNING) pierde el turno\n"); n != 1 {
		t.Errorf("the finding is printed %d times, want once:\n%s", n, carriedRows(got))
	}
	if n := count(got, "- a.go (reliability) audited clean\n"); n != 1 {
		t.Errorf("the clean row is printed %d times, want once:\n%s", n, carriedRows(got))
	}
}

// TestCarriedSectionKeepsRowsThatDiffer: the dedupe is about the same row, not
// about the same file. Two defects on one path, one line apart, or under another
// lens, are different answers and the reviewer must hear both.
func TestCarriedSectionKeepsRowsThatDiffer(t *testing.T) {
	base := review.Finding{
		Path: "a.go", Line: 3, Lens: review.Reliability,
		Severity: review.Warning, Why: "pierde el turno", Blob: "b",
	}
	otherWhy := base
	otherWhy.Why = "devuelve otro turno"
	otherLine := base
	otherLine.Line = 7
	otherLens := base
	otherLens.Lens = review.Risk

	got := carriedPrompt([]review.Finding{base, otherWhy, otherLine, otherLens}, nil)

	for _, want := range []string{
		"(reliability, WARNING) pierde el turno",
		"(reliability, WARNING) devuelve otro turno",
		"a.go:7 (reliability, WARNING)",
		"(risk, WARNING) pierde el turno",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("a row that differs was dropped: missing %q\n%s", want, carriedRows(got))
		}
	}
}

// TestCarriedSectionOrderIsTheStores: two runs against the same store print the
// same prompt, so two reviews are comparable and a re-run does not look like a
// different question.
func TestCarriedSectionOrderIsTheStores(t *testing.T) {
	first := review.Finding{Path: "a.go", Line: 3, Lens: review.Reliability, Severity: review.Warning, Why: "uno", Blob: "b"}
	second := review.Finding{Path: "b.go", Line: 1, Lens: review.Reliability, Severity: review.Warning, Why: "dos", Blob: "c"}
	clean := review.Pass{Path: "c.go", Lens: review.Reliability, Blob: "d"}

	rows := []review.Finding{first, second, first}
	want := carriedPrompt(rows, []review.Pass{clean})
	again := carriedPrompt(rows, []review.Pass{clean})

	if want != again {
		t.Error("two prompts over the same store differ")
	}
	if a, b := strings.Index(want, "a.go:3"), strings.Index(want, "b.go:1"); a > b {
		t.Errorf("the rows are not in the store's order:\n%s", carriedRows(want))
	}
}

// carriedRows is the section alone, to keep a failure readable.
func carriedRows(prompt string) string {
	const header = "## What the store already carries"
	start := strings.Index(prompt, header)
	if start < 0 {
		return "(no carried section)"
	}
	end := strings.Index(prompt[start:], "\n## ")
	if end < 0 {
		return prompt[start:]
	}
	return prompt[start : start+end]
}
