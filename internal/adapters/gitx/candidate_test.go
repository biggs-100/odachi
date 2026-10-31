package gitx_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/gitx"
)

// TestCandidateAnswersTheThreeQuestionsInOneRead: the gates need the tree, the
// changed paths and the diff of the same worktree. One read answers all three,
// and the tree is the same one TreeHash reports. The ignored file proves the
// read respects the repository's own rules.
func TestCandidateAnswersTheThreeQuestionsInOneRead(t *testing.T) {
	t.Parallel()
	root, _ := newRepo(t)

	write(t, filepath.Join(root, "notes.txt"), "primera línea\nsegunda línea\n")
	write(t, filepath.Join(root, "nuevo.go"), "package nuevo\n")
	write(t, filepath.Join(root, ".gitignore"), "ignorado.txt\n")
	write(t, filepath.Join(root, "ignorado.txt"), "no cuenta\n")

	tree, paths, diff, err := gitx.Candidate(root, "HEAD")
	if err != nil {
		t.Fatalf("Candidate: %v", err)
	}

	if want, err := gitx.TreeHash(root); err != nil {
		t.Fatalf("TreeHash: %v", err)
	} else if tree != want {
		t.Errorf("tree = %q, want the worktree tree %q", tree, want)
	}

	wantPaths := []string{".gitignore", "notes.txt", "nuevo.go"}
	if !slices.Equal(paths, wantPaths) {
		t.Errorf("paths = %v, want %v", paths, wantPaths)
	}

	for _, want := range []string{"+segunda línea", "+++ b/nuevo.go"} {
		if !strings.Contains(diff, want) {
			t.Errorf("diff does not contain %q:\n%s", want, diff)
		}
	}
	if strings.Contains(diff, "+++ b/ignorado.txt") {
		t.Errorf("diff includes an ignored file:\n%s", diff)
	}
}

// TestCandidateSeesADeletedFile: a deletion is part of the candidate. A read
// that only staged what exists would report a clean tree for a change that
// removes code.
func TestCandidateSeesADeletedFile(t *testing.T) {
	t.Parallel()
	root, _ := newRepo(t)

	if err := os.Remove(filepath.Join(root, "notes.txt")); err != nil {
		t.Fatalf("remove notes.txt: %v", err)
	}

	_, paths, diff, err := gitx.Candidate(root, "HEAD")
	if err != nil {
		t.Fatalf("Candidate: %v", err)
	}
	if want := []string{"notes.txt"}; !slices.Equal(paths, want) {
		t.Errorf("paths = %v, want %v", paths, want)
	}
	if !strings.Contains(diff, "deleted file mode") {
		t.Errorf("diff does not record the deletion:\n%s", diff)
	}
}

// TestCandidateAgainstNothingSeesEveryFile: a project with no commit compares
// against the empty tree, and every file in the worktree is an addition. It is
// the first day of a project, not an error.
func TestCandidateAgainstNothingSeesEveryFile(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	git(t, root, "init")
	git(t, root, "config", "user.name", "odachi test")
	git(t, root, "config", "user.email", "test@odachi.local")
	write(t, filepath.Join(root, "a.txt"), "a\n")
	write(t, filepath.Join(root, "b.txt"), "b\n")

	tree, paths, diff, err := gitx.Candidate(root, gitx.EmptyTree)
	if err != nil {
		t.Fatalf("Candidate: %v", err)
	}
	if tree == "" {
		t.Error("tree is empty: a worktree with files has a tree")
	}
	if want := []string{"a.txt", "b.txt"}; !slices.Equal(paths, want) {
		t.Errorf("paths = %v, want %v", paths, want)
	}
	for _, want := range []string{"+++ b/a.txt", "+++ b/b.txt"} {
		if !strings.Contains(diff, want) {
			t.Errorf("diff does not contain %q:\n%s", want, diff)
		}
	}
}

// TestCorrectionIsTheDeltaAgainstOneTree: a correction reviews what moved since the
// bytes a verdict judged, and the tree it measures from is a tree, not a branch: no
// merge-base is asked, so a commit that never existed for those bytes cannot hide
// the delta or invent one.
func TestCorrectionIsTheDeltaAgainstOneTree(t *testing.T) {
	t.Parallel()
	root, base := newRepo(t)

	judged, err := gitx.TreeHash(root)
	if err != nil {
		t.Fatalf("TreeHash: %v", err)
	}

	write(t, filepath.Join(root, "notes.txt"), "segunda línea\n")
	write(t, filepath.Join(root, "nuevo.txt"), "algo\n")

	paths, diff, err := gitx.Correction(root, judged)
	if err != nil {
		t.Fatalf("Correction: %v", err)
	}

	want := []string{"notes.txt", "nuevo.txt"}
	if !slices.Equal(paths, want) {
		t.Errorf("paths = %v, want the delta %v", paths, want)
	}
	for _, wantText := range []string{"+segunda línea", "+++ b/nuevo.txt"} {
		if !strings.Contains(diff, wantText) {
			t.Errorf("the delta does not carry %q:\n%s", wantText, diff)
		}
	}

	// The bytes that did not move are not in the delta, and a commit made after the
	// verdict does not change what the correction measures: the tree is the truth.
	git(t, root, "add", "-A")
	git(t, root, "commit", "-m", "after the verdict")
	paths, _, err = gitx.Correction(root, judged)
	if err != nil {
		t.Fatalf("Correction after a commit: %v", err)
	}
	if !slices.Equal(paths, want) {
		t.Errorf("paths after a commit = %v, want the same delta %v", paths, want)
	}
	if base == "" {
		t.Error("newRepo returned no base commit")
	}
}
