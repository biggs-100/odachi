package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// TestStartLightRefusesASecondLine: the lane is one line, so the second positional
// argument is refused before anything is read from disk, and the answer names the
// line it dropped.
func TestStartLightRefusesASecondLine(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	code, stdout, stderr := runDir(t, dir, []string{"start", "--light", "uno", "dos"}, "")
	if code != 1 {
		t.Fatalf("start --light with two lines: exit %d, want 1 (stdout=%s stderr=%s)", code, stdout, stderr)
	}
	if !strings.Contains(stderr, "extra argument: dos") {
		t.Errorf("stderr %q does not name the second line", stderr)
	}
	if _, err := os.Stat(filepath.Join(dir, "odd")); err == nil {
		t.Error("a refused run wrote odd/, want nothing")
	}
}

// TestStartLightTakesALineThatStartsWithADash is 0270: a requirement may name a
// flag, and Go's own separator is how it reaches the verb. Without this, `un --runs
// 0 se rechaza` -a requirement we wrote twice in one day- could not be opened at all.
func TestStartLightTakesALineThatStartsWithADash(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepoWithCommit(t, dir)

	const line = "--runs 0 se rechaza"
	code, _, stderr := runDir(t, dir, []string{"start", "--light", "--", line}, "")
	if code != 0 {
		t.Fatalf("start --light -- %q: exit %d, stderr=%s", line, code, stderr)
	}

	entries, err := os.ReadDir(filepath.Join(dir, "odd", "tasks"))
	if err != nil {
		t.Fatalf("read the task documents: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("task documents = %d, want one", len(entries))
	}
	data, err := os.ReadFile(filepath.Join(dir, "odd", "tasks", entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	doc := taskdoc.Parse(string(data))
	if len(doc.Requirements) != 1 || doc.Requirements[0].Text != line {
		t.Errorf("Requirements = %+v, want the line verbatim", doc.Requirements)
	}
	if len(doc.Steps) != 1 || doc.Steps[0].Text != line {
		t.Errorf("Steps = %+v, want one step with the same words", doc.Steps)
	}
}

// TestStartLightNamesTheSeparatorWhenTheLineStartsWithADash: the refusal a human
// meets first must say what to type, not what the flag package thought.
func TestStartLightNamesTheSeparatorWhenTheLineStartsWithADash(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	code, _, stderr := runDir(t, dir, []string{"start", "--light", "--runs 0 se rechaza"}, "")
	if code != 1 {
		t.Fatalf("start --light without the separator: exit %d, want 1", code)
	}
	if !strings.Contains(stderr, "--light -- ") {
		t.Errorf("stderr %q does not name the form that works", stderr)
	}
	if _, err := os.Stat(filepath.Join(dir, "odd")); err == nil {
		t.Error("a refused run wrote odd/, want nothing")
	}
}

// TestStartLightEndToEnd: one line in, one document out, and no brief beside it.
func TestStartLightEndToEnd(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepoWithCommit(t, dir)

	const line = "el saludo dice la versión"
	code, stdout, stderr := runDir(t, dir, []string{"start", "--light", line}, "")
	if code != 0 {
		t.Fatalf("start --light: exit %d, stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "light feature is one line") {
		t.Errorf("stdout %q does not say what it opened", stdout)
	}

	if _, err := os.Stat(filepath.Join(dir, "brief.md")); err == nil {
		t.Error("a light run wrote a brief")
	}

	entries, err := os.ReadDir(filepath.Join(dir, "odd", "tasks"))
	if err != nil {
		t.Fatalf("read the task documents: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("task documents = %d, want one", len(entries))
	}
	data, err := os.ReadFile(filepath.Join(dir, "odd", "tasks", entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	doc := taskdoc.Parse(string(data))
	if !doc.Light {
		t.Error("the document does not carry the lane")
	}
	if len(doc.Requirements) != 1 || doc.Requirements[0].Text != line {
		t.Errorf("Requirements = %+v, want the line verbatim", doc.Requirements)
	}
	if len(doc.Steps) != 1 || doc.Steps[0].Text != line || doc.Steps[0].Requirement != "R1" {
		t.Errorf("Steps = %+v, want one step with the same words", doc.Steps)
	}
	if doc.Base == "" {
		t.Error("Base is empty: the closing gate would not know what to review")
	}

	// The lane is the line: a run without one is refused, and names the call.
	code, _, stderr = runDir(t, dir, []string{"start", "--light"}, "")
	if code != 1 {
		t.Fatalf("start --light without a line: exit %d, want 1", code)
	}
	if !strings.Contains(stderr, "--light") {
		t.Errorf("stderr %q does not name the call that works", stderr)
	}
}
