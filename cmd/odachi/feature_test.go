package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// TestAFeatureCanBeChosenByItsNumber: two features live in one tree, and the
// writer acts on the one named instead of the highest. Without the flag nothing
// changes: the highest is the feature the human is working on now.
func TestAFeatureCanBeChosenByItsNumber(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)
	writeFile(t, dir, "odd/tasks/0002-otro.md", taskdoc.TaskDoc{
		Title:        "otro",
		Requirements: []taskdoc.Requirement{{ID: "R1", Text: "otro"}},
		DoneWhen:     []string{"otro"},
		History:      []string{`L1 · "otro"`},
	}.Render())
	commitAll(t, dir, "two features")

	// The writer with the flag writes the lower one.
	if code, out, errOut := runDir(t, dir, []string{"plan", "un paso", "--feature", "1"}, ""); code != 0 {
		t.Fatalf("plan --feature 1: %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if !strings.Contains(featureText(t, dir, "0001-turnos.md"), "un paso") {
		t.Error("0001 did not get the step")
	}
	if strings.Contains(featureText(t, dir, "0002-otro.md"), "un paso") {
		t.Error("0002 got a step it was not asked for")
	}

	// A question goes to the same one. The documented shape puts the question
	// first and the flags after it.
	if code, out, errOut := runDir(t, dir, []string{"ask", "¿y esto?", "--feature", "1"}, ""); code != 0 {
		t.Fatalf("ask --feature 1: %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if !strings.Contains(featureText(t, dir, "0001-turnos.md"), "¿y esto?") {
		t.Error("0001 did not get the question")
	}

	// No flag keeps today's answer: the highest.
	if code, out, errOut := runDir(t, dir, []string{"plan", "otro paso"}, ""); code != 0 {
		t.Fatalf("plan: %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if !strings.Contains(featureText(t, dir, "0002-otro.md"), "otro paso") {
		t.Error("without the flag the highest did not get the step")
	}

	// A number the tree does not have names the ones it does.
	code, _, errOut := runDir(t, dir, []string{"plan", "x", "--feature", "9"}, "")
	if code != 1 {
		t.Fatalf("plan --feature 9 = %d, want 1", code)
	}
	for _, want := range []string{"0009", "0001-turnos", "0002-otro"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("stderr %q does not name %q", errOut, want)
		}
	}

	// The gate judges the one named: 0001 holds the open question, 0002 does not.
	code, _, errOut = runDir(t, dir, []string{"check", "--feature", "1"}, "")
	if code != 2 {
		t.Fatalf("check --feature 1 = %d, want 2: its own question blocks", code)
	}
	if !strings.Contains(errOut, "¿y esto?") {
		t.Errorf("check --feature 1 does not name its question: %q", errOut)
	}
	code, _, errOut = runDir(t, dir, []string{"check", "--feature", "2"}, "")
	if code != 1 {
		t.Fatalf("check --feature 2 = %d, want 1: its proof is missing", code)
	}
	if strings.Contains(errOut, "¿y esto?") {
		t.Errorf("check --feature 2 judged 0001's question: %q", errOut)
	}

	// The report describes the one named, as data too.
	code, out, errOut := runDir(t, dir, []string{"progress", "--feature", "1", "--json"}, "")
	if code != 0 {
		t.Fatalf("progress --feature 1: %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	var machine struct {
		State struct {
			Number int `json:"number"`
		} `json:"state"`
	}
	if err := json.Unmarshal([]byte(out), &machine); err != nil {
		t.Fatalf("progress --json is not JSON: %v\n%s", err, out)
	}
	if machine.State.Number != 1 {
		t.Errorf("progress --feature 1 reports feature %d, want 1", machine.State.Number)
	}
}

// featureText reads one task document the test wrote.
func featureText(t *testing.T, dir, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "odd", "tasks", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(data)
}
