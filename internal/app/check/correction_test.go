package check_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/check"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
)

// TestTheGateNamesTheCorrectionWhenAPredecessorExists: a tree with no verdict may be
// the correction of one the store holds for the same feature. The gate names the
// command that pays for the delta, so the human does not have to go looking for a
// hash the store already knows.
func TestTheGateNamesTheCorrectionWhenAPredecessorExists(t *testing.T) {
	const base, predecessor = "abc1234", "2222222222222222222222222222222222222222"

	files := readyFiles()
	files["odd/tasks/0007-turnos.md"] = reviewableDoc(base, true)
	seeded := newFakeFiles(files)
	seeded.files[reviewPath(predecessor)] = []byte(verdictAt(t, predecessor, base))

	result := run(t, seeded, &fakeGit{tree: currentTree, changed: []string{"internal/app/check/check.go"}},
		check.Options{RequireReview: true})

	if result.Code != 2 {
		t.Fatalf("code = %d, want 2 (%s)", result.Code, result.Message)
	}
	lines := strings.Join(result.Lines, "\n")
	if !strings.Contains(lines, "odachi review --since "+predecessor) {
		t.Errorf("the refusal does not name the correction:\n%s", lines)
	}
}

// TestTheGateDoesNotOfferACorrectionItCannotName: with no verdict for this feature's
// base, there is nothing to correct, and a suggestion naming no tree would be noise.
func TestTheGateDoesNotOfferACorrectionItCannotName(t *testing.T) {
	files := readyFiles()
	files["odd/tasks/0007-turnos.md"] = reviewableDoc("abc1234", true)

	result := run(t, newFakeFiles(files), &fakeGit{tree: currentTree, changed: []string{"internal/app/check/check.go"}},
		check.Options{RequireReview: true})

	if result.Code != 2 {
		t.Fatalf("code = %d, want 2 (%s)", result.Code, result.Message)
	}
	if strings.Contains(strings.Join(result.Lines, "\n"), "odachi review --since") {
		t.Errorf("the refusal offers a correction that does not exist:\n%s", strings.Join(result.Lines, "\n"))
	}
}

// verdictAt seals one verdict for a tree and the base a document names.
func verdictAt(t *testing.T, tree, base string) string {
	t.Helper()
	sealed, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "before",
		Model:    "other/model",
		Lenses:   []string{domainreview.Reliability},
	}, tree, base, reviewInstant)
	if err != nil {
		t.Fatalf("sealing the predecessor: %v", err)
	}
	return string(domainreview.Render(sealed))
}
