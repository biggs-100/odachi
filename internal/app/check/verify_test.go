package check_test

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/app/check"
	domainverify "github.com/biggs-100/odachi/internal/domain/verify"
)

// verifyInstant is the instant a seeded verification records.
var verifyInstant = time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)

// verifyPath is where a record for tree lives, in the fake's spelling.
func verifyPath(tree string) string {
	return filepath.Join(store, filepath.FromSlash(domainverify.Dir), tree+".json")
}

// verificationFile seals a record for tree and renders the bytes that live on disk.
func verificationFile(t *testing.T, tree, by, verdict string, notes []string) string {
	t.Helper()
	sealed := domainverify.Record{Tree: tree, By: by, Verdict: verdict, Notes: notes}.Seal(verifyInstant)
	if err := sealed.Validate(); err != nil {
		t.Fatalf("sealing the seeded verification: %v", err)
	}
	return string(domainverify.Render(sealed))
}

// closedRiskFixture is a feature that closed, whose commits declared a risk: the
// one state a verification is owed in.
func closedRiskFixture() (*fakeFiles, *fakeGit) {
	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, "odd", "tasks", "0007-turnos.md")] = []byte(reviewableDoc("abc1234", true))
	git := &fakeGit{
		tree:     currentTree,
		changed:  []string{"internal/gastos/x.go"},
		messages: []string{"pay the invoice\n\nRisk: contract\n"},
	}
	return files, git
}

// TestAClosedRiskOwesItsVerification: the commits said what the change can hurt,
// and the tree that closes carries no record of anyone reading it besides the
// person who wrote it. The refusal names the risk and the command that produces
// the record (0315).
func TestAClosedRiskOwesItsVerification(t *testing.T) {
	files, git := closedRiskFixture()

	result := run(t, files, git, check.Options{NoReview: true, RequireReview: true})

	if result.Code != 1 {
		t.Fatalf("code = %d, want 1: a declared risk closed without a verification (%s)", result.Code, result.Message)
	}
	lines := result.Message + "\n" + result.Next
	for _, want := range []string{"contract", "odachi verify"} {
		if !strings.Contains(lines, want) {
			t.Errorf("the refusal does not name %q: %s", want, lines)
		}
	}
}

// TestARefutedVerificationBlocksNamingWhatShowedIt: the record exists and says the
// requirement does not hold. The close stops, and the note is what the verifier
// ran, so the fix is not a guess (0315).
func TestARefutedVerificationBlocksNamingWhatShowedIt(t *testing.T) {
	files, git := closedRiskFixture()
	files.files[verifyPath(currentTree)] = []byte(verificationFile(t, currentTree, "otro/modelo", domainverify.Refuted,
		[]string{"odachi verify --exec with no command prints a prompt and exits 0"}))

	result := run(t, files, git, check.Options{NoReview: true, RequireReview: true})

	if result.Code != 1 {
		t.Fatalf("code = %d, want 1: a refuted tree closed (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, "no command prints a prompt") {
		t.Errorf("the refusal does not carry the note that showed it: %s", result.Message)
	}
}

// TestAVerifiedRecordForThisTreeCloses: the same commits with the record the
// verifier signed close, so the demand is a door and not a wall.
func TestAVerifiedRecordForThisTreeCloses(t *testing.T) {
	files, git := closedRiskFixture()
	files.files[verifyPath(currentTree)] = []byte(verificationFile(t, currentTree, "otro/modelo", domainverify.Verified, nil))

	if result := run(t, files, git, check.Options{NoReview: true, RequireReview: true}); result.Code != 0 {
		t.Fatalf("the record did not close the tree: %s", result.Message)
	}
}

// TestARecordOfAnotherTreeCoversNothing: a verification speaks for the bytes it
// read, and this tree is not those bytes.
func TestARecordOfAnotherTreeCoversNothing(t *testing.T) {
	files, git := closedRiskFixture()
	other := "fedcba9876543210fedcba9876543210fedcba98"
	files.files[verifyPath(currentTree)] = []byte(verificationFile(t, other, "otro/modelo", domainverify.Verified, nil))

	result := run(t, files, git, check.Options{NoReview: true, RequireReview: true})

	if result.Code != 1 {
		t.Fatalf("code = %d, want 1: a record for another tree closed this one (%s)", result.Code, result.Message)
	}
}

// TestARecordWithoutAnAuthorDoesNotCount: the file parses and says nobody verified
// it, which is the sentence the rule exists to refuse.
func TestARecordWithoutAnAuthorDoesNotCount(t *testing.T) {
	files, git := closedRiskFixture()
	files.files[verifyPath(currentTree)] = []byte(`{"schema":1,"tree":"` + currentTree + `","by":"","at":"2026-11-16T12:00:00Z","verdict":"verified","notes":[]}`)

	result := run(t, files, git, check.Options{NoReview: true, RequireReview: true})

	if result.Code != 1 {
		t.Fatalf("code = %d, want 1: a record that names nobody closed the tree (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, "cannot be read") && !strings.Contains(result.Message, "who verified") {
		t.Errorf("the refusal does not say what is wrong with the record: %s", result.Message)
	}
}

// TestAnUnreadableRecordIsNotAnApproval: the store is plain files a human can
// break, and a record the gate cannot read leaves the tree unverified.
func TestAnUnreadableRecordIsNotAnApproval(t *testing.T) {
	files, git := closedRiskFixture()
	files.files[verifyPath(currentTree)] = []byte("basura")

	result := run(t, files, git, check.Options{NoReview: true, RequireReview: true})

	if result.Code != 1 {
		t.Fatalf("code = %d, want 1: an unreadable record closed the tree (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, "cannot be read") {
		t.Errorf("the refusal does not name the record it could not read: %s", result.Message)
	}
}

// TestADocumentOnlyChangeAsksForNoVerification: a risk is a declaration about the
// code, and a close that only moved documents has no bytes a verifier would probe
// (0301's rule, kept for the verifier).
func TestADocumentOnlyChangeAsksForNoVerification(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, "odd", "tasks", "0007-turnos.md")] = []byte(reviewableDoc("abc1234", true))
	git := &fakeGit{
		tree:     currentTree,
		changed:  []string{"docs/next-session.md"},
		messages: []string{"the note carries the release\n\nRisk: contract\n"},
	}

	if result := run(t, files, git, check.Options{NoReview: true, RequireReview: true}); result.Code != 0 {
		t.Fatalf("a document-only close asked for a verification: %s", result.Message)
	}
}

// TestARiskTheTableDoesNotCarryAsksForNothing: a word the table does not know is
// already named by the ready note, and inventing a demand for it would make the
// gate ask for a verifier nobody can name.
func TestARiskTheTableDoesNotCarryAsksForNothing(t *testing.T) {
	files, git := closedRiskFixture()
	git.messages = []string{"a change\n\nRisk: whatever\n"}

	if result := run(t, files, git, check.Options{NoReview: true, RequireReview: true}); result.Code != 0 {
		t.Fatalf("an unknown risk asked for a verification: %s", result.Message)
	}
}

// TestTheOpenFeatureOwesTheVerificationAtItsClose: mid-feature the tree moves at
// every step, so the record is owed by the tree that closes it and not by every
// check in between.
func TestTheOpenFeatureOwesTheVerificationAtItsClose(t *testing.T) {
	files := newFakeFiles(readyFiles())
	git := &fakeGit{
		tree:     currentTree,
		changed:  []string{"internal/gastos/x.go"},
		messages: []string{"pay the invoice\n\nRisk: contract\n"},
	}

	if result := run(t, files, git, check.Options{NoReview: true, RequireReview: true}); result.Code != 0 {
		t.Fatalf("an open feature asked for a verification: %s", result.Message)
	}
}

// TestAMissingVerificationIsANoteUnlessTheHumanAsks: a clean checkout cannot carry
// a record that lives in the clone, so a gate that refused it would make every pull
// request red. It is named -the hook is where a reminder is read- and
// `--require-review` is the human asking to be stopped (0317).
func TestAMissingVerificationIsANoteUnlessTheHumanAsks(t *testing.T) {
	files, git := closedRiskFixture()

	result := run(t, files, git, check.Options{NoReview: true})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0: a missing record stopped the tree without being asked (%s)", result.Code, result.Message)
	}
	lines := strings.Join(result.Lines, "\n")
	for _, want := range []string{"verify:", "contract", "odachi verify"} {
		if !strings.Contains(lines, want) {
			t.Errorf("the ready line does not name %q:\n%s", want, lines)
		}
	}
}

// TestTheRefutedRecordIsAlsoANoteUnlessTheHumanAsks: the same rule for the answer
// that exists and says no. Who wrote it and that it did not pass stay readable; the
// exit code stays the human's decision.
func TestTheRefutedRecordIsAlsoANoteUnlessTheHumanAsks(t *testing.T) {
	files, git := closedRiskFixture()
	files.files[verifyPath(currentTree)] = []byte(verificationFile(t, currentTree, "otro/modelo", domainverify.Refuted,
		[]string{"the gate closes a tree with no record"}))

	result := run(t, files, git, check.Options{NoReview: true})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0: a refuted record stopped the tree without being asked (%s)", result.Code, result.Message)
	}
	if !strings.Contains(strings.Join(result.Lines, "\n"), "the gate closes a tree with no record") {
		t.Errorf("the note does not carry what the verifier said:\n%s", strings.Join(result.Lines, "\n"))
	}
}
