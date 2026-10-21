package main

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/start"
)

// TestVersionSurfacesPrintTheBuildVersion pins the two shapes a human types and
// the default of a plain `go build`: without ldflags the binary says "dev", so
// whoever builds from source never gets a version that is not theirs.
func TestVersionSurfacesPrintTheBuildVersion(t *testing.T) {
	for _, arg := range []string{"--version", "version"} {
		t.Run(arg, func(t *testing.T) {
			dir := t.TempDir()

			code, out, errOut := runDir(t, dir, []string{arg}, "")

			if code != start.CodeOK {
				t.Fatalf("code = %d, want %d (stderr: %s)", code, start.CodeOK, errOut)
			}
			if out != "odachi dev\n" {
				t.Errorf("stdout = %q, want %q: a build without ldflags says dev", out, "odachi dev\n")
			}
			if errOut != "" {
				t.Errorf("stderr = %q, want nothing", errOut)
			}
			if files := tree(t, dir); len(files) != 0 {
				t.Errorf("version wrote files: %v", files)
			}
		})
	}
}

// TestVersionIsThePackageVariableLDflagsSet guards the seam goreleaser uses:
// `-X main.version=<tag>` must land on the variable the two surfaces print.
func TestVersionIsThePackageVariableLDflagsSet(t *testing.T) {
	if version != "dev" {
		t.Fatalf("version = %q, want %q: the test binary is built without ldflags", version, "dev")
	}

	original := version
	t.Cleanup(func() { version = original })
	version = "1.2.3"

	dir := t.TempDir()
	code, out, errOut := runDir(t, dir, []string{"--version"}, "")
	if code != start.CodeOK {
		t.Fatalf("code = %d, want %d (stderr: %s)", code, start.CodeOK, errOut)
	}
	if strings.TrimSpace(out) != "odachi 1.2.3" {
		t.Errorf("stdout = %q, want %q", out, "odachi 1.2.3\n")
	}
}

// TestUnknownCommandStillBlocks keeps the new verb from swallowing the error
// path: a word that is not a command is still blocked, not printed as version.
func TestUnknownCommandStillBlocks(t *testing.T) {
	dir := t.TempDir()

	code, out, errOut := runDir(t, dir, []string{"versions"}, "")

	if code != start.CodeBlocked {
		t.Fatalf("code = %d, want %d (stdout: %s)", code, start.CodeBlocked, out)
	}
	if !strings.Contains(errOut, "comando desconocido") {
		t.Errorf("stderr = %q, want the unknown-command message", errOut)
	}
	if strings.Contains(out, "odachi dev") {
		t.Errorf("stdout = %q: an unknown command must not print the version", out)
	}
}
