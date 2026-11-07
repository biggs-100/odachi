//go:build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
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

	// A human can write ODACHI_BIN_DIR with forward slashes, and Windows reports
	// the process's image with backslashes. The comparison has to canonicalize, or
	// the PID disappears exactly when it is needed.
	installDir := strings.ReplaceAll(binDir, `\`, "/")
	code, output := runPowerShellInstaller(t, powershell, fixture, installDir, powershellShim(t, 0))

	// The brief promises exit 1, not "anything but zero": a different code would
	// mean the script failed somewhere else and still printed a message.
	if code != 1 {
		t.Fatalf("install.ps1 exit = %d while %s is running, want 1:\n%s", code, target, output)
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

// TestPowerShellInstallerSaysWhatToDoWhenNothingHoldsTheFile: a sharing violation
// can come from a handle odachi cannot name -an indexer, an antivirus-, and the
// brief still owes the path and the action. The test holds the target with no
// sharing from its own process, whose image is not the target, so the holder
// search comes back empty and the generic sentence is the one under test.
func TestPowerShellInstallerSaysWhatToDoWhenNothingHoldsTheFile(t *testing.T) {
	powershell := requirePowerShell(t)
	fixture := windowsRelease(t)
	binDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("create %s: %v", binDir, err)
	}
	target := filepath.Join(binDir, "odachi.exe")
	placeholder := []byte("a file the installer cannot replace")
	if err := os.WriteFile(target, placeholder, 0o755); err != nil {
		t.Fatalf("write the placeholder: %v", err)
	}

	handle := holdWithoutSharing(t, target)
	code, output := runPowerShellInstaller(t, powershell, fixture, binDir, powershellShim(t, 0))
	syscall.CloseHandle(handle)

	if code != 1 {
		t.Fatalf("install.ps1 exit = %d with an unshareable target, want 1:\n%s", code, output)
	}
	for _, want := range []string{canonicalPath(t, target), "no pude reemplazar"} {
		if !strings.Contains(output, want) {
			t.Errorf("the refusal does not name %q:\n%s", want, output)
		}
	}
	after, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read the target after the failed install: %v", err)
	}
	if string(after) != string(placeholder) {
		t.Errorf("the target changed after a failed install: the placeholder was replaced")
	}
}

// canonicalPath is the name Windows reports for path. The installer compares and
// prints the canonical form, so a runner whose temp folder is short-named
// (RUNNER~1) must not make the test blame the message for its spelling: that is
// the same defect the script had, and the script is the one being fixed.
func canonicalPath(t *testing.T, path string) string {
	t.Helper()

	name, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		t.Fatalf("encode %s: %v", path, err)
	}
	buf := make([]uint16, 32768)
	n, err := syscall.GetLongPathName(name, &buf[0], uint32(len(buf)))
	if err != nil {
		t.Fatalf("canonicalize %s: %v", path, err)
	}
	if n == 0 || int(n) >= len(buf) {
		t.Fatalf("canonicalize %s: the name does not fit its buffer", path)
	}
	return syscall.UTF16ToString(buf[:n])
}

// holdWithoutSharing opens path with no sharing at all, which is what an indexer
// or an antivirus does and what makes a copy fail with no process to name.
func holdWithoutSharing(t *testing.T, path string) syscall.Handle {
	t.Helper()

	name, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		t.Fatalf("encode %s: %v", path, err)
	}
	handle, err := syscall.CreateFile(name, syscall.GENERIC_READ, 0, nil, syscall.OPEN_EXISTING, syscall.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		t.Fatalf("open %s without sharing: %v", path, err)
	}
	return handle
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
