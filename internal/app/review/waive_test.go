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

// TestWaivingARowRecordsTheFingerprintAndThePerson: the decision about one blocking
// row is a different decision from "this feature lands without a verdict", and it
// names the row it answers, who answered it and when.
func TestWaivingARowRecordsTheFingerprintAndThePerson(t *testing.T) {
	p := newProject()
	seedTask(p, "odd/tasks/0161-turnos.md", "turnos")

	result := reviewapp.Waive(reviewapp.WaiveOptions{
		Dir:     dir,
		Because: "the tree ships with this row open",
		Row:     "3F2A1B4C5D6E7F80",
		By:      "biggs",
		Now:     now,
	}, p.files)
	if result.Code != reviewapp.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	stored := string(p.files.files[fullPath("odd/tasks/0161-turnos.md")])
	doc := taskdoc.Parse(stored)
	if !doc.WaivedRow("3f2a1b4c5d6e7f80") {
		t.Fatalf("the document does not carry the row:\n%s", stored)
	}
	for _, want := range []string{"by biggs", "2026-10-21T20:31:00Z", "the tree ships with this row open"} {
		if !strings.Contains(stored, want) {
			t.Errorf("the record does not carry %q:\n%s", want, stored)
		}
	}
	if doc.ReviewWaived {
		t.Error("a row waiver waived the whole feature's verdict")
	}
}

// TestARowWaiverNamesARowAndSaysWhy: a word that is not a fingerprint names nothing,
// and a decision without a reason is silence. Both are refused before the document
// is touched.
func TestARowWaiverNamesARowAndSaysWhy(t *testing.T) {
	cases := map[string]reviewapp.WaiveOptions{
		"no reason": {Because: "  ", Row: "3f2a1b4c5d6e7f80"},
		"not a row": {Because: "because", Row: "the one the review found"},
		"too short": {Because: "because", Row: "3f2a"},
	}

	for name, opts := range cases {
		t.Run(name, func(t *testing.T) {
			p := newProject()
			seedTask(p, "odd/tasks/0161-turnos.md", "turnos")
			opts.Dir = dir

			result := reviewapp.Waive(opts, p.files)
			if result.Code != reviewapp.CodeBlocked {
				t.Fatalf("code = %d, want 1 (%s)", result.Code, result.Message)
			}
			if strings.Contains(string(p.files.files[fullPath("odd/tasks/0161-turnos.md")]), "waived the row") {
				t.Error("a refused waiver wrote the line")
			}
		})
	}
}
