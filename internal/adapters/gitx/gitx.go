// Package gitx is the git edge. It answers four questions the gates need:
// which tree is in the worktree right now, which paths changed since the base
// commit, which modules were added to go.mod, and what the candidate looks like
// as a diff.
//
// Every command is local. There is no fetch, no pull, no remote: a gate must
// work offline and must never wait on the network.
package gitx

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

// run executes git in root and returns stdout. The error carries git's own
// stderr, because "exit status 128" alone tells a human nothing.
func run(root string, args ...string) (string, error) {
	return runEnv(root, nil, args...)
}

// runEnv is run with extra environment variables, used to point git at a
// throwaway index file so a tree can be read without touching the real one.
func runEnv(root string, env []string, args ...string) (string, error) {
	cmd := execGit(root, args...)
	cmd.Env = append(cmd.Env, env...)

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

// TreeHash returns the tree of the CURRENT worktree, not just HEAD.
//
// Why it matters: a proof is only valid for the exact bytes it was produced
// from. Hashing HEAD would accept a proof recorded before an uncommitted edit,
// which is precisely the mistake this tool exists to prevent.
//
// The tree is written through a THROWAWAY index (GIT_INDEX_FILE inside a temp
// directory): the real index, the worktree and the human's git configuration
// are never touched. Untracked files are part of the worktree and are included;
// files git ignores stay out, so build output cannot invalidate a proof.
// A repository with no commit yet hashes the same way, from an empty index.
//
// `git stash create` was the first implementation and it was wrong twice: it
// ignores untracked files, so a new file left the hash unchanged and a stale
// proof still looked valid, and it fails before the first commit.
func TreeHash(root string) (string, error) {
	dir, err := os.MkdirTemp("", "odachi-index-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)

	env := []string{"GIT_INDEX_FILE=" + filepath.Join(dir, "index")}

	if _, err := runEnv(root, env, "rev-parse", "--verify", "HEAD"); err != nil {
		if _, err := runEnv(root, env, "read-tree", "--empty"); err != nil {
			return "", err
		}
	} else if _, err := runEnv(root, env, "read-tree", "HEAD"); err != nil {
		return "", err
	}

	if _, err := runEnv(root, env, "add", "--all", "--", "."); err != nil {
		return "", err
	}

	tree, err := runEnv(root, env, "write-tree")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(tree), nil
}

// Between names the paths that differ between two trees. Both are tree ids, as
// git writes them; the empty tree is one of them when the proof predates the
// first commit. A tree git no longer has in its object database is an error, and
// the caller keeps the generic message it already had.
func Between(root, from, to string) ([]string, error) {
	out, err := run(root, "diff", "--name-only", from, to)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	paths := make([]string, 0, 8)
	for _, line := range strings.Split(out, "\n") {
		path := strings.TrimSpace(strings.TrimSuffix(line, "\r"))
		if path == "" || seen[path] {
			continue
		}
		seen[path] = true
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths, nil
}

// Head is the commit the worktree is on: what a feature starts from when its
// task doc is written. A repository with no commit has no HEAD, and that error is
// the answer "nothing to record".
func Head(root string) (string, error) {
	out, err := run(root, "rev-parse", "--verify", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// Resolve names the commit a spelling points at, in full. A base is a commit, not
// a text: `6aeb877` and its forty characters describe the same diff, and a gate
// that compared the strings refused a verdict that judged exactly those bytes.
// The `^{commit}` suffix refuses a tree or a blob: what a base must name is the
// commit itself.
func Resolve(root, ref string) (string, error) {
	trimmed := strings.TrimSpace(ref)
	if trimmed == "" {
		return "", fmt.Errorf("no ref to resolve: a base names a commit")
	}
	out, err := run(root, "rev-parse", "--verify", trimmed+"^{commit}")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// ChangedPaths returns every path that differs between the base's fork point and
// the worktree: tracked files changed or staged since that merge-base, plus
// untracked files that are not ignored. The result is sorted and de-duplicated
// so callers and error messages are deterministic.
func ChangedPaths(root, base string) ([]string, error) {
	target, err := compareBase(root, base)
	if err != nil {
		return nil, err
	}
	if target == "" {
		target = EmptyTree
	}

	diff, err := run(root, "diff", "--name-only", target)
	if err != nil {
		return nil, err
	}
	untracked, err := run(root, "ls-files", "--others", "--exclude-standard")
	if err != nil {
		return nil, err
	}

	paths := map[string]bool{}
	for _, out := range []string{diff, untracked} {
		for _, line := range strings.Split(out, "\n") {
			path := strings.TrimSpace(strings.TrimSuffix(line, "\r"))
			if path == "" {
				continue
			}
			paths[path] = true
		}
	}

	result := make([]string, 0, len(paths))
	for path := range paths {
		result = append(result, path)
	}
	sort.Strings(result)
	return result, nil
}

// BlobHash returns the content hash of one worktree file: the blob id git
// already owns, computed from the bytes with `git hash-object`. It is the
// receipt a finding carries — identity by content, not by environment.
//
// A path that is not an existing worktree file returns "" with no error: the
// candidate deleted it, or the reviewer invented it. That answer is a fact too,
// and the store must be able to tell "no file here" apart from "git broke", so a
// real git failure is an err and blocks the caller instead.
func BlobHash(root, path string) (string, error) {
	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(path)))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("gitx: could not stat %s: %w", path, err)
	}
	if info.IsDir() {
		return "", nil
	}

	out, err := run(root, "hash-object", "--", path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// AddedDeps returns the module paths that go.mod requires now but did not
// require at base. A missing go.mod on either side means no dependencies, not
// an error: the first module in a project is not a failure.
func AddedDeps(root, base string) ([]string, error) {
	before, err := goModAt(root, base)
	if err != nil {
		return nil, err
	}
	after, err := worktreeFile(root, "go.mod")
	if err != nil {
		return nil, err
	}

	existing := map[string]bool{}
	for _, path := range modulePaths(before) {
		existing[path] = true
	}

	added := map[string]bool{}
	for _, path := range modulePaths(after) {
		if !existing[path] {
			added[path] = true
		}
	}

	result := make([]string, 0, len(added))
	for path := range added {
		result = append(result, path)
	}
	slices.Sort(result)
	return result, nil
}

// goModAt returns the go.mod content at rev, or nil when the revision has none.
// `cat-file -e` decides existence with its exit code, so a missing file is not
// confused with a real git failure.
func goModAt(root, rev string) ([]byte, error) {
	if _, err := run(root, "cat-file", "-e", rev+":go.mod"); err != nil {
		return nil, nil
	}
	out, err := run(root, "show", rev+":go.mod")
	if err != nil {
		return nil, err
	}
	return []byte(out), nil
}

// worktreeFile reads a file from the worktree, returning nil when it does not
// exist.
func worktreeFile(root, name string) ([]byte, error) {
	data, err := os.ReadFile(filepath.Join(root, name))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("gitx: could not read %s: %w", name, err)
	}
	return data, nil
}

// modulePaths extracts the module paths of every require directive, single-line
// or block. Versions, comments and indirect markers are irrelevant here: the
// fence cares about which dependency appeared, not which version.
func modulePaths(gomod []byte) []string {
	var paths []string
	inBlock := false

	for _, raw := range strings.Split(string(gomod), "\n") {
		line := strings.TrimSpace(strings.TrimSuffix(raw, "\r"))
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		if inBlock {
			if strings.HasPrefix(line, ")") {
				inBlock = false
				continue
			}
			if path := firstField(line); path != "" {
				paths = append(paths, path)
			}
			continue
		}

		if !strings.HasPrefix(line, "require") {
			continue
		}
		rest := strings.TrimSpace(strings.TrimPrefix(line, "require"))
		if strings.HasPrefix(rest, "(") {
			inBlock = true
			continue
		}
		if path := firstField(rest); path != "" {
			paths = append(paths, path)
		}
	}

	return paths
}

// firstField returns the first whitespace-separated token of a go.mod directive,
// without quotes or a trailing comment.
func firstField(line string) string {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return ""
	}
	path := fields[0]
	if path == ")" || strings.HasPrefix(path, "//") {
		return ""
	}
	return strings.Trim(path, `"`)
}
