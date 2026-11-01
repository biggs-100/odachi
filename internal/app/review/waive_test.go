package review_test

import (
	"strings"
	"testing"

	reviewapp "github.com/biggs-100/odachi/internal/app/review"
	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// seedTask puts one open feature document in the project, the way `start` writes it.
func seedTask(p *project, rel, title string) {
	p.files.files[fullPath(rel)] = []byte(taskdoc.TaskDoc{
		Title:        title,
		Requirements: []taskdoc.Requirement{{ID: "R1", Text: "reservar turnos"}},
		DoneWhen:     []string{"reservo un turno"},
		History:      []string{`L1 · "reservar turnos"`},
	}.Render())
}

// TestWaivingRecordsTheDecisionAndItsReason: the lane for a project with one model
// is a decision on the document, not a flag that dies with the shell. The gate reads
// the preamble, and the reason is what a reader a month later needs.
func TestWaivingRecordsTheDecisionAndItsReason(t *testing.T) {
	p := newProject()
	seedTask(p, "odd/tasks/0161-turnos.md", "turnos")

	result := reviewapp.Waive(reviewapp.WaiveOptions{Dir: dir, Because: "the project owns one model"}, p.files)
	if result.Code != reviewapp.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	stored := string(p.files.files[fullPath("odd/tasks/0161-turnos.md")])
	if !strings.Contains(stored, "Review: waived") {
		t.Errorf("the document does not carry the waiver:\n%s", stored)
	}
	if !strings.Contains(stored, "waived the review of this feature: the project owns one model") {
		t.Errorf("the document does not carry the reason:\n%s", stored)
	}

	// A second run says so and writes nothing: the decision is already made.
	before := string(p.files.files[fullPath("odd/tasks/0161-turnos.md")])
	again := reviewapp.Waive(reviewapp.WaiveOptions{Dir: dir, Because: "again"}, p.files)
	if again.Code != reviewapp.CodeOK || !strings.Contains(again.Message, "already waived") {
		t.Fatalf("second waiver = %d (%s), want 0 and 'already waived'", again.Code, again.Message)
	}
	if after := string(p.files.files[fullPath("odd/tasks/0161-turnos.md")]); after != before {
		t.Errorf("the second run rewrote the document")
	}
}

// TestAWaiverWithoutAReasonIsRefused: a waiver without a reason is silence wearing
// a decision's clothes, and the refusal names the command that carries one.
func TestAWaiverWithoutAReasonIsRefused(t *testing.T) {
	p := newProject()
	seedTask(p, "odd/tasks/0161-turnos.md", "turnos")

	result := reviewapp.Waive(reviewapp.WaiveOptions{Dir: dir, Because: "   "}, p.files)

	if result.Code != reviewapp.CodeBlocked {
		t.Fatalf("code = %d, want 1 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, "--because") {
		t.Errorf("the refusal %q does not name --because", result.Message)
	}
	if strings.Contains(string(p.files.files[fullPath("odd/tasks/0161-turnos.md")]), "Review: waived") {
		t.Error("a refused waiver wrote the line")
	}
}
