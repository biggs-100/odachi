package gitx_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/gitx"
)

// sameDir reports whether two paths name the same existing directory. The tests
// compare directories by identity, not by spelling: macOS hands out /var where
// /private/var was asked, and Windows accepts both separators.
func sameDir(t *testing.T, left, right string) bool {
	t.Helper()

	leftInfo, err := os.Stat(left)
	if err != nil {
		t.Fatalf("stat %s: %v", left, err)
	}
	rightInfo, err := os.Stat(right)
	if err != nil {
		t.Fatalf("stat %s: %v", right, err)
	}
	return os.SameFile(leftInfo, rightInfo)
}

// TestCommonDirIsTheClonesGitDirectory: for an ordinary repository the answer is
// its .git, and that is the directory a clone-scoped store hangs from.
func TestCommonDirIsTheClonesGitDirectory(t *testing.T) {
	root, _ := newRepo(t)

	common, err := gitx.CommonDir(root)
	if err != nil {
		t.Fatalf("CommonDir: %v", err)
	}
	if !sameDir(t, common, filepath.Join(root, ".git")) {
		t.Errorf("CommonDir = %q, want the repository's .git", common)
	}
}

// TestCommonDirSharesTheStoreAcrossWorktrees: a linked worktree has its own
// top-level and its own .git file, and the same common directory. That is what
// makes the store one per clone instead of one per checked-out folder.
func TestCommonDirSharesTheStoreAcrossWorktrees(t *testing.T) {
	root, _ := newRepo(t)

	worktree := filepath.Join(t.TempDir(), "wt")
	git(t, root, "worktree", "add", "-b", "side", worktree)

	fromWorktree, err := gitx.CommonDir(worktree)
	if err != nil {
		t.Fatalf("CommonDir(worktree): %v", err)
	}
	fromMain, err := gitx.CommonDir(root)
	if err != nil {
		t.Fatalf("CommonDir(main): %v", err)
	}
	if !sameDir(t, fromWorktree, fromMain) {
		t.Errorf("CommonDir(worktree) = %q, CommonDir(main) = %q; want one common directory", fromWorktree, fromMain)
	}
}

// TestCommonDirAnswersFromASubdirectory: every verb resolves the store from the
// project root, and it has to be the same answer from anywhere inside it.
func TestCommonDirAnswersFromASubdirectory(t *testing.T) {
	root, _ := newRepo(t)

	sub := filepath.Join(root, "internal", "app")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	fromSub, err := gitx.CommonDir(sub)
	if err != nil {
		t.Fatalf("CommonDir(sub): %v", err)
	}
	fromRoot, err := gitx.CommonDir(root)
	if err != nil {
		t.Fatalf("CommonDir(root): %v", err)
	}
	if !sameDir(t, fromSub, fromRoot) {
		t.Errorf("CommonDir(sub) = %q, CommonDir(root) = %q; want one answer", fromSub, fromRoot)
	}
}

// TestCommonDirRefusesADirectoryThatIsNotARepository: a folder with no git gets
// git's own refusal, which names the fix. Guessing a store for it is the defect
// this adapter exists to prevent.
func TestCommonDirRefusesADirectoryThatIsNotARepository(t *testing.T) {
	plain := t.TempDir()

	if _, err := gitx.CommonDir(plain); err == nil {
		t.Fatal("CommonDir outside a repository = nil error, want git's refusal")
	}
}
