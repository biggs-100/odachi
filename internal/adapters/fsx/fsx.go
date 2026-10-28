// Package fsx is the filesystem edge: content digests, atomic replacement of a
// file the human owns, and backups before the agent rewrites anything. Backups
// go under the clone's store; a file outside a project keeps its backup beside
// it.
package fsx

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Digest returns the SHA-256 of data as lowercase hex. It identifies content,
// never a path, so a proof can be tied to the exact bytes that were tested.
func Digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// WriteAtomic replaces path with data, never leaving a half-written file in its
// place: the bytes go to a temporary file in the same directory (same
// filesystem, so the rename is atomic), are flushed to disk, and only then do
// they take the target name. Any failure removes the temporary file.
func WriteAtomic(path string, data []byte, perm fs.FileMode) error {
	dir := filepath.Dir(path)

	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("fsx: could not create the temporary file next to %s: %w", path, err)
	}
	tmpName := tmp.Name()

	// committed flips only after a successful rename; until then the temporary
	// file is ours to delete.
	committed := false
	defer func() {
		if !committed {
			// Best effort: the original error is the one that matters.
			_ = os.Remove(tmpName)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("fsx: could not write %s: %w", tmpName, err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("fsx: could not sync %s: %w", tmpName, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("fsx: could not close %s: %w", tmpName, err)
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		return fmt.Errorf("fsx: could not apply the permissions to %s: %w", tmpName, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("fsx: could not replace %s: %w", path, err)
	}
	committed = true

	return nil
}

// Backup copies path to "<path>.bak-<first 8 hex of its digest>" and returns
// the new path. The digest names the bytes that were saved, so two backups of
// the same content share a name and a backup can never be mistaken for a
// different version. The original file is only read, never modified.
//
// It is the low-level shape: the backup lands beside the file. Project backups
// go through BackupIn instead, so they do not clutter the project.
func Backup(path string) (string, error) {
	return backupTo(path, path)
}

// BackupIn copies path to store/backups, preserving its path relative to
// project. The two roots are separate because the artifact belongs to the
// project and the version belongs to the clone: a teammate clones the worktree
// and nothing else, so the versions they generate are theirs.
//
// A file outside project has no project to belong to, so it falls back to the
// backup beside it instead of refusing to save a version.
func BackupIn(project, store, path string) (string, error) {
	projectAbs, err := filepath.Abs(project)
	if err != nil {
		return "", fmt.Errorf("fsx: could not resolve the project %s: %w", project, err)
	}
	pathAbs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("fsx: could not resolve %s: %w", path, err)
	}

	rel, err := filepath.Rel(projectAbs, pathAbs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return Backup(path)
	}

	target := filepath.Join(store, "backups", rel)
	return backupTo(path, target)
}

// backupTo writes the bytes of path under base, with the content digest in the
// name. It is the one place that reads, names and writes a backup; the two
// public shapes only decide base.
func backupTo(path, base string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("fsx: could not read %s to back it up: %w", path, err)
	}

	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("fsx: could not read the permissions of %s: %w", path, err)
	}

	backup := base + ".bak-" + Digest(data)[:8]
	if err := os.MkdirAll(filepath.Dir(backup), 0o755); err != nil {
		return "", fmt.Errorf("fsx: could not create the directory for %s: %w", backup, err)
	}
	if err := WriteAtomic(backup, data, info.Mode().Perm()); err != nil {
		return "", err
	}
	return backup, nil
}
