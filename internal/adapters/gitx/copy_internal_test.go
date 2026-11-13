package gitx

import (
	"archive/tar"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestExtractTarRefusesAnEntryThatEscapesTheCopy is the guard the whole row
// rests on: a tar from git carries repository paths, and a name with `..` would
// write outside the copy the row judges. Windows is the case that motivated it:
// `..\` and `../` are the same escape after one normalization each.
func TestExtractTarRefusesAnEntryThatEscapesTheCopy(t *testing.T) {
	dir := t.TempDir()
	tarPath := filepath.Join(t.TempDir(), "malicious.tar")

	file, err := os.Create(tarPath)
	if err != nil {
		t.Fatalf("create tar: %v", err)
	}
	writer := tar.NewWriter(file)
	if err := writer.WriteHeader(&tar.Header{Name: "../escape.txt", Mode: 0o644, Size: 4}); err != nil {
		t.Fatalf("header: %v", err)
	}
	if _, err := writer.Write([]byte("evil")); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close file: %v", err)
	}

	err = extractTar(tarPath, dir)
	if err == nil {
		t.Fatal("an entry that escapes the copy was extracted")
	}
	if !strings.Contains(err.Error(), "escapes") {
		t.Fatalf("the refusal does not name the escape: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(filepath.Dir(dir), "escape.txt")); statErr == nil {
		t.Fatal("the entry was written outside the copy")
	}
}
