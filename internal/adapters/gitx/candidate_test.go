package gitx_test

import (
	"bytes"
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

// TestCandidateNamesTheTreeCheckNamesForAnIgnoredFileItsBaseTracked: `review`
// names the candidate and `check` names the worktree, and a verdict only counts
// for the second id when both name the same bytes. A file the base committed and
// a later commit untracked keeps its index entry when the index is read from that
// base, so `add --all` staged it again and wrote a tree `check` never asks about.
// Git ignores the file now, so nothing about it may move the id.
func TestCandidateNamesTheTreeCheckNamesForAnIgnoredFileItsBaseTracked(t *testing.T) {
	t.Parallel()
	root, _ := newRepo(t)

	if err := os.MkdirAll(filepath.Join(root, ".atl"), 0o755); err != nil {
		t.Fatalf("mkdir .atl: %v", err)
	}
	write(t, filepath.Join(root, ".atl", "skill-registry.md"), "registro\n")
	git(t, root, "add", "-A")
	git(t, root, "commit", "-m", "the registry is tracked")
	// The feature forks here, so the tree it is reviewed against still tracks the
	// file: that is the entry the bug came from.
	base := git(t, root, "rev-parse", "HEAD")

	git(t, root, "rm", "--cached", "-q", ".atl/skill-registry.md")
	write(t, filepath.Join(root, ".gitignore"), ".atl/\n")
	git(t, root, "add", "-A")
	git(t, root, "commit", "-m", "the registry is machine-local")

	tree, _, _, err := gitx.Candidate(root, base)
	if err != nil {
		t.Fatalf("Candidate: %v", err)
	}
	want, err := gitx.TreeHash(root)
	if err != nil {
		t.Fatalf("TreeHash: %v", err)
	}
	if tree != want {
		t.Errorf("Candidate = %q, want the tree check names %q", tree, want)
	}

	// A machine-local edit is the human's business: git ignores the file, so the
	// id both verbs name stays put.
	write(t, filepath.Join(root, ".atl", "skill-registry.md"), "registro\notra línea\n")
	touched, _, _, err := gitx.Candidate(root, base)
	if err != nil {
		t.Fatalf("Candidate after touching the ignored file: %v", err)
	}
	if touched != tree {
		t.Errorf("an ignored file moved the candidate: %q -> %q", tree, touched)
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

// TestStagedCandidateReadsTheCommitInsteadOfTheWorktree: an unstaged deletion
// cannot hide a staged file from the commit-msg gate.
func TestStagedCandidateReadsTheCommitInsteadOfTheWorktree(t *testing.T) {
	t.Parallel()
	root := newReadmeRepo(t)
	base := git(t, root, "rev-parse", "HEAD")

	write(t, filepath.Join(root, "code.go"), "package code\n")
	git(t, root, "add", "--", "code.go")
	if err := os.Remove(filepath.Join(root, "code.go")); err != nil {
		t.Fatalf("remove code.go from the worktree: %v", err)
	}

	_, paths, diff, err := gitx.Candidate(root, base)
	if err != nil {
		t.Fatalf("Candidate: %v", err)
	}
	if slices.Contains(paths, "code.go") || strings.Contains(diff, "+++ b/code.go") {
		t.Errorf("Candidate saw staged code missing from the worktree: paths=%v diff=%s", paths, diff)
	}

	tree, paths, diff, err := gitx.StagedCandidate(root, base)
	if err != nil {
		t.Fatalf("StagedCandidate: %v", err)
	}
	if want := git(t, root, "write-tree"); tree != want {
		t.Errorf("tree = %q, want the real index tree %q", tree, want)
	}
	if want := []string{"code.go"}; !slices.Equal(paths, want) {
		t.Errorf("paths = %v, want %v", paths, want)
	}
	if !strings.Contains(diff, "+++ b/code.go") {
		t.Errorf("diff does not carry the staged code:\n%s", diff)
	}
}

// TestStagedBlobHashReadsCodeMissingFromTheWorktree: coverage receipts describe
// the bytes the commit will carry, not the files an editor left behind.
func TestStagedBlobHashReadsCodeMissingFromTheWorktree(t *testing.T) {
	t.Parallel()
	root := newReadmeRepo(t)

	write(t, filepath.Join(root, "code.go"), "package code\n")
	git(t, root, "add", "--", "code.go")
	if err := os.Remove(filepath.Join(root, "code.go")); err != nil {
		t.Fatalf("remove code.go from the worktree: %v", err)
	}

	got, err := gitx.StagedBlobHash(root, "code.go")
	if err != nil {
		t.Fatalf("StagedBlobHash: %v", err)
	}
	if want := git(t, root, "rev-parse", ":code.go"); got != want {
		t.Errorf("hash = %q, want the staged blob %q", got, want)
	}
}

// TestStagedAddedDepsReadsTheIndex: restoring go.mod in the worktree cannot hide
// a dependency that the commit will carry.
func TestStagedAddedDepsReadsTheIndex(t *testing.T) {
	t.Parallel()
	root := newReadmeRepo(t)
	base := git(t, root, "rev-parse", "HEAD")

	write(t, filepath.Join(root, "go.mod"), "module example.test/project\n\nrequire example.test/new v1.0.0\n")
	git(t, root, "add", "--", "go.mod")
	if err := os.Remove(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("remove go.mod from the worktree: %v", err)
	}

	got, err := gitx.StagedAddedDeps(root, base)
	if err != nil {
		t.Fatalf("StagedAddedDeps: %v", err)
	}
	if want := []string{"example.test/new"}; !slices.Equal(got, want) {
		t.Errorf("dependencies = %v, want %v", got, want)
	}
}

// TestStagedFileReadsTheIndexInsteadOfWorktreeEdits: the commit hook reads the
// bytes Git will commit. An unstaged rewrite or deletion must not change them.
func TestStagedFileReadsTheIndexInsteadOfWorktreeEdits(t *testing.T) {
	t.Parallel()
	const name = "odd/tasks/0001-close.md"
	const staged = "- [x] T1 · close the feature\n"

	for _, tc := range []struct {
		name     string
		remove   bool
		worktree string
	}{
		{name: "different worktree copy", worktree: "- [ ] T1 · still open\n"},
		{name: "absent worktree copy", remove: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			root := newReadmeRepo(t)
			path := filepath.Join(root, filepath.FromSlash(name))
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatalf("mkdir task directory: %v", err)
			}
			write(t, path, staged)
			git(t, root, "add", "--", name)
			if tc.remove {
				if err := os.Remove(path); err != nil {
					t.Fatalf("remove the worktree copy: %v", err)
				}
			} else {
				write(t, path, tc.worktree)
			}

			got, ok, err := gitx.StagedFile(root, name)
			if err != nil {
				t.Fatalf("StagedFile: %v", err)
			}
			if !ok || string(got) != staged {
				t.Errorf("StagedFile = %q, %v; want %q, true", got, ok, staged)
			}
			if got, ok, err := gitx.StagedFile(root, "missing.md"); err != nil || ok || got != nil {
				t.Errorf("StagedFile(missing) = %q, %v, %v; want nil, false, nil", got, ok, err)
			}
			if _, _, err := gitx.StagedFile(root, "odd/tasks"); err == nil {
				t.Error("StagedFile named a directory as a file without an error")
			}
		})
	}
}

// TestStagedListListsTheIndexInsteadOfWorktreeEntries: task discovery must name
// the documents in the commit, not files an editor left unstaged.
func TestStagedListListsTheIndexInsteadOfWorktreeEntries(t *testing.T) {
	t.Parallel()
	root := newReadmeRepo(t)
	const staged = "odd/tasks/0001-close.md"
	stagedPath := filepath.Join(root, filepath.FromSlash(staged))
	if err := os.MkdirAll(filepath.Dir(stagedPath), 0o755); err != nil {
		t.Fatalf("mkdir task directory: %v", err)
	}
	write(t, stagedPath, "- [x] T1 · close the feature\n")
	git(t, root, "add", "--", staged)
	write(t, filepath.Join(root, "odd", "tasks", "0002-unstaged.md"), "- [ ] T1 · open\n")
	if err := os.Remove(stagedPath); err != nil {
		t.Fatalf("remove the staged file from the worktree: %v", err)
	}

	got, err := gitx.StagedList(root, "odd/tasks")
	if err != nil {
		t.Fatalf("StagedList: %v", err)
	}
	if want := []string{"0001-close.md"}; !slices.Equal(got, want) {
		t.Errorf("StagedList = %v, want %v", got, want)
	}

	got, err = gitx.StagedList(root, "odd/missing")
	if err != nil {
		t.Fatalf("StagedList(missing): %v", err)
	}
	if len(got) != 0 {
		t.Errorf("StagedList(missing) = %v, want no entries", got)
	}
	if _, err := gitx.StagedList(root, "README.md"); err == nil {
		t.Error("StagedList named a file as a directory without an error")
	}
}

// TestStagedReadsDoNotChangeTheRealIndex: write-tree may update its cache, so the
// read must use a throwaway index rather than merely avoid git add. A split index
// proves the copy still resolves the repository's shared index without writing it.
func TestStagedReadsDoNotChangeTheRealIndex(t *testing.T) {
	t.Parallel()
	root := newReadmeRepo(t)
	write(t, filepath.Join(root, "code.go"), "package code\n")
	git(t, root, "add", "--", "code.go")
	git(t, root, "update-index", "--split-index")
	indexPath := filepath.Join(root, ".git", "index")
	before, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("read the real index: %v", err)
	}

	if _, _, err := gitx.StagedFile(root, "code.go"); err != nil {
		t.Fatalf("StagedFile: %v", err)
	}
	if _, err := gitx.StagedList(root, "."); err != nil {
		t.Fatalf("StagedList: %v", err)
	}

	after, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("read the real index after the staged reads: %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Error("staged reads rewrote the real index")
	}
}

// newReadmeRepo is a repository with one committed file: the base every staged
// snapshot test compares against.
func newReadmeRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	git(t, root, "init")
	git(t, root, "config", "user.name", "odachi test")
	git(t, root, "config", "user.email", "test@odachi.local")
	write(t, filepath.Join(root, "README.md"), "# project\n")
	git(t, root, "add", "--", "README.md")
	git(t, root, "commit", "-m", "initial")
	return root
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
