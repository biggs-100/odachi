//go:build windows

package shellcmd_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/adapters/shellcmd"
)

// The two streams arrive separately: the proof composes them, and a reviewer's
// answer is stdout only. The seam is cmd /c on Windows.
func TestRunSeparatesStdoutAndStderr(t *testing.T) {
	stdout, stderr, exitCode, _, err := shellcmd.Run(".", `(echo out) & (echo err 1>&2)`, 30)
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

func TestRunTimesOutInsteadOfHanging(t *testing.T) {
	start := time.Now()
	_, _, exitCode, timedOut, err := shellcmd.Run(".", `ping -n 6 127.0.0.1 >nul`, 1)
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

// TestRunATimeoutKillsTheWholeTree: `start /b` leaves a child that outlives a
// kill of cmd.exe alone; it would write after the timeout. taskkill /T must end
// the whole tree, not only the shell.
//
// The grandchild waits five seconds before writing and the test watches for seven:
// under a full-suite load taskkill can be delayed past a one-second margin, and a
// test that fails when the behaviour is right is a gate that cries wolf. A real
// survivor still writes inside the window.
func TestRunATimeoutKillsTheWholeTree(t *testing.T) {
	dir := t.TempDir()
	command := `start /b cmd /c "ping -n 6 127.0.0.1 >nul & echo late>late.txt" & ping -n 12 127.0.0.1 >nul`
	_, _, _, timedOut, err := shellcmd.Run(dir, command, 1)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !timedOut {
		t.Fatal("Run did not report the timeout")
	}
	time.Sleep(7 * time.Second)
	if _, err := os.Stat(filepath.Join(dir, "late.txt")); err == nil {
		t.Fatal("the grandchild outlived the timeout and wrote its file")
	}
}

// The prompt reaches the reviewer as its stdin, because a reviewer that reads
// stdin needs no quoting and no knowledge of the human's shell.
func TestRunInputFeedsThePromptOnStdin(t *testing.T) {
	output, _, exitCode, _, err := shellcmd.RunInput(".", "findstr /r .*", "revisa esto\nsegunda linea\n", 30)
	if err != nil {
		t.Fatalf("RunInput: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("exit code = %d, want 0 (output: %q)", exitCode, output)
	}
	if !strings.Contains(output, "revisa esto") || !strings.Contains(output, "segunda linea") {
		t.Errorf("output = %q, want the prompt verbatim", output)
	}
}

// The second way in is a path: a reviewer that takes a file (an agent CLI) gets
// one, quoted by the adapter because cmd quoting is the adapter's platform
// knowledge, not the human's.
func TestRunInputReplacesThePromptTokenWithAReadablePath(t *testing.T) {
	output, _, exitCode, _, err := shellcmd.RunInput(".", "type {prompt}", "el prompt del revisor\n", 30)
	if err != nil {
		t.Fatalf("RunInput: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("exit code = %d, want 0 (output: %q)", exitCode, output)
	}
	if !strings.Contains(output, "el prompt del revisor") {
		t.Errorf("output = %q, want the prompt file's content", output)
	}
}

// The path is always quoted, not only when it happens to have a space: the
// adapter cannot know what the temporary directory will be called on another
// machine, so the rule is unconditional.
func TestRunInputAlwaysQuotesThePromptPath(t *testing.T) {
	output, _, _, _, err := shellcmd.RunInput(".", "echo {prompt}", "x\n", 30)
	if err != nil {
		t.Fatalf("RunInput: %v", err)
	}
	if !strings.Contains(output, `"`) {
		t.Errorf("path = %q, want it quoted", output)
	}
}

// cmd.exe is not an argv program: it parses its own command line and does not
// understand the backslash escaping Go applies to a quoted argument, so before
// this rule a quoted argument arrived mangled. `--run "go test -run \"TestX\""`
// and every `--exec` with a quoted message were broken on Windows only.
func TestRunKeepsAQuotedArgumentIntact(t *testing.T) {
	output, _, exitCode, _, err := shellcmd.Run(".", `echo "hola mundo"`, 30)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("exit code = %d, want 0 (output: %q)", exitCode, output)
	}
	if strings.Contains(output, `\"`) {
		t.Errorf("output = %q, want the quotes the command wrote, not backslash escapes", output)
	}
	if !strings.Contains(output, `"hola mundo"`) {
		t.Errorf("output = %q, want the quoted argument intact", output)
	}
}

// A path with a space is the case quoting exists for: a directory named
// "proyecto de prueba" is what an ordinary person names a folder.
func TestRunReachesAPathWithASpace(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "hola.txt"), []byte("contenido del archivo\n"), 0o600); err != nil {
		t.Fatalf("writing the file: %v", err)
	}

	output, _, exitCode, _, err := shellcmd.Run(dir, `type "hola.txt"`, 30)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("exit code = %d, want 0 (output: %q)", exitCode, output)
	}
	if !strings.Contains(output, "contenido del archivo") {
		t.Errorf("output = %q, want the file's content", output)
	}
}
