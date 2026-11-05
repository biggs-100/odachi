//go:build windows

package main

import "syscall"

// detached is how the canvas child outlives the command: no console to inherit and
// its own process group, so closing the terminal that ran `odachi design open` does
// not take the page with it. The child is stopped by the command, never by a signal.
func detached() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{CreationFlags: 0x00000008 | 0x00000200} // DETACHED_PROCESS | CREATE_NEW_PROCESS_GROUP
}
