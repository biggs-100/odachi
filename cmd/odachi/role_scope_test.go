package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/layout"
)

// writeRole leaves one complete role source where `role new` would, so a test can
// render it without writing the eight sections again.
func writeRole(t *testing.T, dir string) {
	t.Helper()
	if code, _, stderr := runDir(t, dir, []string{"role", "new", "marketing"}, ""); code != 0 {
		t.Fatalf("role new: exit %d, stderr=%s", code, stderr)
	}
	writeFile(t, dir, filepath.ToSlash(filepath.Join(layout.Roles, "marketing.md")), roleSource)
}

// TestARoleFollowsTheScopeTheStoreRecorded: a project that keeps the craft in the
// machine gets the role's skill there, and the project keeps none. Before, the
// skill landed in .agents/skills whatever the scope said.
func TestARoleFollowsTheScopeTheStoreRecorded(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepo(t, dir)
	if code, _, errOut := runDir(t, dir, []string{"init", "--skills", "home"}, ""); code != 0 {
		t.Fatalf("init --skills home: %s", errOut)
	}
	writeRole(t, dir)

	code, out, errOut := runDir(t, dir, []string{"role", "marketing"}, "")
	if code != 0 {
		t.Fatalf("role marketing: exit %d (stdout %q, stderr %q)", code, out, errOut)
	}

	if _, err := os.Stat(filepath.Join(home, ".agents", "skills", "marketing", "SKILL.md")); err != nil {
		t.Errorf("the skill did not go to the machine: %v\n%s", err, out+errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "marketing", "SKILL.md")); err == nil {
		t.Errorf("the role wrote a copy into a project whose craft lives in the machine:\n%s", out)
	}
	if !strings.Contains(out, "~/.agents/skills/marketing/SKILL.md") {
		t.Errorf("the run does not name the folder the skill went to:\n%s", out)
	}

	// The listing reads the same destination: the bytes are in the machine, so the
	// role is up to date and not stale.
	code, out, _ = runDir(t, dir, []string{"role"}, "")
	if code != 0 {
		t.Fatalf("role: exit %d", code)
	}
	if !strings.Contains(out, "marketing: up to date") {
		t.Errorf("the listing does not read the machine's copy:\n%s", out)
	}
}

// TestARoleDoesNotFlipTheAdoptionOfARepositoryThatAlreadyLives is the row this
// feature exists for: the role's own copy in .agents/skills was the fact the
// adoption inference read, so one role made the next plain init install the whole
// curriculum in a project that keeps it in the machine.
func TestARoleDoesNotFlipTheAdoptionOfARepositoryThatAlreadyLives(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepoWithCommit(t, dir)
	if code, _, errOut := runDir(t, dir, []string{"init"}, ""); code != 0 {
		t.Fatalf("init: %s", errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills")); err == nil {
		t.Fatalf("the adoption wrote a project copy before the role had anything to do")
	}
	writeRole(t, dir)

	code, out, errOut := runDir(t, dir, []string{"role", "marketing"}, "")
	if code != 0 {
		t.Fatalf("role marketing: exit %d (stdout %q, stderr %q)", code, out, errOut)
	}
	if _, err := os.Stat(filepath.Join(home, ".agents", "skills", "marketing", "SKILL.md")); err != nil {
		t.Errorf("the skill did not go to the machine: %v", err)
	}

	code, out, errOut = runDir(t, dir, []string{"init"}, "")
	if code != 0 {
		t.Fatalf("the plain init after the role: exit %d (stdout %q, stderr %q)", code, out, errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills")); err == nil {
		t.Errorf("one role flipped the adoption: the next plain init installed the whole curriculum\n%s", out)
	}
	if !strings.Contains(out, "this repository already lives") {
		t.Errorf("the adoption stopped being inferred:\n%s", out)
	}
	block, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("reading AGENTS.md: %v", err)
	}
	if !strings.Contains(string(block), "`~/.agents/skills/`") {
		t.Errorf("the block stopped naming the machine's copy:\n%s", block)
	}
}

// TestARoleStaysInTheProjectThatKeepsItsCraft: the other way of the same rule.
// Naming the project copy is a decision, and the role honours it.
func TestARoleStaysInTheProjectThatKeepsItsCraft(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepo(t, dir)
	if code, _, errOut := runDir(t, dir, []string{"init", "--skills", "project"}, ""); code != 0 {
		t.Fatalf("init --skills project: %s", errOut)
	}
	writeRole(t, dir)

	code, out, errOut := runDir(t, dir, []string{"role", "marketing"}, "")
	if code != 0 {
		t.Fatalf("role marketing: exit %d (stdout %q, stderr %q)", code, out, errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "marketing", "SKILL.md")); err != nil {
		t.Errorf("the role did not write the project copy the store chose: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".agents", "skills", "marketing", "SKILL.md")); err == nil {
		t.Errorf("the role wrote into the machine for a project that keeps its own copy")
	}
	if !strings.Contains(out, "./.agents/skills/marketing/SKILL.md") {
		t.Errorf("the run does not name the project's copy:\n%s", out)
	}
}
