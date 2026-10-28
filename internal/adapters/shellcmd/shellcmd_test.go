package shellcmd_test

import (
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/adapters/shellcmd"
)

// These two commands behave the same in sh and in cmd, so the shared contract
// is tested once for every platform that has a shell.
func TestRunCapturesTheOutputOfASuccessfulCommand(t *testing.T) {
	output, _, exitCode, timedOut, err := shellcmd.Run(".", "echo ok", 30)
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
	_, _, exitCode, timedOut, err := shellcmd.Run(".", "exit 7", 30)
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
	if _, _, _, _, err := shellcmd.Run(".", "   ", 30); err == nil {
		t.Fatal("Run accepted an empty command")
	}
}

func TestRunRejectsANonPositiveTimeout(t *testing.T) {
	if _, _, _, _, err := shellcmd.Run(".", "echo ok", 0); err == nil {
		t.Fatal("Run accepted a timeout of zero seconds")
	}
}

// A prompt is much bigger than a pipe buffer, and a reviewer that never reads its
// stdin must not be able to block odachi: the input goes to a file, not a pipe.
func TestRunInputDoesNotHangWhenTheCommandIgnoresItsStdin(t *testing.T) {
	input := strings.Repeat("linea del prompt que el comando ignora\n", 20000)

	done := make(chan struct{})
	var output string
	var err error
	go func() {
		defer close(done)
		output, _, _, _, err = shellcmd.RunInput(".", "exit 0", input, 30)
	}()

	select {
	case <-done:
	case <-time.After(20 * time.Second):
		t.Fatal("RunInput blocked on a command that never read its stdin")
	}

	if err != nil {
		t.Fatalf("RunInput: %v", err)
	}
	if output != "" {
		t.Errorf("output = %q, want nothing", output)
	}
}

// TestQuoteSpellsAPathForTheShell: the review substitutes {verdict} itself, so it
// asks this adapter for the spelling {prompt} already gets. A path with a space is
// the case that matters: an unquoted one ends at the space in every shell odachi
// supports.
func TestQuoteSpellsAPathForTheShell(t *testing.T) {
	quoted := shellcmd.Quote(`C:\Users\ana perez\clone\.git\odachi\answers\x.json`)
	if !strings.Contains(quoted, "ana perez") {
		t.Fatalf("Quote = %q, want the path itself", quoted)
	}
	if quoted == `C:\Users\ana perez\clone\.git\odachi\answers\x.json` {
		t.Errorf("Quote = %q, want it quoted", quoted)
	}
}

// A command without the token is untouched, and a command with the token but no
// prompt is left alone too: nothing is invented.
func TestRunInputLeavesTheTokenAloneWhenThereIsNoPrompt(t *testing.T) {
	output, _, exitCode, _, err := shellcmd.RunInput(".", "echo {prompt}", "", 30)
	if err != nil {
		t.Fatalf("RunInput: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("exit code = %d, want 0", exitCode)
	}
	if !strings.Contains(output, "{prompt}") {
		t.Errorf("output = %q, want the token untouched", output)
	}
}
