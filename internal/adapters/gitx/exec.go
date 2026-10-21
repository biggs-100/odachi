package gitx

import (
	"os"
	"os/exec"
)

// execGit builds one git command rooted at dir. The environment is pinned so a
// gate never prompts for credentials and never inherits a localized message
// that tests cannot assert on.
func execGit(dir string, args ...string) *exec.Cmd {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), "LC_ALL=C", "GIT_TERMINAL_PROMPT=0", "GIT_ASKPASS=echo")
	return cmd
}
