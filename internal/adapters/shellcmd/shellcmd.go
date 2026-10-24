// Package shellcmd is the shell edge: it runs the proof command through the
// platform shell and reports stdout and stderr separately. A proof is what the
// command said and both streams belong in it; a reviewer's answer, on the other
// hand, is stdout only — mixing the streams is how a CLI banner poisons a JSON
// answer.
//
// The platform split lives in one file per platform: sh -c on Unix, cmd /c on
// Windows, and a refusal everywhere else (birth decision 11: refuse instead of
// degrade).
package shellcmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// waitDelay bounds Wait after the process was killed or exited while a child
// still holds the output pipes. Without it a timeout can hang on the very
// command it was supposed to stop.
const waitDelay = 2 * time.Second

// promptToken is the placeholder a command uses to receive the prompt as a file
// path. It exists because shell quoting is not portable: `$(cat)` does not exist
// in cmd /c, and a reviewer that takes a file needs a path, not a stream.
const promptToken = "{prompt}"

// ErrUnsupported is returned on a platform with no known shell.
var ErrUnsupported = errors.New("cannot run commands on this platform: there is no known shell; run the proof by hand and pass it with odachi check --proof <file>")

// Run executes command through the platform shell in dir and reports its two
// streams, its exit code and whether the timeout expired.
//
// A non-zero exit is NOT an adapter error: it is the command's result, and the
// caller decides what a failing proof means. err is reserved for "this could
// not be run at all". A timeout returns timedOut true and a negative exit code.
func Run(dir, command string, timeoutSeconds int) (stdout, stderr string, exitCode int, timedOut bool, err error) {
	return RunInput(dir, command, "", timeoutSeconds)
}

// RunInput is Run for a command that needs something to work on: the input is
// written to a temporary file, that file is the command's stdin, and every
// {prompt} in the command is replaced by the file's path, quoted for the shell.
//
// Both delivery channels come from the same file on purpose. stdin is the
// natural pipe for a command that reads one, and the path is what an agent CLI
// that takes a file needs. A file, not a pipe, feeds stdin: a command that
// ignores its input then cannot block the caller by leaving the buffer full.
func RunInput(dir, command, input string, timeoutSeconds int) (stdout, stderr string, exitCode int, timedOut bool, err error) {
	if strings.TrimSpace(command) == "" {
		return "", "", 0, false, errors.New("there is no command to run: pass one with odachi check --run \"<command>\"")
	}
	if timeoutSeconds <= 0 {
		return "", "", 0, false, errors.New("the timeout must be greater than zero seconds")
	}

	stdin, prepared, cleanup, err := prepareInput(command, input)
	if err != nil {
		return "", "", 0, false, err
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	cmd, err := newCommand(ctx, dir, prepared)
	if err != nil {
		return "", "", 0, false, err
	}
	cmd.Stdin = stdin

	// Each stream in its own buffer: the proof composes both, the reviewer reads
	// its answer from stdout, and neither has to guess what the other meant.
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	cmd.WaitDelay = waitDelay

	runErr := cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()

	if ctx.Err() != nil {
		// The context expired: the process was killed, and the partial output
		// is what the command had said so far.
		return stdout, stderr, -1, true, nil
	}

	if runErr != nil {
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			return stdout, stderr, exitErr.ExitCode(), false, nil
		}
		return stdout, stderr, -1, false, fmt.Errorf("could not run the command: %w", runErr)
	}

	return stdout, stderr, 0, false, nil
}

// prepareInput writes the input to a temporary file and returns it opened as the
// command's stdin, the command with every token replaced by its quoted path, and
// the cleanup that removes the directory afterwards.
//
// With no input nothing is written and no token is replaced: a command that asks
// for a prompt it was not given keeps its literal text instead of being pointed
// at a file that does not exist.
func prepareInput(command, input string) (stdin *os.File, prepared string, cleanup func(), err error) {
	if input == "" {
		return nil, command, func() {}, nil
	}

	dir, err := os.MkdirTemp("", "odachi-prompt-")
	if err != nil {
		return nil, "", func() {}, fmt.Errorf("could not prepare the prompt in a temporary file: %w", err)
	}
	cleanup = func() { os.RemoveAll(dir) }

	path := filepath.Join(dir, "prompt.md")
	if err := os.WriteFile(path, []byte(input), 0o600); err != nil {
		return nil, "", cleanup, fmt.Errorf("could not write the prompt to a temporary file: %w", err)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, "", cleanup, fmt.Errorf("could not open the temporary prompt: %w", err)
	}
	cleanup = func() {
		file.Close()
		os.RemoveAll(dir)
	}

	return file, strings.ReplaceAll(command, promptToken, quotePath(path)), cleanup, nil
}
