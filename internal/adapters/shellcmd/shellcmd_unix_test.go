//go:build unix

package shellcmd_test

import (
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/adapters/shellcmd"
)

// A proof is what the command said, so both streams have to arrive. The seam is
// sh -c on every Unix.
func TestRunCombinesStdoutAndStderr(t *testing.T) {
	output, exitCode, _, err := shellcmd.Run(".", "echo out; echo err 1>&2", 30)
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

// sh -c is the platform shell, not a list of arguments: quoting stays the
// human's business.
func TestRunUsesShDashC(t *testing.T) {
	if _, _, _, err := shellcmd.Run(".", `printf '%s' "hola mundo"`, 30); err != nil {
		t.Fatalf("Run: %v", err)
	}
}

func TestRunTimesOutInsteadOfHanging(t *testing.T) {
	start := time.Now()
	_, exitCode, timedOut, err := shellcmd.Run(".", "sleep 5", 1)
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
