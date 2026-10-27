package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPackInstallsTheProjectSkillEndToEnd is the CLI contract for the join: a
// skill created with `odachi skill new` lives in skills/ and the agent reads
// .agents/skills, so pack is what puts it there, next to the bundled ones.
func TestPackInstallsTheProjectSkillEndToEnd(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	if code, _, stderr := runDir(t, dir, []string{"skill", "new", "marketing"}, ""); code != 0 {
		t.Fatalf("skill new: exit %d, stderr=%s", code, stderr)
	}

	code, stdout, stderr := runDir(t, dir, []string{"pack", "--agents", "standard"}, "")
	if code != 0 {
		t.Fatalf("pack: exit %d, stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "marketing") {
		t.Errorf("stdout %q does not name the project skill", stdout)
	}

	files := tree(t, dir)
	for _, path := range []string{
		".agents/skills/marketing/SKILL.md",
		".agents/skills/drive/SKILL.md",
	} {
		if _, ok := files[path]; !ok {
			t.Errorf("%s was not installed", path)
		}
	}
}

// TestPackRefusesABrokenProjectSkill: a catalog that does not pass its check is
// refused by the file a human has to open, instead of installing half a skill.
func TestPackRefusesABrokenProjectSkill(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "skills", "rota", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("no frontmatter\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	code, _, stderr := runDir(t, dir, []string{"pack", "--agents", "standard"}, "")

	if code == 0 {
		t.Fatal("pack accepted a broken project skill")
	}
	if !strings.Contains(stderr, "skills/rota/SKILL.md") {
		t.Errorf("stderr %q does not name the file", stderr)
	}
	if !strings.Contains(stderr, "odachi skills --check") {
		t.Errorf("stderr %q does not name the command that lists the problems", stderr)
	}
	if _, ok := tree(t, dir)[".agents/skills/rota/SKILL.md"]; ok {
		t.Error("the broken skill was installed anyway")
	}
}

// TestPackRefusesACatalogItCannotRead: a catalog that cannot be read is not an
// empty catalog, and installing only the bundled set would hide the project's
// own skills without saying so.
func TestPackRefusesACatalogItCannotRead(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "skills"), []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	code, _, stderr := runDir(t, dir, []string{"pack", "--agents", "standard"}, "")

	if code == 0 {
		t.Fatal("pack installed the bundled set over an unreadable catalog")
	}
	if !strings.Contains(stderr, "skills/") {
		t.Errorf("stderr %q does not name the catalog", stderr)
	}
}
