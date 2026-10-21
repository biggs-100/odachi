package main

import (
	"strings"
	"testing"
)

// TestStartTwiceReusesTheSameTaskDocument pins the human-facing rule: one
// feature, one task document. A beginner fixes a typo in the brief by running
// start again; that must not leave two documents for the agent to choose from.
func TestStartTwiceReusesTheSameTaskDocument(t *testing.T) {
	dir := t.TempDir()

	if code, _, stderr := runDir(t, dir, allFlags(), ""); code != 0 {
		t.Fatalf("primera corrida: exit %d, stderr=%s", code, stderr)
	}
	before := tree(t, dir)
	taskBefore := onlyTask(t, before)

	if code, _, stderr := runDir(t, dir, append(allFlags(), "--force"), ""); code != 0 {
		t.Fatalf("segunda corrida: exit %d, stderr=%s", code, stderr)
	}
	after := tree(t, dir)
	taskAfter := onlyTask(t, after)

	if taskAfter != taskBefore {
		t.Fatalf("correr start dos veces cambió el task doc:\nantes: %s\ndespués: %s", taskBefore, taskAfter)
	}
	if before[taskBefore] != after[taskAfter] {
		t.Fatalf("correr start dos veces reescribió el task doc; sus pasos y su historial son la memoria del trabajo y no se tocan")
	}
}

// TestADifferentOutcomeGetsItsOwnNumber keeps the fix honest: reuse must be by
// slug, never by "whatever document exists".
func TestADifferentOutcomeGetsItsOwnNumber(t *testing.T) {
	dir := t.TempDir()

	if code, _, stderr := runDir(t, dir, allFlags(), ""); code != 0 {
		t.Fatalf("primera feature: exit %d, stderr=%s", code, stderr)
	}

	other := allFlags()
	for i := 0; i < len(other); i++ {
		if other[i] == "--outcome" {
			other[i+1] = "Quiero facturar a mis obras sociales"
		}
	}
	if code, _, stderr := runDir(t, dir, append(other, "--force"), ""); code != 0 {
		t.Fatalf("segunda feature: exit %d, stderr=%s", code, stderr)
	}

	var tasks []string
	for path := range tree(t, dir) {
		if strings.HasPrefix(path, "odd/tasks/") && strings.HasSuffix(path, ".md") {
			tasks = append(tasks, path)
		}
	}

	if len(tasks) != 2 {
		t.Fatalf("dos features distintas dejaron %d task docs (%v); cada una necesita el suyo", len(tasks), tasks)
	}
}

// onlyTask returns the single task document path, failing when there is not
// exactly one.
func onlyTask(t *testing.T, files map[string]string) string {
	t.Helper()
	var tasks []string
	for path := range files {
		if strings.HasPrefix(path, "odd/tasks/") && strings.HasSuffix(path, ".md") {
			tasks = append(tasks, path)
		}
	}
	if len(tasks) != 1 {
		t.Fatalf("esperaba un task doc, encontré %d (%v)", len(tasks), tasks)
	}
	return tasks[0]
}
