package archtest_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/skills"
)

// TestTheInstalledBinary answers one question the source tree cannot: does the
// binary a person installs carry the same skill bytes the source holds? It asks
// the binary from a directory with no skills and no SKILLS.md, so a build that
// silently reads the checkout cannot pass. The digest travels through stdout;
// the flag needs no project, no argument and no model call.
func TestTheInstalledBinary(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skipf("the Go toolchain is unavailable (%v); this check unblocks when go is on PATH", err)
	}

	want, err := skills.Digest()
	if err != nil {
		t.Fatalf("skills.Digest(): %v", err)
	}

	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolving the repository root: %v", err)
	}
	tmp := t.TempDir()
	binary := filepath.Join(tmp, "odachi")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}

	build := exec.Command("go", "build", "-o", binary, "github.com/biggs-100/odachi/cmd/odachi")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, output)
	}

	// The run directory holds the binary and nothing a source checkout provides.
	// If a SKILL.md were here the binary could be reading it instead of its own
	// embedded payload, and the check would prove nothing.
	if found := skillFilesIn(tmp); len(found) != 0 {
		t.Fatalf("the clean directory holds SKILL.md files: %s", strings.Join(found, ", "))
	}

	run := exec.Command(binary, "skills", "--digest")
	run.Dir = tmp
	stdout, err := run.Output()
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			t.Fatalf("odachi skills --digest exited %d: %s", exit.ExitCode(), exit.Stderr)
		}
		t.Fatalf("running odachi skills --digest: %v", err)
	}
	if got := strings.TrimSpace(string(stdout)); got != want {
		t.Fatalf("installed binary printed %q, want the source digest %q", got, want)
	}
}

// skillFilesIn lists every SKILL.md under dir, to prove the run directory is a
// clean installation and not a checkout disguised as one.
func skillFilesIn(dir string) []string {
	found := []string{}
	_ = filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		if entry.Name() == "SKILL.md" {
			found = append(found, path)
		}
		return nil
	})
	return found
}
