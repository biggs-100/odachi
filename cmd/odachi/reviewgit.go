package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/biggs-100/odachi/internal/adapters/gitx"
	reviewapp "github.com/biggs-100/odachi/internal/app/review"
)

// reviewGit is the Git adapter for `odachi review`. Two facts come from gitx, so
// the review sees exactly the tree and the changed paths `check` sees; the diff
// needs a third command, which lives here because it is the only place that owns
// the throwaway index this adapter builds.
type reviewGit struct{}

func (reviewGit) TreeHash(root string) (string, error) { return gitx.TreeHash(root) }

func (reviewGit) ChangedPaths(root, base string) ([]string, error) {
	return gitx.ChangedPaths(root, base)
}

func (reviewGit) Diff(root, base string) (string, error) { return worktreeDiff(root, base) }

// worktreeDiff renders the candidate as one unified diff: everything that differs
// between base and the CURRENT worktree, including files that are not tracked
// yet, and excluding files git ignores.
//
// It reuses the technique `TreeHash` already uses and for the same reason: a
// throwaway index in a temporary directory. `git diff base` alone would miss every
// new file — a reviewer that cannot see the new file cannot review the change that
// matters most — and staging into the real index would touch the human's git.
func worktreeDiff(root, base string) (string, error) {
	dir, err := os.MkdirTemp("", "odachi-diff-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)

	env := []string{"GIT_INDEX_FILE=" + filepath.Join(dir, "index")}

	// The empty tree is "compare against nothing": there is no base commit, so
	// the index starts empty and every file in the worktree is an addition.
	target := strings.TrimSpace(base)
	readTree := []string{"read-tree", target}
	if target == "" || target == reviewapp.EmptyTree {
		target = reviewapp.EmptyTree
		readTree = []string{"read-tree", "--empty"}
	}

	if _, err := gitOutput(root, env, readTree...); err != nil {
		return "", err
	}
	if _, err := gitOutput(root, env, "add", "--all", "--", "."); err != nil {
		return "", err
	}

	return gitOutput(root, env, "diff", "--cached", "--no-color", "--no-ext-diff", "--no-textconv", target)
}

// gitOutput runs one git command in root with extra environment and returns its
// stdout. The error carries git's own stderr, because "exit status 128" alone
// tells a human nothing. The environment is pinned the same way gitx pins it: a
// gate never prompts for credentials and never inherits a localized message.
func gitOutput(root string, env []string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	cmd.Env = append(os.Environ(), append([]string{"LC_ALL=C", "GIT_TERMINAL_PROMPT=0", "GIT_ASKPASS=echo"}, env...)...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, msg)
	}
	return stdout.String(), nil
}
