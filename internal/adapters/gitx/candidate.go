package gitx

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Candidate reads the worktree once and answers the three questions every gate
// asks about it: the content tree, the paths that differ from base, and the diff.
//
// The verbs used to answer them with three reads, and each one built its own
// throwaway index and asked git for the merge-base again. On Windows a git
// process costs about 50ms, so a check spent most of its second measuring the
// same bytes. The three answers were never independent: they describe one
// worktree.
//
// One index, in a temp directory that is removed before returning, means the
// real index, the worktree and the human's git are never touched. Ignored files
// stay out; untracked files and deletions are part of the candidate.
func Candidate(root, base string) (tree string, paths []string, diff string, err error) {
	dir, err := os.MkdirTemp("", "odachi-candidate-")
	if err != nil {
		return "", nil, "", err
	}
	defer os.RemoveAll(dir)

	env := []string{"GIT_INDEX_FILE=" + filepath.Join(dir, "index")}

	// The empty tree is "compare against nothing": there is no base commit, so
	// the index starts empty and every file in the worktree is an addition.
	target, err := compareBase(root, base)
	if err != nil {
		return "", nil, "", err
	}
	readTree := []string{"read-tree", target}
	if target == "" {
		target = EmptyTree
		readTree = []string{"read-tree", "--empty"}
	}

	if _, err := runEnv(root, env, readTree...); err != nil {
		return "", nil, "", err
	}
	if _, err := runEnv(root, env, "add", "--all", "--", "."); err != nil {
		return "", nil, "", err
	}

	written, err := runEnv(root, env, "write-tree")
	if err != nil {
		return "", nil, "", err
	}

	// The index is the worktree snapshot now, so --cached answers what the old
	// worktree diff plus ls-files answered, in one process.
	names, err := runEnv(root, env, "diff", "--cached", "--name-only", target)
	if err != nil {
		return "", nil, "", err
	}

	text, err := runEnv(root, env, "diff", "--cached", "--no-color", "--no-ext-diff", "--no-textconv", target)
	if err != nil {
		return "", nil, "", err
	}

	return strings.TrimSpace(written), splitPaths(names), text, nil
}

// splitPaths reads one path per line, sorted and de-duplicated like the old
// ChangedPaths, so callers and error messages stay deterministic.
func splitPaths(out string) []string {
	seen := map[string]bool{}
	paths := make([]string, 0, strings.Count(out, "\n")+1)
	for _, line := range strings.Split(out, "\n") {
		path := strings.TrimSpace(strings.TrimSuffix(line, "\r"))
		if path == "" || seen[path] {
			continue
		}
		seen[path] = true
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}
