//go:build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestPowerShellInstallerNamesTheProcessThatHoldsTheBinary: Windows refuses to
// replace an executable another process is running, and the raw Copy-Item error
// names neither the process nor the way out. The installer answers with the PID
// and ODACHI_BIN_DIR, and leaves the target exactly as it was.
func TestPowerShellInstallerNamesTheProcessThatHoldsTheBinary(t *testing.T) {
	powershell := requirePowerShell(t)
	fixture := windowsRelease(t)
	binDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("create %s: %v", binDir, err)
	}

	target := filepath.Join(binDir, "odachi.exe")
	buildSleeper(t, target)
	before, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read the target before the install: %v", err)
	}

	holder := exec.Command(target)
	if err := holder.Start(); err != nil {
		t.Fatalf("start the process that holds %s: %v", target, err)
	}
	t.Cleanup(func() {
		_ = holder.Process.Kill()
		_, _ = holder.Process.Wait()
	})
	// CreateProcess returns after the loader mapped the image: the lock is there.
	// The wait only gives the scheduler the moment it needs on a busy machine.
	time.Sleep(150 * time.Millisecond)

	code, output := runPowerShellInstaller(t, powershell, fixture, binDir, powershellShim(t, 0))

	if code == 0 {
		t.Fatalf("install.ps1 exit = 0 while %s is running, want non-zero:\n%s", target, output)
	}
	for _, want := range []string{"PID " + strconv.Itoa(holder.Process.Pid), "ODACHI_BIN_DIR"} {
		if !strings.Contains(output, want) {
			t.Errorf("the refusal does not name %q:\n%s", want, output)
		}
	}
	after, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read the target after the failed install: %v", err)
	}
	if string(before) != string(after) {
		t.Errorf("the target changed after a failed install: the running binary was replaced or truncated")
	}
}

// buildSleeper writes the smallest program that stays alive and builds it at
// path: a process whose own image is the target, which is what locks it.
func buildSleeper(t *testing.T, path string) {
	t.Helper()

	dir := t.TempDir()
	source := filepath.Join(dir, "main.go")
	body := "package main\n\nimport \"time\"\n\nfunc main() { time.Sleep(2 * time.Minute) }\n"
	if err := os.WriteFile(source, []byte(body), 0o644); err != nil {
		t.Fatalf("write the sleeper: %v", err)
	}
	build := exec.Command("go", "build", "-o", path, "main.go")
	build.Dir = dir
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build the sleeper: %v\n%s", err, out)
	}
}
