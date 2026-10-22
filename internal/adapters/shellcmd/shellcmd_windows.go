//go:build windows

package shellcmd

import (
	"context"
	"os/exec"
	"syscall"
)

// newCommand runs the human's command through cmd /c, so quoting and
// redirection are the shell's business and the human's, not ours.
//
// cmd.exe is not an argv program: it parses its own command line and does not
// understand the backslash escaping Go applies to an argument that contains
// quotes, so `--run "go test -run \"TestX\""` arrived mangled (measured: the
// shell reported a syntax error, or echoed the backslashes literally). /s /c
// "<command>" is the documented way to hand cmd one string verbatim: it strips
// exactly the outer pair of quotes and runs the rest as written.
func newCommand(ctx context.Context, dir, command string) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, "cmd")
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: `/s /c "` + command + `"`}
	cmd.Dir = dir
	return cmd, nil
}

// quotePath makes one cmd argument out of a path. Double quotes are enough: a
// Windows path cannot contain one, so there is nothing to escape.
func quotePath(path string) string {
	return `"` + path + `"`
}
