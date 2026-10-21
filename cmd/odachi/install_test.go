package main

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// installVersion is the version the release fixture pretends to be. It has to be
// injected with ldflags, because a real release binary gets it that way: the
// test proves the seam between goreleaser and `odachi --version`.
const installVersion = "1.2.3"

// installerTarget is the platform the installer under test is led to believe it
// runs on, plus the uname(1) answers that make it believe so.
//
// scripts/install.sh covers linux and darwin only (Windows has install.ps1, and
// pretending install.sh handles it would be a lie). On a Windows machine the
// shim stands in for the platform the script does support, so the test is still
// the real script, the real tarball, the real sha256 and the real binary.
func installerTarget(t *testing.T) (osName, unameS, unameM string) {
	t.Helper()

	arch := map[string]string{
		"amd64": "x86_64",
		"arm64": "aarch64",
	}[runtime.GOARCH]
	if arch == "" {
		t.Skipf("la release no publica %s: este test cubre amd64 y arm64", runtime.GOARCH)
	}

	switch runtime.GOOS {
	case "linux":
		return "linux", "Linux", arch
	case "darwin":
		return "darwin", "Darwin", arch
	default:
		return "linux", "Linux", arch
	}
}

// requireBash returns the bash to run the installer with, or skips: the
// installer is a shell script and cannot be exercised without a shell.
func requireBash(t *testing.T) string {
	t.Helper()

	path, err := exec.LookPath("bash")
	if err != nil {
		t.Skipf("bash no está disponible en esta máquina (%v): el instalador es un script de shell y no se puede probar sin shell", err)
	}
	return path
}

// unameShim writes a uname(1) that answers like the target platform, and returns
// the directory to put first in PATH.
func unameShim(t *testing.T, unameS, unameM string) string {
	t.Helper()

	dir := t.TempDir()
	script := "#!/bin/sh\n" +
		"case \"$1\" in\n" +
		"  -s) echo " + unameS + " ;;\n" +
		"  -m) echo " + unameM + " ;;\n" +
		"  *)  echo " + unameS + " ;;\n" +
		"esac\n"
	if err := os.WriteFile(filepath.Join(dir, "uname"), []byte(script), 0o755); err != nil {
		t.Fatalf("write uname shim: %v", err)
	}
	return dir
}

// releaseFixture is one fabricated release: the tarball and its checksums.txt,
// exactly as goreleaser would publish them.
type releaseFixture struct {
	dir   string // where the release files live (the ODACHI_BASE_URL root)
	asset string // odachi_<version>_<os>_<arch>.tar.gz
}

// buildRelease builds a real binary with a real version and packs it the way a
// release does. The binary is built, not copied, so `odachi --version` after the
// install proves the whole chain.
func buildRelease(t *testing.T) releaseFixture {
	t.Helper()

	osName, _, _ := installerTarget(t)
	asset := fmt.Sprintf("odachi_%s_%s_%s.tar.gz", installVersion, osName, runtime.GOARCH)

	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}

	buildDir := t.TempDir()
	binary := filepath.Join(buildDir, "odachi")
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
	writeArchive(t, archivePath, "odachi", data)

	// goreleaser hashes the archive, not the binary inside it: that is what
	// checksums.txt describes, and what the installer has to check.
	archive, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("read the fixture archive: %v", err)
	}
	sum := sha256.Sum256(archive)
	checksums := hex.EncodeToString(sum[:]) + "  " + asset + "\n"
	if err := os.WriteFile(filepath.Join(releaseDir, "checksums.txt"), []byte(checksums), 0o644); err != nil {
		t.Fatalf("write checksums.txt: %v", err)
	}

	return releaseFixture{dir: releaseDir, asset: asset}
}

// writeArchive writes a tar.gz holding one executable entry.
func writeArchive(t *testing.T, path, name string, data []byte) {
	t.Helper()

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	defer file.Close()

	gz := gzip.NewWriter(file)
	tw := tar.NewWriter(gz)
	header := &tar.Header{
		Name:     name,
		Mode:     0o755,
		Size:     int64(len(data)),
		Typeflag: tar.TypeReg,
		ModTime:  time.Unix(0, 0),
	}
	if err := tw.WriteHeader(header); err != nil {
		t.Fatalf("tar header: %v", err)
	}
	if _, err := tw.Write(data); err != nil {
		t.Fatalf("tar body: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("close gzip: %v", err)
	}
}

// fileBaseURL turns a directory into a download root the installer can fetch
// from. file:// is what makes the installer testable with no network at all.
func fileBaseURL(dir string) string {
	slash := filepath.ToSlash(dir)
	if runtime.GOOS == "windows" {
		// curl understands file:///C:/... and not the msys /c/... form.
		return "file:///" + strings.TrimPrefix(slash, "/")
	}
	return "file://" + slash
}

// envWith replaces the given keys instead of appending them: a duplicated PATH
// is resolved differently on each platform, and this test cannot afford that.
func envWith(keys map[string]string) []string {
	env := make([]string, 0, len(os.Environ())+len(keys))
	for _, entry := range os.Environ() {
		name, _, ok := strings.Cut(entry, "=")
		if !ok {
			env = append(env, entry)
			continue
		}
		if _, replaced := lookupFold(keys, name); replaced {
			continue
		}
		env = append(env, entry)
	}
	for key, value := range keys {
		env = append(env, key+"="+value)
	}
	return env
}

func lookupFold(keys map[string]string, name string) (string, bool) {
	value, ok := keys[name]
	if ok {
		return value, true
	}
	for key, value := range keys {
		if strings.EqualFold(key, name) {
			return value, true
		}
	}
	return "", false
}

// runInstaller runs the real script against the local release with no network
// available to it, and returns its exit code and its combined output.
func runInstaller(t *testing.T, bash string, fixture releaseFixture, binDir string) (int, string) {
	t.Helper()

	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	script := filepath.Join(repoRoot, "scripts", "install.sh")

	_, unameS, unameM := installerTarget(t)
	shim := unameShim(t, unameS, unameM)

	// HOME is redirected so a default install location can never touch the real
	// one, and the working directory is a throwaway: the script's last-resort
	// location is ./odachi, and the test asserts that nothing lands there on a
	// failed verification.
	workDir := t.TempDir()
	home := t.TempDir()

	command := exec.Command(bash, script)
	command.Dir = workDir
	command.Env = envWith(map[string]string{
		"ODACHI_VERSION":  installVersion,
		"ODACHI_BASE_URL": fileBaseURL(fixture.dir),
		"ODACHI_BIN_DIR":  binDir,
		"PATH":            shim + string(os.PathListSeparator) + os.Getenv("PATH"),
		"HOME":            home,
	})
	out, err := command.CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode(), string(out)
		}
		t.Fatalf("run the installer: %v\n%s", err, out)
	}
	return 0, string(out)
}

// TestInstallScriptInstallsTheVerifiedBinary is the whole point of T9: someone
// without Go installs it, and the binary they end up with says which version it
// is. No network: the release is a local tarball served over file://.
func TestInstallScriptInstallsTheVerifiedBinary(t *testing.T) {
	bash := requireBash(t)
	fixture := buildRelease(t)
	binDir := filepath.Join(t.TempDir(), "bin")

	code, output := runInstaller(t, bash, fixture, binDir)

	if code != 0 {
		t.Fatalf("installer exit = %d, want 0:\n%s", code, output)
	}

	installed := filepath.Join(binDir, "odachi")
	if _, err := os.Stat(installed); err != nil {
		t.Fatalf("the installer did not leave %s: %v\n%s", installed, err, output)
	}
	if !strings.Contains(output, "instalado en") {
		t.Errorf("output does not say where it installed:\n%s", output)
	}

	versionOut, err := installedVersion(bash, installed)
	if err != nil {
		t.Fatalf("run the installed binary: %v\n%s", err, output)
	}
	if got := strings.TrimSpace(versionOut); got != "odachi "+installVersion {
		t.Errorf("installed binary says %q, want %q: the version survived the ldflags and the install", got, "odachi "+installVersion)
	}
}

// installedVersion runs the installed file through the same shell that
// installed it. On Windows the file is a PE binary without the .exe extension
// the loader wants, and bash — the platform install.sh targets — execs it fine.
func installedVersion(bash, binary string) (string, error) {
	command := exec.Command(bash, "-c", shellQuote(filepath.ToSlash(binary))+" --version")
	out, err := command.Output()
	return string(out), err
}

// shellQuote wraps a path so bash reads it literally, spaces and quotes included.
func shellQuote(path string) string {
	return "'" + strings.ReplaceAll(path, "'", `'\''`) + "'"
}

// TestInstallScriptRefusesAModifiedChecksum is the case that matters: a file
// that is not the one the release describes must not reach the machine. Real
// script, real tampered checksums.txt, and the assertion is on the disk: nothing
// installed, not even the last-resort ./odachi.
func TestInstallScriptRefusesAModifiedChecksum(t *testing.T) {
	bash := requireBash(t)
	fixture := buildRelease(t)

	bogus := strings.Repeat("0", 64) + "  " + fixture.asset + "\n"
	if err := os.WriteFile(filepath.Join(fixture.dir, "checksums.txt"), []byte(bogus), 0o644); err != nil {
		t.Fatalf("tamper checksums.txt: %v", err)
	}

	binDir := filepath.Join(t.TempDir(), "bin")
	code, output := runInstaller(t, bash, fixture, binDir)

	if code == 0 {
		t.Fatalf("installer exit = 0 with a modified checksum, want non-zero:\n%s", output)
	}
	if !strings.Contains(output, "checksum") {
		t.Errorf("output does not say the checksum is what failed:\n%s", output)
	}
	if _, err := os.Stat(filepath.Join(binDir, "odachi")); err == nil {
		t.Errorf("%s exists after a failed verification: nothing may be installed", filepath.Join(binDir, "odachi"))
	}
	if entries, err := os.ReadDir(binDir); err == nil && len(entries) > 0 {
		t.Errorf("%s is not empty after a failed verification: %v", binDir, entries)
	}
}
