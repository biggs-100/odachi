package fsx

import (
	"cmp"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

// Snapshot returns the deterministic SHA-256 content identity of a frozen
// directory: every relative path, every regular file's bytes, and every symlink's
// target, in one order on every platform. It reads types and bytes, never
// modification times or permission bits, so two copies of the same tree answer
// the same digest and a changed or moved file answers a different one.
//
// It is the comparison the managed verifier makes before and after a run: a
// digest that changed means the candidate moved under the verifier and nothing
// may be sealed. A missing root or an entry that is not a regular file, a
// directory or a symlink is a refusal, never an empty identity.
func Snapshot(root string) (string, error) {
	info, err := os.Lstat(root)
	if err != nil {
		return "", fmt.Errorf("fsx: could not read the snapshot root %s: %w", root, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("fsx: the snapshot root %s is not a directory", root)
	}

	type entry struct {
		path string
		rel  string
		mode fs.FileMode
	}
	var entries []entry
	err = filepath.WalkDir(root, func(path string, item fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		entries = append(entries, entry{path: path, rel: filepath.ToSlash(rel), mode: item.Type()})
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("fsx: could not walk the snapshot root %s: %w", root, err)
	}
	slices.SortFunc(entries, func(a, b entry) int { return cmp.Compare(a.rel, b.rel) })

	hasher := sha256.New()
	for _, item := range entries {
		switch {
		case item.mode.IsDir():
			writeSnapshotField(hasher, "directory")
			writeSnapshotField(hasher, item.rel)
		case item.mode&fs.ModeSymlink != 0:
			target, err := os.Readlink(item.path)
			if err != nil {
				return "", fmt.Errorf("fsx: could not read the link %s: %w", item.path, err)
			}
			writeSnapshotField(hasher, "link")
			writeSnapshotField(hasher, item.rel)
			writeSnapshotField(hasher, target)
		case item.mode.IsRegular():
			data, err := os.ReadFile(item.path)
			if err != nil {
				return "", fmt.Errorf("fsx: could not read the snapshot file %s: %w", item.path, err)
			}
			writeSnapshotField(hasher, "file")
			writeSnapshotField(hasher, item.rel)
			var size [8]byte
			binary.BigEndian.PutUint64(size[:], uint64(len(data)))
			_, _ = hasher.Write(size[:])
			_, _ = hasher.Write(data)
		default:
			return "", fmt.Errorf("fsx: the snapshot entry %s is not a regular file, a directory or a symlink", item.path)
		}
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// writeSnapshotField writes one length-prefixed field, so a path that contains
// the separator cannot collide with a different split of the same bytes.
func writeSnapshotField(writer io.Writer, value string) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	_, _ = writer.Write(length[:])
	_, _ = writer.Write([]byte(value))
}
