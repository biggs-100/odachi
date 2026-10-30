package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// TestSetupDryRunWritesNothingAndNamesWhatItWouldWrite: the rehearsal has to be
// believable, and believing it means zero bytes: the home is exactly what it was,
// and the report names every destination it would write.
func TestSetupDryRunWritesNothingAndNamesWhatItWouldWrite(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	writeFile(t, home, ".config/opencode/opencode.json", "{}")
	before := homeTree(t, home)

	dir := t.TempDir()
	initRepo(t, dir)

	code, out, errOut := runDir(t, dir, []string{"setup", "--dry-run"}, "")
	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout: %q, stderr: %q)", code, out, errOut)
	}

	if !slices.Equal(homeTree(t, home), before) {
		t.Errorf("--dry-run touched the home: %v, want %v", homeTree(t, home), before)
	}
	for _, want := range []string{"dry run", "would write", "~/.config/opencode/agents", "~/.agents/skills"} {
		if !strings.Contains(out+errOut, want) {
			t.Errorf("the report does not say %q:\n%s%s", want, out, errOut)
		}
	}
	if strings.Contains(out, "wrote ") {
		t.Errorf("the rehearsal claims it wrote:\n%s", out)
	}
}

// TestSetupAfterARealRunHasNothingToWrite: the same command twice is a no-op, and
// the rehearsal of an up-to-date machine says so instead of naming files.
func TestSetupAfterARealRunHasNothingToWrite(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	writeFile(t, home, ".config/opencode/opencode.json", "{}")

	dir := t.TempDir()
	initRepo(t, dir)

	if code, _, errOut := runDir(t, dir, []string{"setup"}, ""); code != 0 {
		t.Fatalf("the real setup failed: %s", errOut)
	}
	settled := homeTree(t, home)

	code, out, errOut := runDir(t, dir, []string{"setup", "--dry-run"}, "")
	if code != 0 {
		t.Fatalf("the rehearsal failed: %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if !slices.Equal(homeTree(t, home), settled) {
		t.Errorf("the rehearsal touched a settled home")
	}
	if !strings.Contains(out, "unchanged") {
		t.Errorf("the report does not say the home is already current:\n%s", out)
	}
	if strings.Contains(out, "would write") {
		t.Errorf("an up-to-date home still reports files to write:\n%s", out)
	}

	// And the real run again changes nothing either.
	if code, _, errOut := runDir(t, dir, []string{"setup"}, ""); code != 0 {
		t.Fatalf("the second real setup failed: %s", errOut)
	}
	if !slices.Equal(homeTree(t, home), settled) {
		t.Errorf("the second real setup touched the home")
	}
}

// TestSetupDryRunLeavesAStaleDoorAlone: the case the rehearsal exists for -a home
// written by an older binary. The bytes are not rewritten, and the run names the
// file it would replace.
func TestSetupDryRunLeavesAStaleDoorAlone(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	writeFile(t, home, ".config/opencode/opencode.json", "{}")
	stale := filepath.Join(home, ".config", "opencode", "agents", "odachi.md")
	writeFile(t, home, ".config/opencode/agents/odachi.md", "a door from an older odachi\n")

	dir := t.TempDir()
	initRepo(t, dir)

	code, out, errOut := runDir(t, dir, []string{"setup", "--dry-run"}, "")
	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout: %q, stderr: %q)", code, out, errOut)
	}

	got, err := os.ReadFile(stale)
	if err != nil {
		t.Fatalf("the stale door is gone: %v", err)
	}
	if string(got) != "a door from an older odachi\n" {
		t.Errorf("--dry-run rewrote the door: %q", got)
	}
	if !strings.Contains(out, "would write") || !strings.Contains(out, "odachi.md") {
		t.Errorf("the report does not name the door it would replace:\n%s", out)
	}
}
