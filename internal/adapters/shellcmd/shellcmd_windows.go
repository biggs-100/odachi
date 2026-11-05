//go:build windows

package shellcmd

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

var taskkillCommand = "taskkill"

// newCommand runs the human's command through cmd /c, so quoting and
// redirection are the shell's business and the human's, not ours.
//
// cmd.exe is not an argv program: it parses its own command line and does not
// understand the backslash escaping Go applies to an argument that contains
// quotes, so `--run "go test -run \"TestX\""` arrived mangled (measured: the
// shell reported a syntax error, or echoed the backslashes literally). /s /c
// "<command>" is the documented way to hand cmd one string verbatim: it strips
// exactly the outer pair of quotes and runs the rest as written.
//
// A .ps1 script is not a program cmd can run: it hands it to the file
// association, which is not PowerShell, and the proof then runs something nobody
// asked for. The refusal names the invocation that does run it.
func newCommand(ctx context.Context, dir, command string) (*exec.Cmd, error) {
	if script, rest, ok := powerShellScript(command); ok {
		return nil, fmt.Errorf(
			"a .ps1 command is not run through cmd /c; run powershell -NoProfile -ExecutionPolicy Bypass -File %s%s",
			quotePath(script), rest)
	}
	cmd := exec.CommandContext(ctx, "cmd")
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: `/s /c "` + command + `"`}
	cmd.Dir = dir
	return cmd, nil
}

// powerShellScript reads the first token of a command and says whether it is a
// .ps1 script. A quoted path is one token even when it has spaces; everything
// after it is the script's own arguments and travels unchanged.
func powerShellScript(command string) (script, rest string, ok bool) {
	trimmed := strings.TrimLeft(command, " \t\r\n")
	if trimmed == "" {
		return "", "", false
	}

	if trimmed[0] == '"' {
		end := strings.IndexByte(trimmed[1:], '"')
		if end < 0 {
			return "", "", false
		}
		script = trimmed[1 : 1+end]
		rest = trimmed[1+end+1:]
	} else {
		end := strings.IndexAny(trimmed, " \t\r\n")
		if end < 0 {
			script = trimmed
		} else {
			script = trimmed[:end]
			rest = trimmed[end:]
		}
	}
	if !strings.EqualFold(filepath.Ext(script), ".ps1") {
		return "", "", false
	}
	return script, rest, true
}

// killTree ends cmd.exe and every process it started. Windows has no process
// group to signal: taskkill /T walks the tree, and /F is because a timeout is not
// a request. Killing cmd alone leaves the command's children writing to a pipe
// nobody reads.
func killTree(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	taskkillErr := exec.Command(taskkillCommand, "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid)).Run()
	killErr := cmd.Process.Kill()
	if taskkillErr != nil {
		if killErr != nil {
			return killErr
		}
		return taskkillErr
	}
	return killErr
}

// quotePath makes one cmd argument out of a path. Double quotes are enough: a
// Windows path cannot contain one, so there is nothing to escape.
func quotePath(path string) string {
	return `"` + path + `"`
}
