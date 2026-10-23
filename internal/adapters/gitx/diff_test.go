package gitx_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/gitx"
)

// TestDiffShowsChangedTrackedAndUntrackedFiles is the reason this adapter exists
// at all: the reviewer has to see the candidate, and the candidate includes the
// file nobody has added to git yet. `git diff base` alone would hide it, so a
// reviewer could approve a change it never read.
func TestDiffShowsChangedTrackedAndUntrackedFiles(t *testing.T) {
	root, base := newRepo(t)

	write(t, filepath.Join(root, "notes.txt"), "primera línea\nsegunda línea sin commit\n")
	write(t, filepath.Join(root, "nuevo.txt"), "archivo sin trackear\n")

	diff, err := gitx.Diff(root, base)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}

	for _, want := range []string{"notes.txt", "nuevo.txt", "segunda línea sin commit", "archivo sin trackear"} {
		if !strings.Contains(diff, want) {
			t.Errorf("the diff does not mention %q. Diff:\n%s", want, diff)
		}
	}
}

// TestDiffExcludesIgnoredFiles keeps the diff equal to what git would commit:
// build output must never become review noise, and an ignored file is not part
// of the candidate.
func TestDiffExcludesIgnoredFiles(t *testing.T) {
	root, base := newRepo(t)

	write(t, filepath.Join(root, ".gitignore"), "*.log\n")
	git(t, root, "add", "-A")
	git(t, root, "commit", "-m", "ignore logs")

	write(t, filepath.Join(root, "debug.log"), "ruido de build\n")

	diff, err := gitx.Diff(root, base)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	if strings.Contains(diff, "debug.log") {
		t.Errorf("the diff includes a file git ignores. Diff:\n%s", diff)
	}
}

// TestDiffAgainstNothingShowsEveryFileAsAnAddition is the first day of a
// project: there is no commit to compare against, so every file in the worktree
// is new and every file has to be reviewable.
func TestDiffAgainstNothingShowsEveryFileAsAnAddition(t *testing.T) {
	root := t.TempDir()
	git(t, root, "init")
	git(t, root, "config", "user.name", "odachi test")
	git(t, root, "config", "user.email", "test@odachi.local")

	write(t, filepath.Join(root, "README.md"), "# hola\n")
	write(t, filepath.Join(root, "main.go"), "package main\n")

	diff, err := gitx.Diff(root, "")
	if err != nil {
		t.Fatalf("Diff without commits: %v", err)
	}

	for _, want := range []string{"README.md", "main.go", "# hola", "package main", "new file mode"} {
		if !strings.Contains(diff, want) {
			t.Errorf("the diff against nothing does not mention %q. Diff:\n%s", want, diff)
		}
	}
	if strings.Contains(diff, "deleted file mode") {
		t.Errorf("the diff against nothing deletes files that exist. Diff:\n%s", diff)
	}
}

// TestDiffEmptyTreeBehavesLikeNoBase protects the caller: the review app passes
// the empty-tree hash when the project has no commit, and gitx decides that both
// spellings mean the same thing.
func TestDiffEmptyTreeBehavesLikeNoBase(t *testing.T) {
	if want := "4b825dc642cb6eb9a060e54bf8d69288fbee4904"; gitx.EmptyTree != want {
		t.Fatalf("gitx.EmptyTree = %q, want %q", gitx.EmptyTree, want)
	}

	root := t.TempDir()
	git(t, root, "init")
	git(t, root, "config", "user.name", "odachi test")
	git(t, root, "config", "user.email", "test@odachi.local")

	write(t, filepath.Join(root, "README.md"), "# hola\n")

	none, err := gitx.Diff(root, "")
	if err != nil {
		t.Fatalf("Diff(\"\"): %v", err)
	}
	empty, err := gitx.Diff(root, gitx.EmptyTree)
	if err != nil {
		t.Fatalf("Diff(EmptyTree): %v", err)
	}
	if empty != none {
		t.Errorf("Diff(EmptyTree) != Diff(\"\"):\n--- \"\" ---\n%s\n--- EmptyTree ---\n%s", none, empty)
	}
}

// TestDiffLeavesTheRepositoryUntouched is the safety property: the technique
// builds a throwaway index, so the human's real index, their staged work and
// their untracked files are exactly as they were after a review.
func TestDiffLeavesTheRepositoryUntouched(t *testing.T) {
	root, base := newRepo(t)

	write(t, filepath.Join(root, "notes.txt"), "primera línea\nsegunda línea sin commit\n")
	write(t, filepath.Join(root, "nuevo.txt"), "archivo sin trackear\n")

	before := git(t, root, "status", "--porcelain")

	if _, err := gitx.Diff(root, base); err != nil {
		t.Fatalf("Diff: %v", err)
	}

	after := git(t, root, "status", "--porcelain")
	if after != before {
		t.Errorf("Diff changed the repository state:\nbefore:\n%s\nafter:\n%s", before, after)
	}
	if staged := git(t, root, "diff", "--cached", "--name-only"); staged != "" {
		t.Errorf("Diff left files staged: %q", staged)
	}
}

// TestDiffErrorCarriesGitsStderr: "exit status 128" tells a human nothing. A
// base that does not resolve has to say so in git's own words.
func TestDiffErrorCarriesGitsStderr(t *testing.T) {
	root, _ := newRepo(t)

	_, err := gitx.Diff(root, "no-such-revision")
	if err == nil {
		t.Fatal("Diff accepted a base that does not exist")
	}
	msg := err.Error()
	if !strings.Contains(msg, "fatal:") {
		t.Errorf("the error does not carry git's stderr: %q", msg)
	}
	if !strings.Contains(msg, "no-such-revision") {
		t.Errorf("the error does not name the base that failed: %q", msg)
	}
	if strings.Contains(msg, os.ErrNotExist.Error()) && !strings.Contains(msg, "git") {
		t.Errorf("the error lost git's context: %q", msg)
	}
}
