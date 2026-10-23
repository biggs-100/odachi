package fsx_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/fsx"
)

// TestBackupInPutsTheBackupUnderTheProjectStore is the order this feature
// exists for: the backup is still a file, but it lives in the project's store
// and not beside the artifact the human reads.
func TestBackupInPutsTheBackupUnderTheProjectStore(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "odd", "tasks", "0001-turnos.md")
	content := []byte("# turnos\n")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	backup, err := fsx.BackupIn(root, path)
	if err != nil {
		t.Fatalf("BackupIn: %v", err)
	}

	want := filepath.Join(root, ".odachi", "backups", "odd", "tasks", "0001-turnos.md.bak-"+fsx.Digest(content)[:8])
	if backup != want {
		t.Errorf("BackupIn path = %q, want %q", backup, want)
	}
	got, err := os.ReadFile(backup)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("backup = %q, want %q", got, content)
	}
	if _, err := os.Stat(path + ".bak-" + fsx.Digest(content)[:8]); !os.IsNotExist(err) {
		t.Errorf("BackupIn also left a backup beside the file: %v", err)
	}
}

// TestBackupInLeavesOutsideFilesBesideThem: a file in another tree has no
// project store to join. Saving it beside is noisy but safe; refusing would be
// worse.
func TestBackupInLeavesOutsideFilesBesideThem(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	path := filepath.Join(outside, "SKILL.md")
	content := []byte("# skill\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	backup, err := fsx.BackupIn(root, path)
	if err != nil {
		t.Fatalf("BackupIn: %v", err)
	}

	want := path + ".bak-" + fsx.Digest(content)[:8]
	if backup != want {
		t.Errorf("BackupIn(outside) = %q, want %q", backup, want)
	}
}
