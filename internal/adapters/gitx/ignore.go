package gitx

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Ignored reports whether git ignores path under root, using every mechanism the
// repository has: the .gitignore files, .git/info/exclude and the human's global
// excludes.
//
// It answers for a path that does not exist yet, which is the whole point: the
// evidence this tool writes (proof/) has to be ignored BEFORE it is written, or
// writing it moves the very tree it describes. The verdicts are not asked about
// any more: they live in the clone's store, outside the worktree.
//
// A directory that is not on disk is asked with and without its trailing slash
// in ONE process. Git only matches a directory rule (`/proof/`) when it can tell
// the path is a directory, and it cannot tell that from a name nothing created
// yet: measured, `git check-ignore -q proof` exits 1 while `proof/` exits 0 under
// the same rule. Asking both spellings together is what makes the answer the
// caller expects without paying two git processes for it.
//
// Exit 0 means ignored and exit 1 means not ignored; both are answers. Anything
// else is an error carrying git's own stderr, because "exit status 128" alone
// tells nobody that the directory is not a repository.
func Ignored(root, path string) (bool, error) {
	if path == "" || strings.HasSuffix(path, "/") {
		return checkIgnore(root, path)
	}
	return checkIgnore(root, path, path+"/")
}

// checkIgnore is the probe itself: one `git check-ignore` for every spelling.
// Without -q it accepts more than one path and still exits 0 when any of them is
// ignored, so the answer is one process. -q is not an option: git refuses it
// with more than one pathname.
func checkIgnore(root string, paths ...string) (bool, error) {
	args := append([]string{"check-ignore", "--"}, paths...)
	cmd := execGit(root, args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		return true, nil
	}

	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() == 1 {
		return false, nil
	}

	msg := strings.TrimSpace(stderr.String())
	if msg == "" {
		msg = err.Error()
	}
	return false, fmt.Errorf("git check-ignore -- %s: %w: %s", strings.Join(paths, " "), err, msg)
}
