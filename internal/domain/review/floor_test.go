package review_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/review"
)

// floorDiff is a real-shaped diff: two files, two hunks, context lines that
// must not be mistaken for the candidate's own writing.
const floorDiff = `diff --git a/a.go b/a.go
--- a/a.go
+++ b/a.go
@@ -1,3 +1,4 @@
 package a
+func Nueva() {}
 func Vieja() {}
+// fin
diff --git a/b.go b/b.go
--- a/b.go
+++ b/b.go
@@ -10,2 +10,3 @@
 func B() {}
+func C() { go func() { sync.Mutex{} }() }
`

// TestFloorReadsTheAddedLines: the floor's audit set is the new file's added
// lines. Context lines belong to nobody's citation; a finding that cites them
// cites code this candidate did not write.
func TestFloorReadsTheAddedLines(t *testing.T) {
	floor := review.FloorFor([]string{"a.go", "b.go"}, floorDiff)

	a := review.FloorRow{}
	b := review.FloorRow{}
	for _, row := range floor.Rows {
		switch row.Path {
		case "a.go":
			a = row
		case "b.go":
			b = row
		}
	}

	if !slices.Equal(a.Lines, []int{2, 4}) {
		t.Errorf("a.go lines = %v, want the two added lines 2 and 4", a.Lines)
	}
	if !slices.Equal(b.Lines, []int{11}) {
		t.Errorf("b.go lines = %v, want the added line 11: the hunk starts at 10", b.Lines)
	}
}

// TestFloorNamesTheSignals: the floor reads the added lines with a fixed
// vocabulary and says which categories they touch — the same categories the
// lenses exist for. Deterministic, no model, no filesystem.
func TestFloorNamesTheSignals(t *testing.T) {
	floor := review.FloorFor([]string{"a.go", "b.go"}, floorDiff)
	for _, row := range floor.Rows {
		switch row.Path {
		case "a.go":
			if len(row.Signals) != 0 {
				t.Errorf("a.go signals = %v, want none: a plain function and a comment say nothing", row.Signals)
			}
		case "b.go":
			if !slices.Contains(row.Signals, "concurrency") {
				t.Errorf("b.go signals = %v, want concurrency: the added line spawns a goroutine and takes a mutex", row.Signals)
			}
		}
	}
}

// TestFloorClassFollowsTheAssessment: the report says the class the lenses
// were chosen from. A document stays passive; ordinary code whose added lines
// carry a raised category reports full.
func TestFloorClassFollowsTheAssessment(t *testing.T) {
	docs := review.FloorFor([]string{"docs/notas.md"}, floorDiff)
	if docs.Class != review.Passive {
		t.Errorf("floor class for a doc = %q, want passive", docs.Class)
	}

	code := review.FloorFor([]string{"a.go", "b.go"}, floorDiff)
	if code.Class != review.Full {
		t.Errorf("floor class for code with a goroutine = %q, want full: content raises the path class", code.Class)
	}
}

// TestEarnedBlockingNeedsACitation: a finding that can block must cite the
// candidate's own writing or its immediate context. Without that citation it
// goes down one tier — blocking is earned, never declared.
func TestEarnedBlockingNeedsACitation(t *testing.T) {
	floor := review.FloorFor([]string{"a.go", "b.go"}, floorDiff)
	findings := []review.Finding{
		{Path: "a.go", Line: 2, Severity: review.Critical, Causal: true, Why: "cita una linea cambiada"},
		{Path: "a.go", Line: 20, Severity: review.Critical, Causal: true, Why: "cita lejos"},
		{Path: "a.go", Line: 0, Severity: review.Blocker, Causal: true, Why: "no cita nada"},
		{Path: "a.go", Line: 4, Severity: review.Warning, Causal: true, Why: "no bloquea, no baja"},
		{Path: "a.go", Line: 20, Severity: review.Suggestion, Causal: true, Why: "no bloquea, no baja"},
		{Path: "a.go", Line: 20, Severity: review.Critical, Causal: false, Why: "ya estaba, no bloquea"},
		{Path: "otro.go", Line: 20, Severity: review.Critical, Causal: true, Why: "no esta en el piso"},
		{Path: "a.go", Line: 7, Severity: review.Critical, Causal: true, Why: "cita el contexto inmediato"},
	}

	earned, downgraded := review.Earned(findings, floor)

	want := []review.Severity{review.Critical, review.Warning, review.Critical, review.Warning, review.Suggestion, review.Critical, review.Critical, review.Critical}
	for i, severity := range want {
		if earned[i].Severity != severity {
			t.Errorf("finding %d (%s) = %s, want %s", i+1, findings[i].Why, earned[i].Severity, severity)
		}
	}
	if downgraded != 2 {
		t.Errorf("downgraded = %d, want 2: the ones that cannot cite and can block", downgraded)
	}
}

// TestAFloorWithNoRowsJudgesNothing: a diff the floor cannot read (binary, or
// a test fixture that is not a diff) never downgrades a finding. Silence is not
// evidence of an uncited finding.
func TestAFloorWithNoRowsJudgesNothing(t *testing.T) {
	floor := review.FloorFor([]string{"a.go"}, "Binary files a/x.png and b/x.png differ\n")
	findings := []review.Finding{
		{Path: "a.go", Line: 40, Severity: review.Critical, Causal: true, Why: "no hay piso que lo juzgue"},
	}

	earned, downgraded := review.Earned(findings, floor)
	if downgraded != 0 || earned[0].Severity != review.Critical {
		t.Errorf("severity = %s, downgraded = %d; want the finding untouched", earned[0].Severity, downgraded)
	}
}

// TestFloorLinesAreCompressedForReading: the prompt says what to cite in a
// shape a human and a model read the same way.
func TestFloorLinesAreCompressedForReading(t *testing.T) {
	got := review.FloorLines([]int{1, 2, 3, 7, 9, 10})
	if got != "1-3, 7, 9-10" {
		t.Errorf("lines = %q, want the compressed ranges", got)
	}
	if text := strings.Join([]string{got}, ""); strings.Contains(text, "  ") {
		t.Errorf("lines = %q, want no double separators", got)
	}
}
