package gitx_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/gitx"
)

// The ledger asks gitx three questions the fakes in the app tests cannot answer:
// which tree a commit carries, whether a commit is on this branch, and which
// commit the bytes landed in. The functions are small and their edge cases are
// git's (exit 1 is an answer, not a failure; the oldest match wins), so they are
// pinned here against a real repository.

// TestTreeOfReadsTheCommitsTree: the tree of a commit, full sha, and the same
// answer for a short spelling.
func TestTreeOfReadsTheCommitsTree(t *testing.T) {
	root, base := newRepo(t)
	want := git(t, root, "rev-parse", base+"^{tree}")

	got, err := gitx.TreeOf(root, base)
	if err != nil {
		t.Fatalf("TreeOf: %v", err)
	}
	if got != want {
		t.Errorf("TreeOf(%s) = %q, want %q", base[:7], got, want)
	}

	short, err := gitx.TreeOf(root, base[:7])
	if err != nil {
		t.Fatalf("TreeOf(short): %v", err)
	}
	if short != want {
		t.Errorf("TreeOf(%s) = %q, want the same tree %q", base[:7], short, want)
	}
}

// TestTreeOfRefusesACommitItCannotRead: a sha this clone does not have is an
// error, not an empty answer, because the caller decides what a missing commit
// means.
func TestTreeOfRefusesACommitItCannotRead(t *testing.T) {
	root, _ := newRepo(t)

	if _, err := gitx.TreeOf(root, "0123456789abcdef0123456789abcdef01234567"); err == nil {
		t.Fatal("TreeOf(unknown) = nil error, want git's refusal")
	}
	if _, err := gitx.TreeOf(root, "  "); err == nil {
		t.Fatal("TreeOf(empty) = nil error, want a refusal")
	}
}

// TestHasAnswersNoWithoutFailing: a commit that is not an ancestor of HEAD is
// the answer "no" - exit 1 - and a caller that treats it as an error would make
// every old step look broken.
func TestHasAnswersNoWithoutFailing(t *testing.T) {
	root, base := newRepo(t)

	onBranch, err := gitx.Has(root, base)
	if err != nil {
		t.Fatalf("Has(HEAD): %v", err)
	}
	if !onBranch {
		t.Errorf("Has(%s) = false, want true: the base is an ancestor of HEAD", base[:7])
	}

	// A commit on another branch that HEAD does not contain.
	side := filepath.Join(t.TempDir(), "side")
	git(t, root, "worktree", "add", "-b", "side", side)
	write(t, filepath.Join(side, "side.txt"), "otra rama\n")
	git(t, side, "add", "-A")
	git(t, side, "commit", "-m", "side")
	offBranch := git(t, side, "rev-parse", "HEAD")

	reachable, err := gitx.Has(root, offBranch)
	if err != nil {
		t.Fatalf("Has(side commit) = error %v, want the answer no", err)
	}
	if reachable {
		t.Errorf("Has(side commit) = true, want false: it is not on this branch")
	}
}

// TestCommitWithTreeNamesTheOldestMatch: two commits can carry the same tree (an
// empty commit, a return to earlier bytes). The bytes landed when they first
// appeared, so the oldest match is the answer, not the newest.
func TestCommitWithTreeNamesTheOldestMatch(t *testing.T) {
	root, base := newRepo(t)
	tree := git(t, root, "rev-parse", base+"^{tree}")

	// The same tree again, in a later commit: nothing changed, so git records
	// the empty tree delta.
	git(t, root, "commit", "--allow-empty", "-m", "the same bytes again")
	again := git(t, root, "rev-parse", "HEAD")

	found, err := gitx.CommitWithTree(root, tree)
	if err != nil {
		t.Fatalf("CommitWithTree: %v", err)
	}
	if found != base {
		t.Errorf("CommitWithTree = %q, want the oldest match %q (the later commit is %s)", found, base, again[:7])
	}

	if _, err := os.Stat(filepath.Join(root, "notes.txt")); err != nil {
		t.Fatalf("the repository was left unusable: %v", err)
	}
}

// TestCommitWithTreeAnswersNothingWithoutHistory: a repository with no commit
// has no history to search, and the empty answer is not an error.
func TestCommitWithTreeAnswersNothingWithoutHistory(t *testing.T) {
	root := t.TempDir()
	git(t, root, "init")

	found, err := gitx.CommitWithTree(root, strings.Repeat("a", 40))
	if err != nil {
		t.Fatalf("CommitWithTree on an empty repository: %v", err)
	}
	if found != "" {
		t.Errorf("CommitWithTree = %q, want the empty answer", found)
	}
}
