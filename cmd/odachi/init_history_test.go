package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestInitRefusesARepositoryGitCannotRead: a .git that git cannot use is not a
// repository with no commit. The two used to be one answer, and the wrong one:
// the run chose the full install for a project that may already live.
func TestInitRefusesARepositoryGitCannotRead(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}

	code, out, errOut := runDir(t, dir, []string{"init"}, "")

	report := out + errOut
	if code == 0 {
		t.Fatalf("code = 0, want a refusal: a footprint was decided from a repository git cannot read (stdout: %q, stderr: %q)", out, errOut)
	}
	if !strings.Contains(report, "could not read this repository's history") {
		t.Errorf("the refusal does not say what it could not read:\n%s", report)
	}
	for _, rel := range []string{"AGENTS.md", ".gitignore", ".agents"} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err == nil {
			t.Errorf("%s was written before the repository could be read", rel)
		}
	}
}

// TestInitTreatsAFirstDayAsAFirstDay: a repository the human created with no
// commit yet is a state, not a failure. It keeps the full install -it is a
// project starting now- and the refusal above must not swallow it. The agent and
// the craft live in the home, so the project keeps the rules and the hook.
func TestInitTreatsAFirstDayAsAFirstDay(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepo(t, dir)

	code, out, errOut := runDir(t, dir, []string{"init"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout: %q, stderr: %q)", code, out, errOut)
	}
	for _, rel := range []string{"AGENTS.md", ".gitignore", ".git/hooks/pre-commit"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel))); err != nil {
			t.Errorf("%s was not written in a project starting now: %v", rel, err)
		}
	}
	for _, rel := range []string{".agents/skills", ".opencode/agents/odachi.md"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel))); err == nil {
			t.Errorf("%s was written in the project: the agent and the craft live in the home", rel)
		}
	}
	if _, err := os.Stat(filepath.Join(home, filepath.FromSlash(".agents/skills/acceptance/SKILL.md"))); err != nil {
		t.Errorf("the craft did not go to the machine: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, filepath.FromSlash(".config/opencode/agents/odachi.md"))); err != nil {
		t.Errorf("the door did not go to the home: %v", err)
	}
}
