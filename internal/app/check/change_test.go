package check_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/check"
	"github.com/biggs-100/odachi/internal/domain/gate"
	"github.com/biggs-100/odachi/internal/domain/proof"
)

// changeLine is the trailer the fixture writes: one line, no document.
const changeLine = "bring back the wait state on the pay button"

// changeProof is the proof a document-less change owes, named after its line.
const changeProof = "proof/change-bring-back-the-wait-state-on-the-pay-button.txt"

// claimedGit is a branch whose commits name a change, and the fork point it is
// measured from.
func claimedGit() *fakeGit {
	return &fakeGit{
		tree:      currentTree,
		changed:   []string{"cmd/odachi/main.go"},
		messages:  []string{"the wait state\n\nOdachi-Change: " + changeLine + "\n"},
		mergeBase: otherTree,
	}
}

// A change that names itself in a trailer is the feature: the gate judges it
// against that line, measures it from the merge-base, and owes it the proof a
// document owes - the line is the vision, so no brief is asked either.
func TestAClaimedChangeIsJudgedAgainstItsTrailer(t *testing.T) {
	seed := closedFiles()
	seed[changeProof] = proof.Render(proof.Proof{
		Tree:    currentTree,
		Command: "go test ./...",
		Covers:  []string{"R1"},
		Output:  "ok\n",
	})
	files := newFakeFiles(seed)

	got := run(t, files, claimedGit(), check.Options{ChangeBase: "origin/main", Branch: "feat/wait"})

	if got.Code != int(gate.Ready) {
		t.Fatalf("Code = %d (%s): %s\n%s", got.Code, got.Reason, got.Message, strings.Join(got.Lines, "\n"))
	}
	if !strings.Contains(strings.Join(got.Lines, "\n"), "change-bring-back-the-wait-state") {
		t.Fatalf("the ready tree does not name the change's own proof:\n%s", strings.Join(got.Lines, "\n"))
	}
}

// The proof is named after the line, not after the last document: a change with
// no document has no document to borrow a name from.
func TestAClaimedChangeOwesItsOwnProof(t *testing.T) {
	files := newFakeFiles(closedFiles())

	got := run(t, files, claimedGit(), check.Options{ChangeBase: "origin/main", Branch: "feat/wait"})

	if got.Code != int(gate.NeedsWork) || got.Reason != gate.ReasonProofMissing {
		t.Fatalf("Code = %d (%s): %s", got.Code, got.Reason, got.Message)
	}
}

// The review of a document-less change is measured from the same fork point the
// proof is: the verdict names the commit, not the ref the caller happened to type.
func TestAClaimedChangeOwesAVerdictAtItsMergeBase(t *testing.T) {
	seed := closedFiles()
	seed[changeProof] = proof.Render(proof.Proof{
		Tree:    currentTree,
		Command: "go test ./...",
		Covers:  []string{"R1"},
		Output:  "ok\n",
	})
	files := newFakeFiles(seed)

	got := run(t, files, claimedGit(), check.Options{
		ChangeBase:    "origin/main",
		Branch:        "feat/wait",
		ReviewMode:    true,
		RequireReview: true,
	})

	if got.Code == int(gate.Ready) {
		t.Fatalf("a tree with no verdict was ready:\n%s", strings.Join(got.Lines, "\n"))
	}
	if !strings.Contains(got.Next, otherTree) {
		t.Fatalf("the next line does not name the merge-base %s: %q", otherTree, got.Next)
	}
}

// A project with a closed document and no trailer keeps the answer it always had:
// the lane is a door, not a new default.
func TestAClosedDocumentWithoutATrailerKeepsTheAnswer(t *testing.T) {
	git := claimedGit()
	git.messages = []string{"a plain commit, no trailer\n"}

	got := run(t, newFakeFiles(closedFiles()), git, check.Options{ChangeBase: "origin/main", Branch: "feat/wait"})

	if got.Code != int(gate.Ready) {
		t.Fatalf("Code = %d (%s): %s", got.Code, got.Reason, got.Message)
	}
}

// A change that names itself still lives on a branch: committing straight to the
// default branch is the one place a change is not a change.
func TestATrailerOnTheDefaultBranchIsRefused(t *testing.T) {
	got := run(t, newFakeFiles(closedFiles()), claimedGit(), check.Options{ChangeBase: "origin/main", Branch: "main"})

	if got.Code == int(gate.Ready) {
		t.Fatalf("a trailer on the default branch was ready")
	}
	if !strings.Contains(got.Message, "branch") || !strings.Contains(got.Message, "main") {
		t.Fatalf("the refusal does not name the branch: %q", got.Message)
	}
}

func TestANamedFeatureIsNotReplacedByATrailer(t *testing.T) {
	// The lane's own proof is deliberately absent: if a trailer overrode an
	// explicit --feature, this run would block for a proof the human did not ask
	// about.
	got := run(t, newFakeFiles(closedFiles()), claimedGit(), check.Options{
		ChangeBase: "origin/main",
		Branch:     "feat/wait",
		Feature:    7,
	})

	if got.Code != int(gate.Ready) {
		t.Fatalf("an explicit --feature was replaced by a trailer: %s", got.Message)
	}
	if !strings.Contains(strings.Join(got.Lines, "\n"), "proof/0007-turnos.txt") {
		t.Fatalf("the run does not name the document it was told to judge:\n%s", strings.Join(got.Lines, "\n"))
	}
}

// A trailer whose fork point cannot be found is refused by name: a change that
// cannot say where it started cannot be reviewed against anything.
func TestATrailerWithNoMergeBaseIsRefused(t *testing.T) {
	git := claimedGit()
	git.mergeBaseErr = errors.New("no merge base")

	got := run(t, newFakeFiles(closedFiles()), git, check.Options{ChangeBase: "origin/main", Branch: "feat/wait"})

	if got.Code == int(gate.Ready) {
		t.Fatalf("a trailer with no merge-base was ready")
	}
	if !strings.Contains(got.Message, "origin/main") {
		t.Fatalf("the refusal does not name the base it could not resolve: %q", got.Message)
	}
}
