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
func closeWithMutation(t *testing.T, git *fakeGit, shell *fakeShell, guard, remove, expect string) check.Result {
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
		Mutate: "internal/app/pay.go", MutateRemove: remove, MutateExpect: expect,
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

	got := closeWithMutation(t, git, shell, "func cobrar() {\n\tif sinFirma {\n\t\treturn err\n\t}\n}\n", "if sinFirma {", "la firma no se verifica")

	if got.Code != int(gate.NeedsWork) || got.Reason != gate.ReasonMutationFailed {
		t.Fatalf("a guard that stayed green closed the tree: %d %s %q", got.Code, got.Reason, got.Message)
	}
	if !strings.Contains(got.Message, "stayed green with the guard taken out") {
		t.Errorf("the block does not say what happened: %q", got.Message)
	}
	if !strings.Contains(got.Next, `--mutate "internal/app/pay.go"`) || !strings.Contains(got.Next, "--remove \"if sinFirma {\"") {
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

	got := closeWithMutation(t, git, shell, "func cobrar() {\n\tif sinFirma {\n\t\treturn err\n\t}\n}\n", "if sinFirma {", "la firma no se verifica")

	if got.Reason == gate.ReasonMutationFailed {
		t.Fatalf("a guard that fired blocked the close: %q", got.Message)
	}
}

// TestTheMutationRowBlocksWhenTheFailureNamesSomethingElse: a red for another cause
// is a red, not evidence about this guard. The reviewer named this case as untested,
// and a regression that dropped it from the blocking list would leave the suite green
// (0297).
func TestTheMutationRowBlocksWhenTheFailureNamesSomethingElse(t *testing.T) {
	git := &fakeGit{tree: currentTree, changed: []string{"internal/app/pay.go"}, diff: "diff --git a/internal/app/pay.go b/internal/app/pay.go\n"}
	shell := &fakeShell{runs: []fakeRun{
		{output: "ok\n"},
		{output: "--- FAIL: TestPago\nFAIL\n", exitCode: 1},
		{output: "--- FAIL: TestOtra\n    want 2, got 3\nFAIL\n", exitCode: 1},
	}}

	got := closeWithMutation(t, git, shell, "func cobrar() {\n\tif sinFirma {\n\t\treturn err\n\t}\n}\n", "if sinFirma {", "la firma no se verifica")

	if got.Code != int(gate.NeedsWork) || got.Reason != gate.ReasonMutationFailed {
		t.Fatalf("a red that named something else passed: %d %s %q", got.Code, got.Reason, got.Message)
	}
	if !strings.Contains(got.Message, "without naming") {
		t.Errorf("the block does not say why the red did not count: %q", got.Message)
	}
}

// TestTheMutationRowRefusesAStringThatIsNotInTheFile: a mutation that does not mutate
// proves nothing, and running it anyway would find the suite green and call a real
// guard missing. The row blocks and says the question did not apply (0297).
func TestTheMutationRowRefusesAStringThatIsNotInTheFile(t *testing.T) {
	git := &fakeGit{tree: currentTree, changed: []string{"internal/app/pay.go"}, diff: "diff --git a/internal/app/pay.go b/internal/app/pay.go\n"}
	shell := &fakeShell{runs: []fakeRun{
		{output: "ok\n"},
		{output: "--- FAIL: TestPago\nFAIL\n", exitCode: 1},
		{output: "ok\n"},
	}}

	got := closeWithMutation(t, git, shell, "func cobrar() {\n\tif sinFirma {\n\t\treturn err\n\t}\n}\n", "no existe esta cadena", "la firma no se verifica")

	if got.Code != int(gate.NeedsWork) || got.Reason != gate.ReasonMutationFailed {
		t.Fatalf("a string that is not in the file passed: %d %s %q", got.Code, got.Reason, got.Message)
	}
	if !strings.Contains(got.Message, "the string is not in the file") {
		t.Errorf("the block does not say the question did not apply: %q", got.Message)
	}
}

// TestAStoredMutationAnswersOnlyItsOwnQuestion: the stored row is reused only when it
// answered the same file, string, expected text and command. Otherwise a new question
// would pass on old evidence -a false green- or repeat a mistyped flag until the tree
// changes (0297).
func TestAStoredMutationAnswersOnlyItsOwnQuestion(t *testing.T) {
	git := &fakeGit{tree: currentTree, changed: []string{"internal/app/pay.go"}, diff: "diff --git a/internal/app/pay.go b/internal/app/pay.go\n"}
	shell := &fakeShell{runs: []fakeRun{
		{output: "ok\n"},
		{output: "--- FAIL: TestPago\nFAIL\n", exitCode: 1},
		{output: "--- FAIL: TestPago\n    la firma no se verifica\nFAIL\n", exitCode: 1},
	}}

	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, "odd", "tasks", "0007-turnos.md")] = []byte(reviewableDoc("abc1234", true))
	delete(files.files, filepath.Join(dir, "proof", "0007-turnos.txt"))
	delete(files.files, filepath.Join(dir, "proof", "0007-turnos.undone.txt"))
	files.files[filepath.Join(dir, "internal", "app", "pay.go")] = []byte("func cobrar() {\n\tif sinFirma {\n\t\treturn err\n\t}\n}\n")
	files.files[filepath.Join(dir, ".undo-copy", "internal", "app", "pay.go")] = []byte("func cobrar() {\n\tif sinFirma {\n\t\treturn err\n\t}\n}\n")
	// A row for this tree that answered a different question: it must not be reused.
	files.files[filepath.Join(dir, "proof", "0007-turnos.mutated.txt")] = []byte(
		"# mutated " + currentTree + "\n# command go test ./...\n# file internal/app/pay.go\n" +
			"# removed \"otra cadena\"\n# expected \"otro texto\"\n# verdict noticed\n# output\nok\n")

	got := check.Run(check.Options{
		Dir: dir, Store: store, Run: "go test ./...", Base: "HEAD", TimeoutSeconds: 45,
		Mutate: "internal/app/pay.go", MutateRemove: "if sinFirma {", MutateExpect: "la firma no se verifica",
	}, files, git, shell)

	if got.Reason == gate.ReasonMutationFailed {
		t.Fatalf("the run asked the new question and still blocked: %q", got.Message)
	}
}
