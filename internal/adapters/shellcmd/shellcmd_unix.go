//go:build unix

package shellcmd

import (
	"context"
	"os/exec"
	"strings"
)

// newCommand runs the human's command through sh -c, so quoting and redirection
// are the shell's business and the human's, not ours.
func newCommand(ctx context.Context, dir, command string) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = dir
	return cmd, nil
}

// quotePath makes one shell argument out of a path: single quotes stop sh from
// expanding anything inside, and an embedded quote is glued back in the only way
// sh lets it be written literally.
func quotePath(path string) string {
	return "'" + strings.ReplaceAll(path, "'", `'\''`) + "'"
}
