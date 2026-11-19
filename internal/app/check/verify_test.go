package check_test

import (
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/app/check"
	"github.com/biggs-100/odachi/internal/domain/gate"
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
		// The feature wrote code, whatever the closing candidate carries: the
		// question the gate asks is the feature's diff (0318).
		between: []string{"internal/gastos/x.go"},
	}
	return files, git
}

// sliceDiff is a candidate that authors n lines of one file: the count a close
// that owes a verifier measures before it pays for the record (0331).
func sliceDiff(n int) string {
	var b strings.Builder
	b.WriteString("diff --git a/internal/gastos/x.go b/internal/gastos/x.go\n")
	b.WriteString("--- a/internal/gastos/x.go\n")
	b.WriteString("+++ b/internal/gastos/x.go\n")
	b.WriteString("@@ -1,0 +1," + strconv.Itoa(n) + " @@\n")
	for i := 1; i <= n; i++ {
		b.WriteString("+linea " + strconv.Itoa(i) + "\n")
	}
	return b.String()
}

// TestACloseThatOwesAVerifierAndPassesTheSliceBlocks: a verifier spent on a tree
// nobody can finish is a verdict that cannot land, so the close measures its
// candidate before spending it and names the number it read, the limit, and the
// door: the rest of the work is a feature of its own (0331).
func TestACloseThatOwesAVerifierAndPassesTheSliceBlocks(t *testing.T) {
	files, git := closedRiskFixture()
	git.diff = sliceDiff(401)

	result := run(t, files, git, check.Options{NoReview: true, RequireReview: true})

	if result.Code != 1 {
		t.Fatalf("code = %d, want 1: a 401-line close that owes a verifier passed (%s)", result.Code, result.Message)
	}
	if result.Reason != gate.ReasonVerifySlice {
		t.Fatalf("reason = %q, want %q: the close paid the record before weighing the tree", result.Reason, gate.ReasonVerifySlice)
	}
	lines := result.Message + "\n" + result.Next
	for _, want := range []string{"401", "400"} {
		if !strings.Contains(lines, want) {
			t.Errorf("the refusal does not name %q: %s", want, lines)
		}
	}
	if !strings.HasPrefix(result.Next, "odachi ") || !strings.Contains(result.Next, "odachi start") {
		t.Errorf("next = %q, want the door that starts the rest as a feature of its own", result.Next)
	}
}

// TestACandidateUnderTheSliceStillOwesItsRecord: the slice is the number a close
// that owes a verifier pays before spending it, and under the limit there is
// nothing to weigh. The candidate carries no record for this tree, so the reason
// is the missing one -not the slice, and not a ready tree- and it is the same at
// 399 lines and at the limit itself, because the slice starts past it (0331).
func TestACandidateUnderTheSliceStillOwesItsRecord(t *testing.T) {
	for _, n := range []int{399, 400} {
		t.Run(strconv.Itoa(n)+" lines", func(t *testing.T) {
			files, git := closedRiskFixture()
			git.diff = sliceDiff(n)

			result := run(t, files, git, check.Options{NoReview: true, RequireReview: true})

			if result.Code != 1 {
				t.Fatalf("code = %d, want 1: a %d-line close that owes a verifier and has no record passed (%s)", result.Code, n, result.Message)
			}
			if result.Reason != gate.ReasonVerifyMissing {
				t.Fatalf("reason = %q, want %q: the %d-line close was not answered with the record it owes (%s / next %s)",
					result.Reason, gate.ReasonVerifyMissing, n, result.Message, result.Next)
			}
			if !strings.Contains(result.Message, "contract") {
				t.Errorf("the refusal does not name the risk that asked for the verifier: %s", result.Message)
			}
			if !strings.Contains(result.Next, "odachi verify") {
				t.Errorf("next = %q, want the command that writes the record", result.Next)
			}
		})
	}
}

// TestACandidatePastTheSliceWithoutAVerifierToOweIsNotRefused: the slice exists
// only for a close that owes a verifier. A big change whose commits declared no
// risk that asks for one passes exactly as it did (0331).
func TestACandidatePastTheSliceWithoutAVerifierToOweIsNotRefused(t *testing.T) {
	files, git := closedRiskFixture()
	git.diff = sliceDiff(401)
	git.messages = []string{"a change with no risk line\n"}

	if result := run(t, files, git, check.Options{NoReview: true, RequireReview: true}); result.Code != 0 {
		t.Fatalf("a close that owes no verifier paid a slice: %s", result.Message)
	}
}

// TestTheSliceIsANoteUnlessTheHumanAsks: the same tree that closes without
// `--require-review` is a note with the same two numbers. A project that does not
// owe a verifier does not owe a slice, so the exit code stays the human's (0331).
func TestTheSliceIsANoteUnlessTheHumanAsks(t *testing.T) {
	files, git := closedRiskFixture()
	git.diff = sliceDiff(401)

	result := run(t, files, git, check.Options{NoReview: true})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0: the slice stopped a tree without being asked (%s)", result.Code, result.Message)
	}
	lines := strings.Join(result.Lines, "\n")
	for _, want := range []string{"verify:", "401", "400"} {
		if !strings.Contains(lines, want) {
			t.Errorf("the ready line does not name %q:\n%s", want, lines)
		}
	}
}

// TestACandidatePastTheSliceWithARecordForAnotherTreeOwesTheRecord: the store
// holds a file under this tree's name and it speaks for another tree, so it
// answers nothing here. The close owes the record for the tree that stands and
// `odachi verify` is what writes it; `odachi start` cannot, because the work is
// already here (0331).
func TestACandidatePastTheSliceWithARecordForAnotherTreeOwesTheRecord(t *testing.T) {
	files, git := closedRiskFixture()
	git.diff = sliceDiff(401)
	other := "fedcba9876543210fedcba9876543210fedcba98"
	files.files[verifyPath(currentTree)] = []byte(verificationFile(t, other, "otro/modelo", domainverify.Verified, nil))

	result := run(t, files, git, check.Options{NoReview: true, RequireReview: true})

	if result.Code != 1 {
		t.Fatalf("code = %d, want 1: a 401-line close whose record answers another tree passed (%s)", result.Code, result.Message)
	}
	if result.Reason != gate.ReasonVerifyMissing {
		t.Fatalf("reason = %q, want %q: the slice named a door that cannot write the record (%s / next %s)",
			result.Reason, gate.ReasonVerifyMissing, result.Message, result.Next)
	}
	if !strings.Contains(result.Next, "odachi verify") {
		t.Errorf("next = %q, want the command that writes the record for this tree", result.Next)
	}
}

// TestACandidatePastTheSliceWithAnUnreadableRecordNamesTheFile: the file is
// there and nobody can read a verification out of it, so the close keeps the row
// that names the file and the command that writes it again. Starting the work as
// a feature of its own would leave the broken record where it is (0331).
func TestACandidatePastTheSliceWithAnUnreadableRecordNamesTheFile(t *testing.T) {
	files, git := closedRiskFixture()
	git.diff = sliceDiff(401)
	files.files[verifyPath(currentTree)] = []byte("basura")

	result := run(t, files, git, check.Options{NoReview: true, RequireReview: true})

	if result.Code != 1 {
		t.Fatalf("code = %d, want 1: an unreadable record closed a 401-line tree (%s)", result.Code, result.Message)
	}
	if result.Reason != gate.ReasonVerifyMissing {
		t.Fatalf("reason = %q, want %q: the slice answers a record that exists and cannot be read with a door that leaves it there (%s / next %s)",
			result.Reason, gate.ReasonVerifyMissing, result.Message, result.Next)
	}
	if !strings.Contains(result.Message, "cannot be read") {
		t.Errorf("the refusal does not name the record nobody can read: %s", result.Message)
	}
	if !strings.Contains(result.Next, "odachi verify") {
		t.Errorf("next = %q, want the command that writes the record again", result.Next)
	}
}

// TestACandidatePastTheSliceWithItsRecordCloses: the slice is what the close pays
// before spending a verifier, and this close already has the record the verifier
// signed for this exact tree. There is nothing left to weigh and nothing left to
// spend, so 401 authored lines close, and the named next is not `odachi start`
// (a door that cannot shrink a diff somebody already reviewed) (0331).
func TestACandidatePastTheSliceWithItsRecordCloses(t *testing.T) {
	files, git := closedRiskFixture()
	git.diff = sliceDiff(401)
	files.files[verifyPath(currentTree)] = []byte(verificationFile(t, currentTree, "otro/modelo", domainverify.Verified, nil))

	result := run(t, files, git, check.Options{NoReview: true, RequireReview: true})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0: a 401-line close whose record exists was refused as a slice (%s / next %s)", result.Code, result.Message, result.Next)
	}
	if result.Reason == gate.ReasonVerifySlice {
		t.Fatalf("reason = %q: the closed tree was sent to odachi start to shrink a diff already signed", result.Reason)
	}
}

// TestARefutedRecordPastTheSliceStillAnswersRefuted: the refutation leaves the
// missing list empty, so the slice's guard cannot swallow it. A 401-line close
// with a record that says the requirement does not hold keeps the sentence that
// names what showed it (0315, 0331).
func TestARefutedRecordPastTheSliceStillAnswersRefuted(t *testing.T) {
	files, git := closedRiskFixture()
	git.diff = sliceDiff(401)
	files.files[verifyPath(currentTree)] = []byte(verificationFile(t, currentTree, "otro/modelo", domainverify.Refuted,
		[]string{"odachi verify --exec with no command prints a prompt and exits 0"}))

	result := run(t, files, git, check.Options{NoReview: true, RequireReview: true})

	if result.Reason != gate.ReasonVerifyRefuted {
		t.Fatalf("reason = %q, want %q: the refuted record lost to the slice (%s)", result.Reason, gate.ReasonVerifyRefuted, result.Message)
	}
	if !strings.Contains(result.Message, "no command prints a prompt") {
		t.Errorf("the refusal does not carry the note that showed it: %s", result.Message)
	}
}

// TestADiffGitCannotReadNamesNoNumber: the measurement failed, so this row says
// nothing. The missing record still names the risk, and no number is invented for
// a tree nobody weighed (0331).
func TestADiffGitCannotReadNamesNoNumber(t *testing.T) {
	files, git := closedRiskFixture()
	git.diffErrByBase = map[string]error{"abc1234": errors.New("bad object abc1234")}

	result := run(t, files, git, check.Options{NoReview: true, RequireReview: true})

	if result.Reason == gate.ReasonVerifySlice {
		t.Fatalf("a diff git could not read refused the close as a slice: %s", result.Message)
	}
	if strings.Contains(result.Message, "401") || strings.Contains(result.Message, "400") {
		t.Errorf("the refusal invented a number: %s", result.Message)
	}
	if !strings.Contains(result.Message, "no verification") {
		t.Errorf("the close lost the row it owed: %s", result.Message)
	}
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

// TestAStoreNobodyCanProbeIsNamedAndDoesNotFireTheSlice: the close owes a
// verifier and its candidate is past the slice, but the store cannot answer
// whether this tree's record exists. An answered "there is none" is a slice; a
// question nobody could ask is not, so the slice must not fire - and the close
// says what happened instead of claiming there is no verification it can read.
// `odachi doctor` reports the health of a clone and is the verb that names why a
// store cannot be read; `odachi start` writes no record and `odachi verify` would
// write into the store nobody can look at (0332).
func TestAStoreNobodyCanProbeIsNamedAndDoesNotFireTheSlice(t *testing.T) {
	files, git := closedRiskFixture()
	git.diff = sliceDiff(401)
	files.denyExists = func(path string) error {
		if filepath.ToSlash(path) == filepath.ToSlash(verifyPath(currentTree)) {
			return errors.New("permission denied")
		}
		return nil
	}

	result := run(t, files, git, check.Options{NoReview: true, RequireReview: true})

	if result.Reason == gate.ReasonVerifySlice {
		t.Fatalf("reason = %q: a store nobody could probe fired the slice, so the close claims the store is empty (%s)", result.Reason, result.Message)
	}
	if !strings.Contains(result.Message, "store could not be looked at") {
		t.Errorf("the refusal does not say the store could not be looked at: %s", result.Message)
	}
	if !strings.Contains(result.Message, "permission denied") {
		t.Errorf("the refusal does not carry what the store answered: %s", result.Message)
	}
	if !strings.Contains(result.Next, "odachi check") {
		t.Errorf("next = %q, want the command that asks the question again once a person has looked at the store", result.Next)
	}
}

// TestTheStoreNobodyCanProbeIsNamedOnTheOrdinaryClose: the requirement names the
// close a person runs, which is the one without --require-review. The sentence the
// store earns is written before the softening, and the softening rebuilds its note
// from the verdict it already had, so that order is what makes the note true. A
// test that only runs with --require-review never sees it (0332).
func TestTheStoreNobodyCanProbeIsNamedOnTheOrdinaryClose(t *testing.T) {
	files, git := closedRiskFixture()
	git.diff = sliceDiff(401)
	files.denyExists = func(path string) error {
		if filepath.ToSlash(path) == filepath.ToSlash(verifyPath(currentTree)) {
			return errors.New("permission denied")
		}
		return nil
	}

	result := run(t, files, git, check.Options{NoReview: true})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0: a store nobody could probe stopped an ordinary close (%s)", result.Code, result.Message)
	}
	lines := strings.Join(result.Lines, "\n")
	for _, want := range []string{"verify:", "store could not be looked at", "permission denied", "odachi check"} {
		if !strings.Contains(lines, want) {
			t.Errorf("the note does not name %q:\n%s", want, lines)
		}
	}
}

// TestATreeThatCannotNameARecordAgreesWithAnUnprobeableStore: the two answers a
// close could not get -the path a tree's record would live at, and whether that
// file exists- are the same "could not answer", so they resolve the same way: no
// slice, and the sentence the store earns. Resolving them in opposite directions
// is how a store nobody can read earned the number and a door that cannot write
// the record (0332).
func TestATreeThatCannotNameARecordAgreesWithAnUnprobeableStore(t *testing.T) {
	files, git := closedRiskFixture()
	git.diff = sliceDiff(401)
	git.tree = "no-es-un-hash"

	result := run(t, files, git, check.Options{NoReview: true, RequireReview: true})

	if result.Reason == gate.ReasonVerifySlice {
		t.Fatalf("reason = %q: a tree that cannot name its record fired the slice, the opposite answer to a store nobody can probe (%s)", result.Reason, result.Message)
	}
	if !strings.Contains(result.Message, "store could not be looked at") {
		t.Errorf("the refusal does not say the store could not be looked at: %s", result.Message)
	}
	if !strings.Contains(result.Message, "must be a hash") {
		t.Errorf("the refusal does not carry what could not answer: %s", result.Message)
	}
	if !strings.Contains(result.Next, "odachi check") {
		t.Errorf("next = %q, want the command that asks the question again once a person has looked at the store", result.Next)
	}
}

// TestAPassiveFeatureAsksForNoVerification: a feature that only wrote documents
// has no bytes a lens would read, and its declared risk asks for no verification
// (0301's rule, now asked of the feature's own diff).
func TestAPassiveFeatureAsksForNoVerification(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, "odd", "tasks", "0007-turnos.md")] = []byte(reviewableDoc("abc1234", true))
	git := &fakeGit{
		tree:     currentTree,
		changed:  []string{"docs/next-session.md"},
		messages: []string{"the note carries the release\n\nRisk: contract\n"},
		between:  []string{"docs/next-session.md"},
	}

	if result := run(t, files, git, check.Options{NoReview: true, RequireReview: true}); result.Code != 0 {
		t.Fatalf("a document-only feature asked for a verification: %s", result.Message)
	}
}

// TestARiskTheTableDoesNotCarryAsksForNothing: a word the table does not know is
// already named by the ready note, and inventing a demand for it would make the
// gate ask for a verifier nobody can name.
func TestARiskTheTableDoesNotCarryAsksForNothing(t *testing.T) {
	files, git := closedRiskFixture()
	git.messages = []string{"a change\n\nRisk: whatever\n"}
	git.between = []string{"internal/gastos/x.go"}

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

// TestACloseThatOnlyMovesDocumentsStillNamesTheRecord: the question is the
// feature's, not the closing candidate's. A feature that wrote code and closed with
// the mark and the note has a diff a lens would read, and its declared risk owes a
// verification even though the closing tree carries documents (0318).
func TestACloseThatOnlyMovesDocumentsStillNamesTheRecord(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, "odd", "tasks", "0007-turnos.md")] = []byte(reviewableDoc("abc1234", true))
	git := &fakeGit{
		tree:          currentTree,
		changed:       []string{"docs/next-session.md"},
		messages:      []string{"the note carries the close\n\nRisk: contract\n"},
		betweenByFrom: map[string][]string{"abc1234": {"internal/gastos/x.go"}},
	}

	result := run(t, files, git, check.Options{NoReview: true, RequireReview: true})

	if result.Code != 1 {
		t.Fatalf("code = %d, want 1: a close of documents hid the feature's code (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, "contract") {
		t.Errorf("the refusal does not name the risk: %s", result.Message)
	}
}

// TestAPassiveFeatureAsksForNothing: a feature that only wrote documents has no
// bytes a lens would read, and its declared risk asks for no verification.
func TestAPassiveFeatureAsksForNothing(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, "odd", "tasks", "0007-turnos.md")] = []byte(reviewableDoc("abc1234", true))
	git := &fakeGit{
		tree:          currentTree,
		changed:       []string{"docs/next-session.md"},
		messages:      []string{"the note carries the close\n\nRisk: contract\n"},
		betweenByFrom: map[string][]string{"abc1234": {"docs/next-session.md"}},
	}

	if result := run(t, files, git, check.Options{NoReview: true, RequireReview: true}); result.Code != 0 {
		t.Fatalf("a document-only feature asked for a verification: %s", result.Message)
	}
}
