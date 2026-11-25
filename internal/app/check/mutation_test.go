package check_test

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/adapters/gitx"
	"github.com/biggs-100/odachi/internal/adapters/shellcmd"
	"github.com/biggs-100/odachi/internal/app/check"
	"github.com/biggs-100/odachi/internal/domain/brief"
	"github.com/biggs-100/odachi/internal/domain/gate"
	"github.com/biggs-100/odachi/internal/domain/proof"
	"github.com/biggs-100/odachi/internal/domain/taskdoc"
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

// TestAProofCommandThatChangesTheTreeWritesNoGreenProof runs a real command that
// edits a tracked candidate file. The command's success is not evidence about the
// bytes it left behind, so check must refuse instead of stamping the old tree.
func TestAProofCommandThatChangesTheTreeWritesNoGreenProof(t *testing.T) {
	root := newProofRepo(t)
	before, _, _, err := gitx.Candidate(root, "HEAD")
	if err != nil {
		t.Fatalf("measure the candidate before the command: %v", err)
	}

	const command = "echo changed> candidate.txt"
	git := &proofGit{fakeGit: &fakeGit{}}
	got := check.Run(check.Options{
		Dir: root, Run: command, Base: "HEAD", TimeoutSeconds: 30,
	}, proofDiskFiles{}, git, proofShell{})

	after, _, _, err := gitx.Candidate(root, "HEAD")
	if err != nil {
		t.Fatalf("measure the candidate after the command: %v", err)
	}
	if after == before {
		t.Fatalf("the proof command did not change the tracked candidate: %s", before)
	}

	proofPath := filepath.Join(root, "proof", "0001-proof-identity.txt")
	if data, err := os.ReadFile(proofPath); err == nil {
		recorded, parseErr := proof.Parse(string(data))
		if parseErr != nil {
			t.Fatalf("the command wrote an unreadable green proof: %v\n%s", parseErr, data)
		}
		t.Errorf("a mutating command wrote a green proof for tree %s (before %s, after %s)", recorded.Tree, before, after)
	} else if !os.IsNotExist(err) {
		t.Fatalf("read the proof path: %v", err)
	}
	if got.Code != int(gate.NeedsWork) || got.Reason != gate.ReasonProofFailed {
		t.Errorf("Code/Reason = %d/%q, want %d/%q: a changed candidate has no proof", got.Code, got.Reason, gate.NeedsWork, gate.ReasonProofFailed)
	}
	if !strings.Contains(got.Message, "changed the candidate tree") {
		t.Errorf("Message = %q, want the candidate mutation named", got.Message)
	}
	if !strings.Contains(got.Message, "candidate.txt") {
		t.Errorf("Message = %q, want the changed path named", got.Message)
	}
	if want := `odachi check --run "echo changed> candidate.txt"`; got.Next != want {
		t.Errorf("Next = %q, want %q", got.Next, want)
	}
	if git.candidateCalls != 2 {
		t.Errorf("candidate reads = %d, want the pre-command and post-command snapshots", git.candidateCalls)
	}
}

// TestAProofCommandThatLeavesTheCandidateUnchangedWritesGreenProof covers both
// ordinary no-op commands and commands whose only writes are ignored evidence.
func TestAProofCommandThatLeavesTheCandidateUnchangedWritesGreenProof(t *testing.T) {
	for _, tc := range []struct {
		name        string
		command     string
		hasEvidence bool
	}{
		{name: "unchanged tracked bytes", command: "echo ok"},
		{name: "ignored evidence only", command: "echo ignored> proof/evidence.txt", hasEvidence: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := newProofRepo(t)
			before, _, _, err := gitx.Candidate(root, "HEAD")
			if err != nil {
				t.Fatalf("measure the candidate before the command: %v", err)
			}

			git := &proofGit{fakeGit: &fakeGit{}}
			got := check.Run(check.Options{
				Dir: root, Run: tc.command, Base: "HEAD", TimeoutSeconds: 30,
			}, proofDiskFiles{}, git, proofShell{})

			after, _, _, err := gitx.Candidate(root, "HEAD")
			if err != nil {
				t.Fatalf("measure the candidate after the command: %v", err)
			}
			if after != before {
				t.Fatalf("candidate changed: %s -> %s", before, after)
			}
			if got.Code != int(gate.Ready) {
				t.Fatalf("Code = %d, want %d: %s", got.Code, gate.Ready, got.Message)
			}
			if git.candidateCalls != 2 {
				t.Errorf("candidate reads = %d, want the pre-command and post-command snapshots", git.candidateCalls)
			}

			proofPath := filepath.Join(root, "proof", "0001-proof-identity.txt")
			data, err := os.ReadFile(proofPath)
			if err != nil {
				t.Fatalf("read the green proof: %v", err)
			}
			recorded, err := proof.Parse(string(data))
			if err != nil {
				t.Fatalf("parse the green proof: %v", err)
			}
			if recorded.Tree != before || recorded.Command != tc.command {
				t.Errorf("proof tree/command = %q/%q, want %q/%q", recorded.Tree, recorded.Command, before, tc.command)
			}
			if tc.hasEvidence {
				evidence, err := os.ReadFile(filepath.Join(root, "proof", "evidence.txt"))
				if err != nil || strings.TrimSpace(string(evidence)) != "ignored" {
					t.Errorf("ignored evidence = %q, %v; want the command's write", evidence, err)
				}
			}
		})
	}
}

// TestAProofCommandWhoseTreeCannotBeRemeasuredWritesNoGreenProof fails closed on
// an unreadable post-command snapshot instead of publishing the pre-command hash.
func TestAProofCommandWhoseTreeCannotBeRemeasuredWritesNoGreenProof(t *testing.T) {
	root := newProofRepo(t)
	git := &proofGit{fakeGit: &fakeGit{}, afterErr: os.ErrPermission}
	got := check.Run(check.Options{
		Dir: root, Run: "echo ok", Base: "HEAD", TimeoutSeconds: 30,
	}, proofDiskFiles{}, git, proofShell{})

	if got.Code != int(gate.NeedsWork) || got.Reason != gate.ReasonProofFailed {
		t.Fatalf("Code/Reason = %d/%q, want %d/%q", got.Code, got.Reason, gate.NeedsWork, gate.ReasonProofFailed)
	}
	if !strings.Contains(got.Message, "could not measure the candidate tree after the proof command") {
		t.Errorf("Message = %q, want the unreadable post-command tree named", got.Message)
	}
	if want := `odachi check --run "echo ok"`; got.Next != want {
		t.Errorf("Next = %q, want %q", got.Next, want)
	}
	if _, err := os.Stat(filepath.Join(root, "proof", "0001-proof-identity.txt")); !os.IsNotExist(err) {
		t.Errorf("an unreadable post-command tree left a green proof: %v", err)
	}
}

// proofGit keeps the fake port's unrelated answers but uses real Git for the
// candidate, ignore rules and tree delta this regression depends on.
type proofGit struct {
	*fakeGit
	afterErr error
}

func (g *proofGit) Candidate(root, base string) (string, []string, string, error) {
	g.candidateCalls++
	g.bases = append(g.bases, base)
	if g.candidateCalls == 2 && g.afterErr != nil {
		return "", nil, "", g.afterErr
	}
	return gitx.Candidate(root, base)
}

func (g *proofGit) Between(root, from, to string) ([]string, error) {
	g.betweenCalls++
	return gitx.Between(root, from, to)
}

func (g *proofGit) Ignored(root, path string) (bool, error) {
	g.ignoredAs = append(g.ignoredAs, path)
	return gitx.Ignored(root, path)
}

// proofShell executes the command through the same platform adapter as the CLI.
type proofShell struct{}

func (proofShell) Run(dir, command string, timeoutSeconds int) (string, string, int, bool, error) {
	return shellcmd.Run(dir, command, timeoutSeconds)
}

func (proofShell) RunStream(dir, command string, timeoutSeconds int, w io.Writer) (string, string, int, bool, error) {
	return shellcmd.RunStream(dir, command, timeoutSeconds, w)
}

// proofDiskFiles is the disk port for the real-repository regressions.
type proofDiskFiles struct{}

func (proofDiskFiles) Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (proofDiskFiles) Read(path string) ([]byte, error) { return os.ReadFile(path) }

func (proofDiskFiles) List(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names, nil
}

func (proofDiskFiles) Write(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (proofDiskFiles) ModTime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

func (proofDiskFiles) Remove(path string) error { return os.Remove(path) }

// newProofRepo creates one committed candidate with an open feature and ignored
// evidence directory, so a real check reaches the proof command without owing a
// review or a close-only row.
func newProofRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	proofGitCommand(t, root, "init")
	proofGitCommand(t, root, "config", "user.name", "odachi test")
	proofGitCommand(t, root, "config", "user.email", "test@odachi.local")

	writeProofFile(t, filepath.Join(root, ".gitignore"), "proof/\n")
	writeProofFile(t, filepath.Join(root, "candidate.txt"), "before\n")
	writeProofFile(t, filepath.Join(root, "brief.md"), brief.Render(brief.Brief{
		Name:     "proof identity",
		Outcome:  []string{"proofs name the bytes the command left behind"},
		DoneWhen: []string{"a mutating proof command is refused"},
		NonGoals: []string{"new proof formats"},
	}))
	writeProofFile(t, filepath.Join(root, "odd", "tasks", "0001-proof-identity.md"), taskdoc.TaskDoc{
		Title:        "proof identity",
		Requirements: []taskdoc.Requirement{{ID: "R1", Text: "proofs name the post-command candidate"}},
		DoneWhen:     []string{"candidate mutation refuses the proof"},
		Steps:        []taskdoc.Step{{ID: "T1", Text: "bind proof to the post-command tree", Requirement: "R1"}},
	}.Render())
	proofGitCommand(t, root, "add", "-A")
	proofGitCommand(t, root, "commit", "-m", "initial candidate")
	if err := os.MkdirAll(filepath.Join(root, "proof"), 0o755); err != nil {
		t.Fatalf("create the ignored evidence directory: %v", err)
	}
	return root
}

func writeProofFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func proofGitCommand(t *testing.T, root string, args ...string) string {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), "git", args...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}
