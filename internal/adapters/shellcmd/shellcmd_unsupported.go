//go:build !unix && !windows

package shellcmd

import (
	"context"
	"os/exec"
)

// newCommand refuses on a platform with no known shell. Degrading here would
// mean running the proof command some other way and calling the result a proof;
// the tool says it cannot instead (birth decision 11).
func newCommand(ctx context.Context, dir, command string) (*exec.Cmd, error) {
	return nil, ErrUnsupported
}

// quotePath is unreachable: newCommand refuses on this platform before any
// command is prepared.
func quotePath(path string) string {
	return path
}
