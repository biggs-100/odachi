package check_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/check"
	"github.com/biggs-100/odachi/internal/domain/gate"
)

// closeWithMutation builds a closed feature whose proof and mutation row this run
// writes, and returns the result with the answers the shell gives in order: the
// proof, the undo copy, and the mutation copy.
func closeWithMutation(t *testing.T, git *fakeGit, shell *fakeShell, guard string) check.Result {
	t.Helper()

	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, "odd", "tasks", "0007-turnos.md")] = []byte(reviewableDoc("abc1234", true))
	delete(files.files, filepath.Join(dir, "proof", "0007-turnos.txt"))
	delete(files.files, filepath.Join(dir, "proof", "0007-turnos.undone.txt"))
	delete(files.files, filepath.Join(dir, "proof", "0007-turnos.mutated.txt"))
	files.files[filepath.Join(dir, "internal", "app", "pay.go")] = []byte(guard)
	// The copy the row exports carries the same bytes; the row takes the string out
	// of this one and runs there.
	files.files[filepath.Join(dir, ".undo-copy", "internal", "app", "pay.go")] = []byte(guard)

	return check.Run(check.Options{
		Dir: dir, Store: store, Run: "go test ./...", Base: "HEAD", TimeoutSeconds: 45,
		Mutate: "internal/app/pay.go", MutateRemove: "if sinFirma {", MutateExpect: "la firma no se verifica",
	}, files, git, shell)
}

// TestAMutationWhoseSuiteStaysGreenBlocksTheClose is the row this feature exists
// for: the guard was taken out and the suite stayed green, so the test does not
// cover the rule it names. The block names the tree, the file and the question to
// ask again (0297).
func TestAMutationWhoseSuiteStaysGreenBlocksTheClose(t *testing.T) {
	git := &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/app/pay.go"},
		diff:    "diff --git a/internal/app/pay.go b/internal/app/pay.go\n",
	}
	shell := &fakeShell{runs: []fakeRun{
		{output: "ok\n"},
		{output: "--- FAIL: TestPago\nFAIL\n", exitCode: 1},
		{output: "ok\n"},
	}}

	got := closeWithMutation(t, git, shell, "func cobrar() {\n\tif sinFirma {\n\t\treturn err\n\t}\n}\n")

	if got.Code != int(gate.NeedsWork) || got.Reason != gate.ReasonMutationFailed {
		t.Fatalf("a guard that stayed green closed the tree: %d %s %q", got.Code, got.Reason, got.Message)
	}
	if !strings.Contains(got.Message, "stayed green with the guard taken out") {
		t.Errorf("the block does not say what happened: %q", got.Message)
	}
	if !strings.Contains(got.Next, "--mutate internal/app/pay.go") || !strings.Contains(got.Next, "--remove \"if sinFirma {\"") {
		t.Errorf("the continuation does not name the question: %q", got.Next)
	}
}

// TestAMutationThatFiresDoesNotBlock: the guard was taken out, the suite failed and
// its output named what the run expected. That is the answer the row exists for, and
// it leaves the close alone (0297).
func TestAMutationThatFiresDoesNotBlock(t *testing.T) {
	git := &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/app/pay.go"},
		diff:    "diff --git a/internal/app/pay.go b/internal/app/pay.go\n",
	}
	shell := &fakeShell{runs: []fakeRun{
		{output: "ok\n"},
		{output: "--- FAIL: TestPago\nFAIL\n", exitCode: 1},
		{output: "--- FAIL: TestPago\n    la firma no se verifica\nFAIL\n", exitCode: 1},
	}}

	got := closeWithMutation(t, git, shell, "func cobrar() {\n\tif sinFirma {\n\t\treturn err\n\t}\n}\n")

	if got.Reason == gate.ReasonMutationFailed {
		t.Fatalf("a guard that fired blocked the close: %q", got.Message)
	}
}
