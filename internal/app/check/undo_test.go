package check_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/check"
	"github.com/biggs-100/odachi/internal/domain/gate"
)

// closeWithUndo builds a closed feature whose proof this run writes, and returns
// the result of running it with the two answers the shell gives: the proof and
// the reverted copy.
func closeWithUndo(t *testing.T, git *fakeGit, shell *fakeShell) check.Result {
	t.Helper()

	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, "odd", "tasks", "0007-turnos.md")] = []byte(reviewableDoc("abc1234", true))
	delete(files.files, filepath.Join(dir, "proof", "0007-turnos.txt"))
	delete(files.files, filepath.Join(dir, "proof", "0007-turnos.undone.txt"))

	return check.Run(check.Options{Dir: dir, Store: store, Run: "go test ./...", Base: "HEAD", TimeoutSeconds: 45},
		files, git, shell)
}

// TestACloseWhoseSuiteStaysGreenWithoutTheChangeDoesNotClose is the row this
// feature exists for: the command passed, and it passes again with the change
// reverted, so nothing proves the tests notice the code.
func TestACloseWhoseSuiteStaysGreenWithoutTheChangeDoesNotClose(t *testing.T) {
	git := &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/app/pay.go", "internal/app/pay_test.go"},
		diff:    "diff --git a/internal/app/pay.go b/internal/app/pay.go\n",
	}
	shell := &fakeShell{runs: []fakeRun{
		{output: "ok\n"},
		{output: "ok\n"},
	}}

	got := closeWithUndo(t, git, shell)

	if got.Code != int(gate.NeedsWork) || got.Reason != gate.ReasonUndoMissed {
		t.Fatalf("a suite that passed without the change closed the tree: %d %s %q", got.Code, got.Reason, got.Message)
	}
	if len(git.patched) != 1 || git.patched[0] != "internal/app/pay.go" {
		t.Fatalf("the row reverted %v; it must revert the production files and keep the witness", git.patched)
	}
	if shell.dir != filepath.Join(dir, ".undo-copy") {
		t.Fatalf("the row ran in %q; it must run in the copy, never in the worktree", shell.dir)
	}
	for _, line := range got.Lines {
		if strings.HasPrefix(line, "next: ") && !strings.Contains(line, "go test ./...") {
			t.Fatalf("the block does not name the command that ran: %q", line)
		}
	}
}

// TestACloseWhoseSuiteNoticesTheChangeCloses: the same tree, and the reverted copy
// fails, so the suite is doing its job and the close stands.
func TestACloseWhoseSuiteNoticesTheChangeCloses(t *testing.T) {
	git := &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/app/pay.go"},
		diff:    "diff --git a/internal/app/pay.go b/internal/app/pay.go\n",
	}
	shell := &fakeShell{runs: []fakeRun{
		{output: "ok\n"},
		{output: "--- FAIL: TestPay (0.00s)\n", exitCode: 1},
	}}

	got := closeWithUndo(t, git, shell)

	if got.Code != int(gate.Ready) {
		t.Fatalf("a suite that notices the change blocked the close: %d %s %q", got.Code, got.Reason, got.Message)
	}
}

// TestACloseWaivedByTheHumanSaysSoAndSkipsTheRow: a refactor stays green when
// reverted because the behavior did not change, and only the human can say which
// one it is. The close names the waiver instead of skipping the row in silence.
func TestACloseWaivedByTheHumanSaysSoAndSkipsTheRow(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, "odd", "tasks", "0007-turnos.md")] = []byte(strings.Replace(
		reviewableDoc("abc1234", true),
		"Base: abc1234\n",
		"Base: abc1234\nUndo: waived - solo cambie nombres\n",
		1,
	))
	delete(files.files, filepath.Join(dir, "proof", "0007-turnos.txt"))
	delete(files.files, filepath.Join(dir, "proof", "0007-turnos.undone.txt"))

	git := &fakeGit{tree: currentTree, changed: []string{"internal/app/pay.go"}}
	shell := &fakeShell{runs: []fakeRun{{output: "ok\n"}}}

	got := check.Run(check.Options{Dir: dir, Store: store, Run: "go test ./...", Base: "HEAD", TimeoutSeconds: 45},
		files, git, shell)

	if got.Code != int(gate.Ready) {
		t.Fatalf("a waived row blocked the close: %d %s %q", got.Code, got.Reason, got.Message)
	}
	found := false
	for _, line := range got.Lines {
		if strings.Contains(line, "undo: waived by the human: solo cambie nombres") {
			found = true
		}
	}
	if !found {
		t.Fatalf("the close did not name the waiver: %v", got.Lines)
	}
	if git.archived != "" {
		t.Fatalf("a waived feature copied a tree: %q", git.archived)
	}
}

// TestACloseWithNothingToUndoClosesAndSaysSo: a change of documents or tests only
// has no production file to revert, and a silent row is the silence this gate
// exists to avoid. The row is named, never a green nobody can tell from a pass.
func TestACloseWithNothingToUndoClosesAndSaysSo(t *testing.T) {
	git := &fakeGit{
		tree:    currentTree,
		changed: []string{"odd/tasks/0007-turnos.md", "internal/app/pay_test.go"},
	}
	shell := &fakeShell{runs: []fakeRun{
		{output: "ok\n"},
	}}

	got := closeWithUndo(t, git, shell)

	if got.Code != int(gate.Ready) {
		t.Fatalf("a change with nothing to undo blocked the close: %d %s", got.Code, got.Message)
	}
	found := false
	for _, line := range got.Lines {
		if strings.Contains(line, "undo: not owed") {
			found = true
		}
	}
	if !found {
		t.Fatalf("the close did not say the row was not owed: %v", got.Lines)
	}
	if git.archived != "" {
		t.Fatalf("the row copied a tree it did not need: %q", git.archived)
	}
}

// TestACloseWhoseCopyDoesNotBuildNamesWhatTheChangeAdded: the row reverts the
// production files and keeps the tests, so a change whose tests call what it added
// leaves a copy that cannot compile. When the compiler names that symbol, the revert
// took something away and the tests said so: the evidence is `noticed`, and the file
// carries the name. A failure that names nothing stays unknown (0300).
func TestACloseWhoseCopyDoesNotBuildNamesWhatTheChangeAdded(t *testing.T) {
	git := &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/app/pay.go", "internal/app/pay_test.go"},
		diff: "diff --git a/internal/app/pay.go b/internal/app/pay.go\n" +
			"--- a/internal/app/pay.go\n" +
			"+++ b/internal/app/pay.go\n" +
			"@@ -1 +1,3 @@\n" +
			"+func Pay(amount int) string {\n" +
			"+\treturn \"\"\n" +
			"+}\n",
	}
	shell := &fakeShell{runs: []fakeRun{
		{output: "ok\n"},
		{exitCode: 1, output: "# github.com/x/y [build failed]\n\tinternal/app/pay_test.go:12:19: undefined: Pay\n"},
	}}

	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, "odd", "tasks", "0007-turnos.md")] = []byte(reviewableDoc("abc1234", true))
	delete(files.files, filepath.Join(dir, "proof", "0007-turnos.txt"))
	delete(files.files, filepath.Join(dir, "proof", "0007-turnos.undone.txt"))

	got := check.Run(check.Options{Dir: dir, Store: store, Run: "go test ./...", Base: "HEAD", TimeoutSeconds: 45},
		files, git, shell)

	if got.Code != int(gate.Ready) {
		t.Fatalf("a build failure that names the change blocked the close: %d %s %q", got.Code, got.Reason, got.Message)
	}
	stored := string(files.files[filepath.Join(dir, "proof", "0007-turnos.undone.txt")])
	for _, want := range []string{"# verdict noticed", "Pay"} {
		if !strings.Contains(stored, want) {
			t.Fatalf("the evidence does not carry %q:\n%s", want, stored)
		}
	}
}

// TestACloseWhoseCopyDoesNotBuildForAnotherReasonStaysUnknown: a build failure the
// change did not cause says nothing about the tests, and the row keeps saying so.
func TestACloseWhoseCopyDoesNotBuildForAnotherReasonStaysUnknown(t *testing.T) {
	git := &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/app/pay.go", "internal/app/pay_test.go"},
		diff: "diff --git a/internal/app/pay.go b/internal/app/pay.go\n" +
			"--- a/internal/app/pay.go\n" +
			"+++ b/internal/app/pay.go\n" +
			"+func Pay(amount int) string {\n",
	}
	shell := &fakeShell{runs: []fakeRun{
		{output: "ok\n"},
		{exitCode: 1, output: "# github.com/x/y [build failed]\n\tinternal/app/pay_test.go:12:19: undefined: SomethingElse\n"},
	}}

	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, "odd", "tasks", "0007-turnos.md")] = []byte(reviewableDoc("abc1234", true))
	delete(files.files, filepath.Join(dir, "proof", "0007-turnos.txt"))
	delete(files.files, filepath.Join(dir, "proof", "0007-turnos.undone.txt"))

	got := check.Run(check.Options{Dir: dir, Store: store, Run: "go test ./...", Base: "HEAD", TimeoutSeconds: 45},
		files, git, shell)

	if got.Code != int(gate.Ready) {
		t.Fatalf("an unrelated build failure blocked the close: %d %s %q", got.Code, got.Reason, got.Message)
	}
	stored := string(files.files[filepath.Join(dir, "proof", "0007-turnos.undone.txt")])
	for _, want := range []string{"# verdict unknown", "the copy did not build"} {
		if !strings.Contains(stored, want) {
			t.Fatalf("the evidence does not carry %q:\n%s", want, stored)
		}
	}
}
