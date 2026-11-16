package check_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/check"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
)

// attachedVerdict is a verdict a human or an agent wrote and attached: its own
// sentence, without the findings that would derive it.
func attachedVerdict(verdict string, findings []domainreview.Finding) string {
	return string(domainreview.Render(domainreview.Verdict{
		Reviewer: "claude",
		Base:     "abc1234",
		Lenses:   []string{domainreview.Reliability},
		Findings: findings,
		Verdict:  verdict,
	}))
}

// TestAVerdictThatAsksForChangesStopsTheClose: the artifact's own sentence is
// evidence. A verdict that says changes_requested with no row that blocks closes as
// ready today, and the gate would be contradicting the file it just read (0319).
func TestAVerdictThatAsksForChangesStopsTheClose(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, "odd", "tasks", "0007-turnos.md")] = []byte(reviewableDoc("abc1234", true))
	files.files[reviewPath(currentTree)] = []byte(attachedVerdict(domainreview.ChangesRequested, nil))

	git := &fakeGit{tree: currentTree, changed: []string{"internal/gastos/x.go"}}
	result := run(t, files, git, check.Options{RequireReview: true})

	if result.Code != 2 {
		t.Fatalf("code = %d, want 2: a verdict that asks for changes closed the tree (%s)", result.Code, result.Message)
	}
	lines := result.Message + "\n" + result.Next
	for _, want := range []string{"asks for changes", "odachi review waive"} {
		if !strings.Contains(lines, want) {
			t.Errorf("the refusal does not name %q: %s", want, lines)
		}
	}
}

// TestAnApprovedVerdictClosesNothingMore: the stop is about the sentence that asks
// for changes, and an approving verdict is not asked for anything.
func TestAnApprovedVerdictClosesNothingMore(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, "odd", "tasks", "0007-turnos.md")] = []byte(reviewableDoc("abc1234", true))
	files.files[reviewPath(currentTree)] = []byte(attachedVerdict(domainreview.Approved, nil))

	git := &fakeGit{tree: currentTree, changed: []string{"internal/gastos/x.go"}}
	result := run(t, files, git, check.Options{RequireReview: true})

	if strings.Contains(result.Message, "asks for changes") {
		t.Fatalf("an approved verdict was read as a refusal: %s", result.Message)
	}
}

// TestTheWrittenWaiverIsTheDoor: the sentence stops the close, and the human who
// decides the tree ships with it names that decision in the document. The review
// gate reads the waiver and stops asking, which is the door 0319 promises.
func TestTheWrittenWaiverIsTheDoor(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, "odd", "tasks", "0007-turnos.md")] = []byte(waivedReviewableDoc("abc1234", true))
	files.files[reviewPath(currentTree)] = []byte(attachedVerdict(domainreview.ChangesRequested, nil))

	git := &fakeGit{tree: currentTree, changed: []string{"internal/gastos/x.go"}}
	result := run(t, files, git, check.Options{RequireReview: true})

	if strings.Contains(result.Message, "asks for changes") {
		t.Fatalf("the written waiver did not answer the verdict: %s", result.Message)
	}
}
