package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/schema"
)

// storePath is the file the clone's choices live in.
func storePath(dir string) string {
	return filepath.Join(dir, ".git", "odachi", "agents.toml")
}

// TestAPlainInitKeepsTheScopeTheStoreRecorded is the first acceptance line: a
// project that put the craft in the home must not get .agents/skills rebuilt by
// the next plain init. Before the store carried the key, every upgrade did.
func TestAPlainInitKeepsTheScopeTheStoreRecorded(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepo(t, dir)

	if code, _, errOut := runDir(t, dir, []string{"init", "--skills", "home"}, ""); code != 0 {
		t.Fatalf("the first init failed: %s", errOut)
	}

	code, out, errOut := runDir(t, dir, []string{"init"}, "")
	if code != 0 {
		t.Fatalf("the plain init failed: %d (stdout %q, stderr %q)", code, out, errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills")); err == nil {
		t.Errorf("a plain init rebuilt the project copy the store said not to:\n%s", out)
	}
	if !strings.Contains(out, "from the store") {
		t.Errorf("the output does not say where the scope came from:\n%s", out)
	}
}

// TestTheSkillsFlagPutsTheProjectCopyBack: naming the scope is the way out of the
// home choice, and the store keeps the new answer.
func TestTheSkillsFlagPutsTheProjectCopyBack(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepo(t, dir)

	if code, _, errOut := runDir(t, dir, []string{"init", "--skills", "home"}, ""); code != 0 {
		t.Fatalf("the first init failed: %s", errOut)
	}

	code, out, errOut := runDir(t, dir, []string{"init", "--skills", "project"}, "")
	if code != 0 {
		t.Fatalf("the second init failed: %d (stdout %q, stderr %q)", code, out, errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "acceptance", "SKILL.md")); err != nil {
		t.Errorf("--skills project did not write the project copy: %v", err)
	}
	if !strings.Contains(out, "from --skills") {
		t.Errorf("the output does not say the flag was the source:\n%s", out)
	}

	store, err := os.ReadFile(storePath(dir))
	if err != nil {
		t.Fatalf("the store was not written: %v", err)
	}
	if !strings.Contains(string(store), `skills = "project"`) {
		t.Errorf("the store does not keep the new scope:\n%s", store)
	}
}

// TestAProjectThatNeverChoseKeepsTheProject: absence is not a third value. The
// run answers what the tool always answered and writes no store to say it.
func TestAProjectThatNeverChoseKeepsTheProject(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepo(t, dir)

	code, out, errOut := runDir(t, dir, []string{"init"}, "")
	if code != 0 {
		t.Fatalf("code = %d (stdout %q, stderr %q)", code, out, errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "acceptance", "SKILL.md")); err != nil {
		t.Errorf("a plain init wrote no project copy: %v", err)
	}
	if !strings.Contains(out, "skills: project (the default)") {
		t.Errorf("the output does not name the default:\n%s", out)
	}
	if _, err := os.Stat(storePath(dir)); err == nil {
		t.Errorf("a project that never chose gained a store")
	}
}

// TestTheStoreSaysWhichSchemaWroteIt: the fourth acceptance line. The key arrived
// with a schema, so an installed binary that predates it is told to upgrade
// instead of being told the key is unknown.
func TestTheStoreSaysWhichSchemaWroteIt(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepo(t, dir)

	if code, _, errOut := runDir(t, dir, []string{"init", "--skills", "home"}, ""); code != 0 {
		t.Fatalf("the init failed: %s", errOut)
	}

	store, err := os.ReadFile(storePath(dir))
	if err != nil {
		t.Fatalf("the store was not written: %v", err)
	}
	text := string(store)
	if !strings.Contains(text, `skills = "home"`) {
		t.Errorf("the store does not keep the scope:\n%s", text)
	}
	if !strings.Contains(text, "schema = "+strconv.Itoa(schema.Current)) {
		t.Errorf("the store does not declare schema %d:\n%s", schema.Current, text)
	}
}
