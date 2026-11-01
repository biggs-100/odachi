package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestReviewWaiveWritesTheDecisionIntoTheOpenFeature: the lane a project with one
// model needs, from the command line, and it lands where the gate reads it.
func TestReviewWaiveWritesTheDecisionIntoTheOpenFeature(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)

	code, out, errOut := runDir(t, dir, []string{"review", "waive", "--because", "the project owns one model"}, "")
	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout: %q, stderr: %q)", code, out, errOut)
	}

	data, err := os.ReadFile(filepath.Join(dir, "odd", "tasks", "0001-turnos.md"))
	if err != nil {
		t.Fatalf("read the task doc: %v", err)
	}
	doc := string(data)
	if !strings.Contains(doc, "Review: waived") {
		t.Errorf("the document does not carry the waiver:\n%s", doc)
	}
	if !strings.Contains(doc, "the project owns one model") {
		t.Errorf("the document does not carry the reason:\n%s", doc)
	}
}

// TestReviewWaiveNeedsAReason: the refusal names the flag, because a waiver without
// one is a decision nobody can read.
func TestReviewWaiveNeedsAReason(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)

	code, _, errOut := runDir(t, dir, []string{"review", "waive"}, "")
	if code != 1 {
		t.Fatalf("code = %d, want 1 (stderr: %q)", code, errOut)
	}
	if !strings.Contains(errOut, "--because") {
		t.Errorf("stderr %q does not name --because", errOut)
	}
}
