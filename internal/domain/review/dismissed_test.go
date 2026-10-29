package review_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/review"
)

// gitFinding is the row these tests judge: one finding with everything the
// fingerprint reads.
func gitFinding() review.Finding {
	return review.Finding{
		Path:     "internal/app/turnos.go",
		Line:     12,
		Severity: review.Warning,
		Causal:   true,
		Why:      "the propagation is pinned by no test",
		Lens:     "reliability",
	}
}

// TestDismissedPathIsBesideTheVerdict: the judgment lives next to what it judges,
// named by the same tree, so a reader of the store sees the verdict and the human's
// answer to it without a second lookup.
func TestDismissedPathIsBesideTheVerdict(t *testing.T) {
	got, err := review.DismissedPath(tree)
	if err != nil {
		t.Fatalf("DismissedPath: %v", err)
	}
	want := review.Dir + "/" + tree + ".dismissed.json"
	if got != want {
		t.Errorf("DismissedPath = %q, want %q", got, want)
	}
	if _, err := review.DismissedPath("not a hash"); err == nil {
		t.Error("a name that is not a tree was accepted")
	}
}

// TestADismissalNeedsTheRowAndTheReason: the fingerprint binds the judgment to one
// findings row, and the reason is the evidence. Both are refused missing, and a row
// judged twice is a row with two answers.
func TestADismissalNeedsTheRowAndTheReason(t *testing.T) {
	valid := `{"dismissed":[{"fingerprint":"` + review.Fingerprint(gitFinding()) + `","path":"internal/app/turnos.go","reason":"the test is there"}]}`

	if _, err := review.ParseDismissals([]byte(valid)); err != nil {
		t.Fatalf("the valid document was refused: %v", err)
	}

	cases := map[string]string{
		"no reason":      `{"dismissed":[{"fingerprint":"abcdef0123456789","path":"a.go"}]}`,
		"no path":        `{"dismissed":[{"fingerprint":"abcdef0123456789","reason":"because"}]}`,
		"not a hash":     `{"dismissed":[{"fingerprint":"nope","path":"a.go","reason":"because"}]}`,
		"unknown field":  `{"dismissed":[{"fingerprint":"abcdef0123456789","path":"a.go","reason":"because","class":"typo"}]}`,
		"not rfc3339":    `{"dismissed":[{"fingerprint":"abcdef0123456789","path":"a.go","reason":"because","at":"ayer"}]}`,
		"negative line":  `{"dismissed":[{"fingerprint":"abcdef0123456789","path":"a.go","line":-2,"reason":"because"}]}`,
		"twice the row":  `{"dismissed":[{"fingerprint":"abcdef0123456789","path":"a.go","reason":"uno"},{"fingerprint":"abcdef0123456789","path":"a.go","reason":"dos"}]}`,
		"a verdict here": `{"reviewer":"x","findings":[]}`,
	}
	for name, data := range cases {
		if _, err := review.ParseDismissals([]byte(data)); err == nil {
			t.Errorf("%s passed", name)
		}
	}
}

// TestADismissalRetiresOnlyItsOwnRow: the fingerprint is the row's identity. A
// judgment written about one defect never retires another, and a second judgment
// about the same row replaces the first instead of stacking.
func TestADismissalRetiresOnlyItsOwnRow(t *testing.T) {
	row := review.Dismissal{
		Fingerprint: review.Fingerprint(gitFinding()),
		Path:        "internal/app/turnos.go",
		Reason:      "the test exists at line 139",
		By:          "the human",
	}
	dismissals := review.Dismissals{}.With(row)

	if got, ok := dismissals.Retires(gitFinding()); !ok || got.Reason != row.Reason {
		t.Errorf("the row was not retired: %+v %v", got, ok)
	}

	other := gitFinding()
	other.Why = "pierde el turno en la caja"
	if _, ok := dismissals.Retires(other); ok {
		t.Error("a different defect was retired by this judgment")
	}

	corrected := dismissals.With(review.Dismissal{Fingerprint: row.Fingerprint, Path: row.Path, Reason: "the test is at line 139, and it fails without the line"})
	if len(corrected.Rows) != 1 {
		t.Fatalf("rows = %d, want one judgment per row", len(corrected.Rows))
	}
	if !strings.Contains(corrected.Rows[0].Reason, "fails without") {
		t.Errorf("the corrected reason was not kept: %q", corrected.Rows[0].Reason)
	}
}

// TestDismissalsRenderCanonically: two renders of the same decision are the same
// bytes, so the file can be compared and its immutability checked.
func TestDismissalsRenderCanonically(t *testing.T) {
	document := review.Dismissals{Rows: []review.Dismissal{{
		Fingerprint: "abcdef0123456789",
		Path:        "a.go",
		Line:        12,
		Reason:      "the reviewer read another revision",
	}}}
	first := review.RenderDismissals(document)
	second := review.RenderDismissals(document)
	if string(first) != string(second) {
		t.Error("two renders differ")
	}
	parsed, err := review.ParseDismissals(first)
	if err != nil {
		t.Fatalf("the rendered document does not parse: %v", err)
	}
	if len(parsed.Rows) != 1 || parsed.Rows[0].Line != 12 {
		t.Errorf("parsed = %+v, want the row back", parsed)
	}
}
