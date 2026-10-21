package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/biggz/odachi/internal/domain/taskdoc"
)

// startedProject runs `odachi start` once and returns the directory plus the
// task document path it wrote.
func startedProject(t *testing.T) (dir, taskPath string) {
	t.Helper()
	dir = t.TempDir()
	if code, _, stderr := runDir(t, dir, allFlags(), ""); code != 0 {
		t.Fatalf("start: exit %d, stderr=%s", code, stderr)
	}
	taskPath = onlyTask(t, tree(t, dir))
	return dir, taskPath
}

// historyOf reads the task document and returns its history lines.
func historyOf(t *testing.T, dir, taskPath string) []string {
	t.Helper()
	files := tree(t, dir)
	content, ok := files[taskPath]
	if !ok {
		t.Fatalf("%s disappeared; files: %v", taskPath, files)
	}
	return taskdoc.Parse(content).History
}

// TestAskRecordsTheThreeShapesEndToEnd is the CLI contract for T6: the agent's
// question, the human's answer and the agent's own decision all live in the
// same append-only log, in that order.
func TestAskRecordsTheThreeShapesEndToEnd(t *testing.T) {
	dir, taskPath := startedProject(t)

	steps := []struct {
		args []string
		want string
	}{
		{[]string{"ask", "¿Los turnos se pueden cancelar?"}, `L2 · PREGUNTA P1 · "¿Los turnos se pueden cancelar?"`},
		{[]string{"ask", "--answer", "No en la primera versión"}, `L3 · VOS P1 · "No en la primera versión"`},
		{[]string{"ask", "--craft", "Guardé los turnos en SQLite", "--because", "dos personas no pueden escribir el mismo archivo"}, `L4 · OFICIO · "Guardé los turnos en SQLite" (porque dos personas no pueden escribir el mismo archivo)`},
	}

	for _, step := range steps {
		code, stdout, stderr := runDir(t, dir, step.args, "")
		if code != 0 {
			t.Fatalf("%v: exit %d, stderr=%s", step.args, code, stderr)
		}
		if !strings.Contains(stdout, taskPath) {
			t.Errorf("%v: stdout %q does not name the task document", step.args, stdout)
		}

		history := historyOf(t, dir, taskPath)
		if got := history[len(history)-1]; got != step.want {
			t.Fatalf("%v: appended line = %q, want %q", step.args, got, step.want)
		}
	}

	if history := historyOf(t, dir, taskPath); len(history) != 4 {
		t.Errorf("history has %d lines, want 4: %v", len(history), history)
	}
}

// TestAskBlocksWithoutATaskDocument: a project with no feature cannot record
// anything, and the message names `odachi start`.
func TestAskBlocksWithoutATaskDocument(t *testing.T) {
	dir := t.TempDir()

	code, stdout, stderr := runDir(t, dir, []string{"ask", "¿Y esto?"}, "")

	if code != 1 {
		t.Fatalf("exit = %d, want 1 (stdout=%s stderr=%s)", code, stdout, stderr)
	}
	if !strings.Contains(stderr, "odachi start") {
		t.Errorf("stderr %q does not name odachi start", stderr)
	}
	if files := tree(t, dir); len(files) != 0 {
		t.Errorf("a blocked ask wrote files: %v", files)
	}
}

// TestAskCraftWithoutBecauseIsBlocked keeps the tutor rule at the CLI: every
// blocked message names the flag that unblocks it.
func TestAskCraftWithoutBecauseIsBlocked(t *testing.T) {
	dir, _ := startedProject(t)

	code, _, stderr := runDir(t, dir, []string{"ask", "--craft", "usé SQLite"}, "")

	if code != 1 {
		t.Fatalf("exit = %d, want 1 (stderr=%s)", code, stderr)
	}
	if !strings.Contains(stderr, "--because") {
		t.Errorf("stderr %q does not name --because", stderr)
	}
}

// TestAskJSONReportsTheTaskDocument keeps the machine output shape.
func TestAskJSONReportsTheTaskDocument(t *testing.T) {
	dir, taskPath := startedProject(t)

	code, stdout, stderr := runDir(t, dir, []string{"ask", "¿Y esto?", "--json"}, "")
	if code != 0 {
		t.Fatalf("exit = %d, stderr=%s", code, stderr)
	}

	var payload struct {
		Code    int      `json:"code"`
		Message string   `json:"message"`
		Task    string   `json:"task"`
		Files   []string `json:"files"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, stdout)
	}
	if payload.Code != 0 {
		t.Errorf("code = %d, want 0", payload.Code)
	}
	if payload.Task != taskPath {
		t.Errorf("task = %q, want %q", payload.Task, taskPath)
	}
	if len(payload.Files) != 1 || payload.Files[0] != taskPath {
		t.Errorf("files = %v, want [%s]", payload.Files, taskPath)
	}
}
