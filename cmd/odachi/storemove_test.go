package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestStoreMoveAdoptsTheEvidenceAndNamesTheGitMoves is the acceptance of T5: a
// project that still carries the old folder gets its evidence adopted into the
// clone's store, byte for byte, and the rules are named for git — the tool never
// moves a tracked file.
func TestStoreMoveAdoptsTheEvidenceAndNamesTheGitMoves(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)

	if code, _, stderr := runDir(t, dir, []string{"init"}, ""); code != 0 {
		t.Fatalf("init: exit %d: %s", code, stderr)
	}
	if code, _, stderr := runDir(t, dir, append([]string{"start"}, allFlags()...), ""); code != 0 {
		t.Fatalf("start: exit %d: %s", code, stderr)
	}

	// What an older release left behind.
	const verdict = `{"tree":"abc123","findings":[]}`
	writeFile(t, dir, ".odachi/reviews/abc123.json", verdict)
	writeFile(t, dir, ".odachi/agents.toml", "agents = \"opencode\"\n")
	writeFile(t, dir, ".odachi/fence.toml", "[[rule]]\nnot = \"pagos\"\n")

	code, stdout, stderr := runDir(t, dir, []string{"store", "move"}, "")
	if code != 0 {
		t.Fatalf("store move: exit %d (stdout: %q, stderr: %q)", code, stdout, stderr)
	}

	storeDir := filepath.Join(dir, ".git", "odachi")
	if got, err := os.ReadFile(filepath.Join(storeDir, "reviews", "abc123.json")); err != nil || string(got) != verdict {
		t.Errorf("the verdict was not adopted byte for byte: %q (%v)", got, err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".odachi", "reviews", "abc123.json")); !os.IsNotExist(err) {
		t.Errorf("the old verdict is still there: %v", err)
	}
	if _, err := os.ReadFile(filepath.Join(storeDir, "agents.toml")); err != nil {
		t.Errorf("the client choice was not copied into the store: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".odachi", "agents.toml")); err != nil {
		t.Errorf("the tracked choice was deleted by the tool: %v", err)
	}

	for _, want := range []string{
		"git mv .odachi/fence.toml odd/fence.toml",
		"git rm --cached .odachi/agents.toml",
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout does not carry %q:\n%s", want, stdout)
		}
	}

	// A second run has nothing to adopt: the migration is safe to repeat.
	code, second, stderr := runDir(t, dir, []string{"store", "move"}, "")
	if code != 0 {
		t.Fatalf("second store move: exit %d: %s", code, stderr)
	}
	if strings.Contains(second, "moved 1 file") && strings.Contains(second, "abc123.json ->") {
		t.Errorf("the second run moved the verdict again:\n%s", second)
	}
}

// TestStoreMoveWithoutARepositoryNamesTheFix: there is no store outside git, and
// the refusal says so instead of inventing a location.
func TestStoreMoveWithoutARepositoryNamesTheFix(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	code, _, stderr := runDir(t, dir, []string{"store", "move"}, "")

	if code == 0 {
		t.Fatal("store move outside a repository = 0, want a refusal")
	}
	if !strings.Contains(stderr, "git repository") {
		t.Errorf("stderr %q does not name the fix", stderr)
	}
}
