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

// TestAPlainInitWritesTheHomeDoorAndNoProjectAgent: a plain init is the whole
// job since 0242 -the rules and the hook here, the agent and the craft in the
// machine- and the store records both answers so the next run answers the same.
func TestAPlainInitWritesTheHomeDoorAndNoProjectAgent(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepo(t, dir)

	code, out, errOut := runDir(t, dir, []string{"init"}, "")
	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout %q, stderr %q)", code, out, errOut)
	}

	for _, rel := range []string{".opencode", ".claude"} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err == nil {
			t.Errorf("a plain init wrote %s in the project: the agent lives in the home", rel)
		}
	}
	for _, file := range persona.GlobalFiles() {
		if _, err := os.Stat(filepath.Join(home, filepath.FromSlash(file.Path))); err != nil {
			t.Errorf("%s was not installed in the home: %v", file.Path, err)
		}
	}
	if _, err := os.Stat(filepath.Join(home, filepath.FromSlash(".agents/skills/acceptance/SKILL.md"))); err != nil {
		t.Errorf("the craft was not installed in the home: %v", err)
	}
	agents, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("AGENTS.md: %v", err)
	}
	if !strings.Contains(string(agents), "`~/.agents/skills/`") {
		t.Errorf("the block does not name the machine's copy:\n%s", agents)
	}

	store, err := os.ReadFile(filepath.Join(dir, ".git", "odachi", "agents.toml"))
	if err != nil {
		t.Fatalf("the choice was not stored: %v", err)
	}
	for _, want := range []string{`project_agents = "none"`, `skills = "home"`} {
		if !strings.Contains(string(store), want) {
			t.Errorf("the store does not keep %q:\n%s", want, store)
		}
	}
}

// TestAgentsSelectsTheClientsOfTheDoor: --agents names which clients receive the
// door in the home. It says nothing about the project, which never carries an
// agent file in this version.
func TestAgentsSelectsTheClientsOfTheDoor(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepo(t, dir)

	if code, _, errOut := runDir(t, dir, []string{"init", "--agents", "opencode"}, ""); code != 0 {
		t.Fatalf("init --agents failed: %s", errOut)
	}

	if _, err := os.Stat(filepath.Join(home, filepath.FromSlash(".config/opencode/agents/odachi.md"))); err != nil {
		t.Errorf("the OpenCode door was not installed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, filepath.FromSlash(".claude/output-styles/odachi.md"))); err == nil {
		t.Error("the run installed the Claude door although the choice was opencode")
	}
	if _, err := os.Stat(filepath.Join(dir, ".opencode", "agents", "odachi.md")); err == nil {
		t.Error("--agents put an agent back in the project")
	}
	store, err := os.ReadFile(filepath.Join(dir, ".git", "odachi", "agents.toml"))
	if err != nil {
		t.Fatalf("the client choice was not stored: %v", err)
	}
	if !strings.Contains(string(store), `agents = "opencode"`) {
		t.Errorf("the store does not keep the client choice:\n%s", store)
	}

	// A plain init keeps the stored choice instead of bringing Claude back.
	code, out, errOut := runDir(t, dir, []string{"init"}, "")
	if code != 0 {
		t.Fatalf("the plain init failed: %s", errOut)
	}
	if _, err := os.Stat(filepath.Join(home, filepath.FromSlash(".claude/output-styles/odachi.md"))); err == nil {
		t.Errorf("the Claude door came back on a plain init:\n%s", out)
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
