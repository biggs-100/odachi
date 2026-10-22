// Package shellcmd is the shell edge: it runs the proof command through the
// platform shell and returns stdout+stderr combined, because a proof is what the
// command said, not what one stream happened to carry.
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
var ErrUnsupported = errors.New("no sé correr comandos en esta plataforma: no hay shell conocido; corré la prueba a mano y pasala con odachi check --proof <archivo>")

// Run executes command through the platform shell in dir and reports its
// combined output, its exit code and whether the timeout expired.
//
// A non-zero exit is NOT an adapter error: it is the command's result, and the
// caller decides what a failing proof means. err is reserved for "this could
// not be run at all". A timeout returns timedOut true and a negative exit code.
func Run(dir, command string, timeoutSeconds int) (output string, exitCode int, timedOut bool, err error) {
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
func RunInput(dir, command, input string, timeoutSeconds int) (output string, exitCode int, timedOut bool, err error) {
	if strings.TrimSpace(command) == "" {
		return "", 0, false, errors.New("no hay comando para correr: pasá uno con odachi check --run \"<comando>\"")
	}
	if timeoutSeconds <= 0 {
		return "", 0, false, errors.New("el tiempo límite tiene que ser mayor que cero segundos")
	}

	stdin, prepared, cleanup, err := prepareInput(command, input)
	if err != nil {
		return "", 0, false, err
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	cmd, err := newCommand(ctx, dir, prepared)
	if err != nil {
		return "", 0, false, err
	}
	cmd.Stdin = stdin

	// Both streams into one buffer: the order is the order the command wrote,
	// which is the order a human reads the failure in.
	var combined bytes.Buffer
	cmd.Stdout = &combined
	cmd.Stderr = &combined
	cmd.WaitDelay = waitDelay

	runErr := cmd.Run()
	output = combined.String()

	if ctx.Err() != nil {
		// The context expired: the process was killed, and the partial output
		// is what the command had said so far.
		return output, -1, true, nil
	}

	if runErr != nil {
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			return output, exitErr.ExitCode(), false, nil
		}
		return output, -1, false, fmt.Errorf("no pude correr el comando: %w", runErr)
	}

	return output, 0, false, nil
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
		return nil, "", func() {}, fmt.Errorf("no pude preparar el prompt en un archivo temporal: %w", err)
	}
	cleanup = func() { os.RemoveAll(dir) }

	path := filepath.Join(dir, "prompt.md")
	if err := os.WriteFile(path, []byte(input), 0o600); err != nil {
		return nil, "", cleanup, fmt.Errorf("no pude escribir el prompt en un archivo temporal: %w", err)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, "", cleanup, fmt.Errorf("no pude abrir el prompt temporal: %w", err)
	}
	cleanup = func() {
		file.Close()
		os.RemoveAll(dir)
	}

	return file, strings.ReplaceAll(command, promptToken, quotePath(path)), cleanup, nil
}
