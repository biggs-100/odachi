package check_test

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/app/check"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
)

// reviewInstant is the instant a seeded verdict records.
var reviewInstant = time.Date(2026, 10, 21, 20, 31, 0, 0, time.UTC)

// verdictFile seals a verdict for tree and renders the bytes that live on disk.
func verdictFile(t *testing.T, tree string, findings []domainreview.Finding) string {
	t.Helper()
	sealed, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "claude",
		Lenses:   []string{domainreview.Reliability},
		Findings: findings,
	}, tree, "HEAD", reviewInstant)
	if err != nil {
		t.Fatalf("sealing the seeded verdict: %v", err)
	}
	return string(domainreview.Render(sealed))
}

// reviewPath is where a verdict for tree lives, in the fake's spelling.
func reviewPath(tree string) string {
	return filepath.Join(dir, filepath.FromSlash(domainreview.Dir), tree+".json")
}

// filesWithVerdict is a ready project plus one stored verdict.
func filesWithVerdict(t *testing.T, tree string, findings []domainreview.Finding) *fakeFiles {
	t.Helper()
	files := newFakeFiles(readyFiles())
	files.files[reviewPath(tree)] = []byte(verdictFile(t, tree, findings))
	return files
}

// TestReviewDoesNotInfluenceTheGateWithoutTheFlag is the opt-in rule: a beginner
// who has never heard of a review must not be blocked by one.
func TestReviewDoesNotInfluenceTheGateWithoutTheFlag(t *testing.T) {
	files := filesWithVerdict(t, currentTree, []domainreview.Finding{
		{Path: "a.go", Severity: domainreview.Blocker, Causal: true, Why: "pierde el turno"},
	})

	result := run(t, files, &fakeGit{tree: currentTree}, check.Options{})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0: without the flag a review never blocks (%s)", result.Code, result.Message)
	}
	if result.Reason != "ready" {
		t.Errorf("reason = %q, want the gate's own ready", result.Reason)
	}
}

// TestRequireReviewWithoutAVerdictBlocks: the flag is on, this tree has no
// verdict, so a human decision is missing: exit 2, and the continuation is the
// command that produces one.
func TestRequireReviewWithoutAVerdictBlocks(t *testing.T) {
	files := newFakeFiles(readyFiles())

	result := run(t, files, &fakeGit{tree: currentTree, changed: []string{"internal/app/check/check.go"}}, check.Options{RequireReview: true})

	if result.Code != 2 {
		t.Fatalf("code = %d, want 2 (%s)", result.Code, result.Message)
	}
	if result.Message != "no review for this tree" {
		t.Errorf("message = %q, want the documented sentence", result.Message)
	}
	if result.Next != "odachi review" {
		t.Errorf("next = %q, want odachi review", result.Next)
	}
	if len(files.writes) != 0 {
		t.Fatalf("writes = %v, want none: the gate writes nothing", files.writes)
	}
}

// TestRequireReviewWithApprovedContinues: an approved verdict for THIS tree is
// the only thing that lets the gate continue.
func TestRequireReviewWithApprovedContinues(t *testing.T) {
	files := filesWithVerdict(t, currentTree, nil)

	result := run(t, files, &fakeGit{tree: currentTree, changed: []string{"internal/app/check/check.go"}}, check.Options{RequireReview: true})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if result.Reason != "ready" {
		t.Errorf("reason = %q, want ready", result.Reason)
	}
}

// TestRequireReviewWithChangesRequestedNamesTheFirstFinding: exit 2, and the
// sentence names the finding that blocks, so the human knows what to fix.
func TestRequireReviewWithChangesRequestedNamesTheFirstFinding(t *testing.T) {
	files := filesWithVerdict(t, currentTree, []domainreview.Finding{
		{Path: "estilo.go", Severity: domainreview.Warning, Causal: true, Why: "nombres largos"},
		{Path: "internal/app/check/check.go", Line: 42, Severity: domainreview.Critical, Causal: true, Why: "un fallo de lectura se informa como listo"},
		{Path: "otro.go", Severity: domainreview.Blocker, Causal: true, Why: "cobra dos veces"},
	})

	result := run(t, files, &fakeGit{tree: currentTree, changed: []string{"internal/app/check/check.go"}}, check.Options{RequireReview: true})

	if result.Code != 2 {
		t.Fatalf("code = %d, want 2 (%s)", result.Code, result.Message)
	}
	for _, want := range []string{"internal/app/check/check.go:42", "CRITICAL", "un fallo de lectura se informa como listo"} {
		if !strings.Contains(result.Message, want) {
			t.Errorf("message %q does not name %q", result.Message, want)
		}
	}
	if strings.Contains(result.Message, "cobra dos veces") {
		t.Errorf("message %q names the second blocking finding, want the first", result.Message)
	}
	if !strings.Contains(result.Next, "odachi review") {
		t.Errorf("next = %q, want a command that shows the review", result.Next)
	}
}

// TestRequireReviewIsPerTree: a verdict belongs to one tree. A different tree is
// a different candidate, and approval does not travel.
func TestRequireReviewIsPerTree(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[reviewPath(otherTree)] = []byte(verdictFile(t, otherTree, nil))

	result := run(t, files, &fakeGit{tree: currentTree, changed: []string{"internal/app/check/check.go"}}, check.Options{RequireReview: true})

	if result.Code != 2 {
		t.Fatalf("code = %d, want 2: another tree's approval must not count", result.Code)
	}
	if result.Message != "no review for this tree" {
		t.Errorf("message = %q, want the documented sentence", result.Message)
	}
}

// TestRequireReviewReportsAMangledVerdict: the store is a plain file, so a human
// can break one. A verdict that cannot be read is not an approval.
func TestRequireReviewReportsAMangledVerdict(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[reviewPath(currentTree)] = []byte("basura")

	result := run(t, files, &fakeGit{tree: currentTree, changed: []string{"internal/app/check/check.go"}}, check.Options{RequireReview: true})

	if result.Code != 2 {
		t.Fatalf("code = %d, want 2 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, currentTree) {
		t.Errorf("message = %q, want the tree named", result.Message)
	}
	if !strings.Contains(result.Message, "odachi review --attach") {
		t.Errorf("message = %q, want the command that rewrites it", result.Message)
	}
}

// TestRequireReviewComesAfterTheGatesTheAgentCanActOn: a missing proof is
// actionable, so it is reported before a missing review. Sending an agent to
// `odachi review` while the gate is failing for another reason is a dead end.
func TestRequireReviewComesAfterTheGatesTheAgentCanActOn(t *testing.T) {
	files := newFakeFiles(readyFiles())
	delete(files.files, filepath.Join(dir, "proof", "0007-turnos.txt"))

	result := run(t, files, &fakeGit{tree: currentTree}, check.Options{RequireReview: true})

	if result.Code != 1 {
		t.Fatalf("code = %d, want 1: the proof gate comes first", result.Code)
	}
	if result.Reason != "proof_missing" {
		t.Errorf("reason = %q, want proof_missing", result.Reason)
	}
}
