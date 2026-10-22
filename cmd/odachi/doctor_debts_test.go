package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A linked worktree has .git as a file pointing at the main repository. Telling
// that project "no hay .git, corré git init" is a lie and a dead end: git init is
// already done, in the main repository.
func TestDoctorUnderstandsAWorktree(t *testing.T) {
	dir := t.TempDir()
	pointer := "gitdir: /otro/repo/.git/worktrees/rama\n"
	if err := os.WriteFile(filepath.Join(dir, ".git"), []byte(pointer), 0o644); err != nil {
		t.Fatalf("write .git: %v", err)
	}

	_, stdout, stderr := runDir(t, dir, []string{"doctor"}, "")
	out := stdout + stderr

	if !strings.Contains(out, "worktree") {
		t.Fatalf("el reporte no reconoce el worktree: %q", out)
	}
	if strings.Contains(out, "corré git init") {
		t.Fatalf("el reporte manda a git init en un worktree: %q", out)
	}
}

// A profile whose name nobody shipped and that inherits from nothing resolves to
// a profile with every field empty: it validates and configures nothing. Doctor
// has to call that a problem, because every verb would run with a blank ceremony.
func TestDoctorRefusesAProfileThatInheritsFromNobody(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".odachi"), 0o755); err != nil {
		t.Fatalf("mkdir .odachi: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".odachi", "profile.toml"), []byte("name = \"inventado\"\n"), 0o644); err != nil {
		t.Fatalf("write profile.toml: %v", err)
	}

	code, stdout, stderr := runDir(t, dir, []string{"doctor"}, "")
	out := stdout + stderr

	if code != 1 {
		t.Fatalf("code = %d, want 1: un perfil que no hereda de nadie es un problema", code)
	}
	if !strings.Contains(out, "no hereda de nadie") {
		t.Fatalf("el reporte no dice por qué: %q", out)
	}
	if !strings.Contains(out, "extends") {
		t.Fatalf("el reporte no dice cómo arreglarlo: %q", out)
	}
}

// A profile that inherits from a shipped one is a legitimate configuration, and
// must stay green so the rule only catches the empty ceremony.
func TestDoctorAcceptsAProfileThatInherits(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".odachi"), 0o755); err != nil {
		t.Fatalf("mkdir .odachi: %v", err)
	}
	content := "name = \"mio\"\nextends = \"senior\"\n"
	if err := os.WriteFile(filepath.Join(dir, ".odachi", "profile.toml"), []byte(content), 0o644); err != nil {
		t.Fatalf("write profile.toml: %v", err)
	}

	_, stdout, stderr := runDir(t, dir, []string{"doctor"}, "")
	out := stdout + stderr

	if !strings.Contains(out, "perfil mio") {
		t.Fatalf("el reporte no resuelve el perfil heredado: %q", out)
	}
	if strings.Contains(out, "no hay archivo") {
		t.Fatalf("el reporte dice que no hay archivo cuando sí lo hay: %q", out)
	}
}
