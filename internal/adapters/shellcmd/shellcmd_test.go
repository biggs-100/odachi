package shellcmd_test

import (
	"strings"
	"testing"

	"github.com/biggz/odachi/internal/adapters/shellcmd"
)

// These two commands behave the same in sh and in cmd, so the shared contract
// is tested once for every platform that has a shell.
func TestRunCapturesTheOutputOfASuccessfulCommand(t *testing.T) {
	output, exitCode, timedOut, err := shellcmd.Run(".", "echo ok", 30)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if timedOut {
		t.Fatal("Run reported a timeout for `echo ok`")
	}
	if exitCode != 0 {
		t.Errorf("exit code = %d, want 0", exitCode)
	}
	if !strings.Contains(output, "ok") {
		t.Errorf("output = %q, want the command's output", output)
	}
}

// A non-zero exit is a result of the command, not a failure of the adapter: the
// caller decides what a failing proof means.
func TestRunReportsTheExitCodeWithoutAnError(t *testing.T) {
	_, exitCode, timedOut, err := shellcmd.Run(".", "exit 7", 30)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if timedOut {
		t.Fatal("Run reported a timeout")
	}
	if exitCode != 7 {
		t.Errorf("exit code = %d, want 7", exitCode)
	}
}

func TestRunRejectsAnEmptyCommand(t *testing.T) {
	if _, _, _, err := shellcmd.Run(".", "   ", 30); err == nil {
		t.Fatal("Run accepted an empty command")
	}
}

func TestRunRejectsANonPositiveTimeout(t *testing.T) {
	if _, _, _, err := shellcmd.Run(".", "echo ok", 0); err == nil {
		t.Fatal("Run accepted a timeout of zero seconds")
	}
}
