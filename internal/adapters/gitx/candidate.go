package gitx

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
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
	// The empty tree is "compare against nothing": there is no base commit, so
	// the index starts empty and every file in the worktree is an addition.
	target, err := compareBase(root, base)
	if err != nil {
		return "", nil, "", err
	}
	return snapshot(root, target)
}

// StagedCandidate reads the real index without modifying it and answers the same
// three questions about the tree a commit will carry. The index is first written
// as a tree, then read into a throwaway index for the cached diff: unstaged
// worktree changes can neither hide nor invent commit bytes.
func StagedCandidate(root, base string) (tree string, paths []string, diff string, err error) {
	target, err := compareBase(root, base)
	if err != nil {
		return "", nil, "", err
	}

	dir, err := os.MkdirTemp("", "odachi-staged-")
	if err != nil {
		return "", nil, "", err
	}
	defer func() { _ = os.RemoveAll(dir) }()
	env := []string{"GIT_INDEX_FILE=" + filepath.Join(dir, "index")}

	written, err := stagedTree(root)
	if err != nil {
		return "", nil, "", err
	}
	if _, err := runEnv(root, env, "read-tree", strings.TrimSpace(written)); err != nil {
		return "", nil, "", err
	}
	return indexSnapshot(root, env, target, written)
}

// StagedFile reads one file from the tree the real index will commit. It writes
// that index as a tree without taking an index lock, then reads the blob by object
// id: unstaged edits, deletions and pathspec metacharacters cannot change the
// answer. A missing path is absent, not an error; a directory is not a file.
func StagedFile(root, path string) ([]byte, bool, error) {
	name, err := stagedTreePath(path)
	if err != nil {
		return nil, false, err
	}
	if name == "." {
		return nil, false, fmt.Errorf("gitx: the staged root is not a file")
	}

	tree, err := stagedTree(root)
	if err != nil {
		return nil, false, err
	}
	entries, err := stagedTreeEntries(root, tree, name)
	if err != nil {
		return nil, false, err
	}
	if len(entries) == 0 {
		return nil, false, nil
	}
	entry := entries[0]
	if entry.path != name || entry.kind != "blob" {
		return nil, false, fmt.Errorf("gitx: staged path %s is not a file", path)
	}

	out, err := run(root, "cat-file", "blob", entry.object)
	if err != nil {
		return nil, false, err
	}
	return []byte(out), true, nil
}

// StagedList returns the immediate entry names of one directory in the tree the
// real index will commit, like os.ReadDir. A missing directory has no entries; a
// path that names a file is an error rather than an empty catalog.
func StagedList(root, dir string) ([]string, error) {
	name, err := stagedTreePath(dir)
	if err != nil {
		return nil, err
	}
	tree, err := stagedTree(root)
	if err != nil {
		return nil, err
	}
	entries, err := stagedTreeEntries(root, tree, name)
	if err != nil {
		return nil, err
	}
	if name != "." && len(entries) == 1 && entries[0].path == name {
		if entries[0].kind != "tree" {
			return nil, fmt.Errorf("%s is not a directory", dir)
		}
		entries, err = stagedTreeEntries(root, tree, name+"/")
		if err != nil {
			return nil, err
		}
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, filepath.Base(filepath.FromSlash(entry.path)))
	}
	slices.Sort(names)
	return names, nil
}

// stagedTree writes a throwaway copy of the real index as a tree. The copy keeps
// split-index links readable from the repository while every write-tree cache
// update lands outside the human's index. The object database gains the tree, but
// the staged bytes and the index file itself stay untouched.
func stagedTree(root string) (string, error) {
	indexPath, err := run(root, "rev-parse", "--git-path", "index")
	if err != nil {
		return "", err
	}
	indexPath = strings.TrimSpace(indexPath)
	if !filepath.IsAbs(indexPath) {
		indexPath = filepath.Join(root, indexPath)
	}

	dir, err := os.MkdirTemp("", "odachi-staged-tree-")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.RemoveAll(dir) }()
	copiedIndex := filepath.Join(dir, "index")
	data, err := os.ReadFile(indexPath)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("gitx: could not copy the staged index: %w", err)
	}
	if err == nil {
		if err := os.WriteFile(copiedIndex, data, 0o600); err != nil {
			return "", fmt.Errorf("gitx: could not copy the staged index: %w", err)
		}
	}

	out, err := runEnv(root, []string{
		"GIT_INDEX_FILE=" + copiedIndex,
		"GIT_OPTIONAL_LOCKS=0",
	}, "write-tree")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// stagedTreePath normalizes the repository-relative spelling Git receives and
// refuses a path that could escape the tree.
func stagedTreePath(name string) (string, error) {
	if name == "" {
		name = "."
	}
	if !filepath.IsLocal(name) {
		return "", fmt.Errorf("gitx: staged path %s is outside the repository", name)
	}
	clean := filepath.Clean(name)
	if clean == "." {
		return ".", nil
	}
	return filepath.ToSlash(clean), nil
}

// stagedTreeEntry is one NUL-delimited ls-tree record.
type stagedTreeEntry struct {
	kind   string
	object string
	path   string
}

// stagedTreeEntries lists one path without recursion. Literal pathspecs keep a
// tracked name such as :(glob)* from becoming a Git query instead of a filename.
func stagedTreeEntries(root, tree, name string) ([]stagedTreeEntry, error) {
	args := []string{"--literal-pathspecs", "ls-tree", "-z", tree}
	if name != "." {
		args = append(args, "--", name)
	}
	out, err := run(root, args...)
	if err != nil {
		return nil, err
	}

	var entries []stagedTreeEntry
	for raw := range strings.SplitSeq(out, "\x00") {
		if raw == "" {
			continue
		}
		metadata, path, ok := strings.Cut(raw, "\t")
		if !ok {
			return nil, fmt.Errorf("gitx: could not read a staged tree entry")
		}
		fields := strings.Fields(metadata)
		if len(fields) != 3 {
			return nil, fmt.Errorf("gitx: could not read a staged tree entry for %s", path)
		}
		entries = append(entries, stagedTreeEntry{
			kind: fields[1], object: fields[2], path: path,
		})
	}
	return entries, nil
}

// Correction is the worktree compared against one exact tree, without asking git
// for a fork point: the tree is the bytes a verdict judged, and a correction
// reviews what moved since those bytes. It returns the candidate tree, the paths
// that differ and their diff from the same throwaway index, so a caller never has
// to re-read a worktree that may have moved.
//
// It reuses the one throwaway index Candidate builds, so the real index, the
// worktree and the human's git are never touched.
func Correction(root, tree string) (candidateTree string, paths []string, diff string, err error) {
	target := strings.TrimSpace(tree)
	if target == "" {
		return "", nil, "", fmt.Errorf("no tree to correct against")
	}
	return snapshot(root, target)
}

// snapshot writes the worktree into a throwaway index and answers the content
// tree the index holds, the paths that differ from target, and their diff.
//
// The index is built from HEAD, not from target: see stageWorktree. The
// comparison still runs against target, which is what the candidate adds.
func snapshot(root, target string) (tree string, paths []string, diff string, err error) {
	dir, err := os.MkdirTemp("", "odachi-candidate-")
	if err != nil {
		return "", nil, "", err
	}
	defer func() { _ = os.RemoveAll(dir) }()
	env := []string{"GIT_INDEX_FILE=" + filepath.Join(dir, "index")}

	if target == "" {
		target = EmptyTree
	}

	if err := stageWorktree(root, env); err != nil {
		return "", nil, "", err
	}

	written, err := runEnv(root, env, "write-tree")
	if err != nil {
		return "", nil, "", err
	}
	return indexSnapshot(root, env, target, written)
}

// indexSnapshot reads paths and diff from an index that already holds the
// candidate tree. The worktree and staged snapshots differ only in how they fill
// that index; both answers describe the same written tree.
func indexSnapshot(root string, env []string, target, written string) (tree string, paths []string, diff string, err error) {
	if target == "" {
		target = EmptyTree
	}

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
	for line := range strings.SplitSeq(out, "\n") {
		path := strings.TrimSpace(strings.TrimSuffix(line, "\r"))
		if path == "" || seen[path] {
			continue
		}
		seen[path] = true
		paths = append(paths, path)
	}
	slices.Sort(paths)
	return paths
}
