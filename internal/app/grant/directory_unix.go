//go:build unix

package grant

import (
	"context"
	"errors"
	"os"
)

// syncParentDirectory flushes the directory entry of name. On unix a directory
// is a file whose fsync makes its entries durable; without it a rename or link
// can still be lost by a crash after publication reported success.
func syncParentDirectory(ctx context.Context, root *os.Root, name string) error {
	directory, err := openParentDirectory(ctx, root, name)
	if err != nil {
		return err
	}
	syncErr := directory.Sync()
	closeErr := directory.Close()
	if err := ctx.Err(); err != nil {
		return errors.Join(syncErr, closeErr, err)
	}
	return errors.Join(syncErr, closeErr)
}
