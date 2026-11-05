//go:build !windows

package main

import "syscall"

// detached is how the canvas child outlives the command: a new session, so closing
// the terminal that ran `odachi design open` does not take the page with it. The
// child is stopped by the command, never by a signal.
func detached() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}
