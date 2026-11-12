package taskdoc_test

import (
	"testing"

	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// waiving is a document whose history carries the lines under test.
func waiving(lines ...string) taskdoc.TaskDoc {
	return taskdoc.TaskDoc{
		Title:   "turnos",
		History: append([]string{`L1 · "reservar turnos"`}, lines...),
	}
}

// TestAWaivedRowIsTheFingerprintItNames: the human's decision travels with the
// repository and answers one row, so a later reader can tell which finding the
// human looked at when the gate stopped.
func TestAWaivedRowIsTheFingerprintItNames(t *testing.T) {
	doc := waiving(`L2 · "waived the row 3f2a1b4c5d6e7f80 (internal/app/check/check.go:12): the tree ships with it open, by biggs at 2026-11-12T09:00:00Z"`)

	if got := doc.WaivedRows(); len(got) != 1 || got[0] != "3f2a1b4c5d6e7f80" {
		t.Fatalf("WaivedRows = %v, want [3f2a1b4c5d6e7f80]", got)
	}
	if !doc.WaivedRow("3f2a1b4c5d6e7f80") {
		t.Error("the row the line names reads as waived")
	}
	if doc.WaivedRow("0000000000000000") {
		t.Error("a row nobody answered reads as waived")
	}
}

// TestAWaiverThatNamesNoRowWaivesNothing: prose that sounds like a decision is not
// one, and the gate must not open on a sentence.
func TestAWaiverThatNamesNoRowWaivesNothing(t *testing.T) {
	doc := waiving(`L2 · "waived the row that the review found, because it is fine"`)

	if got := doc.WaivedRows(); len(got) != 0 {
		t.Errorf("WaivedRows = %v, want none: the line names no fingerprint", got)
	}
	if doc.WaivedRow("") {
		t.Error("an empty fingerprint reads as waived")
	}
}

// TestTheSameRowWaivedTwiceIsOneDecision: the log is append-only and a human may
// write it twice; the gate reads one answer.
func TestTheSameRowWaivedTwiceIsOneDecision(t *testing.T) {
	doc := waiving(
		`L2 · "waived the row 3f2a1b4c5d6e7f80: first"`,
		`L3 · "waived the row 3F2A1B4C5D6E7F80: again, in capitals"`,
	)

	if got := doc.WaivedRows(); len(got) != 1 {
		t.Fatalf("WaivedRows = %v, want the row once", got)
	}
}

// TestAClosedFeatureSaysNothingAboutRows: the feature waiver and the row waiver are
// different decisions, and one does not read as the other.
func TestAClosedFeatureSaysNothingAboutRows(t *testing.T) {
	doc := taskdoc.TaskDoc{Title: "turnos", ReviewWaived: true, History: []string{`L1 · "waived the review of this feature: one model"`}}

	if len(doc.WaivedRows()) != 0 {
		t.Errorf("WaivedRows = %v, want none: the feature waiver names no row", doc.WaivedRows())
	}
}
