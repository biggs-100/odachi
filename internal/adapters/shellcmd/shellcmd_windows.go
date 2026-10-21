//go:build windows

package shellcmd

import (
	"context"
	"os/exec"
)

// newCommand runs the human's command through cmd /c, so quoting and
// redirection are the shell's business and the human's, not ours.
func newCommand(ctx context.Context, dir, command string) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, "cmd", "/c", command)
	cmd.Dir = dir
	return cmd, nil
}
