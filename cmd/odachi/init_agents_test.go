package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/rules"
)

// TestInitAgentsOpenCodeLeavesClaudeAlone: the real command, on a real tree. The
// client that was not chosen never reaches the disk, and a plain init after it
// keeps the stored choice instead of bringing Claude back.
func TestInitAgentsOpenCodeLeavesClaudeAlone(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)

	code, out, errOut := runDir(t, dir, []string{"init", "--agents", "opencode"}, "")
	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout %q, stderr %q)", code, out, errOut)
	}
	if !strings.Contains(out, "opencode") {
		t.Errorf("stdout %q does not name the chosen client", out)
	}
	if _, err := os.Stat(filepath.Join(dir, ".claude")); !os.IsNotExist(err) {
		t.Errorf(".claude exists after --agents opencode: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".opencode", "agents", persona.FileName)); err != nil {
		t.Errorf("the OpenCode agent was not written: %v", err)
	}

	store, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(persona.SelectionPath)))
	if err != nil {
		t.Fatalf("%s was not written: %v", persona.SelectionPath, err)
	}
	if !strings.Contains(string(store), `agents = "opencode"`) {
		t.Errorf("%s = %q, want agents = \"opencode\"", persona.SelectionPath, store)
	}

	code, out, errOut = runDir(t, dir, []string{"init"}, "")
	if code != 0 {
		t.Fatalf("second run code = %d, want 0 (stdout %q, stderr %q)", code, out, errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, ".claude")); !os.IsNotExist(err) {
		t.Errorf(".claude came back on a plain init: %v", err)
	}
	if strings.Contains(out, ".claude") {
		t.Errorf("the second run names .claude although it did not write it:\n%s", out)
	}
}

// TestInitAgentsBothRestoresTheOtherClient: the flag always wins over the stored
// choice, so choosing both again installs the files that were skipped.
func TestInitAgentsBothRestoresTheOtherClient(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)

	if code, _, errOut := runDir(t, dir, []string{"init", "--agents", "opencode"}, ""); code != 0 {
		t.Fatalf("first run code = %d (stderr %q)", code, errOut)
	}
	code, _, errOut := runDir(t, dir, []string{"init", "--agents", "both"}, "")
	if code != 0 {
		t.Fatalf("second run code = %d, want 0 (stderr %q)", code, errOut)
	}

	if _, err := os.Stat(filepath.Join(dir, ".claude", "output-styles", persona.FileName)); err != nil {
		t.Errorf("the Claude agent was not restored: %v", err)
	}
	store, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(persona.SelectionPath)))
	if err != nil {
		t.Fatalf("%s was not written: %v", persona.SelectionPath, err)
	}
	if !strings.Contains(string(store), `agents = "both"`) {
		t.Errorf("%s = %q, want agents = \"both\"", persona.SelectionPath, store)
	}
}

// TestInitAgentsRefusesAnUnknownClient: a typo exits before writing anything, so
// it cannot leave half an install behind.
func TestInitAgentsRefusesAnUnknownClient(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)

	code, _, errOut := runDir(t, dir, []string{"init", "--agents", "todos"}, "")
	if code != 1 {
		t.Fatalf("code = %d, want 1 (stderr %q)", code, errOut)
	}
	for _, want := range []string{"opencode", "claude", "both"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("stderr %q does not name %q", errOut, want)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); !os.IsNotExist(err) {
		t.Errorf("init wrote AGENTS.md with an invalid --agents: %v", err)
	}
}

// TestLevelKeepsTheUnchosenClient: the stored choice also governs the refresh a
// profile change does. A stale file from the other client keeps its bytes.
func TestLevelKeepsTheUnchosenClient(t *testing.T) {
	dir, _ := startedProject(t)
	if code, _, errOut := runDir(t, dir, []string{"init", "--agents", "opencode"}, ""); code != 0 {
		t.Fatalf("init: exit %d, stderr=%s", code, errOut)
	}
	writeFile(t, dir, ".claude/output-styles/odachi.md", "viejo\n")

	code, _, errOut := runDir(t, dir, []string{"level", "--to", "senior"}, "")

	if code != 0 {
		t.Fatalf("level: exit %d, stderr=%s", code, errOut)
	}
	got, err := os.ReadFile(filepath.Join(dir, ".claude", "output-styles", persona.FileName))
	if err != nil {
		t.Fatalf("read the Claude agent: %v", err)
	}
	if string(got) != "viejo\n" {
		t.Errorf("level refreshed the unchosen client: %q", got)
	}
	body, err := rules.Body(persona.Senior)
	if err != nil {
		t.Fatalf("rules.Body(senior): %v", err)
	}
	chosen, _ := persona.FilesFor(persona.SelectionOpenCode, body)
	opencode, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(chosen.Path)))
	if err != nil {
		t.Fatalf("read the OpenCode agent: %v", err)
	}
	if string(opencode) != chosen.Content {
		t.Errorf("the OpenCode agent did not follow the new voice")
	}
}

// TestInitAgentsJSONOnlyReportsTheChosenClient: the machine report has exactly
// the destinations this run owns, so a consumer cannot think Claude was written.
func TestInitAgentsJSONOnlyReportsTheChosenClient(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)

	code, out, errOut := runDir(t, dir, []string{"init", "--agents", "opencode", "--json"}, "")
	if code != 0 {
		t.Fatalf("code = %d, want 0 (stderr %q)", code, errOut)
	}

	var payload struct {
		Personas []struct {
			Client string `json:"client"`
			Scope  string `json:"scope"`
			Dir    string `json:"dir"`
		} `json:"personas"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, out)
	}
	if len(payload.Personas) == 0 {
		t.Fatal("personas is empty")
	}
	for _, row := range payload.Personas {
		if row.Client != persona.SelectionOpenCode {
			t.Errorf("persona row = %+v, want only opencode", row)
		}
		if strings.Contains(row.Dir, ".claude") {
			t.Errorf("persona row = %+v, want no Claude destination", row)
		}
	}
}
