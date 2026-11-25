package check_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/check"
	"github.com/biggs-100/odachi/internal/domain/gate"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
	domainverify "github.com/biggs-100/odachi/internal/domain/verify"
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

// citationDiff is the candidate evidence shared by the refutation fixtures.
const citationDiff = "diff --git a/internal/auth/session.go b/internal/auth/session.go\n--- a/internal/auth/session.go\n+++ b/internal/auth/session.go\n@@ -12 +12 @@\n-viejo\n+verified signature\n"

func citationGit() *fakeGit {
	return &fakeGit{tree: currentTree, changed: []string{"internal/auth/session.go"}, diff: citationDiff}
}

// refutationPath is where a refutation for tree lives, in the fake's spelling.
func refutationPath(tree string) string {
	return filepath.Join(store, filepath.FromSlash(domainreview.Dir), tree+".refuted.json")
}

// blockingFinding is the one finding the refutation tests argue about.
func blockingFinding() domainreview.Finding {
	return domainreview.Finding{
		Path:     "internal/auth/session.go",
		Line:     12,
		Severity: domainreview.Blocker,
		Causal:   true,
		Why:      "la sesión se acepta sin verificar la firma",
		Evidence: domainreview.EvidenceCited,
	}
}

// refutedFile is a stored refutation for one finding, with evidence.
func refutedFile(t *testing.T, tree string, reasons ...string) string {
	t.Helper()
	return refutedFileBy(t, "claude", tree, reasons...)
}

// refutedFileBy is the same answer signed by a named refuter: the quorum counts
// accounts, so two witnesses are two names.
func refutedFileBy(t *testing.T, refuter, tree string, reasons ...string) string {
	t.Helper()
	judgments := make([]domainreview.Judgment, 0, len(reasons))
	for i, why := range reasons {
		judgments = append(judgments, domainreview.Judgment{ID: "F" + string(rune('1'+i)), Verdict: domainreview.Refuted, Why: why})
	}

	sealed, err := domainreview.SealRefutation(domainreview.Refutation{Refuter: refuter, Refutations: judgments},
		[]domainreview.Finding{blockingFinding()}, domainreview.CitationEvidence{Tree: tree, Diff: citationDiff}, reviewInstant)
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
	files.files[reviewPath(currentTree)] = []byte(verdictAtBase(t, currentTree, "abc1234", []string{domainreview.Reliability}))

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
	files.files[reviewPath(currentTree)] = []byte(verdictAtBase(t, currentTree, "abc1234", []string{domainreview.Reliability}))

	git := citationGit()
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

// TestRequireReviewDoesNotRequireAVerdictOnADocsChange: passive is not a review
// that passed; it is a candidate that asks for no review. Requiring a verdict
// here would make the human pay for a question nobody asked.
func TestRequireReviewDoesNotRequireAVerdictOnADocsChange(t *testing.T) {
	files := newFakeFiles(readyFiles())

	git := &fakeGit{tree: currentTree, changed: []string{"README.md", "docs/guia.md"}}
	result := run(t, files, git, check.Options{RequireReview: true})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if len(files.writes) != 0 {
		t.Fatalf("writes = %v, want none: the gate writes nothing", files.writes)
	}
}

// TestARefutedBlockerStopsBlocking is the whole reason the refuter exists: a
// finding nobody can defend with the diff does not hold a candidate.
func TestARefutedBlockerStopsBlocking(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[reviewPath(currentTree)] = []byte(verdictWithLenses(t, currentTree, domainreview.Lenses(domainreview.Full), []domainreview.Finding{blockingFinding()}))
	files.files[refutationPath(currentTree)] = []byte(refutedFile(t, currentTree, "internal/auth/session.go:12 el diff de la línea 12 muestra que la firma ya se validaba"))

	git := citationGit()
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

	git := citationGit()
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
	files.files[refutationPath(otherTree)] = []byte(refutedFile(t, otherTree, "internal/auth/session.go:12 el diff de la línea 12 muestra que ya estaba"))

	git := citationGit()
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

	git := citationGit()
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

// TestADeclaredRiskAsksForItsLens: the commits say what the change can hurt, and
// the gate asks the lens that looks for it. A verdict that answered reliability and
// nothing else does not close a range whose commits declared `security`, and the
// refusal names the command: `odachi review --lens risk` (0301).
func TestADeclaredRiskAsksForItsLens(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, "odd", "tasks", "0007-turnos.md")] = []byte(reviewableDoc("abc1234", true))
	files.files[reviewPath(currentTree)] = []byte(verdictAtBase(t, currentTree, "abc1234", []string{domainreview.Reliability}))
	// The risk now also owes an independent verifier (0315), and this test is about
	// the lens: the record is here so the tree lacks exactly one thing.
	files.files[verifyPath(currentTree)] = []byte(verificationFile(t, currentTree, "otro/modelo", domainverify.Verified, nil))

	git := &fakeGit{
		tree:     currentTree,
		changed:  []string{"internal/gastos/x.go"},
		messages: []string{"pay the invoice\n\nRisk: security\n"},
	}
	result := run(t, files, git, check.Options{RequireReview: true})

	if result.Code != 2 {
		t.Fatalf("code = %d, want 2: a declared risk that nobody answered closed the tree (%s)", result.Code, result.Message)
	}
	lines := strings.Join(append([]string{result.Message}, result.Lines...), "\n")
	for _, want := range []string{"risk", "--lens risk"} {
		if !strings.Contains(lines, want) {
			t.Errorf("the refusal does not name %q:\n%s", want, lines)
		}
	}

	// The same commits with the lens answered close: the demand adds, it does not
	// become a wall.
	files.files[reviewPath(currentTree)] = []byte(verdictAtBase(t, currentTree, "abc1234", []string{domainreview.Risk, domainreview.Reliability}))
	if again := run(t, files, git, check.Options{RequireReview: true}); again.Code == 2 {
		t.Fatalf("the lens the risk asked for did not close: %s", again.Message)
	}
}

// verdictAtBase seals a verdict for tree against the base the feature started from:
// the gate compares the two, and a verdict sealed with another base answers another
// diff.
func verdictAtBase(t *testing.T, tree, base string, lenses []string) string {
	t.Helper()
	sealed, err := domainreview.Seal(domainreview.Verdict{Reviewer: "claude", Lenses: lenses}, tree, base, reviewInstant)
	if err != nil {
		t.Fatalf("sealing the seeded verdict: %v", err)
	}
	return string(domainreview.Render(sealed))
}

// featureCloseCandidate is the clean-close shape: HEAD carries only the task
// document, while the candidate read from the feature's base carries code.
func featureCloseCandidate(base string, featurePaths []string, featureDiff string) *fakeGit {
	const closingDoc = "odd/tasks/0007-turnos.md"
	return &fakeGit{
		tree:    currentTree,
		changed: []string{closingDoc},
		between: featurePaths,
		diff: "diff --git a/odd/tasks/0007-turnos.md b/odd/tasks/0007-turnos.md\n" +
			"--- a/odd/tasks/0007-turnos.md\n+++ b/odd/tasks/0007-turnos.md\n@@ -1 +1 @@\n-old\n+done\n",
		messages: []string{"close the feature\n\nRisk: none - the closing commit only marks the task complete\n"},
		blobs: map[string]string{
			closingDoc:                    checkBlob,
			"internal/app/check/check.go": checkBlob,
		},
		candidateByBase: map[string]fakeCandidate{
			base: {paths: featurePaths, diff: featureDiff},
		},
	}
}

const ordinaryFeatureDiff = "diff --git a/internal/app/check/check.go b/internal/app/check/check.go\n" +
	"--- a/internal/app/check/check.go\n+++ b/internal/app/check/check.go\n@@ -1 +1 @@\n-old\n+new\n"

func TestCleanCloseReadsRequiredLensesFromTheFeatureDiff(t *testing.T) {
	const base = "8ee3e55f31ee50a088bb1f8aa8bca4afdbc40319"
	const closingDoc = "odd/tasks/0007-turnos.md"
	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, "odd", "tasks", "0007-turnos.md")] = []byte(reviewableDoc(base, true))
	sealed, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "claude",
		Lenses:   []string{domainreview.Readability},
		Passes:   []domainreview.Pass{{Path: closingDoc, Blob: checkBlob, Lens: domainreview.Readability}},
	}, currentTree, base, reviewInstant)
	if err != nil {
		t.Fatalf("sealing the seeded verdict: %v", err)
	}
	files.files[reviewPath(currentTree)] = domainreview.Render(sealed)
	git := featureCloseCandidate(base, []string{"internal/app/check/check.go"}, ordinaryFeatureDiff)

	got := run(t, files, git, check.Options{ReviewMode: true})

	if got.Code != 2 || got.Reason != check.ReasonReviewLensMissing {
		t.Fatalf("clean close = %d/%q, want the feature's missing lens (%s)", got.Code, got.Reason, got.Message)
	}
	if !strings.Contains(got.Message, domainreview.Reliability) {
		t.Errorf("message %q does not name the feature code's reliability lens", got.Message)
	}
}

func TestCleanCloseReadsCoverageFromTheFeatureDiff(t *testing.T) {
	const base = "8ee3e55f31ee50a088bb1f8aa8bca4afdbc40319"
	const closingDoc = "odd/tasks/0007-turnos.md"
	const featureCode = "internal/app/check/check.go"
	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, "odd", "tasks", "0007-turnos.md")] = []byte(reviewableDoc(base, true))
	files.files[reviewPath(currentTree)] = []byte(verdictFileWithRowsAtBase(t, currentTree, base, []domainreview.Pass{
		{Path: closingDoc, Blob: checkBlob, Lens: domainreview.Reliability},
	}))
	git := featureCloseCandidate(base, []string{featureCode}, ordinaryFeatureDiff)

	got := run(t, files, git, check.Options{ReviewMode: true})

	if got.Code != 2 || got.Reason != check.ReasonReviewCoverage {
		t.Fatalf("a verdict covering only the closing doc = %d/%q, want feature coverage refusal (%s)", got.Code, got.Reason, got.Message)
	}
	if !strings.Contains(got.Message, featureCode) {
		t.Errorf("message %q does not name the uncovered feature path", got.Message)
	}

	files.files[reviewPath(currentTree)] = []byte(verdictFileWithRowsAtBase(t, currentTree, base, []domainreview.Pass{
		{Path: featureCode, Blob: checkBlob, Lens: domainreview.Reliability},
	}))
	if got := run(t, files, git, check.Options{ReviewMode: true}); got.Code != int(gate.Ready) {
		t.Fatalf("the feature path's current receipt did not close the tree: %d/%q (%s)", got.Code, got.Reason, got.Message)
	}
}
