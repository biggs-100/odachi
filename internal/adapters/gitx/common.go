package gitx

import (
	"errors"
	"path/filepath"
	"strings"
)

// ErrNoRepository is the answer git gives when a folder is not inside a
// repository. It is a state, not a failure: a project that never ran git init
// has no store, and saying so is not an error to report as a broken install.
var ErrNoRepository = errors.New("not a git repository")

// CommonDir returns the repository's git common directory: the one every
// worktree of a clone shares. For an ordinary repository that is its `.git`; for
// a linked worktree it is the main clone's `.git`, reached through the
// `commondir` file git writes for it.
//
// This is the directory a store that "belongs to the clone" has to hang from.
// `--show-toplevel` answers a different question — which worktree this is — and a
// store derived from it would be duplicated per worktree, which is the opposite
// of what a shared clone wants.
//
// The answer is the operating system's, not the string's: git resolves the
// `gitdir:` file, the `commondir` file and every symlinked ancestor. A relative
// answer is resolved against root, because that is the directory git ran in.
//
// A folder that is not inside a repository gets ErrNoRepository, not a bare
// error: "there is nothing here" and "git failed" are different answers, and a
// caller that cannot tell them apart ends up guessing. execGit pins LC_ALL=C, so
// git's message is the English one this classification reads.
func CommonDir(root string) (string, error) {
	out, err := run(root, "rev-parse", "--git-common-dir")
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not a git repository") {
			return "", ErrNoRepository
		}
		return "", err
	}

	dir := strings.TrimSpace(out)
	if dir == "" {
		return "", errors.New("git rev-parse --git-common-dir printed nothing: this git cannot name the common directory")
	}
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(root, dir)
	}
	return filepath.Clean(dir), nil
}
