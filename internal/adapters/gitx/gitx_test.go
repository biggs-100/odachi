package gitx_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/gitx"
)

// git runs one git command in root and returns its trimmed stdout. Tests use
// it only to build and inspect the temporary repository.
func git(t *testing.T, root string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	cmd.Env = append(os.Environ(), "LC_ALL=C", "GIT_TERMINAL_PROMPT=0", "GIT_ASKPASS=echo")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out))
}

// newRepo builds a real temporary repository with one commit and returns its
// root and the base commit hash.
func newRepo(t *testing.T) (root, base string) {
	t.Helper()
	root = t.TempDir()

	git(t, root, "init")
	git(t, root, "config", "user.name", "odachi test")
	git(t, root, "config", "user.email", "test@odachi.local")

	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/demo\n\ngo 1.25\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "notes.txt"), []byte("primera línea\n"), 0o644); err != nil {
		t.Fatalf("write notes.txt: %v", err)
	}
	git(t, root, "add", "-A")
	git(t, root, "commit", "-m", "base")

	return root, git(t, root, "rev-parse", "HEAD")
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// TestBetweenNamesWhatMovedBetweenTwoTrees: the proof gate knows two hashes
// differ; this is the answer that tells a human which file moved.
func TestBetweenNamesWhatMovedBetweenTwoTrees(t *testing.T) {
	root, base := newRepo(t)
	write(t, filepath.Join(root, "notes.txt"), "segunda línea\n")
	write(t, filepath.Join(root, "nuevo.txt"), "algo\n")
	git(t, root, "add", "-A")
	git(t, root, "commit", "-m", "second")

	from := git(t, root, "rev-parse", base+"^{tree}")
	to := git(t, root, "rev-parse", "HEAD^{tree}")

	paths, err := gitx.Between(root, from, to)
	if err != nil {
		t.Fatalf("Between: %v", err)
	}
	if !slices.Equal(paths, []string{"notes.txt", "nuevo.txt"}) {
		t.Errorf("Between = %v, want the two moved paths, sorted", paths)
	}
}

// TestHeadIsTheCommitAndAnEmptyRepositoryHasNone: the base a feature records is
// the commit that exists when it starts; a repository with no commit has nothing
// to record, and the error is the answer, not a crash.
func TestHeadIsTheCommitAndAnEmptyRepositoryHasNone(t *testing.T) {
	root, base := newRepo(t)

	got, err := gitx.Head(root)
	if err != nil {
		t.Fatalf("Head: %v", err)
	}
	if got != base {
		t.Errorf("Head = %q, want %q", got, base)
	}

	empty := t.TempDir()
	git(t, empty, "init")
	if _, err := gitx.Head(empty); err == nil {
		t.Error("Head on a repository with no commit = nil error, want the failure to record nothing")
	}
}

// TestHeadTreeIsTheBytesTheCommitCarries: a rebase rewrites the commit and never
// the tree, so the tree is what a step records. A repository with no commit has
// no tree to remember.
func TestHeadTreeIsTheBytesTheCommitCarries(t *testing.T) {
	root, _ := newRepo(t)

	got, err := gitx.HeadTree(root)
	if err != nil {
		t.Fatalf("HeadTree: %v", err)
	}
	if want := git(t, root, "rev-parse", "HEAD^{tree}"); got != want {
		t.Errorf("HeadTree = %q, want %q", got, want)
	}

	empty := t.TempDir()
	git(t, empty, "init")
	if _, err := gitx.HeadTree(empty); err == nil {
		t.Error("HeadTree on a repository with no commit = nil error, want none to record")
	}
}

// TestResolveNamesTheCommitASpellingPointsAt: a base is a commit, and the gate
// asks which one a spelling names. A prefix answers the full hash, so a receipt
// written short still describes the diff it judged; a ref that names nothing is
// an error, not an empty answer.
func TestResolveNamesTheCommitASpellingPointsAt(t *testing.T) {
	t.Parallel()
	root, base := newRepo(t)

	got, err := gitx.Resolve(root, base[:7])
	if err != nil {
		t.Fatalf("Resolve(%q): %v", base[:7], err)
	}
	if got != base {
		t.Errorf("Resolve(%q) = %q, want %q", base[:7], got, base)
	}

	if _, err := gitx.Resolve(root, "no-such-ref"); err == nil {
		t.Error("Resolve of a ref that does not exist = nil error, want the failure")
	}
	if _, err := gitx.Resolve(root, "  "); err == nil {
		t.Error("Resolve of an empty ref = nil error, want the failure")
	}
}

// TestTreeHashTracksTheWorktree is the property the whole product rests on: an
// uncommitted edit must change the tree hash, otherwise a proof could be
// accepted for a tree that was never tested.
func TestTreeHashTracksTheWorktree(t *testing.T) {
	t.Parallel()
	root, _ := newRepo(t)

	clean, err := gitx.TreeHash(root)
	if err != nil {
		t.Fatalf("TreeHash(clean): %v", err)
	}
	if want := git(t, root, "rev-parse", "HEAD^{tree}"); clean != want {
		t.Errorf("TreeHash(clean) = %q, want HEAD tree %q", clean, want)
	}

	write(t, filepath.Join(root, "notes.txt"), "primera línea\nsegunda línea sin commit\n")

	dirty, err := gitx.TreeHash(root)
	if err != nil {
		t.Fatalf("TreeHash(dirty): %v", err)
	}
	if dirty == clean {
		t.Fatalf("TreeHash did not change after an uncommitted edit (both %q)", clean)
	}

	// The probe must not touch the worktree: the edit is still there.
	data, err := os.ReadFile(filepath.Join(root, "notes.txt"))
	if err != nil {
		t.Fatalf("read notes.txt: %v", err)
	}
	if !strings.Contains(string(data), "segunda línea sin commit") {
		t.Errorf("TreeHash modified the worktree: %q", data)
	}

	// Committing the edit moves HEAD to that tree.
	git(t, root, "add", "-A")
	git(t, root, "commit", "-m", "edit")
	committed, err := gitx.TreeHash(root)
	if err != nil {
		t.Fatalf("TreeHash(committed): %v", err)
	}
	if committed != dirty {
		t.Errorf("TreeHash(committed) = %q, want the previously dirty tree %q", committed, dirty)
	}
}

func TestChangedPathsReportsEditsAndUntrackedFiles(t *testing.T) {
	t.Parallel()
	root, base := newRepo(t)

	write(t, filepath.Join(root, "notes.txt"), "primera línea\neditada\n")
	write(t, filepath.Join(root, "nuevo.txt"), "sin trackear\n")

	got, err := gitx.ChangedPaths(root, base)
	if err != nil {
		t.Fatalf("ChangedPaths: %v", err)
	}
	want := []string{"notes.txt", "nuevo.txt"}
	if !slices.Equal(got, want) {
		t.Errorf("ChangedPaths = %q, want %q (sorted, de-duplicated)", got, want)
	}
}

func TestChangedPathsIsEmptyForACleanWorktree(t *testing.T) {
	t.Parallel()
	root, base := newRepo(t)

	got, err := gitx.ChangedPaths(root, base)
	if err != nil {
		t.Fatalf("ChangedPaths: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("ChangedPaths(clean) = %q, want none", got)
	}
}

func TestChangedPathsReportsAFileChangedOnce(t *testing.T) {
	t.Parallel()
	root, base := newRepo(t)

	// Untracked and then tracked and then edited: one path, not three.
	write(t, filepath.Join(root, "nuevo.txt"), "uno\n")
	git(t, root, "add", "nuevo.txt")
	write(t, filepath.Join(root, "nuevo.txt"), "dos\n")

	got, err := gitx.ChangedPaths(root, base)
	if err != nil {
		t.Fatalf("ChangedPaths: %v", err)
	}
	if want := []string{"nuevo.txt"}; !slices.Equal(got, want) {
		t.Errorf("ChangedPaths = %q, want %q", got, want)
	}
}

func TestAddedDepsReportsNewRequireLines(t *testing.T) {
	t.Parallel()
	root, base := newRepo(t)

	// Unchanged go.mod: no deps added.
	got, err := gitx.AddedDeps(root, base)
	if err != nil {
		t.Fatalf("AddedDeps(unchanged): %v", err)
	}
	if len(got) != 0 {
		t.Errorf("AddedDeps(unchanged) = %q, want none", got)
	}

	write(t, filepath.Join(root, "go.mod"),
		"module example.com/demo\n\ngo 1.25\n\nrequire github.com/example/thing v1.2.3\n")

	got, err = gitx.AddedDeps(root, base)
	if err != nil {
		t.Fatalf("AddedDeps(added): %v", err)
	}
	if want := []string{"github.com/example/thing"}; !slices.Equal(got, want) {
		t.Errorf("AddedDeps = %q, want %q", got, want)
	}
}

func TestAddedDepsReadsRequireBlocks(t *testing.T) {
	t.Parallel()
	root, base := newRepo(t)

	write(t, filepath.Join(root, "go.mod"),
		"module example.com/demo\n\ngo 1.25\n\nrequire (\n\tgithub.com/zeta/zeta-go v0.1.0\n\tgithub.com/alpha/alpha-go v0.2.0 // indirect\n)\n")

	got, err := gitx.AddedDeps(root, base)
	if err != nil {
		t.Fatalf("AddedDeps: %v", err)
	}
	want := []string{"github.com/alpha/alpha-go", "github.com/zeta/zeta-go"}
	if !slices.Equal(got, want) {
		t.Errorf("AddedDeps = %q, want %q (sorted)", got, want)
	}
}

func TestAddedDepsWithoutGoModIsNotAnError(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	git(t, root, "init")
	git(t, root, "config", "user.name", "odachi test")
	git(t, root, "config", "user.email", "test@odachi.local")
	write(t, filepath.Join(root, "README.md"), "sin go.mod\n")
	git(t, root, "add", "-A")
	git(t, root, "commit", "-m", "base")
	base := git(t, root, "rev-parse", "HEAD")

	got, err := gitx.AddedDeps(root, base)
	if err != nil {
		t.Fatalf("AddedDeps(no go.mod): %v", err)
	}
	if len(got) != 0 {
		t.Errorf("AddedDeps(no go.mod) = %q, want none", got)
	}
}

func TestErrorsIncludeGitStderr(t *testing.T) {
	t.Parallel()
	// A plain directory is not a repository: git must fail, and its stderr
	// must reach the caller instead of a bare "exit status 128".
	dir := t.TempDir()

	if _, err := gitx.TreeHash(dir); err == nil {
		t.Fatalf("TreeHash(non-repo) succeeded, want an error")
	} else if !strings.Contains(err.Error(), "not a git repository") {
		t.Errorf("error %q does not include git stderr", err)
	}

	if _, err := gitx.ChangedPaths(dir, "HEAD"); err == nil {
		t.Fatalf("ChangedPaths(non-repo) succeeded, want an error")
	}
}
