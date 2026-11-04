//go:build unix

package shellcmd

import (
	"context"
	"os/exec"
	"strings"
	"syscall"
)

// newCommand runs the human's command through sh -c, so quoting and redirection
// are the shell's business and the human's, not ours.
func newCommand(ctx context.Context, dir, command string) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	// The child leads its own process group, so a timeout can end the shell and
	// everything it started instead of the shell alone.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Dir = dir
	return cmd, nil
}

// killTree ends the shell and every process it started. The negative pid names
// the group Setpgid created: killing only sh would leave a `go test` writing to a
// pipe nobody reads, which is how a timeout became an orphan.
func killTree(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}

// quotePath makes one shell argument out of a path: single quotes stop sh from
// expanding anything inside, and an embedded quote is glued back in the only way
// sh lets it be written literally.
func quotePath(path string) string {
	return "'" + strings.ReplaceAll(path, "'", `'\''`) + "'"
}
