//go:build unix

package shellcmd

import (
	"context"
	"os/exec"
)

// newCommand runs the human's command through sh -c, so quoting and redirection
// are the shell's business and the human's, not ours.
func newCommand(ctx context.Context, dir, command string) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = dir
	return cmd, nil
}
