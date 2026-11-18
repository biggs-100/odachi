package review

import (
	"strings"
	"testing"

	domain "github.com/biggs-100/odachi/internal/domain/review"
)

// TestTheDeclarationIsJudgedAgainstTheCandidateNotThePart: a prompt too big is
// split, so one reviewer reads a part of the candidate while the verdict answers
// the whole tree. A path the part's diff does not carry but the candidate does is
// not an invention -and refusing it made --exec reject an answer --attach accepted
// for the same bytes-. What is an invention is a path outside the candidate.
func TestTheDeclarationIsJudgedAgainstTheCandidateNotThePart(t *testing.T) {
	t.Parallel()

	candidate := []string{"a.go", "b.go"}
	part := []string{"a.go"}

	// b.go belongs to another part: the reviewer read the frozen candidate, which
	// holds it, and the prompt's slice does not make it somebody else's file.
	read := domain.Verdict{Inspection: &domain.Inspection{
		Status: domain.InspectionCompleted,
		Paths:  []string{"a.go", "b.go"},
	}}
	if message, _, ok := inspectionRefusal(read, part, candidate); !ok {
		t.Errorf("a path the candidate changes was refused as foreign: %s", message)
	}

	// c.go is in no diff: the reviewer invented it.
	invented := domain.Verdict{Inspection: &domain.Inspection{
		Status: domain.InspectionCompleted,
		Paths:  []string{"a.go", "c.go"},
	}}
	message, _, ok := inspectionRefusal(invented, part, candidate)
	if ok {
		t.Fatal("a path the candidate does not change was accepted")
	}
	if !strings.Contains(message, "c.go") {
		t.Errorf("the refusal does not name the invented path: %q", message)
	}

	// The part's own path left out is still a hole: coverage is the slice's.
	hole := domain.Verdict{Inspection: &domain.Inspection{
		Status: domain.InspectionCompleted,
		Paths:  []string{"b.go"},
	}}
	message, _, ok = inspectionRefusal(hole, part, candidate)
	if ok {
		t.Error("a changed path the part carries and the reviewer left out was accepted")
	}
	if !strings.Contains(message, "a.go") {
		t.Errorf("the refusal does not name the path left out: %q", message)
	}
}
