package check_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/check"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
)

// verdictWithLenses seals a verdict that covers the given lenses.
func verdictWithLenses(t *testing.T, tree string, lenses []string, findings []domainreview.Finding) string {
	t.Helper()
	sealed, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "claude",
		Lenses:   lenses,
		Findings: findings,
	}, tree, "HEAD", reviewInstant)
	if err != nil {
		t.Fatalf("sealing the seeded verdict: %v", err)
	}
	return string(domainreview.Render(sealed))
}

// refutationPath is where a refutation for tree lives, in the fake's spelling.
func refutationPath(tree string) string {
	return filepath.Join(dir, filepath.FromSlash(domainreview.Dir), tree+".refuted.json")
}

// blockingFinding is the one finding the refutation tests argue about.
func blockingFinding() domainreview.Finding {
	return domainreview.Finding{
		Path:     "internal/auth/session.go",
		Line:     12,
		Severity: domainreview.Blocker,
		Causal:   true,
		Why:      "la sesión se acepta sin verificar la firma",
	}
}

// refutedFile is a stored refutation for one finding, with evidence.
func refutedFile(t *testing.T, tree string, reasons ...string) string {
	t.Helper()
	judgments := make([]domainreview.Judgment, 0, len(reasons))
	for i, why := range reasons {
		judgments = append(judgments, domainreview.Judgment{ID: "F" + string(rune('1'+i)), Verdict: domainreview.Refuted, Why: why})
	}

	sealed, err := domainreview.SealRefutation(domainreview.Refutation{Refuter: "claude", Refutations: judgments},
		[]domainreview.Finding{blockingFinding()}, tree, reviewInstant)
	if err != nil {
		t.Fatalf("sealing the seeded refutation: %v", err)
	}
	return string(domainreview.RenderRefutation(sealed))
}

// TestRequireReviewAsksForVerdictPerRequiredLens: a code change requires the
// reliability lens, and a verdict that covers nothing does not answer it. The
// message names the lens that is missing, so the fix is not a guess.
func TestRequireReviewAsksForVerdictPerRequiredLens(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[reviewPath(currentTree)] = []byte(verdictWithLenses(t, currentTree, nil, nil))

	git := &fakeGit{tree: currentTree, changed: []string{"internal/gastos/x.go"}}
	result := run(t, files, git, check.Options{RequireReview: true})

	if result.Code != 2 {
		t.Fatalf("code = %d, want 2 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, "reliability") {
		t.Errorf("message %q does not name the missing lens", result.Message)
	}
	if !strings.Contains(result.Next, "odachi review") {
		t.Errorf("next = %q, want a command that resolves it", result.Next)
	}
	if len(files.writes) != 0 {
		t.Fatalf("writes = %v, want none: the gate writes nothing", files.writes)
	}
}

// TestRequireReviewAcceptsACoveredLens: the same verdict with the lens the
// evidence asked for passes.
func TestRequireReviewAcceptsACoveredLens(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[reviewPath(currentTree)] = []byte(verdictWithLenses(t, currentTree, []string{domainreview.Reliability}, nil))

	git := &fakeGit{tree: currentTree, changed: []string{"internal/gastos/x.go"}}
	result := run(t, files, git, check.Options{RequireReview: true})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
}

// TestRequireReviewNamesTheFirstMissingLensOfFour: an auth change asks for the
// four, and the message names one — in the fixed order, so it is the same
// sentence on every run.
func TestRequireReviewNamesTheFirstMissingLensOfFour(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[reviewPath(currentTree)] = []byte(verdictWithLenses(t, currentTree, []string{domainreview.Reliability}, nil))

	git := &fakeGit{tree: currentTree, changed: []string{"internal/auth/session.go"}}
	result := run(t, files, git, check.Options{RequireReview: true})

	if result.Code != 2 {
		t.Fatalf("code = %d, want 2 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, domainreview.Risk) {
		t.Errorf("message %q does not name risk, the first missing lens", result.Message)
	}
}

// TestRequireReviewDoesNotAskForLensesOnADocsChange: a documentation-only
// candidate requires none, so a verdict stored before lenses existed still
// approves it. The gate is opt-in and it never invents a requirement.
func TestRequireReviewDoesNotAskForLensesOnADocsChange(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[reviewPath(currentTree)] = []byte(verdictWithLenses(t, currentTree, nil, nil))

	git := &fakeGit{tree: currentTree, changed: []string{"README.md", "docs/guia.md"}}
	result := run(t, files, git, check.Options{RequireReview: true})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
}

// TestARefutedBlockerStopsBlocking is the whole reason the refuter exists: a
// finding nobody can defend with the diff does not hold a candidate.
func TestARefutedBlockerStopsBlocking(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[reviewPath(currentTree)] = []byte(verdictWithLenses(t, currentTree, domainreview.Lenses(domainreview.Full), []domainreview.Finding{blockingFinding()}))
	files.files[refutationPath(currentTree)] = []byte(refutedFile(t, currentTree, "el diff de la línea 12 muestra que la firma ya se validaba"))

	git := &fakeGit{tree: currentTree, changed: []string{"internal/auth/session.go"}}
	result := run(t, files, git, check.Options{RequireReview: true})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0: a refuted blocker does not block (%s)", result.Code, result.Message)
	}
}

// TestAStandingBlockerStillBlocks: a refutation that answered `stands` changes
// nothing, and the gate keeps naming the finding.
func TestAStandingBlockerStillBlocks(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[reviewPath(currentTree)] = []byte(verdictWithLenses(t, currentTree, domainreview.Lenses(domainreview.Full), []domainreview.Finding{blockingFinding()}))
	files.files[refutationPath(currentTree)] = []byte(refutedFile(t, currentTree, "   "))

	git := &fakeGit{tree: currentTree, changed: []string{"internal/auth/session.go"}}
	result := run(t, files, git, check.Options{RequireReview: true})

	if result.Code != 2 {
		t.Fatalf("code = %d, want 2 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, "internal/auth/session.go:12") {
		t.Errorf("message %q does not name the finding", result.Message)
	}
}

// TestARefutationForAnotherTreeIsIgnored: a refutation is about one tree, like a
// verdict. Another tree's evidence cannot lift this one's blocker.
func TestARefutationForAnotherTreeIsIgnored(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[reviewPath(currentTree)] = []byte(verdictWithLenses(t, currentTree, domainreview.Lenses(domainreview.Full), []domainreview.Finding{blockingFinding()}))
	files.files[refutationPath(otherTree)] = []byte(refutedFile(t, otherTree, "el diff de la línea 12 muestra que ya estaba"))

	git := &fakeGit{tree: currentTree, changed: []string{"internal/auth/session.go"}}
	result := run(t, files, git, check.Options{RequireReview: true})

	if result.Code != 2 {
		t.Fatalf("code = %d, want 2: another tree's refutation must not count (%s)", result.Code, result.Message)
	}
}

// TestAMangledRefutationIsNotAnApproval: the store is a plain file a human can
// break. A refutation that cannot be read leaves the finding standing, and the
// message names the file so it can be fixed.
func TestAMangledRefutationIsNotAnApproval(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[reviewPath(currentTree)] = []byte(verdictWithLenses(t, currentTree, domainreview.Lenses(domainreview.Full), []domainreview.Finding{blockingFinding()}))
	files.files[refutationPath(currentTree)] = []byte("basura")

	git := &fakeGit{tree: currentTree, changed: []string{"internal/auth/session.go"}}
	result := run(t, files, git, check.Options{RequireReview: true})

	if result.Code != 2 {
		t.Fatalf("code = %d, want 2 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, ".refuted.json") {
		t.Errorf("message %q does not name the refutation file", result.Message)
	}
	if !strings.Contains(result.Message, "internal/auth/session.go:12") {
		t.Errorf("message %q does not name the finding that still blocks", result.Message)
	}
}
