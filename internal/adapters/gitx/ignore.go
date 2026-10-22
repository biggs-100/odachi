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
// evidence this tool writes (proof/ and .odachi/reviews/) has to be ignored
// BEFORE it is written, or writing it moves the very tree it describes.
//
// A directory that is not on disk is asked twice, with and without its trailing
// slash. Git only matches a directory rule (`/proof/`) when it can tell the path
// is a directory, and it cannot tell that from a name nothing created yet:
// measured, `git check-ignore -q proof` exits 1 while `proof/` exits 0 under the
// same rule. Asking both is what makes the answer the caller expects.
//
// Exit 0 means ignored and exit 1 means not ignored; both are answers. Anything
// else is an error carrying git's own stderr, because "exit status 128" alone
// tells nobody that the directory is not a repository.
func Ignored(root, path string) (bool, error) {
	ignored, err := checkIgnore(root, path)
	if err != nil || ignored {
		return ignored, err
	}
	if path == "" || strings.HasSuffix(path, "/") {
		return false, nil
	}
	return checkIgnore(root, path+"/")
}

// checkIgnore is the probe itself: one `git check-ignore -q` for one spelling.
func checkIgnore(root, path string) (bool, error) {
	cmd := execGit(root, "check-ignore", "-q", "--", path)

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
	return false, fmt.Errorf("git check-ignore -- %s: %w: %s", path, err, msg)
}
