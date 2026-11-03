//go:build windows

package main

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// requirePowerShell is the Windows side of requireBash: the installer under test
// is a PowerShell script and cannot run without a shell.
func requirePowerShell(t *testing.T) string {
	t.Helper()

	for _, name := range []string{"powershell", "pwsh"} {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}
	t.Skip("PowerShell is not available on this machine: install.ps1 cannot be exercised")
	return ""
}

// windowsRelease builds what goreleaser publishes for this platform: a zip with
// odachi.exe inside, the checksums that describe it, and the sigstore bundle the
// installer verifies before it reads a single line of the checksums.
func windowsRelease(t *testing.T) releaseFixture {
	t.Helper()

	asset := "odachi_" + installVersion + "_windows_" + runtime.GOARCH + ".zip"
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}

	buildDir := t.TempDir()
	binary := filepath.Join(buildDir, "odachi.exe")
	build := exec.Command("go", "build", "-trimpath",
		"-ldflags", "-X main.version="+installVersion,
		"-o", binary, "./cmd/odachi")
	build.Dir = repoRoot
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build the fixture binary: %v\n%s", err, out)
	}
	data, err := os.ReadFile(binary)
	if err != nil {
		t.Fatalf("read the fixture binary: %v", err)
	}

	releaseDir := t.TempDir()
	archivePath := filepath.Join(releaseDir, asset)
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("create %s: %v", asset, err)
	}
	zw := zip.NewWriter(file)
	entry, err := zw.Create("odachi.exe")
	if err != nil {
		t.Fatalf("zip entry: %v", err)
	}
	if _, err := entry.Write(data); err != nil {
		t.Fatalf("zip body: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close %s: %v", asset, err)
	}

	archive, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("read the fixture archive: %v", err)
	}
	sum := sha256.Sum256(archive)
	checksums := hex.EncodeToString(sum[:]) + "  " + asset + "\r\n"
	if err := os.WriteFile(filepath.Join(releaseDir, "checksums.txt"), []byte(checksums), 0o644); err != nil {
		t.Fatalf("write checksums.txt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(releaseDir, "checksums.txt.sigstore.json"), []byte(`{"mediaType":"application/vnd.dev.sigstore.bundle+json;version=0.3"}`), 0o644); err != nil {
		t.Fatalf("write the bundle: %v", err)
	}

	return releaseFixture{dir: releaseDir, asset: asset}
}

// powershellShim answers for cosign: what the test measures is the installer's
// decision, not sigstore's cryptography. It is a .cmd on purpose: cosign is a
// native program and writes "Verified OK" to stderr, and PowerShell 5.1 turns a
// native command's stderr into a terminating error under $ErrorActionPreference
// 'Stop'. A PowerShell shim would not reproduce that; a batch file does.
func powershellShim(t *testing.T, exitCode int) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "cosign.cmd")
	message := "Verified OK"
	if exitCode != 0 {
		message = "signature verification failed"
	}
	script := "@echo " + message + " 1>&2\r\n@exit /b " + strconv.Itoa(exitCode) + "\r\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write the cosign shim: %v", err)
	}
	return path
}

// runPowerShellInstaller runs the real script against the local release, with no
// network available to it, and returns its exit code and its output.
func runPowerShellInstaller(t *testing.T, powershell string, fixture releaseFixture, binDir, cosign string) (int, string) {
	t.Helper()

	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	script := filepath.Join(repoRoot, "scripts", "install.ps1")

	command := exec.Command(powershell, "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", script)
	command.Dir = t.TempDir()
	command.Env = envWith(map[string]string{
		"ODACHI_VERSION":  installVersion,
		"ODACHI_BASE_URL": fileBaseURL(fixture.dir),
		"ODACHI_BIN_DIR":  binDir,
		"ODACHI_COSIGN":   cosign,
	})
	out, err := command.CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode(), string(out)
		}
		t.Fatalf("run install.ps1: %v\n%s", err, out)
	}
	return 0, string(out)
}

// TestPowerShellInstallerRefusesASignatureThatDoesNotVerify is the Windows half
// of the row: a checksum that arrived with the archive proves nothing until the
// signature that names this workflow and this tag verifies.
func TestPowerShellInstallerRefusesASignatureThatDoesNotVerify(t *testing.T) {
	powershell := requirePowerShell(t)
	fixture := windowsRelease(t)
	binDir := filepath.Join(t.TempDir(), "bin")

	code, output := runPowerShellInstaller(t, powershell, fixture, binDir, powershellShim(t, 1))

	if code == 0 {
		t.Fatalf("install.ps1 exit = 0 with a signature that does not verify, want non-zero:\n%s", output)
	}
	if !strings.Contains(output, "firma") {
		t.Errorf("output does not say the signature is what failed:\n%s", output)
	}
	if _, err := os.Stat(filepath.Join(binDir, "odachi.exe")); err == nil {
		t.Errorf("odachi.exe exists after a failed signature: nothing may be installed")
	}
	if entries, err := os.ReadDir(binDir); err == nil && len(entries) > 0 {
		t.Errorf("%s is not empty after a failed signature: %v", binDir, entries)
	}
}

// TestPowerShellInstallerRefusesWhenCosignIsMissing pins the brief's own criterion:
// a check that runs only sometimes is a silent degradation, so the installer refuses
// when the tool it needs is not there, and says which tool and where to get it.
func TestPowerShellInstallerRefusesWhenCosignIsMissing(t *testing.T) {
	powershell := requirePowerShell(t)
	fixture := windowsRelease(t)
	binDir := filepath.Join(t.TempDir(), "bin")
	// A path that exists nowhere: Get-Command and Test-Path both miss it, so the
	// installer faces the same absence a machine without cosign has. Naming the
	// path is the seam the other cases already use, and it keeps this test from
	// depending on whether cosign happens to be on the PATH of whoever runs it.
	missing := filepath.Join(t.TempDir(), "no-cosign-here.exe")

	code, output := runPowerShellInstaller(t, powershell, fixture, binDir, missing)

	if code == 0 {
		t.Fatalf("install.ps1 exit = 0 with no cosign, want non-zero:\n%s", output)
	}
	for _, want := range []string{"cosign", "https://docs.sigstore.dev/cosign"} {
		if !strings.Contains(output, want) {
			t.Errorf("the refusal does not name %q:\n%s", want, output)
		}
	}
	if entries, err := os.ReadDir(binDir); err == nil && len(entries) > 0 {
		t.Errorf("%s is not empty after a missing cosign: %v", binDir, entries)
	}
}

// TestPowerShellInstallerInstallsTheVerifiedBinary: someone without Go installs
// it, and the binary says which version it is.
func TestPowerShellInstallerInstallsTheVerifiedBinary(t *testing.T) {
	powershell := requirePowerShell(t)
	fixture := windowsRelease(t)
	binDir := filepath.Join(t.TempDir(), "bin")

	code, output := runPowerShellInstaller(t, powershell, fixture, binDir, powershellShim(t, 0))

	if code != 0 {
		t.Fatalf("install.ps1 exit = %d, want 0:\n%s", code, output)
	}
	installed := filepath.Join(binDir, "odachi.exe")
	if _, err := os.Stat(installed); err != nil {
		t.Fatalf("the installer did not leave %s: %v\n%s", installed, err, output)
	}
	versionOut, err := exec.Command(installed, "--version").Output()
	if err != nil {
		t.Fatalf("run the installed binary: %v", err)
	}
	if got := strings.TrimSpace(string(versionOut)); got != "odachi "+installVersion {
		t.Errorf("installed binary says %q, want %q", got, "odachi "+installVersion)
	}
}
