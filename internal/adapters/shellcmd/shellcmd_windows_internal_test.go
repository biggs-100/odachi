//go:build windows

package shellcmd

import (
	"os/exec"
	"testing"
)

func TestKillTreeReportsTaskkillFailure(t *testing.T) {
	orig := taskkillCommand
	taskkillCommand = "taskkill-does-not-exist-odachi"
	defer func() { taskkillCommand = orig }()

	cmd := exec.Command("ping", "-n", "10", "127.0.0.1")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer cmd.Process.Kill()

	err := killTree(cmd)
	if err == nil {
		t.Fatal("killTree should report taskkill failure")
	}
}
