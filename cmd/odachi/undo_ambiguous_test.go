package main

import (
	"strings"
	"testing"
)

// Bare `undo` must not guess. The newest backup belongs to whichever artifact the
// tool wrote last — inside start that is fence.toml — so "the most recent" is an
// accident of internal write order, never what the human meant. With more than
// one candidate the honest answer is to ask.
func TestUndoWithoutFilterRefusesWhenSeveralArtifactsHaveBackups(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	args := allFlags()[:len(allFlags())-1] // human output, no --json

	if code, _, stderr := runDir(t, dir, args, ""); code != 0 {
		t.Fatalf("start: exit %d, stderr=%s", code, stderr)
	}
	if code, _, stderr := runDir(t, dir, append(args, "--force"), ""); code != 0 {
		t.Fatalf("start --force: exit %d, stderr=%s", code, stderr)
	}

	code, stdout, stderr := runDir(t, dir, []string{"undo"}, "")

	if code == 0 {
		t.Fatalf("undo guessed and restored something: stdout=%q", stdout)
	}
	if !strings.Contains(stderr, "--file") {
		t.Fatalf("the message does not say how to choose: %q", stderr)
	}
	if !strings.Contains(stderr, "fence") || !strings.Contains(stderr, "brief") {
		t.Fatalf("the message does not list the candidates: %q", stderr)
	}
}

// Restoring a backup whose bytes are already the artifact's bytes changes nothing,
// and saying otherwise prints the same backup name as both source and result.
func TestUndoIsANoOpWhenTheCurrentBytesAreAlreadyTheBackup(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	args := allFlags()[:len(allFlags())-1]

	if code, _, stderr := runDir(t, dir, args, ""); code != 0 {
		t.Fatalf("start: exit %d, stderr=%s", code, stderr)
	}
	// Same answers twice: the second run backs up bytes identical to its own
	// output, so the backup and the artifact agree.
	if code, _, stderr := runDir(t, dir, append(args, "--force"), ""); code != 0 {
		t.Fatalf("start --force: exit %d, stderr=%s", code, stderr)
	}

	code, stdout, stderr := runDir(t, dir, []string{"undo", "--file", "brief"}, "")

	if code != 0 {
		t.Fatalf("undo --file brief: exit %d, stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "no changes") {
		t.Fatalf("stdout = %q, want it to say nothing changed", stdout)
	}
}
