package grant

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// ErrDirectorySyncUnsupported is returned where this package cannot make a
// directory entry durable. The store fails closed instead of promising a
// durability it cannot deliver: an unsynced grant record is replayable after a
// crash, which is exactly what the grant contract forbids.
var ErrDirectorySyncUnsupported = errors.New("grant directory sync is unsupported")

// ensureDurableDirectory creates path if needed and makes the new directory
// entries durable, including the parent of every directory it creates. A crash
// after this returns leaves the namespace present, not half-created.
func ensureDurableDirectory(ctx context.Context, path string) error {
	return ensureDurableDirectoryWith(ctx, path, syncDirectoryPath)
}

// ensureDurableDirectoryWith is ensureDurableDirectory with the directory sync
// supplied, so a test can observe the order and prove a platform that refuses the
// sync makes the store refuse to open.
func ensureDurableDirectoryWith(ctx context.Context, path string, syncDirectory func(context.Context, string) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if syncDirectory == nil {
		return fmt.Errorf("%w: directory sync is required", ErrStorageFailure)
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("%w: resolve directory path: %w", ErrStorageFailure, err)
	}

	// Walk up until an existing ancestor proves where creation starts. Every
	// missing level is remembered so it is created ancestor-first.
	var missing []string
	current := absolute
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		info, err := os.Stat(current)
		if err == nil {
			if !info.IsDir() {
				return fmt.Errorf("%w: %q is not a directory", ErrStorageFailure, current)
			}
			break
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%w: stat directory %q: %w", ErrStorageFailure, current, err)
		}
		missing = append(missing, current)
		parent := filepath.Dir(current)
		if parent == current {
			return fmt.Errorf("%w: %q has no existing ancestor", ErrStorageFailure, absolute)
		}
		current = parent
	}

	if len(missing) == 0 {
		if err := syncDirectory(ctx, absolute); err != nil {
			return fmt.Errorf("%w: sync directory %q: %w", ErrStorageFailure, absolute, err)
		}
		return nil
	}
	for i := len(missing) - 1; i >= 0; i-- {
		if err := ctx.Err(); err != nil {
			return err
		}
		directory := missing[i]
		if err := os.Mkdir(directory, storeMode); err != nil && !errors.Is(err, fs.ErrExist) {
			return fmt.Errorf("%w: create directory %q: %w", ErrStorageFailure, directory, err)
		}
		// The directory itself and the parent that now carries its entry are both
		// made durable, or a crash loses the namespace we just reported.
		if err := syncDirectory(ctx, directory); err != nil {
			return fmt.Errorf("%w: sync created directory %q: %w", ErrStorageFailure, directory, err)
		}
		if err := syncDirectory(ctx, filepath.Dir(directory)); err != nil {
			return fmt.Errorf("%w: sync parent of created directory %q: %w", ErrStorageFailure, directory, err)
		}
	}
	return nil
}

// syncDirectoryPath makes one directory's own entries durable by opening it as a
// rooted filesystem and syncing its root.
func syncDirectoryPath(ctx context.Context, path string) error {
	root, err := os.OpenRoot(path)
	if err != nil {
		return fmt.Errorf("%w: open directory: %w", ErrStorageFailure, err)
	}
	defer func() { _ = root.Close() }()
	return syncParentDirectory(ctx, root, ".")
}

// openParentDirectory opens the directory that holds name inside root, so the
// platform-specific step only has to flush a handle it did not have to find.
func openParentDirectory(ctx context.Context, root *os.Root, name string) (*os.File, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	directory, err := root.Open(filepath.Dir(name))
	if err != nil {
		return nil, fmt.Errorf("%w: open parent directory: %w", ErrStorageFailure, err)
	}
	info, err := directory.Stat()
	if err != nil {
		_ = directory.Close()
		return nil, fmt.Errorf("%w: stat parent directory: %w", ErrStorageFailure, err)
	}
	if !info.IsDir() {
		_ = directory.Close()
		return nil, fmt.Errorf("%w: parent path is not a directory", ErrStorageFailure)
	}
	return directory, nil
}
