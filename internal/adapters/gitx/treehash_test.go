package gitx_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/gitx"
)

// TestTreeHashIncludesUntrackedFiles is the property the product rests on: a
// proof is only valid for the exact worktree it was produced from, and a brand
// new file is part of that worktree. `git stash create` ignores untracked
// files, so a hash built that way would accept a stale proof.
func TestTreeHashIncludesUntrackedFiles(t *testing.T) {
	t.Parallel()
	root, _ := newRepo(t)

	before, err := gitx.TreeHash(root)
	if err != nil {
		t.Fatalf("TreeHash before: %v", err)
	}

	if err := os.WriteFile(filepath.Join(root, "nuevo.txt"), []byte("archivo sin trackear\n"), 0o644); err != nil {
		t.Fatalf("write nuevo.txt: %v", err)
	}

	after, err := gitx.TreeHash(root)
	if err != nil {
		t.Fatalf("TreeHash after: %v", err)
	}

	if after == before {
		t.Fatalf("TreeHash ignored an untracked file: %q == %q. The tree must include every file in the worktree.", after, before)
	}

	// Reading the tree must not stage, add, or otherwise mutate the worktree.
	if staged := git(t, root, "diff", "--cached", "--name-only"); staged != "" {
		t.Fatalf("TreeHash left files staged: %q", staged)
	}
	if status := git(t, root, "status", "--porcelain"); !strings.Contains(status, "?? nuevo.txt") {
		t.Fatalf("TreeHash changed the worktree state: %q", status)
	}
}

// TestTreeHashRespectsGitIgnore keeps the fix honest: the hash covers the files
// git would commit, not the bytes on disk. Build output must not invalidate a
// proof.
func TestTreeHashRespectsGitIgnore(t *testing.T) {
	t.Parallel()
	root, _ := newRepo(t)

	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("*.log\n"), 0o644); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}
	git(t, root, "add", "-A")
	git(t, root, "commit", "-m", "ignore logs")

	before, err := gitx.TreeHash(root)
	if err != nil {
		t.Fatalf("TreeHash before: %v", err)
	}

	if err := os.WriteFile(filepath.Join(root, "debug.log"), []byte("ruido de build\n"), 0o644); err != nil {
		t.Fatalf("write debug.log: %v", err)
	}

	after, err := gitx.TreeHash(root)
	if err != nil {
		t.Fatalf("TreeHash after: %v", err)
	}

	if after != before {
		t.Fatalf("TreeHash included a file git ignores: %q != %q", after, before)
	}
}

// TestTreeHashWorksBeforeTheFirstCommit lets the tool work on day one of a
// project, when there is still no HEAD to read.
func TestTreeHashWorksBeforeTheFirstCommit(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	git(t, root, "init")
	git(t, root, "config", "user.name", "odachi test")
	git(t, root, "config", "user.email", "test@odachi.local")

	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# hola\n"), 0o644); err != nil {
		t.Fatalf("write README.md: %v", err)
	}

	tree, err := gitx.TreeHash(root)
	if err != nil {
		t.Fatalf("TreeHash without commits: %v", err)
	}
	if strings.TrimSpace(tree) == "" {
		t.Fatal("TreeHash returned an empty tree in a repository with no commits")
	}
}
