package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/layout"
)

// TestRoleRefusesABlockItCannotRead: the role mirrors the block the project
// carries, and a block odachi could not read is not a project without one. The
// fallback wrote the default block over a file nobody had seen, so the run stops
// and names the file.
func TestRoleRefusesABlockItCannotRead(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	if code, _, stderr := runDir(t, dir, []string{"role", "new", "marketing"}, ""); code != 0 {
		t.Fatalf("role new: exit %d, stderr=%s", code, stderr)
	}
	if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(layout.Roles), "marketing.md"), []byte(roleSource), 0o644); err != nil {
		t.Fatalf("write the source: %v", err)
	}
	// A directory where the block should be: the read fails, and it is not the
	// absence of a file.
	if err := os.MkdirAll(filepath.Join(dir, "AGENTS.md"), 0o755); err != nil {
		t.Fatalf("mkdir AGENTS.md: %v", err)
	}

	code, stdout, stderr := runDir(t, dir, []string{"role", "marketing"}, "")

	if code == 0 {
		t.Fatalf("code = 0, want a refusal: the role was rendered over a block nobody read (stdout: %q)", stdout)
	}
	if !strings.Contains(stderr, "AGENTS.md") {
		t.Errorf("the refusal does not name the file it could not read: %q", stderr)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "marketing", "SKILL.md")); err == nil {
		t.Error("the role was written from a block that could not be read")
	}
}
