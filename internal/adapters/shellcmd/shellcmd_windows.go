//go:build windows

package shellcmd

import (
	"context"
	"os/exec"
	"strconv"
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

// killTree ends cmd.exe and every process it started. Windows has no process
// group to signal: taskkill /T walks the tree, and /F is because a timeout is not
// a request. Killing cmd alone leaves the command's children writing to a pipe
// nobody reads.
func killTree(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	_ = exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid)).Run()
	return cmd.Process.Kill()
}

// quotePath makes one cmd argument out of a path. Double quotes are enough: a
// Windows path cannot contain one, so there is nothing to escape.
func quotePath(path string) string {
	return `"` + path + `"`
}
