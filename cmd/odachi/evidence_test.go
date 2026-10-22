package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// findIgnores lists every .gitignore under dir, as slash paths relative to it.
func findIgnores(t *testing.T, dir string) []string {
	t.Helper()
	var found []string
	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Name() != ".gitignore" {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		found = append(found, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", dir, err)
	}
	return found
}

// TestInitKeepsTheProofValidAcrossAReview is the whole point of the block: the
// evidence is ignored by the project itself, so taking a review does not move the
// tree, the proof stays valid and the gate does not ask for a retake.
//
// The verdict is attached from OUTSIDE the worktree, so the only thing that can
// move the tree is what the tool writes.
func TestInitKeepsTheProofValidAcrossAReview(t *testing.T) {
	dir := changedProject(t)

	if code, out, errOut := runDir(t, dir, []string{"init"}, ""); code != 0 {
		t.Fatalf("init: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}

	ignore, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf(".gitignore was not written: %v", err)
	}
	for _, want := range []string{"# odachi:begin", "/proof/", "/.odachi/reviews/", "# odachi:end"} {
		if !strings.Contains(string(ignore), want) {
			t.Errorf(".gitignore does not carry %q:\n%s", want, ignore)
		}
	}

	outside := t.TempDir()
	writeFile(t, outside, "veredicto.json", `{"reviewer":"humano","findings":[]}`)
	attach := filepath.Join(outside, "veredicto.json")

	if code, out, errOut := runDir(t, dir, []string{"check", "--run", "echo ok"}, ""); code != 0 {
		t.Fatalf("check --run: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}

	proofPath := filepath.Join(dir, "proof", "0001-turnos.txt")
	before, err := os.ReadFile(proofPath)
	if err != nil {
		t.Fatalf("the proof was not written: %v", err)
	}

	if code, out, errOut := runDir(t, dir, []string{"review", "--attach", attach}, ""); code != 0 {
		t.Fatalf("review --attach: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}

	after, err := os.ReadFile(proofPath)
	if err != nil {
		t.Fatalf("the proof disappeared after the review: %v", err)
	}
	if string(after) != string(before) {
		t.Fatalf("the review rewrote the proof:\n%q\nvs\n%q", after, before)
	}

	code, out, errOut := runDir(t, dir, []string{"check", "--require-review"}, "")
	if code != 0 {
		t.Fatalf("check --require-review: code = %d, want 0 — the proof did not move (stdout: %q, stderr: %q)", code, out, errOut)
	}

	// No retake, and no lazy ignore file either: init's block is the only one.
	if got := findIgnores(t, dir); len(got) != 1 || got[0] != ".gitignore" {
		t.Errorf(".gitignore files = %v, want only the project's own", got)
	}
	if last, err := os.ReadFile(proofPath); err != nil || string(last) != string(before) {
		t.Errorf("the proof changed across the gate: %q vs %q (err %v)", last, before, err)
	}
}

// TestWithoutInitTheFallbackKeepsTheProofValid is the other half: a project that
// never ran init still has its evidence out of the tree, through the interior
// ignore file the verb writes once.
func TestWithoutInitTheFallbackKeepsTheProofValid(t *testing.T) {
	dir := changedProject(t)

	if code, out, errOut := runDir(t, dir, []string{"check", "--run", "echo ok"}, ""); code != 0 {
		t.Fatalf("check --run: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}

	proofPath := filepath.Join(dir, "proof", "0001-turnos.txt")
	before, err := os.ReadFile(proofPath)
	if err != nil {
		t.Fatalf("the proof was not written: %v", err)
	}

	if code, out, errOut := runDir(t, dir, []string{"check", "--run", "echo ok"}, ""); code != 0 {
		t.Fatalf("second check --run: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}

	after, err := os.ReadFile(proofPath)
	if err != nil {
		t.Fatalf("read the proof: %v", err)
	}
	if string(after) != string(before) {
		t.Fatalf("a repeated run rewrote the proof:\n%q\nvs\n%q", after, before)
	}

	ignores := findIgnores(t, dir)
	if len(ignores) != 1 || ignores[0] != "proof/.gitignore" {
		t.Errorf(".gitignore files = %v, want exactly the fallback inside proof/", ignores)
	}
	if _, err := os.Stat(filepath.Join(dir, ".gitignore")); !os.IsNotExist(err) {
		t.Errorf("check created a project .gitignore (err %v), want init to be the only writer of it", err)
	}
}
