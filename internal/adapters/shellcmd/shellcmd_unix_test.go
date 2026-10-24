//go:build unix

package shellcmd_test

import (
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/adapters/shellcmd"
)

// The two streams arrive separately: the proof composes them, and a reviewer's
// answer is stdout only. The seam is sh -c on every Unix.
func TestRunSeparatesStdoutAndStderr(t *testing.T) {
	stdout, stderr, exitCode, _, err := shellcmd.Run(".", "echo out; echo err 1>&2", 30)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("exit code = %d, want 0", exitCode)
	}
	if !strings.Contains(stdout, "out") || strings.Contains(stdout, "err") {
		t.Errorf("stdout = %q, want only the first stream", stdout)
	}
	if !strings.Contains(stderr, "err") || strings.Contains(stderr, "out") {
		t.Errorf("stderr = %q, want only the second stream", stderr)
	}
}

// sh -c is the platform shell, not a list of arguments: quoting stays the
// human's business.
func TestRunUsesShDashC(t *testing.T) {
	if _, _, _, _, err := shellcmd.Run(".", `printf '%s' "hola mundo"`, 30); err != nil {
		t.Fatalf("Run: %v", err)
	}
}

func TestRunTimesOutInsteadOfHanging(t *testing.T) {
	start := time.Now()
	_, _, exitCode, timedOut, err := shellcmd.Run(".", "sleep 5", 1)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !timedOut {
		t.Fatal("Run did not report the timeout")
	}
	if exitCode >= 0 {
		t.Errorf("exit code = %d, want a negative code for a killed command", exitCode)
	}
	if elapsed > 4*time.Second {
		t.Errorf("Run took %s, want it to stop near the 1s timeout", elapsed)
	}
}

// The prompt reaches the reviewer as its stdin, because a reviewer that reads
// stdin needs no quoting and no knowledge of the human's shell.
func TestRunInputFeedsThePromptOnStdin(t *testing.T) {
	output, _, exitCode, _, err := shellcmd.RunInput(".", "cat", "revisá esto\nsegunda línea\n", 30)
	if err != nil {
		t.Fatalf("RunInput: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("exit code = %d, want 0", exitCode)
	}
	if output != "revisá esto\nsegunda línea\n" {
		t.Errorf("output = %q, want the prompt verbatim", output)
	}
}

// The second way in is a path: a reviewer that takes a file (an agent CLI) gets
// one, quoted by the adapter because sh quoting is the adapter's platform
// knowledge, not the human's.
func TestRunInputReplacesThePromptTokenWithAReadablePath(t *testing.T) {
	output, _, exitCode, _, err := shellcmd.RunInput(".", "cat {prompt}", "el prompt del revisor\n", 30)
	if err != nil {
		t.Fatalf("RunInput: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("exit code = %d, want 0 (output: %q)", exitCode, output)
	}
	if output != "el prompt del revisor\n" {
		t.Errorf("output = %q, want the prompt file's content", output)
	}
}

// The path is quoted, so a prompt whose temporary directory has a space in it is
// still one argument.
func TestRunInputQuotesThePromptPath(t *testing.T) {
	output, _, _, _, err := shellcmd.RunInput(".", `printf '%s' {prompt}`, "x\n", 30)
	if err != nil {
		t.Fatalf("RunInput: %v", err)
	}
	if strings.Contains(output, " ") && !strings.HasPrefix(output, "'") {
		t.Errorf("path = %q, want it quoted when it has a space", output)
	}
}
