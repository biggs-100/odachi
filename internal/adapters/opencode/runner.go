package opencode

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Request is the thin production boundary: one OpenCode API call as the CLI
// spells it. The adapter builds it; a runner executes it.
type Request struct {
	Method string
	Path   string
	Body   string
}

// Runner executes one request and returns the response body. It exists so the
// adapter is testable without a live OpenCode service or model.
type Runner interface {
	Run(ctx context.Context, request Request) ([]byte, error)
}

// ExitError reports that OpenCode ran and failed. It is distinct from an error
// that prevented the command from running at all.
type ExitError struct {
	Code   int
	Stderr string
}

// Error spells the exit for a message or a test's failure.
func (err *ExitError) Error() string {
	return fmt.Sprintf("opencode exited with code %d", err.Code)
}

// ExecRunner is the production runner: it invokes `opencode api` so discovery
// and authentication stay owned by the CLI, then reports the response.
type ExecRunner struct {
	Binary string
}

// Run executes one request through the OpenCode CLI.
func (runner ExecRunner) Run(ctx context.Context, request Request) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, runner.binary(), runner.args(request)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return stdout.Bytes(), ctxErr
		}
		if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
			return stdout.Bytes(), &ExitError{Code: exitErr.ExitCode(), Stderr: stderr.String()}
		}
		return stdout.Bytes(), fmt.Errorf("run opencode: %w", err)
	}
	return stdout.Bytes(), nil
}

// args is the exact argv the production runner passes. The fresh-session call
// carries only the request body the adapter fixed; no session, continue, fork or
// resume flag is ever added.
func (runner ExecRunner) args(request Request) []string {
	args := []string{"api", strings.ToLower(request.Method), request.Path}
	if request.Body != "" {
		args = append(args, "--data", request.Body)
	}
	return args
}

func (runner ExecRunner) binary() string {
	if runner.Binary != "" {
		return runner.Binary
	}
	return "opencode"
}
