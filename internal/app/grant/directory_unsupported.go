//go:build !unix && !windows

package grant

import (
	"context"
	"os"
)

// syncParentDirectory fails closed where the platform offers no way to make a
// directory entry durable. The store refuses to open rather than publish a
// record that a crash could erase or replay.
func syncParentDirectory(context.Context, *os.Root, string) error {
	return ErrDirectorySyncUnsupported
}
