package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFilesBackupUsesTheProjectStore: the adapter finds the project by its
// .odachi boundary, so no caller has to know where the store is.
func TestFilesBackupUsesTheProjectStore(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".odachi"), 0o755); err != nil {
		t.Fatalf("mkdir .odachi: %v", err)
	}
	path := filepath.Join(root, "odd", "tasks", "0001-turnos.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir tasks: %v", err)
	}
	if err := os.WriteFile(path, []byte("# turnos\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	backup, err := (osFiles{}).Backup(path)
	if err != nil {
		t.Fatalf("Backup: %v", err)
	}

	wantPrefix := filepath.Join(root, ".odachi", "backups", "odd", "tasks", "0001-turnos.md.bak-")
	if !strings.HasPrefix(backup, wantPrefix) {
		t.Errorf("backup = %q, want it under %q", backup, wantPrefix)
	}
	if _, err := os.Stat(path + ".bak-"); !os.IsNotExist(err) {
		t.Errorf("a backup was left beside the artifact: %v", err)
	}
}

// TestFilesBackupLeavesOutsideFilesBesideThem: the home's skills are not part
// of this project, so they keep the old shape.
func TestFilesBackupLeavesOutsideFilesBesideThem(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	outside := t.TempDir()
	path := filepath.Join(outside, "SKILL.md")
	if err := os.WriteFile(path, []byte("# skill\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".odachi"), 0o755); err != nil {
		t.Fatalf("mkdir .odachi: %v", err)
	}

	backup, err := (osFiles{}).Backup(path)
	if err != nil {
		t.Fatalf("Backup: %v", err)
	}
	if !strings.HasPrefix(backup, path+".bak-") {
		t.Errorf("backup = %q, want it beside %q", backup, path)
	}
}
