package fsx_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/fsx"
)

func TestDigestIsStableSHA256Hex(t *testing.T) {
	// Well-known SHA-256 of "abc".
	const want = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"

	if got := fsx.Digest([]byte("abc")); got != want {
		t.Errorf("Digest(abc) = %q, want %q", got, want)
	}
	if got := fsx.Digest([]byte("abc")); got != want {
		t.Errorf("Digest is not stable: %q then %q", want, got)
	}
	if fsx.Digest([]byte("abc")) == fsx.Digest([]byte("abd")) {
		t.Errorf("Digest collided for different bytes")
	}
	if got := fsx.Digest(nil); len(got) != 64 {
		t.Errorf("Digest(nil) = %q, want a 64-character hex digest", got)
	}
}

func TestWriteAtomicReplacesContentAndLeavesNoTempFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.txt")

	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	if err := fsx.WriteAtomic(path, []byte("new content"), 0o644); err != nil {
		t.Fatalf("WriteAtomic: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read after WriteAtomic: %v", err)
	}
	if string(data) != "new content" {
		t.Errorf("content = %q, want %q", data, "new content")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	if want := []string{"data.txt"}; !slices.Equal(names, want) {
		t.Errorf("directory after WriteAtomic = %q, want exactly %q (no temp file left)", names, want)
	}
}

func TestWriteAtomicCreatesTheFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nuevo.txt")

	if err := fsx.WriteAtomic(path, []byte("hola"), 0o600); err != nil {
		t.Fatalf("WriteAtomic: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read after WriteAtomic: %v", err)
	}
	if string(data) != "hola" {
		t.Errorf("content = %q, want %q", data, "hola")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "nuevo.txt" {
		t.Errorf("directory after WriteAtomic = %v, want only nuevo.txt", entries)
	}
}

func TestBackupCopiesTheBytesAndKeepsTheOriginal(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "brief.md")
	content := []byte("# demo\n\n## Listo cuando\n- algo\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	backup, err := fsx.Backup(path)
	if err != nil {
		t.Fatalf("Backup: %v", err)
	}

	wantPath := path + ".bak-" + fsx.Digest(content)[:8]
	if backup != wantPath {
		t.Errorf("Backup path = %q, want %q", backup, wantPath)
	}
	backupData, err := os.ReadFile(backup)
	if err != nil {
		t.Fatalf("read backup %s: %v", backup, err)
	}
	if string(backupData) != string(content) {
		t.Errorf("backup content = %q, want %q", backupData, content)
	}
	originalData, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read original after Backup: %v", err)
	}
	if string(originalData) != string(content) {
		t.Errorf("original content changed by Backup: %q", originalData)
	}
}

func TestBackupOfAMissingFileErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "no-existe.md")

	if _, err := fsx.Backup(path); err == nil {
		t.Fatalf("Backup(%s) succeeded, want an error", path)
	}
}
