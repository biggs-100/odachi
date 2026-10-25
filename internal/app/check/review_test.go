package check_test

import (
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/app/check"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
)

// reviewInstant is the instant a seeded verdict records.
var reviewInstant = time.Date(2026, 10, 21, 20, 31, 0, 0, time.UTC)

// The two blobs the git port can answer with: the one these bytes have now, and
// one a stale receipt would name.
const (
	checkBlob = "c1a0f9e8d7b6a5c4d3e2f1a0b9c8d7e6f5a4b3c2"
	otherBlob = "7d3c2b1a9e8f7a6b5c4d3e2f1a0b9c8d7e6f5a4b"
)

// verdictFile seals a verdict for tree and renders the bytes that live on disk.
func verdictFile(t *testing.T, tree string, findings []domainreview.Finding) string {
	t.Helper()
	return verdictFileWithRows(t, tree, findings, nil)
}

// verdictFileWithRows seals a verdict that also carries coverage rows (passes).
func verdictFileWithRows(t *testing.T, tree string, findings []domainreview.Finding, passes []domainreview.Pass) string {
	t.Helper()
	sealed, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "claude",
		Lenses:   []string{domainreview.Reliability},
		Findings: findings,
		Passes:   passes,
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

// proofPath is this feature's proof, in the fake's spelling.
func proofPath() string { return filepath.Join(dir, "proof", "0007-turnos.txt") }

// codeChange is a candidate that touches ordinary code, so the evidence asks for
// the reliability lens and the review gate is on.
func codeChange() *fakeGit {
	return &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/app/check/check.go"},
		blobs:   map[string]string{"internal/app/check/check.go": checkBlob},
	}
}

// verdictFileByModel seals a verdict that declares which model judged it.
func verdictFileByModel(t *testing.T, tree, model string) string {
	t.Helper()
	sealed, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "reviewer",
		Model:    model,
		Lenses:   []string{domainreview.Reliability},
	}, tree, "HEAD", reviewInstant)
	if err != nil {
		t.Fatalf("sealing the seeded verdict: %v", err)
	}
	return string(domainreview.Render(sealed))
}

// proofFileByModel seeds this tree's proof with the model that ran it.
func proofFileByModel(tree, model string) string {
	return "# tree " + tree + "\n# command go test ./...\n# model " + model + "\nok\n"
}

// TestRequireReviewRefusesASelfReview: the proof and the verdict name the same
// model, so the tree is approving itself. The gate names the identity and the
// command that fixes it, because "find another reviewer" is not a next step.
func TestRequireReviewRefusesASelfReview(t *testing.T) {
	cases := []struct{ name, proof, verdict string }{
		{
			name:    "the same model",
			proof:   proofFileByModel(currentTree, "provider/model#variant"),
			verdict: verdictFileByModel(t, currentTree, "provider/model#variant"),
		},
		{
			// A self-review must not slip through by spelling the name differently:
			// model ids are lowercase in the wild, and case is not a second model.
			name:    "the same model in another case",
			proof:   proofFileByModel(currentTree, "Provider/Model"),
			verdict: verdictFileByModel(t, currentTree, "provider/model"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			files := newFakeFiles(readyFiles())
			files.files[proofPath()] = []byte(tc.proof)
			files.files[reviewPath(currentTree)] = []byte(tc.verdict)

			got := check.Run(check.Options{Dir: dir, Base: "HEAD", TimeoutSeconds: 120, RequireReview: true},
				files, codeChange(), &fakeShell{})

			if got.Reason != check.ReasonReviewSelf {
				t.Fatalf("Reason = %q, want %q (message: %s)", got.Reason, check.ReasonReviewSelf, got.Message)
			}
			lines := strings.Join(got.Lines, "\n")
			for _, want := range []string{"provider/model", "odachi review --model"} {
				if !strings.Contains(strings.ToLower(lines), want) {
					t.Errorf("the refusal does not carry %q:\n%s", want, lines)
				}
			}
		})
	}
}

// A review by another model, or one that cannot be attributed (a human wrote it,
// or the proof predates the field, is not a self-review.
func TestRequireReviewAcceptsAReviewByAnotherModelOrByNobody(t *testing.T) {
	cases := []struct {
		name    string
		proof   string
		verdict string
	}{
		{"another model", proofFileByModel(currentTree, "provider/builder"), verdictFileByModel(t, currentTree, "provider/reviewer")},
		{"proof without a model", "# tree " + currentTree + "\n# command go test ./...\nok\n", verdictFileByModel(t, currentTree, "provider/reviewer")},
		{"verdict without a model", proofFileByModel(currentTree, "provider/builder"), verdictFile(t, currentTree, nil)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			files := newFakeFiles(readyFiles())
			files.files[proofPath()] = []byte(tc.proof)
			files.files[reviewPath(currentTree)] = []byte(tc.verdict)

			got := check.Run(check.Options{Dir: dir, Base: "HEAD", TimeoutSeconds: 120, RequireReview: true},
				files, codeChange(), &fakeShell{})

			if got.Reason == check.ReasonReviewSelf {
				t.Fatalf("the gate refused a review it cannot attribute to the builder: %s", got.Message)
			}
		})
	}
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

// concurrentCheckDiff is content the path class cannot see: a goroutine and a
// mutex in an otherwise ordinary file.
const concurrentCheckDiff = "diff --git a/internal/app/check/check.go b/internal/app/check/check.go\n--- a/internal/app/check/check.go\n+++ b/internal/app/check/check.go\n@@ -1 +1,2 @@\n package check\n+go func() { sync.Mutex{} }()\n"

// TestRequireReviewAsksTheRaisedClass: a changed file whose added lines spawn
// concurrency asks the four lenses. A verdict that covers only the path class is
// missing the rest, and the gate names it.
func TestRequireReviewAsksTheRaisedClass(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[reviewPath(currentTree)] = []byte(verdictFileWithRows(t, currentTree, nil, []domainreview.Pass{
		{Path: "internal/app/check/check.go", Blob: checkBlob, Lens: domainreview.Reliability},
	}))

	result := run(t, files, &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/app/check/check.go"},
		diff:    concurrentCheckDiff,
		blobs:   map[string]string{"internal/app/check/check.go": checkBlob},
	}, check.Options{RequireReview: true})

	if result.Code != 2 {
		t.Fatalf("code = %d, want 2 (%s)", result.Code, result.Message)
	}
	if result.Reason != "review_lens_missing" {
		t.Errorf("reason = %q, want review_lens_missing", result.Reason)
	}
	if !strings.Contains(result.Message, "risk") {
		t.Errorf("message = %q, want the first raised lens named", result.Message)
	}
}

// TestRequireReviewNeedsTheContentRows: declaring the lens is not enough; the
// rows are per file. A verdict that claims the four lenses but answers only the
// path-class row of the concurrent file has not covered it.
func TestRequireReviewNeedsTheContentRows(t *testing.T) {
	sealed, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "claude",
		Lenses:   []string{domainreview.Risk, domainreview.Readability, domainreview.Reliability, domainreview.Resilience},
		Findings: []domainreview.Finding{},
		Passes: []domainreview.Pass{
			{Path: "internal/app/check/check.go", Blob: checkBlob, Lens: domainreview.Reliability},
		},
	}, currentTree, "HEAD", reviewInstant)
	if err != nil {
		t.Fatalf("sealing the seeded verdict: %v", err)
	}

	files := newFakeFiles(readyFiles())
	files.files[reviewPath(currentTree)] = domainreview.Render(sealed)

	result := run(t, files, &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/app/check/check.go"},
		diff:    concurrentCheckDiff,
		blobs:   map[string]string{"internal/app/check/check.go": checkBlob},
	}, check.Options{RequireReview: true})

	if result.Code != 2 {
		t.Fatalf("code = %d, want 2 (%s)", result.Code, result.Message)
	}
	if result.Reason != "review_coverage" {
		t.Errorf("reason = %q, want review_coverage", result.Reason)
	}
	if !strings.Contains(result.Message, "risk") {
		t.Errorf("message = %q, want the missing raised lens named", result.Message)
	}
}

// TestRequireReviewBlocksWhenTheDiffCannotBeRead: the class cannot be measured
// without the diff, and a measurement nobody can make is not a pass.
func TestRequireReviewBlocksWhenTheDiffCannotBeRead(t *testing.T) {
	files := filesWithVerdict(t, currentTree, nil)

	result := run(t, files, &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/app/check/check.go"},
		diffErr: errors.New("fatal: bad revision"),
	}, check.Options{RequireReview: true})

	if result.Code == 0 {
		t.Fatal("a diff failure must not read as an assessed class")
	}
	if !strings.Contains(result.Message, "could not read") {
		t.Errorf("message = %q, want the read failure named", result.Message)
	}
}

// TestCheckReadsTheWorktreeOnce: the fence, the lens class and the proof tree
// describe the same bytes. One read answers all three, however the gate ends.
func TestCheckReadsTheWorktreeOnce(t *testing.T) {
	files := filesWithVerdict(t, currentTree, nil)
	git := &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/app/check/check.go"},
		diff:    "diff --git a/a.go b/a.go\n--- a/a.go\n+++ b/a.go\n@@ -1 +1 @@\n-viejo\n+nuevo\n",
	}

	run(t, files, git, check.Options{RequireReview: true})

	if git.candidateCalls != 1 {
		t.Errorf("candidate reads = %d, want 1: check asks the worktree once", git.candidateCalls)
	}
}

// TestADismissedFindingDoesNotBlockTheGate: a dismissal was judged with
// evidence in its verdict. The gate records it and lets the tree through — the
// receipt still covers the row, the severity no longer blocks.
func TestADismissedFindingDoesNotBlockTheGate(t *testing.T) {
	dismissed := domainreview.Finding{
		Path: "internal/app/check/check.go", Line: 12, Severity: domainreview.Critical, Causal: true,
		Why: "pierde el turno", Lens: domainreview.Reliability, Blob: checkBlob,
		Status: domainreview.Dismissed, Reason: "la transaccion lo cubre, la cita lo prueba",
	}
	dismissed.Fingerprint = domainreview.Fingerprint(dismissed)

	files := newFakeFiles(readyFiles())
	files.files[reviewPath(currentTree)] = []byte(verdictFileWithRows(t, currentTree, []domainreview.Finding{dismissed}, nil))

	result := run(t, files, &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/app/check/check.go"},
		blobs:   map[string]string{"internal/app/check/check.go": checkBlob},
	}, check.Options{RequireReview: true})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0: a dismissal never blocks (%s)", result.Code, result.Message)
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

// TestRequireReviewWithApprovedContinues continues only when the verdict can
// serve coverage: a current answer — a pass row whose blob is exactly what git
// measures now — for every lens the candidate's files ask for.
func TestRequireReviewWithApprovedContinues(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[reviewPath(currentTree)] = []byte(verdictFileWithRows(t, currentTree, nil, []domainreview.Pass{
		{Path: "internal/app/check/check.go", Blob: checkBlob, Lens: domainreview.Reliability},
	}))

	result := run(t, files, &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/app/check/check.go"},
		blobs:   map[string]string{"internal/app/check/check.go": checkBlob},
	}, check.Options{RequireReview: true})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if result.Reason != "ready" {
		t.Errorf("reason = %q, want ready", result.Reason)
	}
}

// TestRequireReviewRefusesReceiptsForOtherBytes: a pass whose blob describes
// other bytes is the potency boundary — the file changed and the old answer is
// dead. The gate says so instead of reusing it.
func TestRequireReviewRefusesReceiptsForOtherBytes(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[reviewPath(currentTree)] = []byte(verdictFileWithRows(t, currentTree, nil, []domainreview.Pass{
		{Path: "internal/app/check/check.go", Blob: otherBlob, Lens: domainreview.Reliability},
	}))

	result := run(t, files, &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/app/check/check.go"},
		blobs:   map[string]string{"internal/app/check/check.go": checkBlob},
	}, check.Options{RequireReview: true})

	if result.Code != 2 {
		t.Fatalf("code = %d, want 2 (%s)", result.Code, result.Message)
	}
	if result.Reason != "review_coverage" {
		t.Errorf("reason = %q, want review_coverage", result.Reason)
	}
	if !strings.Contains(result.Message, "internal/app/check/check.go") {
		t.Errorf("message = %q, want the file whose answer is dead", result.Message)
	}
	if result.Next != "odachi review" {
		t.Errorf("next = %q, want odachi review", result.Next)
	}
	rel := filepath.ToSlash(filepath.Join(domainreview.Dir, currentTree+".json"))
	for _, want := range []string{"delete", rel} {
		if !slices.ContainsFunc(result.Lines, func(line string) bool { return strings.Contains(line, want) }) {
			t.Errorf("lines %v do not tell the human to delete %s", result.Lines, rel)
		}
	}
}

// TestRequireReviewNamesAnOldVerdictUncomputable: a verdict written before
// content receipts carries no coverage at all. The gate reports it as not
// computable — an approval the tool cannot prove is an approval it must not use.
func TestRequireReviewNamesAnOldVerdictUncomputable(t *testing.T) {
	files := filesWithVerdict(t, currentTree, nil)

	result := run(t, files, &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/app/check/check.go"},
		blobs:   map[string]string{"internal/app/check/check.go": checkBlob},
	}, check.Options{RequireReview: true})

	if result.Code != 2 {
		t.Fatalf("code = %d, want 2 (%s)", result.Code, result.Message)
	}
	if result.Reason != "review_coverage" {
		t.Errorf("reason = %q, want review_coverage", result.Reason)
	}
	if !strings.Contains(result.Message, "before content receipts") {
		t.Errorf("message = %q, want the old-format diagnosis", result.Message)
	}
	if result.Next != "odachi review" {
		t.Errorf("next = %q, want odachi review", result.Next)
	}
}

// TestCoverageNeverAsksAboutAbsentBytes: a path git cannot hash has no bytes in
// the worktree — the candidate deleted it, and there is nothing to answer about.
// The gate demands coverage of changed content, not of absence.
func TestCoverageNeverAsksAboutAbsentBytes(t *testing.T) {
	files := filesWithVerdict(t, currentTree, nil)

	result := run(t, files, &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/app/check/check.go"},
	}, check.Options{RequireReview: true})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0: no bytes, no coverage to ask for (%s)", result.Code, result.Message)
	}
}

// TestRequireReviewBlocksWhenTheBlobCannotBeMeasured: git failing for real is
// not a coverage answer — the measurement is load-bearing, so a failure blocks
// instead of pretending the rows are current.
func TestRequireReviewBlocksWhenTheBlobCannotBeMeasured(t *testing.T) {
	files := filesWithVerdict(t, currentTree, nil)

	result := run(t, files, &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/app/check/check.go"},
		blobs:   map[string]string{"internal/app/check/check.go": checkBlob},
		blobErr: errors.New("fatal: bad object"),
	}, check.Options{RequireReview: true})

	if result.Code == 0 {
		t.Fatal("a git failure must not read as a current receipt")
	}
	if !strings.Contains(result.Message, "could not measure") {
		t.Errorf("message = %q, want the measurement failure", result.Message)
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
