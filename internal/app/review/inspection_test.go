package review_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/review"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
)

// attachInspection attaches one answer to the seeded candidate and returns the
// result. Every test here asks the same question: what did the answer say it
// read, and does that match the tree it judges.
func attachInspection(t *testing.T, p *project, answer string) review.Result {
	t.Helper()
	p.files.files[fullPath("veredicto.json")] = []byte(answer)
	return p.run(review.Options{Base: "HEAD", Attach: "veredicto.json"})
}

// TestAttachRefusesAnAnswerThatNeverSaysWhatItRead: the row is complete, the
// declaration is absent, and the answer is not sealed. This is the case that was
// sealed twice: a caller cannot answer a question only the reviewer can.
func TestAttachRefusesAnAnswerThatNeverSaysWhatItRead(t *testing.T) {
	p := newProject()
	p.git.blobs = map[string]string{"a.go": gitBlob}

	result := attachInspection(t, p, `{"reviewer":"agente","findings":[],"passes":[{"path":"a.go","lens":"reliability"}]}`)

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, "inspection") {
		t.Errorf("message = %q, want the field the answer is missing", result.Message)
	}
	if !strings.Contains(result.Message, "no verdict was written") {
		t.Errorf("message = %q, want the promise that nothing was sealed", result.Message)
	}
	if _, sealed := p.files.files[storePath(currentTree)]; sealed {
		t.Error("a verdict was sealed for an answer that never declared what it read")
	}
}

// TestAttachNamesWhatTheDeclarationLeftOut: a reviewer that read one of two
// changed files says so, and the refusal repeats the path it left out. The
// caller learns which question went unanswered, not that something did.
func TestAttachNamesWhatTheDeclarationLeftOut(t *testing.T) {
	p := newProject()
	p.git.changed = []string{"a.go", "b.go"}
	p.git.blobs = map[string]string{"a.go": gitBlob, "b.go": gitBlob}

	result := attachInspection(t, p, `{"reviewer":"agente","findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`)

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, "b.go") {
		t.Errorf("message = %q, want the changed path the declaration left out", result.Message)
	}
	if _, sealed := p.files.files[storePath(currentTree)]; sealed {
		t.Error("a verdict was sealed for a declaration that left a changed path out")
	}
}

// TestAttachRefusesADeclarationThatNamesAPathNobodyChanged: an inspection is an
// account of this candidate. A path the candidate does not change was not there
// to read, so the answer describes a tree this store does not hold.
func TestAttachRefusesADeclarationThatNamesAPathNobodyChanged(t *testing.T) {
	p := newProject()
	p.git.blobs = map[string]string{"a.go": gitBlob}

	result := attachInspection(t, p, `{"reviewer":"agente","findings":[],"inspection":{"status":"completed","paths":["a.go","notas/x.md"]}}`)

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, "notas/x.md") {
		t.Errorf("message = %q, want the path nobody changed", result.Message)
	}
}

// TestAttachEchoesAnInspectionTheReviewerCouldNotComplete: "unavailable" is the
// honest answer, and it is still not a verdict. The refusal repeats the reviewer's
// reason, so the human reads what happened instead of inventing it.
func TestAttachEchoesAnInspectionTheReviewerCouldNotComplete(t *testing.T) {
	p := newProject()
	p.git.blobs = map[string]string{"a.go": gitBlob}

	result := attachInspection(t, p, `{"reviewer":"agente","findings":[],"inspection":{"status":"unavailable","paths":[],"reason":"the reviewer timed out before it read a.go"}}`)

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, "the reviewer timed out before it read a.go") {
		t.Errorf("message = %q, want the reviewer's own reason", result.Message)
	}
	if _, sealed := p.files.files[storePath(currentTree)]; sealed {
		t.Error("a verdict was sealed for a reviewer that could not read the candidate")
	}
}

// TestAttachStoresAVerdictThatDeclaresWhatItRead: the declaration travels into
// the stored file, because the next reader - a person or a gate - has to be able
// to see what the verdict claims it covered.
func TestAttachStoresAVerdictThatDeclaresWhatItRead(t *testing.T) {
	p := newProject()
	p.git.blobs = map[string]string{"a.go": gitBlob}

	result := attachInspection(t, p, `{"reviewer":"agente","findings":[],"inspection":{"status":"completed","paths":["a.go"]},"passes":[{"path":"a.go","lens":"reliability"}]}`)

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	parsed, err := domainreview.Parse(p.files.files[storePath(currentTree)])
	if err != nil {
		t.Fatalf("the stored verdict does not parse: %v", err)
	}
	if parsed.Inspection == nil {
		t.Fatal("the stored verdict lost the declaration")
	}
	if parsed.Inspection.Status != domainreview.InspectionCompleted || !slices.Equal(parsed.Inspection.Paths, []string{"a.go"}) {
		t.Errorf("inspection = %+v, want completed over a.go", *parsed.Inspection)
	}
}
