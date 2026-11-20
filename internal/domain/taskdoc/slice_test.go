package taskdoc_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// TestTheAcceptedOverageIsReadWithItsReason: the human's decision to let a close
// that owes a verifier carry more than the slice lives in the task document, in
// the same preamble that carries the base and the waivers, and it carries the
// number it accepts and the reason it accepts it (0340).
func TestTheAcceptedOverageIsReadWithItsReason(t *testing.T) {
	doc := taskdoc.Parse("# turnos\nBase: abc1234\nSlice: 438 - the ADR and the tests are the change\n\n## Requirements (verbatim, from the user)\nR1 · \"cualquier cosa\"\n")

	if doc.SliceAccepted != 438 {
		t.Errorf("SliceAccepted = %d, want 438: the number the human accepted is the one the gate weighs", doc.SliceAccepted)
	}
	if doc.SliceReason != "the ADR and the tests are the change" {
		t.Errorf("SliceReason = %q, want the sentence the human wrote: a waiver nobody can read is a silence with a number", doc.SliceReason)
	}
}

// TestAnOverageAcceptedWithoutAReasonAcceptsNothing: the reason is what makes the
// decision legible, so a line that carries the number alone accepts nothing and
// the limit stands (0340).
func TestAnOverageAcceptedWithoutAReasonAcceptsNothing(t *testing.T) {
	for _, line := range []string{"Slice: 438", "Slice: 438 -", "Slice: - the reason", "Slice: many - the reason"} {
		doc := taskdoc.Parse("# turnos\nBase: abc1234\n" + line + "\n\n## Requirements (verbatim, from the user)\nR1 · \"cualquier cosa\"\n")
		if doc.SliceAccepted != 0 || doc.SliceReason != "" {
			t.Errorf("%q accepted %d/%q: a line without a number and a reason accepts nothing", line, doc.SliceAccepted, doc.SliceReason)
		}
	}
}

// TestTheAcceptanceSurvivesTheRender: every verb that writes the document -plan,
// done, ask- renders it from what it parsed. An acceptance the render forgets is a
// decision that disappears the next time a step closes, so the round trip is the
// promise the package states and the one this line must keep (0340).
func TestTheAcceptanceSurvivesTheRender(t *testing.T) {
	doc := taskdoc.Parse("# turnos\nBase: abc1234\nSlice: 438 - the ADR and the tests are the change\n\n## Requirements (verbatim, from the user)\nR1 · \"cualquier cosa\"\n")

	again := taskdoc.Parse(doc.Render())

	if again.SliceAccepted != doc.SliceAccepted || again.SliceReason != doc.SliceReason {
		t.Errorf("the render kept %d/%q of an acceptance that read %d/%q",
			again.SliceAccepted, again.SliceReason, doc.SliceAccepted, doc.SliceReason)
	}
	if !strings.Contains(doc.Render(), "Slice: 438 - the ADR and the tests are the change") {
		t.Errorf("the render does not write the line a human reads:\n%s", doc.Render())
	}
}

// TestASliceLineInTheHistoryIsProse: the acceptance is read where the base and the
// waivers are read, in the preamble. A sentence inside a section is something a
// human wrote, not a decision the gate must obey (0340).
func TestASliceLineInTheHistoryIsProse(t *testing.T) {
	doc := taskdoc.Parse(strings.Join([]string{
		"# turnos",
		"Base: abc1234",
		"",
		"## Requirements (verbatim, from the user)",
		"R1 · \"cualquier cosa\"",
		"",
		"## History",
		"L1 · Slice: 438 - the ADR and the tests are the change",
	}, "\n"))

	if doc.SliceAccepted != 0 {
		t.Errorf("SliceAccepted = %d, want 0: a line in the history is prose, not a decision", doc.SliceAccepted)
	}
}
