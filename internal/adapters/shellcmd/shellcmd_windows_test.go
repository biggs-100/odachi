//go:build windows

package shellcmd_test

import (
	"strings"
	"testing"
	"time"

	"github.com/biggz/odachi/internal/adapters/shellcmd"
)

// A proof is what the command said, so both streams have to arrive. The seam is
// cmd /c on Windows.
func TestRunCombinesStdoutAndStderr(t *testing.T) {
	output, exitCode, _, err := shellcmd.Run(".", `(echo out) & (echo err 1>&2)`, 30)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("exit code = %d, want 0", exitCode)
	}
	if !strings.Contains(output, "out") || !strings.Contains(output, "err") {
		t.Errorf("output = %q, want both streams", output)
	}
}

func TestRunTimesOutInsteadOfHanging(t *testing.T) {
	start := time.Now()
	_, exitCode, timedOut, err := shellcmd.Run(".", `ping -n 6 127.0.0.1 >nul`, 1)
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
	if elapsed > 6*time.Second {
		t.Errorf("Run took %s, want it to stop near the 1s timeout", elapsed)
	}
}
