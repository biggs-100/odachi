package gitx

import (
	"os"
	"path/filepath"
	"strings"
)

// EmptyTree is git's hash of the empty tree. It lives here because gitx is what
// decides that "no base" means "everything is new": a caller passes "" or this
// hash and both render the same candidate.
const EmptyTree = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

// compareBase is what "against this base" means: the merge-base of HEAD and
// base, not base itself. Two dots (`git diff base`) report every commit the
// base gained after the fork as a deletion, so a candidate on a branch looks
// like it deleted code it never touched — the promptfoo bug. The fork point is
// the candidate's parent, and everything since is the candidate.
//
// "" and the empty tree are the project's first day: there is no commit to fork
// from, and the caller wants "everything is new". The answer "" means that.
func compareBase(root, base string) (string, error) {
	target := strings.TrimSpace(base)
	if target == "" || target == EmptyTree {
		return "", nil
	}

	// HEAD's fork point with itself is HEAD, and HEAD is the default base: one
	// git process per verb that nobody needs to pay. A named base still asks
	// merge-base, because another branch's commits must not read as deletions.
	if target == "HEAD" {
		return target, nil
	}

	out, err := run(root, "merge-base", "HEAD", target)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// Diff renders the candidate as one unified diff: everything that differs
// between base and the CURRENT worktree, including files that are not tracked
// yet, and excluding files git ignores.
//
// It uses the same technique as TreeHash and for the same reason: a throwaway
// index (GIT_INDEX_FILE inside a temp directory). `git diff base` alone would
// miss every new file — a reviewer that cannot see the new file cannot review
// the change that matters most — and staging into the real index would touch the
// human's git.
func Diff(root, base string) (string, error) {
	dir, err := os.MkdirTemp("", "odachi-diff-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)

	env := []string{"GIT_INDEX_FILE=" + filepath.Join(dir, "index")}

	// The empty tree is "compare against nothing": there is no base commit, so
	// the index starts empty and every file in the worktree is an addition.
	target, err := compareBase(root, base)
	if err != nil {
		return "", err
	}
	readTree := []string{"read-tree", target}
	if target == "" {
		target = EmptyTree
		readTree = []string{"read-tree", "--empty"}
	}

	if _, err := runEnv(root, env, readTree...); err != nil {
		return "", err
	}
	if _, err := runEnv(root, env, "add", "--all", "--", "."); err != nil {
		return "", err
	}

	return runEnv(root, env, "diff", "--cached", "--no-color", "--no-ext-diff", "--no-textconv", target)
}
