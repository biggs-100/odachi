package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The release body has one source: the changelog section for the tag, printed
// verbatim, plus the install block for that tag. A tag with no section is
// refused, because an empty note is a release nobody can read (0180).
func TestTheReleaseBodyComesFromTheChangelog(t *testing.T) {
	bash := requireBash(t)

	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	script := filepath.Join(repoRoot, "scripts", "release-notes.sh")

	dir := t.TempDir()
	changelog := filepath.Join(dir, "CHANGELOG.md")
	fixture := "# Changelog\n\n## v9.9.9 - 2026-01-01\n\n- the thing that changed for a person (0180)\n\n## v9.9.8 - 2025-12-01\n\n- the previous thing\n"
	if err := os.WriteFile(changelog, []byte(fixture), 0o644); err != nil {
		t.Fatalf("write the changelog fixture: %v", err)
	}

	run := func(tag string) (int, string) {
		t.Helper()
		command := exec.Command(bash, script, tag)
		command.Env = envWith(map[string]string{"ODACHI_CHANGELOG": changelog})
		out, err := command.CombinedOutput()
		if err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				return exitErr.ExitCode(), string(out)
			}
			t.Fatalf("run release-notes.sh: %v\n%s", err, out)
		}
		return 0, string(out)
	}

	code, body := run("v9.9.9")
	if code != 0 {
		t.Fatalf("release-notes.sh exit = %d:\n%s", code, body)
	}
	for _, want := range []string{
		"## v9.9.9 - 2026-01-01",
		"the thing that changed for a person (0180)",
		"## Install",
		"v9.9.9/scripts/install.ps1",
		"v9.9.9/scripts/install.sh",
		"cmd/odachi@v9.9.9",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("the body does not carry %q:\n%s", want, body)
		}
	}
	if strings.Contains(body, "the previous thing") {
		t.Errorf("the body carries a section for another tag:\n%s", body)
	}

	code, refusal := run("v9.9.7")
	if code == 0 {
		t.Fatalf("release-notes.sh exit = 0 for a tag with no section:\n%s", refusal)
	}
	if !strings.Contains(refusal, "v9.9.7") {
		t.Errorf("the refusal does not name the tag:\n%s", refusal)
	}
}
