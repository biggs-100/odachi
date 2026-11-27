//go:build windows

package grant

import (
	"context"
	"errors"
	"fmt"
	"os"
	"syscall"
)

// syncParentDirectory flushes the directory entry of name. Windows cannot fsync
// a handle opened through os.Root (it is opened read-only), so the directory is
// reopened with FILE_FLAG_BACKUP_SEMANTICS and flushed by handle. The path comes
// from the already resolved root handle; the namespace is private to the clone,
// so the reopen is not a place another process can redirect.
func syncParentDirectory(ctx context.Context, root *os.Root, name string) error {
	directory, err := openParentDirectory(ctx, root, name)
	if err != nil {
		return err
	}
	path := directory.Name()
	if err := directory.Close(); err != nil {
		return fmt.Errorf("%w: close parent directory: %w", ErrStorageFailure, err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := flushDirectory(path); err != nil {
		return fmt.Errorf("%w: flush parent directory %q: %w", ErrStorageFailure, path, err)
	}
	return nil
}

// flushDirectory flushes one directory's metadata by handle.
func flushDirectory(path string) error {
	pointer, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	handle, err := syscall.CreateFile(
		pointer,
		syscall.GENERIC_READ|syscall.GENERIC_WRITE,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
	if err != nil {
		return err
	}
	flushErr := syscall.FlushFileBuffers(handle)
	closeErr := syscall.CloseHandle(handle)
	return errors.Join(flushErr, closeErr)
}
