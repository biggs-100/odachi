package check_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/check"
	"github.com/biggs-100/odachi/internal/domain/gate"
	"github.com/biggs-100/odachi/internal/domain/proof"
	"github.com/biggs-100/odachi/internal/domain/undone"
)

// changeLine is the trailer the fixture writes: one line, no document.
const changeLine = "bring back the wait state on the pay button"

// changeProof is the proof a document-less change owes, named after its line.
const changeProof = "proof/change-bring-back-the-wait-state-on-the-pay-button.txt"

// changeUndo is the row the same close owes: the evidence lives beside the proof,
// and a close that does not take it again reads it back (0052).
const changeUndo = "proof/change-bring-back-the-wait-state-on-the-pay-button.undone.txt"

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
	seed[changeUndo] = undone.Render(undone.Evidence{Tree: currentTree, Command: "go test ./...", Verdict: undone.Noticed, Output: "--- FAIL: TestX\n"})
	seed[changeUndo] = undone.Render(undone.Evidence{Tree: currentTree, Command: "go test ./...", Verdict: undone.Noticed, Output: "--- FAIL: TestX\n"})
	files := newFakeFiles(seed)

	got := run(t, files, claimedGit(), check.Options{ChangeBase: "origin/main"})

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

	got := run(t, files, claimedGit(), check.Options{ChangeBase: "origin/main"})

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
	seed[changeUndo] = undone.Render(undone.Evidence{Tree: currentTree, Command: "go test ./...", Verdict: undone.Noticed, Output: "--- FAIL: TestX\n"})
	seed[changeUndo] = undone.Render(undone.Evidence{Tree: currentTree, Command: "go test ./...", Verdict: undone.Noticed, Output: "--- FAIL: TestX\n"})
	files := newFakeFiles(seed)

	got := run(t, files, claimedGit(), check.Options{
		ChangeBase:    "origin/main",
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

	got := run(t, newFakeFiles(closedFiles()), git, check.Options{ChangeBase: "origin/main"})

	if got.Code != int(gate.Ready) {
		t.Fatalf("Code = %d (%s): %s", got.Code, got.Reason, got.Message)
	}
}

// A change that names itself still lives on a branch: committing straight to the
// default branch is the one place a change is not a change. The repository says
// which branch that is; the gate no longer compares a name it was handed.
func TestATrailerOnTheDefaultBranchIsRefused(t *testing.T) {
	git := claimedGit()
	git.onDefault = true
	git.defaultBranch = "main"

	got := run(t, newFakeFiles(closedFiles()), git, check.Options{ChangeBase: "origin/main"})

	if got.Code == int(gate.Ready) {
		t.Fatalf("a trailer on the default branch was ready")
	}
	if !strings.Contains(got.Message, "branch") || !strings.Contains(got.Message, "main") {
		t.Fatalf("the refusal does not name the branch: %q", got.Message)
	}
}

// A repository that cannot say which branch is the default blocks instead of
// judging the trailer: a check that runs only when git answers is a silent
// degradation, and this lane has no second opinion (0174).
func TestATrailerBlocksWhenTheDefaultBranchCannotBeAsked(t *testing.T) {
	git := claimedGit()
	git.onDefaultErr = errors.New("cannot read the refs")

	got := run(t, newFakeFiles(closedFiles()), git, check.Options{ChangeBase: "origin/main"})

	if got.Code == int(gate.Ready) {
		t.Fatalf("a trailer was judged while the default branch was unknown")
	}
	if !strings.Contains(got.Message, "cannot read the refs") {
		t.Fatalf("the refusal does not carry git's reason: %q", got.Message)
	}
}

func TestANamedFeatureIsNotReplacedByATrailer(t *testing.T) {
	// The lane's own proof is deliberately absent: if a trailer overrode an
	// explicit --feature, this run would block for a proof the human did not ask
	// about.
	got := run(t, newFakeFiles(closedFiles()), claimedGit(), check.Options{
		ChangeBase: "origin/main",
		Feature:    7,
	})

	if got.Code != int(gate.Ready) {
		t.Fatalf("an explicit --feature was replaced by a trailer: %s", got.Message)
	}
	if !strings.Contains(strings.Join(got.Lines, "\n"), "proof/0007-turnos.txt") {
		t.Fatalf("the run does not name the document it was told to judge:\n%s", strings.Join(got.Lines, "\n"))
	}
}

// commitMessage is the file git hands the commit-msg hook: the message, before
// it is a commit.
const commitMessage = ".git/COMMIT_EDITMSG"

// withMessage seeds a project with the message git hands the commit-msg hook.
func withMessage(seed map[string]string, message string) *fakeFiles {
	seed[commitMessage] = message
	return newFakeFiles(seed)
}

// messageGit is a branch that is not the default one: the lane lives on a branch.
func messageGit() *fakeGit {
	return &fakeGit{
		tree:          currentTree,
		changed:       []string{"cmd/odachi/main.go"},
		mergeBase:     otherTree,
		defaultBranch: "main",
	}
}

// The hook that sees the message judges the change before the commit exists and
// defers its proof: no commit carries the tree yet, so the proof is owed by the
// one that will.
func TestAMessageThatNamesAChangeDefersItsProof(t *testing.T) {
	files := withMessage(closedFiles(), "the wait state\n\nOdachi-Change: "+changeLine+"\n")

	got := run(t, files, messageGit(), check.Options{ChangeBase: "origin/main", Message: commitMessage})

	if got.Code != int(gate.Ready) {
		t.Fatalf("Code = %d (%s): %s\n%s", got.Code, got.Reason, got.Message, strings.Join(got.Lines, "\n"))
	}
	if !strings.Contains(got.Message, "proof is owed") {
		t.Errorf("Message = %q, want the proof named as owed after the commit", got.Message)
	}
}

// A message with no trailer keeps the answer the pre-commit hook always gave:
// the newest document is the feature and its proof is the one that counts.
func TestAMessageWithNoTrailerKeepsTheAnswer(t *testing.T) {
	files := withMessage(closedFiles(), "a plain commit, no trailer\n")

	got := run(t, files, messageGit(), check.Options{ChangeBase: "origin/main", Message: commitMessage})

	if got.Code != int(gate.Ready) {
		t.Fatalf("Code = %d (%s): %s", got.Code, got.Reason, got.Message)
	}
}

// The lane lives on a branch, and the message is where that is decided before the
// commit exists.
func TestAMessageOnTheDefaultBranchIsRefused(t *testing.T) {
	files := withMessage(closedFiles(), "the wait state\n\nOdachi-Change: "+changeLine+"\n")
	git := messageGit()
	git.onDefault = true

	got := run(t, files, git, check.Options{ChangeBase: "origin/main", Message: commitMessage})

	if got.Code == int(gate.Ready) {
		t.Fatalf("a trailer on the default branch was ready")
	}
	if !strings.Contains(got.Message, "branch") || !strings.Contains(got.Message, "main") {
		t.Fatalf("the refusal does not name the branch: %q", got.Message)
	}
}

// The message is read by the same parser as the commits, so its refusals are the
// same refusals.
func TestAMessageRefusalIsNamed(t *testing.T) {
	cases := []struct {
		name    string
		message string
		want    string
	}{
		{name: "empty line", message: "x\n\nOdachi-Change:   \n", want: "Odachi-Change"},
		{
			name:    "two lines",
			message: "x\n\nOdachi-Change: one line\nOdachi-Change: another line\n",
			want:    "different",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			files := withMessage(closedFiles(), tc.message)

			got := run(t, files, messageGit(), check.Options{ChangeBase: "origin/main", Message: commitMessage})

			if got.Code == int(gate.Ready) {
				t.Fatalf("a refused message was ready")
			}
			if !strings.Contains(got.Message, tc.want) {
				t.Errorf("Message = %q, want it to name %q", got.Message, tc.want)
			}
		})
	}
}

// A message lane never runs the suite: the proof is the tree's business once the
// commit exists. Asking for both is a mistake that would run twice for nothing.
func TestMessageAndRunDoNotGoTogether(t *testing.T) {
	files := withMessage(closedFiles(), "x\n\nOdachi-Change: "+changeLine+"\n")

	got := run(t, files, messageGit(), check.Options{
		ChangeBase: "origin/main",
		Message:    commitMessage,
		Run:        "go test ./...",
	})

	if got.Code == int(gate.Ready) {
		t.Fatalf("--message and --run together were ready")
	}
	if !strings.Contains(got.Message, "--message") || !strings.Contains(got.Message, "--run") {
		t.Errorf("Message = %q, want both flags named", got.Message)
	}
}

// A trailer whose fork point cannot be found is refused by name: a change that
// cannot say where it started cannot be reviewed against anything.
func TestATrailerWithNoMergeBaseIsRefused(t *testing.T) {
	git := claimedGit()
	git.mergeBaseErr = errors.New("no merge base")

	got := run(t, newFakeFiles(closedFiles()), git, check.Options{ChangeBase: "origin/main"})

	if got.Code == int(gate.Ready) {
		t.Fatalf("a trailer with no merge-base was ready")
	}
	if !strings.Contains(got.Message, "origin/main") {
		t.Fatalf("the refusal does not name the base it could not resolve: %q", got.Message)
	}
}
