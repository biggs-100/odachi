package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initRepoWithCommit is a repository that already lives: the case the human
// reported, a project with history that is adopting odachi.
func initRepoWithCommit(t *testing.T, dir string) {
	t.Helper()
	initRepo(t, dir)
	writeFile(t, dir, "README.md", "un proyecto que ya vive\n")
	for _, args := range [][]string{
		{"-C", dir, "add", "-A"},
		{"-C", dir, "-c", "user.email=t@example.com", "-c", "user.name=T", "commit", "-q", "-m", "one"},
	} {
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
}

// TestAdoptingAnExistingRepoKeepsTheMinimalFootprint: a project that already
// lives keeps the rules, the ignore and the hook, and none of odachi's folders.
// The craft stays in the machine, and the run says so.
func TestAdoptingAnExistingRepoKeepsTheMinimalFootprint(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepoWithCommit(t, dir)

	code, out, errOut := runDir(t, dir, []string{"init"}, "")
	if code != 0 {
		t.Fatalf("code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}

	for _, rel := range []string{"AGENTS.md", ".gitignore", ".git/hooks/pre-commit"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel))); err != nil {
			t.Errorf("%s was not written: %v", rel, err)
		}
	}
	for _, rel := range []string{".agents", ".opencode", ".claude"} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err == nil {
			t.Errorf("%s was written in a project that already lives", rel)
		}
	}
	report := out + errOut
	if !strings.Contains(report, "this repository already lives") {
		t.Errorf("the run does not say why the footprint is minimal:\n%s", report)
	}
	if _, err := os.Stat(filepath.Join(home, ".agents", "skills", "acceptance", "SKILL.md")); err != nil {
		t.Errorf("the craft did not go to the machine: %v", err)
	}
	// The inference is not a choice: a store that recorded it would turn the next
	// plain init into a decision the human never made.
	store, err := os.ReadFile(storePath(dir))
	if err == nil && strings.Contains(string(store), "skills =") {
		t.Errorf("the run stored an inference:\n%s", store)
	}
}

// TestAFreshRepoKeepsTheFullInstall: a folder with no history is a project
// starting now, and there the craft travels with the repository.
func TestAFreshRepoKeepsTheFullInstall(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepo(t, dir)

	code, out, errOut := runDir(t, dir, []string{"init"}, "")
	if code != 0 {
		t.Fatalf("code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "acceptance", "SKILL.md")); err != nil {
		t.Errorf("a fresh project did not get the craft: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".opencode", "agents", "odachi.md")); err != nil {
		t.Errorf("a fresh project did not get the native agent: %v", err)
	}
	if strings.Contains(out+errOut, "this repository already lives") {
		t.Errorf("a folder with no history was called a project that lives:\n%s%s", out, errOut)
	}
}

// TestAProjectThatAlreadyCarriesTheCraftIsNotMoved: an inference never moves a
// project that adopted odachi the old way. The folder stays the project's, its
// out-of-date skill is refreshed like any init does (with the old bytes set
// aside), and the run does not send the craft to the machine behind the human.
func TestAProjectThatAlreadyCarriesTheCraftIsNotMoved(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepoWithCommit(t, dir)
	writeFile(t, dir, ".agents/skills/acceptance/SKILL.md", "the old copy\n")

	code, out, errOut := runDir(t, dir, []string{"init"}, "")
	if code != 0 {
		t.Fatalf("code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if strings.Contains(out+errOut, "this repository already lives") {
		t.Errorf("a project that already carries the craft was treated as an adoption:\n%s%s", out, errOut)
	}
	skill := filepath.Join(dir, ".agents", "skills", "acceptance", "SKILL.md")
	if _, err := os.Stat(skill); err != nil {
		t.Fatalf("the project copy is gone: %v", err)
	}
	// pack backs a replaced skill up in the clone's store, not beside the file.
	backups, err := filepath.Glob(filepath.Join(dir, ".git", "odachi", "backups", "*"))
	if err != nil || len(backups) == 0 {
		t.Errorf("the old bytes were replaced without a backup in the store: %v (%v)", backups, err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".opencode", "agents", "odachi.md")); err != nil {
		t.Errorf("the project did not get its native agent: %v", err)
	}
}

// TestTheFlagsAskForTheCopiesAndAreStored: the way out of the minimal footprint
// is naming the choice, and the choice outlives the run.
func TestTheFlagsAskForTheCopiesAndAreStored(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepoWithCommit(t, dir)

	code, out, errOut := runDir(t, dir, []string{"init", "--skills", "project", "--agents", "opencode"}, "")
	if code != 0 {
		t.Fatalf("code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "acceptance", "SKILL.md")); err != nil {
		t.Errorf("--skills project did not put the craft here: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".opencode", "agents", "odachi.md")); err != nil {
		t.Errorf("--agents did not put the native agent here: %v", err)
	}
	store, err := os.ReadFile(storePath(dir))
	if err != nil {
		t.Fatalf("the choice was not stored: %v", err)
	}
	if !strings.Contains(string(store), `skills = "project"`) {
		t.Errorf("the store does not keep the choice:\n%s", store)
	}

	// A plain init afterwards honours it instead of going back to the minimal
	// footprint of an adopted repository.
	code, out, errOut = runDir(t, dir, []string{"init"}, "")
	if code != 0 {
		t.Fatalf("the plain init failed: %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if strings.Contains(out+errOut, "this repository already lives") {
		t.Errorf("the plain init ignored the stored choice:\n%s%s", out, errOut)
	}
}
