package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSameFileAsksTheFilesystem: the two paths a run compares are a draft and
// brief.md. A string comparison cannot answer "one file", and a path that does not
// exist is not the same file (0186).
func TestSameFileAsksTheFilesystem(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	brief := filepath.Join(dir, "brief.md")
	writeFile(t, dir, "brief.md", "# turnos\n")
	writeFile(t, dir, "other.md", "# otra\n")

	same, err := (osFiles{}).SameFile(brief, brief)
	if err != nil {
		t.Fatalf("SameFile: %v", err)
	}
	if !same {
		t.Errorf("a path is not the same file as itself")
	}

	same, err = (osFiles{}).SameFile(brief, filepath.Join(dir, "other.md"))
	if err != nil {
		t.Fatalf("SameFile other: %v", err)
	}
	if same {
		t.Errorf("two different files answered as one")
	}

	same, err = (osFiles{}).SameFile(brief, filepath.Join(dir, "missing.md"))
	if err != nil {
		t.Fatalf("SameFile missing: %v", err)
	}
	if same {
		t.Errorf("a missing path answered as the same file")
	}
}

// TestSameFileFoldsTheCaseTheDiskFolds: on Windows and on a case-insensitive
// macOS volume, Brief.md names brief.md, and the run must not rewrite the artifact
// it was pointed at. The test skips where the filesystem is case-sensitive (0186).
func TestSameFileFoldsTheCaseTheDiskFolds(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFile(t, dir, "brief.md", "# turnos\n")
	upper := filepath.Join(dir, "BRIEF.md")

	if _, err := os.Stat(upper); err != nil {
		t.Skip("this filesystem does not fold case")
	}

	same, err := (osFiles{}).SameFile(filepath.Join(dir, "brief.md"), upper)
	if err != nil {
		t.Fatalf("SameFile: %v", err)
	}
	if !same {
		t.Errorf("BRIEF.md and brief.md answered as two files on a disk that folds case")
	}
}
