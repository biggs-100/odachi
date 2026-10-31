package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// unwritableHome gives this test a home that cannot be written on any platform:
// the path is a file, so nothing can be created inside it.
func unwritableHome(t *testing.T) string {
	t.Helper()
	parent := t.TempDir()
	home := filepath.Join(parent, "home")
	writeFile(t, parent, "home", "no soy una carpeta\n")
	withHome(t, home)
	return home
}

// TestInitFailsWhenTheScopeItChoseCannotBeWritten is the row's case: with
// `skills = home` the machine is the only destination, so a home that refuses the
// write leaves a project whose block names a folder nothing created. That run used
// to exit 0 and call it pending.
func TestInitFailsWhenTheScopeItChoseCannotBeWritten(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	unwritableHome(t)

	code, out, errOut := runDir(t, dir, []string{"init", "--skills", "home"}, "")

	report := out + errOut
	if code == 0 {
		t.Fatalf("code = 0, want a refusal: the only destination could not be written\n%s", report)
	}
	if !strings.Contains(report, "could not write") {
		t.Errorf("the refusal does not say what failed:\n%s", report)
	}
	if !strings.Contains(report, "~/.agents/skills") {
		t.Errorf("the refusal does not name the destination:\n%s", report)
	}
	if !strings.Contains(report, "odachi init") {
		t.Errorf("the refusal does not name the command that retries it:\n%s", report)
	}
	if strings.Contains(report, "pending") {
		t.Errorf("a run that failed calls it pending:\n%s", report)
	}
}

// TestSetupDoesNotClaimTheSkillWhenTheWriteFailed: the closing line of a setup
// that installed nothing used to say the bootstrap skill was installed. This
// fixture reaches that branch -the home is a folder, so no client is found, and
// `.agents` is a file, so the skills write fails- and the line must not claim it.
func TestSetupDoesNotClaimTheSkillWhenTheWriteFailed(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	writeFile(t, home, ".agents", "no soy una carpeta\n")
	dir := t.TempDir()

	code, out, errOut := runDir(t, dir, []string{"setup"}, "")

	report := out + errOut
	if code == 0 {
		t.Fatalf("code = 0, want a refusal: setup wrote no skill into the home\n%s", report)
	}
	if !strings.Contains(report, "no agent client found in the home") {
		t.Fatalf("the fixture does not reach the line under test:\n%s", report)
	}
	if strings.Contains(report, "the bootstrap skill is installed") {
		t.Errorf("the run claims the skill was installed while the write failed:\n%s", report)
	}
}

// TestSetupFailsWhenTheHomeRefusesTheWrite: setup's only destination is the home,
// and its refusal has to name setup. It used to exit 0 with the work left pending.
func TestSetupFailsWhenTheHomeRefusesTheWrite(t *testing.T) {
	unwritableHome(t)
	dir := t.TempDir()

	code, out, errOut := runDir(t, dir, []string{"setup"}, "")

	report := out + errOut
	if code == 0 {
		t.Fatalf("code = 0, want a refusal: setup wrote nothing into the home\n%s", report)
	}
	if !strings.Contains(report, "odachi setup") {
		t.Errorf("the refusal does not name the verb that ran:\n%s", report)
	}
	if strings.Contains(report, "odachi init") {
		t.Errorf("a failed setup tells the human to run init:\n%s", report)
	}
}
