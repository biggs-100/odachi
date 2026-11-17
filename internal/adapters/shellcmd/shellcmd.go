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
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

// Watchdog bounds silence, not duration. FirstOutput is how long a command may
// run before its first byte; Idle is how long it may go quiet between bytes.
// Zero disables that clock. The total timeout still applies in both cases.
//
// A model that is slow but keeps writing is alive; one that stops writing has
// stopped working. A single total deadline cannot tell them apart.
type Watchdog struct {
	FirstOutput time.Duration
	Idle        time.Duration
}

const (
	// FirstOutputTimeout is how long a reviewer may stay silent before its first
	// byte. Two minutes is longer than a slow model's first token and shorter
	// than the ten-minute total deadline it replaces for a silent answer.
	FirstOutputTimeout = 120 * time.Second
	// IdleTimeout is how long a reviewer may go quiet between bytes.
	IdleTimeout = 120 * time.Second
)

// pingWriter tells someone that bytes arrived without changing what is written.
type pingWriter struct {
	io.Writer
	ping func()
}

func (w pingWriter) Write(p []byte) (int, error) {
	n, err := w.Writer.Write(p)
	w.ping()
	return n, err
}

// waitDelay bounds Wait after the process was killed or exited while a child
// still holds the output pipes. Without it a timeout can hang on the very
// command it was supposed to stop.
const waitDelay = 2 * time.Second

// promptToken is the placeholder a command uses to receive the prompt as a file
// path. It exists because shell quoting is not portable: `$(cat)` does not exist
// in cmd /c, and a reviewer that takes a file needs a path, not a stream.
const promptToken = "{prompt}"

// Quote spells one path the way this platform's shell needs it. It is how a
// caller that substitutes a path of its own - the review's {verdict} - gets the
// same spelling the adapter gives {prompt}.
func Quote(path string) string { return quotePath(path) }

// ErrUnsupported is returned on a platform with no known shell.
var ErrUnsupported = errors.New("cannot run commands on this platform: there is no known shell; run the proof by hand and pass it with odachi check --proof <file>")

// Run executes command through the platform shell in dir and reports its two
// streams, its exit code and whether the timeout expired.
//
// A non-zero exit is NOT an adapter error: it is the command's result, and the
// caller decides what a failing proof means. err is reserved for "this could
// not be run at all". A timeout returns timedOut true and a negative exit code.
func Run(dir, command string, timeoutSeconds int) (stdout, stderr string, exitCode int, timedOut bool, err error) {
	return run(dir, command, "", timeoutSeconds, nil, Watchdog{})
}

// RunStream is Run with a mirror: everything the command prints reaches w while
// it is being captured, so a human watching a proof sees the work instead of a
// silent terminal until the process ends. A nil w is Run exactly.
//
// The mirror sees both streams, in whatever order they arrive: it is a window,
// not the proof. The returned strings keep them apart, because a proof composes
// them and a reviewer reads its answer from stdout only. Two goroutines copy the
// streams, so w must be safe for concurrent use -the terminal is.
func RunStream(dir, command string, timeoutSeconds int, w io.Writer) (stdout, stderr string, exitCode int, timedOut bool, err error) {
	return run(dir, command, "", timeoutSeconds, w, Watchdog{})
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
	return RunInputWithLimits(dir, command, input, timeoutSeconds,
		Watchdog{FirstOutput: FirstOutputTimeout, Idle: IdleTimeout})
}

// RunInputWithLimits is RunInput with explicit silence clocks. The defaults in
// RunInput are for a model; a test or a caller with a different budget names its
// own. Run and RunStream never pass a watchdog: a proof command such as go test
// may stay silent for the whole of its total deadline and still be working.
func RunInputWithLimits(dir, command, input string, timeoutSeconds int, watchdog Watchdog) (stdout, stderr string, exitCode int, timedOut bool, err error) {
	return run(dir, command, input, timeoutSeconds, nil, watchdog)
}

// run is the one implementation: the three entry points differ only in the input
// they prepare and the mirror they write to.
func run(dir, command, input string, timeoutSeconds int, mirror io.Writer, watchdog Watchdog) (stdout, stderr string, exitCode int, timedOut bool, err error) {
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
	// The parent keeps the total deadline. The child is what the silence clock
	// cancels, so either clock kills the same process tree through cmd.Cancel.
	watchCtx, cancelSilence := context.WithCancel(ctx)
	defer cancelSilence()

	cmd, err := newCommand(watchCtx, dir, prepared)
	if err != nil {
		return "", "", 0, false, err
	}
	// The context's default cancel ends the shell alone; the work is its
	// children. A timeout that leaves a `go test` writing to a pipe nobody reads
	// is not a timeout.
	cmd.Cancel = func() error { return killTree(cmd) }
	cmd.Stdin = stdin

	// Each stream in its own buffer: the proof composes both, the reviewer reads
	// its answer from stdout, and neither has to guess what the other meant. The
	// mirror, when there is one, is the same bytes on their way to a human: it
	// changes nothing about what is returned.
	var outBuf, errBuf bytes.Buffer
	var outWriter, errWriter io.Writer = &outBuf, &errBuf
	if mirror != nil {
		outWriter = io.MultiWriter(&outBuf, mirror)
		errWriter = io.MultiWriter(&errBuf, mirror)
	}

	var silent atomic.Bool
	if watchdog.FirstOutput > 0 || watchdog.Idle > 0 {
		activity := make(chan struct{}, 1)
		ping := func() {
			select {
			case activity <- struct{}{}:
			default:
			}
		}
		outWriter = pingWriter{Writer: outWriter, ping: ping}
		errWriter = pingWriter{Writer: errWriter, ping: ping}
		stop := make(chan struct{})
		defer close(stop)
		go func() {
			first := watchdog.FirstOutput > 0
			wait := watchdog.FirstOutput
			if !first {
				wait = watchdog.Idle
			}
			timer := time.NewTimer(wait)
			defer timer.Stop()
			for {
				select {
				case <-activity:
					first = false
					if !timer.Stop() {
						select {
						case <-timer.C:
						default:
						}
					}
					timer.Reset(watchdog.Idle)
				case <-timer.C:
					silent.Store(true)
					cancelSilence()
					return
				case <-stop:
					return
				}
			}
		}()
	}
	cmd.Stdout = outWriter
	cmd.Stderr = errWriter
	cmd.WaitDelay = waitDelay

	runErr := cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()

	if ctx.Err() != nil {
		// The total deadline expired: the process was killed, and the partial
		// output is what the command had said so far.
		return stdout, stderr, -1, true, nil
	}
	if silent.Load() {
		// The silence clock fired first. It is the same timedOut signal the
		// callers already read, with a different reason behind it.
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
	cleanup = func() { _ = os.RemoveAll(dir) }

	path := filepath.Join(dir, "prompt.md")
	if err := os.WriteFile(path, []byte(input), 0o600); err != nil {
		return nil, "", cleanup, fmt.Errorf("could not write the prompt to a temporary file: %w", err)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, "", cleanup, fmt.Errorf("could not open the temporary prompt: %w", err)
	}
	cleanup = func() {
		_ = file.Close()
		_ = os.RemoveAll(dir)
	}

	return file, strings.ReplaceAll(command, promptToken, quotePath(path)), cleanup, nil
}
