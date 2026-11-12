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
	"strconv"
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
		t.Skipf("the release does not publish %s: this test covers amd64 and arm64", runtime.GOARCH)
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
		t.Skipf("bash is not available on this machine (%v): the installer is a shell script and cannot be tested without a shell", err)
	}

	// Git Bash reads C:/paths; WSL bash lives in another filesystem namespace
	// and sees the script only under /mnt/c. The suite skips here instead of
	// pretending the failure is the installer's.
	if runtime.GOOS == "windows" {
		if out, err := exec.Command(path, "-c", "uname -s").Output(); err == nil && strings.Contains(string(out), "Linux") {
			t.Skip("bash is WSL and does not share the Windows paths; the installer is tested with Git Bash or in CI")
		}
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
	// The bundle is what makes the checksum trustworthy. The fixture does not
	// sign: the shim answers for cosign, so the installer's own logic is what the
	// test measures. The bytes only have to be there.
	if err := os.WriteFile(filepath.Join(releaseDir, "checksums.txt.sigstore.json"), []byte(`{"mediaType":"application/vnd.dev.sigstore.bundle+json;version=0.3"}`), 0o644); err != nil {
		t.Fatalf("write the bundle: %v", err)
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
	defer func() { _ = file.Close() }()

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

// cosignShim answers for cosign: the test measures what the installer does with
// a verification that passes and one that fails, not sigstore's cryptography.
// The path is handed to the script as ODACHI_COSIGN so the lookup does not depend
// on which `cosign` a runner happens to have.
func cosignShim(t *testing.T, exitCode int) string {
	t.Helper()

	dir := t.TempDir()
	script := "#!/bin/sh\nexit " + strconv.Itoa(exitCode) + "\n"
	path := filepath.Join(dir, "cosign")
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write cosign shim: %v", err)
	}
	return path
}

// runInstaller runs the real script against the local release with no network
// available to it, and returns its exit code and its combined output.
func runInstaller(t *testing.T, bash string, fixture releaseFixture, binDir, cosign string) (int, string) {
	return runInstallerWith(t, bash, fixture, binDir, map[string]string{"ODACHI_COSIGN": cosign})
}

// runInstallerWith is runInstaller with the environment the caller needs to
// change, so a test can leave ODACHI_COSIGN unset and reach the branch that
// refuses when cosign is missing.
func runInstallerWith(t *testing.T, bash string, fixture releaseFixture, binDir string, extra map[string]string) (int, string) {
	t.Helper()

	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	// bash is a Unix shell even on Windows: a native path with backslashes
	// arrives as one long word with the separators eaten.
	script := filepath.ToSlash(filepath.Join(repoRoot, "scripts", "install.sh"))

	osName, unameS, unameM := installerTarget(t)
	shim := unameShim(t, unameS, unameM)

	// HOME is redirected so a default install location can never touch the real
	// one, and the working directory is a throwaway: the script's last-resort
	// location is ./odachi, and the test asserts that nothing lands there on a
	// failed verification.
	workDir := t.TempDir()
	home := t.TempDir()

	command := exec.Command(bash, script)
	command.Dir = workDir
	env := map[string]string{
		"ODACHI_VERSION":  installVersion,
		"ODACHI_BASE_URL": fileBaseURL(fixture.dir),
		"ODACHI_BIN_DIR":  binDir,
		// The target platform is passed explicitly instead of relying on the
		// uname shim winning in PATH: on a Windows runner it does not, and the
		// test would then be measuring the refusal path instead of the install.
		"ODACHI_OS":   osName,
		"ODACHI_ARCH": runtime.GOARCH,
		"PATH":        shim + string(os.PathListSeparator) + os.Getenv("PATH"),
		"HOME":        home,
	}
	for key, value := range extra {
		env[key] = value
	}
	command.Env = envWith(env)
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
	t.Parallel()
	bash := requireBash(t)
	fixture := buildRelease(t)
	binDir := filepath.Join(t.TempDir(), "bin")

	code, output := runInstaller(t, bash, fixture, binDir, cosignShim(t, 0))

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
	t.Parallel()
	bash := requireBash(t)
	fixture := buildRelease(t)

	bogus := strings.Repeat("0", 64) + "  " + fixture.asset + "\n"
	if err := os.WriteFile(filepath.Join(fixture.dir, "checksums.txt"), []byte(bogus), 0o644); err != nil {
		t.Fatalf("tamper checksums.txt: %v", err)
	}

	binDir := filepath.Join(t.TempDir(), "bin")
	code, output := runInstaller(t, bash, fixture, binDir, cosignShim(t, 0))

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

// TestInstallScriptRefusesWhenCosignIsMissing: the brief's own criterion -the
// refusal names cosign and how to install it- and the rule that a check which
// runs only sometimes is no check. A machine without cosign installs nothing.
func TestInstallScriptRefusesWhenCosignIsMissing(t *testing.T) {
	t.Parallel()
	bash := requireBash(t)
	if _, err := exec.LookPath("cosign"); err == nil {
		t.Skip("cosign is installed on this machine: the missing-cosign branch cannot be reached")
	}
	fixture := buildRelease(t)

	binDir := filepath.Join(t.TempDir(), "bin")
	code, output := runInstallerWith(t, bash, fixture, binDir, map[string]string{"ODACHI_COSIGN": ""})

	if code == 0 {
		t.Fatalf("installer exit = 0 without cosign, want non-zero:\n%s", output)
	}
	if !strings.Contains(output, "cosign") {
		t.Errorf("output does not name the tool that is missing:\n%s", output)
	}
	if !strings.Contains(output, "sigstore.dev") {
		t.Errorf("output does not say how to get cosign:\n%s", output)
	}
	if entries, err := os.ReadDir(binDir); err == nil && len(entries) > 0 {
		t.Errorf("%s is not empty without cosign: %v", binDir, entries)
	}
}

// TestInstallScriptRefusesASignatureThatDoesNotVerify is the row this feature
// paid: the checksum and the archive come from the same host, so without the
// signature the chain ends there. A failing cosign must stop the install before
// anything is written.
func TestInstallScriptRefusesASignatureThatDoesNotVerify(t *testing.T) {
	t.Parallel()
	bash := requireBash(t)
	fixture := buildRelease(t)

	binDir := filepath.Join(t.TempDir(), "bin")
	code, output := runInstaller(t, bash, fixture, binDir, cosignShim(t, 1))

	if code == 0 {
		t.Fatalf("installer exit = 0 with a signature that does not verify, want non-zero:\n%s", output)
	}
	if !strings.Contains(output, "firma") {
		t.Errorf("output does not say the signature is what failed:\n%s", output)
	}
	if _, err := os.Stat(filepath.Join(binDir, "odachi")); err == nil {
		t.Errorf("%s exists after a failed signature: nothing may be installed", filepath.Join(binDir, "odachi"))
	}
	if entries, err := os.ReadDir(binDir); err == nil && len(entries) > 0 {
		t.Errorf("%s is not empty after a failed signature: %v", binDir, entries)
	}
}
