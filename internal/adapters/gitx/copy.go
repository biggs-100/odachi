package gitx

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Archive exports a tree into a new directory and returns its path. It is the
// copy the undo row judges: the repository is read, never written, and the
// worktree is not touched. A tree object is not a commit, so a worktree cannot
// hold it; a plain directory can, and the command that runs there needs no git.
//
// The caller owns the directory and must remove it when the row is done.
func Archive(root, tree string) (string, error) {
	tree = strings.TrimSpace(tree)
	if tree == "" {
		return "", fmt.Errorf("archive: no tree to export")
	}

	dir, err := os.MkdirTemp("", "odachi-undo-")
	if err != nil {
		return "", err
	}

	tarPath := filepath.Join(dir, ".odachi-undo.tar")
	if _, err := run(root, "archive", "--format=tar", "-o", tarPath, tree); err != nil {
		_ = os.RemoveAll(dir)
		return "", err
	}

	if err := extractTar(tarPath, dir); err != nil {
		_ = os.RemoveAll(dir)
		return "", err
	}
	_ = os.Remove(tarPath)
	return dir, nil
}

// PatchReverse reverts a diff inside dir, and only for the paths it is given:
// the copy keeps its tests and loses the change.
func PatchReverse(dir, diff string, paths []string) error {
	if strings.TrimSpace(diff) == "" || len(paths) == 0 {
		return nil
	}

	patch, err := os.CreateTemp("", "odachi-undo-*.patch")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(patch.Name()) }()
	if _, err := patch.WriteString(diff); err != nil {
		_ = patch.Close()
		return err
	}
	if err := patch.Close(); err != nil {
		return err
	}

	args := []string{"apply", "-R", "--whitespace=nowarn"}
	for _, path := range paths {
		args = append(args, "--include="+path)
	}
	args = append(args, patch.Name())
	_, err = run(dir, args...)
	return err
}

// extractTar writes every entry under dir. An entry that would escape dir is
// refused: a tar from git carries repository paths, and a copy that writes
// outside the copy is not a copy.
func extractTar(path, dir string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	reader := tar.NewReader(file)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		name := filepath.Clean(filepath.FromSlash(header.Name))
		if name == "." {
			continue
		}
		if filepath.IsAbs(name) || strings.HasPrefix(name, ".."+string(filepath.Separator)) || name == ".." {
			return fmt.Errorf("archive: %q escapes the copy", header.Name)
		}
		target := filepath.Join(dir, name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode)&0o777)
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, reader); err != nil {
				_ = out.Close()
				return err
			}
			if err := out.Close(); err != nil {
				return err
			}
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			if err := os.Symlink(header.Linkname, target); err != nil {
				return err
			}
		}
	}
}

// Discard removes the copy. A directory that is already gone is not an error:
// the row has nothing left to clean.
func Discard(dir string) error {
	if dir == "" {
		return nil
	}
	return os.RemoveAll(dir)
}
