package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestTheDoctorRecognisesAnAdoptedProject is the row end to end: a repository that
// already lives, adopted with a plain init, keeps the rules here and the craft and
// the agent in the machine. The store records where the agent lives, so the
// absence of the mirrors is a choice and not a gap; and a store an older release
// left silent is still the adoption, not a defect.
func TestTheDoctorRecognisesAnAdoptedProject(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	// A machine that has the clients: setup installs the door for the clients it
	// finds, and a home with no config folder is a machine with none.
	for _, dir := range []string{filepath.Join(home, ".config", "opencode"), filepath.Join(home, ".claude")} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
	dir := t.TempDir()
	initRepoWithCommit(t, dir)
	if code, out, errOut := runDir(t, dir, []string{"init"}, ""); code != 0 {
		t.Fatalf("init: exit %d (stdout %q, stderr %q)", code, out, errOut)
	}
	if code, out, errOut := runDir(t, dir, []string{"setup"}, ""); code != 0 {
		t.Fatalf("setup: exit %d (stdout %q, stderr %q)", code, out, errOut)
	}

	_, out, errOut := runDir(t, dir, []string{"doctor"}, "")

	report := out + errOut
	if !strings.Contains(report, "agents.toml keeps no choice") {
		t.Errorf("the doctor does not read the home-only shape:\n%s", report)
	}
	if strings.Contains(report, "does not say so") {
		t.Errorf("the doctor calls the shape a defect:\n%s", report)
	}

	// A project adopted with no store at all: with no command choosing a project
	// agent, the silence is the shape, and the doctor answers it without sending
	// the human to record a choice nobody made.
	_, out, errOut = runDir(t, dir, []string{"doctor"}, "")
	report = out + errOut
	if !strings.Contains(report, "adopted odachi") {
		t.Errorf("the doctor does not recognise the silent adoption:\n%s", report)
	}
	if !strings.Contains(report, "nothing to run") {
		t.Errorf("the doctor does not say the adoption needs nothing:\n%s", report)
	}
	if strings.Contains(report, "run odachi init to record it") {
		t.Errorf("the doctor sends the human to record a key init no longer writes:\n%s", report)
	}
}
