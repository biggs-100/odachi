package taskdoc_test

import (
	"testing"

	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// TestTheUndoWaiverTravelsWithItsReason: the decision is the pair, and a line
// that waves at the row without a reason is prose, not a waiver (0052).
func TestTheUndoWaiverTravelsWithItsReason(t *testing.T) {
	md := "# turnos\nBase: 4a1c9f\nUndo: waived - solo cambie nombres\n\n## Requirements (verbatim, from the user)\nR1 � \"reservar turnos\"\n"

	doc := taskdoc.Parse(md)
	if !doc.UndoWaived {
		t.Fatal("the waiver was not read")
	}
	if doc.UndoWaiverReason != "solo cambie nombres" {
		t.Fatalf("reason = %q", doc.UndoWaiverReason)
	}
	if again := taskdoc.Parse(doc.Render()); !again.UndoWaived || again.UndoWaiverReason != doc.UndoWaiverReason {
		t.Fatalf("the waiver did not survive a round trip: %+v", again)
	}

	bare := taskdoc.Parse("# turnos\nUndo: waived - \n\n## Requirements (verbatim, from the user)\nR1 � \"x\"\n")
	if bare.UndoWaived {
		t.Fatal("a waiver with no reason was read as a decision")
	}
}
