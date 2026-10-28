package check_test

import (
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/check"
	"github.com/biggs-100/odachi/internal/domain/gate"
)

// TestRunWritesNothingWhenTheRepositoryIgnoresProof: a project that ran
// `odachi init` carries the block in its own .gitignore, so the lazy per-directory
// file is not written at all — that file is a new path in the worktree, and a new
// path moves the tree the proof is about to describe.
func TestRunWritesNothingWhenTheRepositoryIgnoresProof(t *testing.T) {
	files := newFakeFiles(readyFiles())
	git := &fakeGit{tree: currentTree, ignored: true}

	got := check.Run(check.Options{Dir: dir, Store: store, Run: "echo ok", Base: "HEAD", TimeoutSeconds: 120},
		files, git, &fakeShell{output: "ok\n"})

	if got.Code != int(gate.Ready) {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	if !slices.Equal(git.ignoredAs, []string{"proof"}) {
		t.Errorf("the git port was asked about %v, want the proof directory", git.ignoredAs)
	}
	if slices.Contains(files.writes, filepath.Join(dir, "proof", ".gitignore")) {
		t.Errorf("writes = %v, want no per-directory ignore file", files.writes)
	}
	if !slices.Contains(files.writes, filepath.Join(dir, "proof", "0007-turnos.txt")) {
		t.Errorf("writes = %v, want the proof itself", files.writes)
	}
}

// TestRunFallsBackToItsOwnIgnoreFileWhenTheProbeFails: a probe that fails is not
// permission to shift the tree. Evidence must never change the tree it certifies,
// so the fallback is to write the ignore file exactly as before.
func TestRunFallsBackToItsOwnIgnoreFileWhenTheProbeFails(t *testing.T) {
	files := newFakeFiles(readyFiles())
	git := &fakeGit{tree: currentTree, ignoredErr: errors.New("git is not installed")}

	got := check.Run(check.Options{Dir: dir, Store: store, Run: "echo ok", Base: "HEAD", TimeoutSeconds: 120},
		files, git, &fakeShell{output: "ok\n"})

	if got.Code != int(gate.Ready) {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	ignore := filepath.Join(dir, "proof", ".gitignore")
	if !slices.Contains(files.writes, ignore) {
		t.Errorf("writes = %v, want the fallback ignore file after a failed probe", files.writes)
	}
	if content, ok := files.files[ignore]; !ok || !strings.Contains(string(content), "*") {
		t.Errorf("proof/.gitignore = %q, want the proof files ignored", content)
	}
}
