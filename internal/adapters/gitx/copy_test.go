package gitx_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/gitx"
)

// TestTheCopyRefusesWhatItCannotJudge: the empty tree, the empty patch and the
// empty path list are asked once and answered without a git call, and a discard
// of nothing is not an error.
func TestTheCopyRefusesWhatItCannotJudge(t *testing.T) {
	t.Parallel()

	if _, err := gitx.Archive(t.TempDir(), "  "); err == nil {
		t.Fatal("an empty tree was archived")
	}
	if err := gitx.PatchReverse(t.TempDir(), "", []string{"pay.go"}); err != nil {
		t.Fatalf("an empty patch is not an error: %v", err)
	}
	if err := gitx.PatchReverse(t.TempDir(), "diff --git a/pay.go b/pay.go\n", nil); err != nil {
		t.Fatalf("no paths means nothing to revert: %v", err)
	}
	if err := gitx.Discard(""); err != nil {
		t.Fatalf("discarding nothing is not an error: %v", err)
	}
}

// TestArchiveAndReverseKeepTheTestsAndLoseTheChange is the copy the undo row
// judges: the production file goes back to what the base had, the test that
// arrived with the change stays, and the worktree is never touched.
func TestArchiveAndReverseKeepTheTestsAndLoseTheChange(t *testing.T) {
	t.Parallel()
	root, base := newRepo(t)

	write(t, filepath.Join(root, "pay.go"), "package demo\n\nfunc Charge() int { return 2 }\n")
	write(t, filepath.Join(root, "pay_test.go"), "package demo\n\nimport \"testing\"\n\nfunc TestCharge(t *testing.T) {\n\tif Charge() != 2 {\n\t\tt.Fatal(\"no\")\n\t}\n}\n")
	git(t, root, "add", "-A")
	git(t, root, "commit", "-m", "charge")

	tree, _, diff, err := gitx.Candidate(root, base)
	if err != nil {
		t.Fatalf("Candidate: %v", err)
	}
	if tree == "" || diff == "" {
		t.Fatalf("the candidate is empty: tree %q, diff %q", tree, diff)
	}

	dir, err := gitx.Archive(root, tree)
	if err != nil {
		t.Fatalf("Archive: %v", err)
	}
	defer func() { _ = gitx.Discard(dir) }()

	copied, err := os.ReadFile(filepath.Join(dir, "pay.go"))
	if err != nil {
		t.Fatalf("read the copy: %v", err)
	}
	if !strings.Contains(string(copied), "return 2") {
		t.Fatalf("the copy is not the tree: %q", copied)
	}

	if err := gitx.PatchReverse(dir, diff, []string{"pay.go"}); err != nil {
		t.Fatalf("PatchReverse: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "pay.go")); err == nil {
		t.Fatal("pay.go survived the reverse: the change was not reverted")
	}
	if _, err := os.Stat(filepath.Join(dir, "pay_test.go")); err != nil {
		t.Fatalf("the test did not survive the reverse: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "pay.go")); err != nil {
		t.Fatalf("the worktree moved: %v", err)
	}
}

// TestDiscardRemovesTheCopy: the row leaves nothing behind, and a directory that
// is already gone is not an error.
func TestDiscardRemovesTheCopy(t *testing.T) {
	t.Parallel()
	root, _ := newRepo(t)

	tree, err := gitx.TreeHash(root)
	if err != nil {
		t.Fatalf("TreeHash: %v", err)
	}
	dir, err := gitx.Archive(root, tree)
	if err != nil {
		t.Fatalf("Archive: %v", err)
	}
	if err := gitx.Discard(dir); err != nil {
		t.Fatalf("Discard: %v", err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("the copy survived Discard: %v", err)
	}
	if err := gitx.Discard(dir); err != nil {
		t.Fatalf("a second Discard is not an error: %v", err)
	}
}
