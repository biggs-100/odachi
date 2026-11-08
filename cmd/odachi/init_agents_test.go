package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/persona"
)

// TestInitAgentsOpenCodeLeavesClaudeAlone: the real command, on a real tree. The
// client that was not chosen never reaches the disk, and a plain init after it
// keeps the stored choice instead of bringing Claude back.
func TestInitAgentsOpenCodeLeavesClaudeAlone(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepo(t, dir)

	code, out, errOut := runDir(t, dir, []string{"init", "--agents", "opencode"}, "")
	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout %q, stderr %q)", code, out, errOut)
	}
	if !strings.Contains(out, "opencode") {
		t.Errorf("stdout %q does not name the chosen client", out)
	}
	claude, _ := persona.GlobalFilesFor(persona.SelectionClaude)
	if _, err := os.Stat(filepath.Join(home, filepath.FromSlash(claude.Path))); !os.IsNotExist(err) {
		t.Errorf("%s exists after --agents opencode: %v", claude.Path, err)
	}
	opencode, _ := persona.GlobalFilesFor(persona.SelectionOpenCode)
	if _, err := os.Stat(filepath.Join(home, filepath.FromSlash(opencode.Path))); err != nil {
		t.Errorf("the OpenCode door was not installed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".opencode", "agents", persona.FileName)); err == nil {
		t.Error("init wrote the OpenCode agent in the project")
	}

	store, err := os.ReadFile(filepath.Join(dir, ".git", "odachi", filepath.FromSlash(persona.SelectionPath)))
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
	if _, err := os.Stat(filepath.Join(home, filepath.FromSlash(claude.Path))); !os.IsNotExist(err) {
		t.Errorf("the Claude door came back on a plain init: %v", err)
	}
	if strings.Contains(out, ".claude") {
		t.Errorf("the second run names .claude although it did not write it:\n%s", out)
	}
}

// TestInitAgentsBothRestoresTheOtherClient: the flag always wins over the stored
// choice, so choosing both again installs the door that was skipped.
func TestInitAgentsBothRestoresTheOtherClient(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepo(t, dir)

	if code, _, errOut := runDir(t, dir, []string{"init", "--agents", "opencode"}, ""); code != 0 {
		t.Fatalf("first run code = %d (stderr %q)", code, errOut)
	}
	code, _, errOut := runDir(t, dir, []string{"init", "--agents", "both"}, "")
	if code != 0 {
		t.Fatalf("second run code = %d, want 0 (stderr %q)", code, errOut)
	}

	claude, _ := persona.GlobalFilesFor(persona.SelectionClaude)
	if _, err := os.Stat(filepath.Join(home, filepath.FromSlash(claude.Path))); err != nil {
		t.Errorf("the Claude door was not restored: %v", err)
	}
	store, err := os.ReadFile(filepath.Join(dir, ".git", "odachi", filepath.FromSlash(persona.SelectionPath)))
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

// TestLevelLeavesTheProjectAgentLeftoversAlone: since 0242 no project carries an
// agent file. A profile change updates the block and leaves a file an older
// release wrote alone, so level cannot keep alive a file doctor calls unread.
func TestLevelLeavesTheProjectAgentLeftoversAlone(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir, _ := startedProject(t)
	if code, _, errOut := runDir(t, dir, []string{"init", "--agents", "opencode"}, ""); code != 0 {
		t.Fatalf("init: exit %d, stderr=%s", code, errOut)
	}
	writeFile(t, dir, ".opencode/agents/odachi.md", "an older agent\n")
	writeFile(t, dir, ".claude/output-styles/odachi.md", "viejo\n")

	code, _, errOut := runDir(t, dir, []string{"level", "--to", "senior"}, "")

	if code != 0 {
		t.Fatalf("level: exit %d, stderr=%s", code, errOut)
	}
	for path, want := range map[string]string{
		".opencode/agents/odachi.md":      "an older agent\n",
		".claude/output-styles/odachi.md": "viejo\n",
	} {
		got, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(path)))
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if string(got) != want {
			t.Errorf("level rewrote %s: %q", path, got)
		}
	}
}

// TestDoctorRespectsTheStoredChoice: after an install that skipped Claude,
// doctor does not ask for Claude's files and says which choice it obeyed.
func TestDoctorRespectsTheStoredChoice(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	if code, _, errOut := runDir(t, dir, []string{"init", "--agents", "opencode"}, ""); code != 0 {
		t.Fatalf("init: exit %d, stderr=%s", code, errOut)
	}

	_, out, errOut := runDir(t, dir, []string{"doctor"}, "")
	combined := out + errOut

	if !strings.Contains(combined, "I only check opencode") {
		t.Errorf("doctor does not say which client it checked:\n%s", combined)
	}
	if strings.Contains(combined, ".claude") {
		t.Errorf("doctor still asks for .claude after --agents opencode:\n%s", combined)
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
