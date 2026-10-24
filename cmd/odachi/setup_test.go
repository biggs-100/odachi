package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	domainopencode "github.com/biggs-100/odachi/internal/domain/opencode"
	"github.com/biggs-100/odachi/internal/domain/persona"
)

// TestSetupLeavesTheFolderAloneAndTheHomeReady: the installer runs setup from
// any folder. The human only installs; the first conversation finds an agent
// that knows the verbs, and the folder it ran in gains nothing.
func TestSetupLeavesTheFolderAloneAndTheHomeReady(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	writeFile(t, home, ".config/opencode/opencode.json", "{}")

	dir := t.TempDir()
	initRepo(t, dir)

	code, out, errOut := runDir(t, dir, []string{"setup"}, "")
	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout: %q, stderr: %q)", code, out, errOut)
	}

	door := filepath.Join(home, ".config", "opencode", "agents", "odachi.md")
	if _, err := os.Stat(door); err != nil {
		t.Errorf("the home door was not installed: %v", err)
	}
	bootstrap := filepath.Join(home, ".agents", "skills", "bootstrap", "SKILL.md")
	if _, err := os.Stat(bootstrap); err != nil {
		t.Errorf("the bootstrap skill was not installed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(home, ".config", "opencode", "opencode.json"))
	if err != nil {
		t.Fatalf("reading the config: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(domainopencode.StripComments(data), &doc); err != nil {
		t.Fatalf("the config is not valid: %v\n%s", err, data)
	}
	if doc["default_agent"] != persona.AgentName {
		t.Errorf("default_agent = %v, want %q", doc["default_agent"], persona.AgentName)
	}

	for _, name := range []string{"AGENTS.md", ".gitignore", ".opencode"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			t.Errorf("setup wrote %s in the folder it ran in", name)
		}
	}
}
