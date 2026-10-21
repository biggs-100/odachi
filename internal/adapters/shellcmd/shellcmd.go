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
	"os/exec"
	"strings"
	"time"
)

// waitDelay bounds Wait after the process was killed or exited while a child
// still holds the output pipes. Without it a timeout can hang on the very
// command it was supposed to stop.
const waitDelay = 2 * time.Second

// ErrUnsupported is returned on a platform with no known shell.
var ErrUnsupported = errors.New("no sé correr comandos en esta plataforma: no hay shell conocido; corré la prueba a mano y pasala con odachi check --proof <archivo>")

// Run executes command through the platform shell in dir and reports its
// combined output, its exit code and whether the timeout expired.
//
// A non-zero exit is NOT an adapter error: it is the command's result, and the
// caller decides what a failing proof means. err is reserved for "this could
// not be run at all". A timeout returns timedOut true and a negative exit code.
func Run(dir, command string, timeoutSeconds int) (output string, exitCode int, timedOut bool, err error) {
	if strings.TrimSpace(command) == "" {
		return "", 0, false, errors.New("no hay comando para correr: pasá uno con odachi check --run \"<comando>\"")
	}
	if timeoutSeconds <= 0 {
		return "", 0, false, errors.New("el tiempo límite tiene que ser mayor que cero segundos")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	cmd, err := newCommand(ctx, dir, command)
	if err != nil {
		return "", 0, false, err
	}

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
