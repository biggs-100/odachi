package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/persona"
)

// TestGlobalInitLeavesTheProjectAgentsAlone: a global install is a machine decision,
// not a project one. The door lives in the home and runs in every folder, so the
// project keeps only the rules and the plan - and the choice is remembered, so a
// later plain init in the same clone honors it instead of rebuilding the mirrors.
func TestGlobalInitLeavesTheProjectAgentsAlone(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepo(t, dir)

	code, out, errOut := runDir(t, dir, []string{"init", "--global"}, "")
	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout %q, stderr %q)", code, out, errOut)
	}

	if _, err := os.Stat(filepath.Join(dir, ".opencode", "agents")); err == nil {
		t.Errorf("init --global wrote the project's agent files:\n%s", out)
	}
	for _, file := range persona.GlobalFiles() {
		if _, err := os.Stat(filepath.Join(home, filepath.FromSlash(file.Path))); err != nil {
			t.Errorf("%s was not installed in the home: %v", file.Path, err)
		}
	}

	store, err := os.ReadFile(filepath.Join(dir, ".git", "odachi", "agents.toml"))
	if err != nil {
		t.Fatalf("the choice was not stored: %v", err)
	}
	if !strings.Contains(string(store), `project_agents = "none"`) {
		t.Errorf("the store does not remember the home-only choice:\n%s", store)
	}
}

// TestAPlainInitWritesTheProjectAgents: the run without --global is the one that
// means "this project", and it writes the project's own copies. The home stays as it
// was found: the door is opt-in.
func TestAPlainInitWritesTheProjectAgents(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepo(t, dir)

	code, out, errOut := runDir(t, dir, []string{"init"}, "")
	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout %q, stderr %q)", code, out, errOut)
	}

	if _, err := os.Stat(filepath.Join(dir, ".opencode", "agents", "odachi.md")); err != nil {
		t.Errorf("a plain init wrote no project agent: %v", err)
	}
	for _, file := range persona.GlobalFiles() {
		if _, err := os.Stat(filepath.Join(home, filepath.FromSlash(file.Path))); err == nil {
			t.Errorf("%s was installed in the home without --global", file.Path)
		}
	}
}

// TestTheFlagThatAskedForTheDefaultIsGone: --no-project-agents existed to ask for
// what --global now does. A flag that cannot change the outcome is a lie about what
// ran, so it is refused, and the refusal names the run that does the same thing.
func TestTheFlagThatAskedForTheDefaultIsGone(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepo(t, dir)

	code, _, errOut := runDir(t, dir, []string{"init", "--no-project-agents"}, "")
	if code == 0 {
		t.Fatal("the old flag was accepted")
	}
	if !strings.Contains(errOut, "no-project-agents") {
		t.Errorf("the refusal does not name the flag:\n%s", errOut)
	}
}

// TestAgentsAsksForTheMirrorsBack: the store remembers the home-only choice, so a
// plain init honors it; naming the clients is the run that puts the project's copies
// back (0081's rule, kept by this change).
func TestAgentsAsksForTheMirrorsBack(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepo(t, dir)

	if code, _, errOut := runDir(t, dir, []string{"init", "--global"}, ""); code != 0 {
		t.Fatalf("the global install failed: %s", errOut)
	}
	if code, _, errOut := runDir(t, dir, []string{"init"}, ""); code != 0 {
		t.Fatalf("the plain init failed: %s", errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, ".opencode", "agents", "odachi.md")); err == nil {
		t.Error("a plain init rebuilt the mirrors over the stored home-only choice")
	}

	if code, _, errOut := runDir(t, dir, []string{"init", "--agents", "opencode"}, ""); code != 0 {
		t.Fatalf("init --agents failed: %s", errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, ".opencode", "agents", "odachi.md")); err != nil {
		t.Errorf("--agents did not put the mirrors back: %v", err)
	}
}
