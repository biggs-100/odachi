package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestInitWritesBothArtifacts: the rules block and the hook, and it says so.
func TestInitWritesBothArtifacts(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)

	code, out, errOut := runDir(t, dir, []string{"init"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if !strings.Contains(out, "AGENTS.md") || !strings.Contains(out, ".git/hooks/pre-commit") {
		t.Errorf("stdout %q does not report both artifacts", out)
	}

	agents, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("AGENTS.md was not written: %v", err)
	}
	if !strings.Contains(string(agents), "<!-- odachi:begin -->") || !strings.Contains(string(agents), "<!-- odachi:end -->") {
		t.Errorf("AGENTS.md has no odachi block:\n%s", agents)
	}

	hook, err := os.ReadFile(filepath.Join(dir, ".git", "hooks", "pre-commit"))
	if err != nil {
		t.Fatalf("the hook was not written: %v", err)
	}
	if !strings.HasPrefix(string(hook), "#!/bin/sh\n") {
		t.Errorf("hook starts with %q, want the shebang", hook)
	}
	if !strings.Contains(string(hook), "check --quiet || exit 1") {
		t.Errorf("hook does not run the gate:\n%s", hook)
	}
}

// A rerun is silent about work it did not do and changes no bytes.
func TestInitRerunChangesNothing(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	if code, _, errOut := runDir(t, dir, []string{"init"}, ""); code != 0 {
		t.Fatalf("first run code = %d (stderr: %q)", code, errOut)
	}
	agentsBefore, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	hookBefore, err := os.ReadFile(filepath.Join(dir, ".git", "hooks", "pre-commit"))
	if err != nil {
		t.Fatalf("read hook: %v", err)
	}

	code, out, errOut := runDir(t, dir, []string{"init"}, "")

	if code != 0 {
		t.Fatalf("second run code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if !strings.Contains(out, "sin cambios") {
		t.Errorf("stdout %q does not say nothing changed", out)
	}
	agentsAfter, _ := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if string(agentsAfter) != string(agentsBefore) {
		t.Errorf("AGENTS.md changed on a rerun:\n%q\nvs\n%q", agentsAfter, agentsBefore)
	}
	hookAfter, _ := os.ReadFile(filepath.Join(dir, ".git", "hooks", "pre-commit"))
	if string(hookAfter) != string(hookBefore) {
		t.Errorf("the hook changed on a rerun:\n%q\nvs\n%q", hookAfter, hookBefore)
	}
}

// A hook someone else wrote is reported, never replaced, and the three lines to
// paste are printed.
func TestInitRefusesAForeignHook(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	foreign := "#!/bin/sh\necho hola\n"
	writeFile(t, dir, ".git/hooks/pre-commit", foreign)

	code, _, errOut := runDir(t, dir, []string{"init"}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1 (stderr: %q)", code, errOut)
	}
	if !strings.Contains(errOut, ".git/hooks/pre-commit") {
		t.Errorf("stderr %q does not name the file", errOut)
	}
	for _, want := range []string{"#!/bin/sh", "# odachi", "check --quiet || exit 1"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("stderr %q does not print the line %q to paste", errOut, want)
		}
	}

	after, err := os.ReadFile(filepath.Join(dir, ".git", "hooks", "pre-commit"))
	if err != nil {
		t.Fatalf("read hook: %v", err)
	}
	if string(after) != foreign {
		t.Errorf("the foreign hook was modified:\n%q", after)
	}
}

// Without a repository the command creates the history itself. Owning a project
// does not require knowing what git is, and refusing with "corré git init" in front
// of a beginner is the opposite of what this tool promises: creating the history is
// dirty work, and dirty work is the agent's job.
func TestInitCreatesTheHistoryWhenThereIsNone(t *testing.T) {
	dir := t.TempDir()

	code, out, errOut := runDir(t, dir, []string{"init"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stderr: %q)", code, errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		t.Fatalf("init did not create the history: %v", err)
	}
	if !strings.Contains(out, "historial del proyecto") {
		t.Errorf("stdout %q does not say what it created", out)
	}
	if !strings.Contains(out, "git init") {
		t.Errorf("stdout %q does not name git in parentheses, so a dev can see what it did", out)
	}
	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); err != nil {
		t.Errorf("AGENTS.md was not written: %v", err)
	}
}

// The block is replaced in place through the real filesystem too: the human's
// lines survive and the old block does not.
func TestInitReplacesTheBlockAndKeepsTheRest(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	writeFile(t, dir, "AGENTS.md", "# Mis reglas\n\nno toques production\n\n<!-- odachi:begin -->\nregla vieja\n<!-- odachi:end -->\n\ny esto va después\n")

	if code, _, errOut := runDir(t, dir, []string{"init"}, ""); code != 0 {
		t.Fatalf("code = %d (stderr: %q)", code, errOut)
	}

	data, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	content := string(data)
	if strings.Contains(content, "regla vieja") {
		t.Errorf("the old block survived:\n%s", content)
	}
	for _, want := range []string{"# Mis reglas", "no toques production", "y esto va después", "Brief antes de código"} {
		if !strings.Contains(content, want) {
			t.Errorf("AGENTS.md lost %q:\n%s", want, content)
		}
	}
	if matches, err := filepath.Glob(filepath.Join(dir, "AGENTS.md.bak-*")); err != nil || len(matches) != 1 {
		t.Errorf("backups = %v (err %v), want exactly one backup of AGENTS.md", matches, err)
	}
}

// Machine output carries the code and the files it touched.
func TestInitJSONListsTheFiles(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)

	code, out, errOut := runDir(t, dir, []string{"init", "--json"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stderr: %q)", code, errOut)
	}

	var payload struct {
		Code  int      `json:"code"`
		Files []string `json:"files"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, out)
	}
	if payload.Code != 0 {
		t.Errorf("JSON code = %d, want 0", payload.Code)
	}
	joined := strings.Join(payload.Files, " ")
	if !strings.Contains(joined, "AGENTS.md") || !strings.Contains(joined, ".git/hooks/pre-commit") {
		t.Errorf("JSON files = %v, want both artifacts", payload.Files)
	}
}

// The human does not commit: the agent does. The closing line used to say
// "commiteá", which assumed the one thing this tool exists to take off their plate.
func TestInitDoesNotAskTheHumanToCommit(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)

	_, out, _ := runDir(t, dir, []string{"init"}, "")

	if strings.Contains(out, "commiteá") {
		t.Fatalf("init asks the human to commit: %q", out)
	}
	if !strings.Contains(out, "el agente guarda el avance") {
		t.Fatalf("init does not say who saves the progress: %q", out)
	}
}
