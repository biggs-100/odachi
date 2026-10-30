package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestSetupNamesTheClientsItDidNotFind: detection looks at the home and picks the
// clients that are there. A client that is not there is not a silent omission: an
// installer that reports three destinations and stays quiet about the fourth looks
// like it forgot.
func TestSetupNamesTheClientsItDidNotFind(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	// OpenCode is here, Claude Code is not.
	writeFile(t, home, ".config/opencode/opencode.json", "{}")

	dir := t.TempDir()
	initRepo(t, dir)

	code, out, errOut := runDir(t, dir, []string{"setup", "--dry-run"}, "")
	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout: %q, stderr: %q)", code, out, errOut)
	}

	report := out + errOut
	for _, want := range []string{"not installed here", "claude", "~/.claude"} {
		if !strings.Contains(report, want) {
			t.Errorf("the report does not say %q:\n%s", want, report)
		}
	}
	if strings.Contains(report, "not installed here: opencode") {
		t.Errorf("the report calls a client that is here absent:\n%s", report)
	}
	if _, err := filepath.Abs(home); err != nil {
		t.Fatalf("abs: %v", err)
	}
}

// TestSetupWithBothClientsHasNoOneToMiss: the note is about detection, so when both
// config folders exist it is not printed at all.
func TestSetupWithBothClientsHasNoOneToMiss(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	writeFile(t, home, ".config/opencode/opencode.json", "{}")
	writeFile(t, home, ".claude/settings.json", "{}")

	dir := t.TempDir()
	initRepo(t, dir)

	code, out, errOut := runDir(t, dir, []string{"setup", "--dry-run"}, "")
	if code != 0 {
		t.Fatalf("code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if strings.Contains(out+errOut, "not installed here") {
		t.Errorf("the report names an absent client when both are here:\n%s%s", out, errOut)
	}
}
