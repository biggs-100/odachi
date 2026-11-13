package check_test

import (
	"path/filepath"
	"testing"

	"github.com/biggs-100/odachi/internal/app/check"
	"github.com/biggs-100/odachi/internal/domain/gate"
	"github.com/biggs-100/odachi/internal/domain/proof"
	"github.com/biggs-100/odachi/internal/domain/undone"
)

// storedClose is a closed feature whose proof for this tree is already written:
// the state a plain `odachi check` sees, without `--run`.
func storedClose(t *testing.T) *fakeFiles {
	t.Helper()
	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, "odd", "tasks", "0007-turnos.md")] = []byte(reviewableDoc("abc1234", true))
	files.files[filepath.Join(dir, "proof", "0007-turnos.txt")] =
		[]byte(proof.Render(proof.Proof{Tree: currentTree, Command: "go test ./...", Output: "ok\n"}))
	delete(files.files, filepath.Join(dir, "proof", "0007-turnos.undone.txt"))
	return files
}

// TestACloseReadsTheStoredUndoWithoutRunningAgain: the row is evidence, not an
// action. A check that does not run the command still blocks the tree the stored
// run judged, or the block would last only as long as one invocation.
func TestACloseReadsTheStoredUndoWithoutRunningAgain(t *testing.T) {
	files := storedClose(t)
	files.files[filepath.Join(dir, "proof", "0007-turnos.undone.txt")] =
		[]byte(undone.Render(undone.Evidence{
			Tree:    currentTree,
			Command: "go test ./...",
			Verdict: undone.Missed,
			Output:  "ok\n",
		}))

	got := check.Run(check.Options{Dir: dir, Store: store, Base: "HEAD"},
		files, &fakeGit{tree: currentTree}, &fakeShell{})

	if got.Code != int(gate.NeedsWork) || got.Reason != gate.ReasonUndoMissed {
		t.Fatalf("a stored miss did not close the tree: %d %s %q", got.Code, got.Reason, got.Message)
	}
}

// TestACloseWithoutTheRowAndWithoutARunOwesIt: a stored proof does not stand in
// for the row. The close names the row it did not take, the way it names a proof
// it does not have.
func TestACloseWithoutTheRowAndWithoutARunOwesIt(t *testing.T) {
	files := storedClose(t)

	got := check.Run(check.Options{Dir: dir, Store: store, Base: "HEAD"},
		files, &fakeGit{tree: currentTree}, &fakeShell{})

	if got.Code != int(gate.NeedsWork) || got.Reason != gate.ReasonUndoMissing {
		t.Fatalf("a close with no row was ready: %d %s %q", got.Code, got.Reason, got.Message)
	}
}

// TestACloseWithNoProductionBehaviorStoresTheNotOwedRow: "there is nothing to
// undo" is a verdict, not a silence, and the next check reads it instead of
// asking again.
func TestACloseWithNoProductionBehaviorStoresTheNotOwedRow(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, "odd", "tasks", "0007-turnos.md")] = []byte(reviewableDoc("abc1234", true))
	delete(files.files, filepath.Join(dir, "proof", "0007-turnos.txt"))
	delete(files.files, filepath.Join(dir, "proof", "0007-turnos.undone.txt"))

	git := &fakeGit{tree: currentTree, changed: []string{"odd/tasks/0007-turnos.md"}}
	got := check.Run(check.Options{Dir: dir, Store: store, Run: "go test ./...", Base: "HEAD", TimeoutSeconds: 45},
		files, git, &fakeShell{runs: []fakeRun{{output: "ok\n"}}})
	if got.Code != int(gate.Ready) {
		t.Fatalf("a change with nothing to undo blocked the close: %d %s", got.Code, got.Message)
	}

	stored := string(files.files[filepath.Join(dir, "proof", "0007-turnos.undone.txt")])
	ev, err := undone.Parse(stored)
	if err != nil {
		t.Fatalf("the not-owed row was not stored: %v", err)
	}
	if ev.Verdict != undone.NotOwed {
		t.Fatalf("stored verdict = %q, want not_owed", ev.Verdict)
	}

	// The next check reads it back and still closes, naming the row.
	got = check.Run(check.Options{Dir: dir, Store: store, Base: "HEAD"},
		files, &fakeGit{tree: currentTree}, &fakeShell{})
	if got.Code != int(gate.Ready) || got.Reason != gate.ReasonReady {
		t.Fatalf("the stored not-owed row blocked the next check: %d %s", got.Code, got.Message)
	}
}
