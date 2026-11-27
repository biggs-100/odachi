package fsx_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/fsx"
)

// writeTree creates one file with its parent directories.
func writeTree(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("make %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// TestSnapshotIsStableAndIndependentOfCreationOrder: the digest names content and
// relative paths, not the order in which a filesystem created the entries.
func TestSnapshotIsStableAndIndependentOfCreationOrder(t *testing.T) {
	first := t.TempDir()
	writeTree(t, first, "a/one.txt", "one\n")
	writeTree(t, first, "b/two.txt", "two\n")

	second := t.TempDir()
	writeTree(t, second, "b/two.txt", "two\n")
	writeTree(t, second, "a/one.txt", "one\n")

	firstDigest, err := fsx.Snapshot(first)
	if err != nil {
		t.Fatalf("Snapshot(first): %v", err)
	}
	again, err := fsx.Snapshot(first)
	if err != nil {
		t.Fatalf("Snapshot(first) again: %v", err)
	}
	if firstDigest != again {
		t.Errorf("two reads of one tree differ: %q then %q", firstDigest, again)
	}
	secondDigest, err := fsx.Snapshot(second)
	if err != nil {
		t.Fatalf("Snapshot(second): %v", err)
	}
	if firstDigest != secondDigest {
		t.Errorf("two trees with the same bytes and paths differ: %q and %q", firstDigest, secondDigest)
	}
}

// TestSnapshotChangesWhenAFileChanges: a frozen candidate that moved is no longer
// the same content identity, even when the new bytes have the same length.
func TestSnapshotChangesWhenAFileChanges(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, "internal/x.go", "package x\n")

	before, err := fsx.Snapshot(root)
	if err != nil {
		t.Fatalf("Snapshot(before): %v", err)
	}
	writeTree(t, root, "internal/x.go", "package y\n")
	after, err := fsx.Snapshot(root)
	if err != nil {
		t.Fatalf("Snapshot(after): %v", err)
	}
	if before == after {
		t.Errorf("a changed file kept the digest %q", before)
	}
}

// TestSnapshotChangesWhenAPathChanges: the path is part of the identity.
func TestSnapshotChangesWhenAPathChanges(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, "one.txt", "same\n")
	before, err := fsx.Snapshot(root)
	if err != nil {
		t.Fatalf("Snapshot(before): %v", err)
	}
	if err := os.Rename(filepath.Join(root, "one.txt"), filepath.Join(root, "two.txt")); err != nil {
		t.Fatalf("rename: %v", err)
	}
	after, err := fsx.Snapshot(root)
	if err != nil {
		t.Fatalf("Snapshot(after): %v", err)
	}
	if before == after {
		t.Errorf("a renamed file kept the digest %q", before)
	}
}

// TestSnapshotIncludesAnEmptyDirectory: an empty directory is a fact of the
// frozen tree, so adding one changes the identity.
func TestSnapshotIncludesAnEmptyDirectory(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, "one.txt", "one\n")
	before, err := fsx.Snapshot(root)
	if err != nil {
		t.Fatalf("Snapshot(before): %v", err)
	}
	if err := os.Mkdir(filepath.Join(root, "empty"), 0o755); err != nil {
		t.Fatalf("make the empty directory: %v", err)
	}
	after, err := fsx.Snapshot(root)
	if err != nil {
		t.Fatalf("Snapshot(after): %v", err)
	}
	if before == after {
		t.Errorf("an added empty directory kept the digest %q", before)
	}
}

// TestSnapshotRefusesAMissingRoot: a directory that is not there has no content
// identity, and guessing one would let a missing freeze look unchanged.
func TestSnapshotRefusesAMissingRoot(t *testing.T) {
	if _, err := fsx.Snapshot(filepath.Join(t.TempDir(), "absent")); err == nil {
		t.Fatal("Snapshot() accepted a missing directory")
	}
}
