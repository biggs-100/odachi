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

// personaLegacyFiles are the per-client files v0.10.0 wrote, pinned by name: they
// are the contract with the previous version, and a change that stops reporting
// them strands whoever installed that version with files nothing reads.
var personaLegacyFiles = []string{
	".opencode/agents/odachi-novato.md",
	".opencode/agents/odachi-senior.md",
	".claude/output-styles/odachi-novato.md",
	".claude/output-styles/odachi-senior.md",
}

// readAgents is AGENTS.md from the project, or a fatal error: the block is the
// artifact this change is about.
func readAgents(t *testing.T, dir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("AGENTS.md was not written: %v", err)
	}
	return string(data)
}

// TestInitWritesTheVoiceIntoTheBlock: one command leaves the voice where every
// client reads it, with no per-client file to select and no keystroke to learn.
func TestInitWritesTheVoiceIntoTheBlock(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)

	code, out, errOut := runDir(t, dir, []string{"init"}, "")
	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout %q, stderr %q)", code, out, errOut)
	}

	want, err := rules.Block(persona.Novato)
	if err != nil {
		t.Fatalf("rules.Block(novato): %v", err)
	}
	if got := readAgents(t, dir); got != want {
		t.Fatalf("the block does not carry the novato voice:\n%q\nvs\n%q", got, want)
	}
	if !strings.Contains(out, persona.Novato) {
		t.Errorf("the output does not say which voice the block was written with:\n%s", out)
	}

	// The native agent mirrors the same block, so the human can pick odachi with
	// the client's own selector. The v0.10.0 files are not rewritten under their
	// old names.
	block, err := rules.Body(persona.Novato)
	if err != nil {
		t.Fatalf("rules.Body(novato): %v", err)
	}
	for _, client := range persona.Clients() {
		file, ok := persona.FilesFor(client.Agent, block)
		if !ok {
			t.Fatalf("persona.FilesFor(%q) does not exist", client.Agent)
		}
		got, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(file.Path)))
		if err != nil {
			t.Errorf("%s was not written: %v", file.Path, err)
			continue
		}
		if string(got) != file.Content {
			t.Errorf("%s is not the block mirror:\n%q\nvs\n%q", file.Path, got, file.Content)
		}
	}

	// The worker and the verifier are native subagents: generated with the same
	// command, read only when the parent invokes them.
	for _, client := range persona.Clients() {
		for _, file := range persona.SubagentFilesFor(client.Agent) {
			got, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(file.Path)))
			if err != nil {
				t.Errorf("%s was not written: %v", file.Path, err)
				continue
			}
			if string(got) != file.Content {
				t.Errorf("%s does not match the generated subagent:\n%q\nvs\n%q", file.Path, got, file.Content)
			}
		}
	}
	for _, want := range []string{"odachi-worker", "odachi-verify"} {
		if !strings.Contains(out, want) {
			t.Errorf("the output does not name %s:\n%s", want, out)
		}
	}
	for _, legacy := range personaLegacyFiles {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(legacy))); err == nil {
			t.Errorf("init wrote %s: this version's agent is %s", legacy, persona.FileName)
		}
	}

	// The door in the home is opt-in and belongs to --global, which now leaves the
	// project's own files alone: TestGlobalInitLeavesTheProjectAgentsAlone drives
	// both halves, and TestInitWithoutGlobalDoesNotWriteTheDoor drives the silence.
}

// TestInitWithoutGlobalDoesNotWriteTheDoor: an init without the flag leaves the
// home exactly as it found it. The door is an agent primary in every folder, so
// installing it is a decision the human takes out loud.
func TestInitWithoutGlobalDoesNotWriteTheDoor(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepo(t, dir)

	code, out, errOut := runDir(t, dir, []string{"init"}, "")
	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout %q, stderr %q)", code, out, errOut)
	}

	for _, file := range persona.GlobalFiles() {
		if _, err := os.Stat(filepath.Join(home, filepath.FromSlash(file.Path))); err == nil {
			t.Errorf("%s was installed without --global", file.Path)
		}
	}
	if !strings.Contains(out, "--global") {
		t.Errorf("the output does not name the flag that installs the door:\n%s", out)
	}
}

// TestInitUsesTheProfileTheHumanWroteDown: init cannot know the person, so the
// only thing that can say the voice is a profile already on disk.
func TestInitUsesTheProfileTheHumanWroteDown(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	writeFile(t, dir, "odd/profile.toml", "name = \"senior\"\n")

	if code, _, errOut := runDir(t, dir, []string{"init"}, ""); code != 0 {
		t.Fatalf("code = %d (stderr %q)", code, errOut)
	}

	want, err := rules.Block(persona.Senior)
	if err != nil {
		t.Fatalf("rules.Block(senior): %v", err)
	}
	if got := readAgents(t, dir); got != want {
		t.Fatalf("the block does not carry the profile voice:\n%q\nvs\n%q", got, want)
	}
}

// TestInitRerunWritesNothing: the block is generated, so a second run touches no
// file, leaves no backup, and says so.
func TestInitRerunWritesNothing(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	if code, _, errOut := runDir(t, dir, []string{"init"}, ""); code != 0 {
		t.Fatalf("first run code = %d (stderr %q)", code, errOut)
	}
	home, err := homeDir()
	if err != nil {
		t.Fatalf("homeDir: %v", err)
	}
	before := tree(t, dir)
	beforeHome := homeTree(t, home)

	code, out, errOut := runDir(t, dir, []string{"init"}, "")

	if code != 0 {
		t.Fatalf("second run code = %d (stdout %q, stderr %q)", code, out, errOut)
	}
	if after := tree(t, dir); strings.Join(keys(before), ",") != strings.Join(keys(after), ",") {
		t.Errorf("the second init touched the project: before %v after %v", keys(before), keys(after))
	}
	if afterHome := homeTree(t, home); strings.Join(beforeHome, ",") != strings.Join(afterHome, ",") {
		t.Errorf("the second init touched the home: before %v after %v", beforeHome, afterHome)
	}
	if !strings.Contains(out, "unchanged") {
		t.Errorf("the output does not say nothing changed:\n%s", out)
	}
}

// TestInitReportsAndKeepsTheLegacyPersonaFiles: they are the human's files now.
// init names them in one line, says they are leftovers, and writes nothing near
// them.
func TestInitReportsAndKeepsTheLegacyPersonaFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	legacy := ".opencode/agents/odachi-senior.md"
	writeFile(t, dir, legacy, "la versión vieja\n")

	code, out, errOut := runDir(t, dir, []string{"init"}, "")
	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout %q, stderr %q)", code, out, errOut)
	}

	for _, want := range []string{"v0.10.0", ".opencode/agents", "AGENTS.md", "delete them"} {
		if !strings.Contains(out, want) {
			t.Errorf("the output does not say %q:\n%s", want, out)
		}
	}
	data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(legacy)))
	if err != nil {
		t.Fatalf("%s was deleted: odachi does not delete files this version did not write", legacy)
	}
	if string(data) != "la versión vieja\n" {
		t.Errorf("%s changed: %q", legacy, data)
	}
	// One line, not a list: a human reads a report, not an inventory.
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	for _, line := range lines {
		if strings.Contains(line, "v0.10.0") && strings.Contains(line, persona.Deprecated()[0]) {
			t.Errorf("the leftovers line lists files instead of folders: %q", line)
		}
	}
}

// TestInitJSONCarriesTheNativeAgent: the machine report says where the agent
// landed and what this run did there, the same two facts the human line carries.
func TestInitJSONCarriesTheNativeAgent(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)

	code, out, errOut := runDir(t, dir, []string{"init", "--json"}, "")
	if code != 0 {
		t.Fatalf("code = %d, want 0 (stderr %q)", code, errOut)
	}

	var payload struct {
		Files  []string `json:"files"`
		Skills []struct {
			Dest string `json:"dest"`
		} `json:"skills"`
		Personas []struct {
			Client    string   `json:"client"`
			Scope     string   `json:"scope"`
			Dir       string   `json:"dir"`
			Written   int      `json:"written"`
			Unchanged int      `json:"unchanged"`
			Files     []string `json:"files"`
		} `json:"personas"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, out)
	}
	// One row per destination, and a plain init is the run that writes the project's
	// own: OpenCode shares the primary and subagent folder, Claude Code reads the
	// primary from output-styles and the subagents from agents. The home is untouched
	// without --global (its rows are the next test's).
	want := map[string]int{
		"project|.opencode/agents":      4,
		"project|.claude/output-styles": 1,
		"project|.claude/agents":        3,
	}
	if len(payload.Personas) != len(want) {
		t.Fatalf("personas = %+v, want %d destination rows", payload.Personas, len(want))
	}
	scopes := map[string]bool{}
	for _, row := range payload.Personas {
		scopes[row.Scope] = true
		key := row.Scope + "|" + row.Dir
		count, ok := want[key]
		if !ok {
			t.Errorf("unexpected row %s = %+v", key, row)
			continue
		}
		if row.Written+row.Unchanged != count || len(row.Files) != count {
			t.Errorf("row %s = %+v, want %d files owned", key, row, count)
			continue
		}
		for _, file := range row.Files {
			if !strings.HasPrefix(file, row.Dir+"/") {
				t.Errorf("row %s owns %s, outside its directory", key, file)
			}
		}
	}
	for _, scope := range []string{"project"} {
		if !scopes[scope] {
			t.Errorf("the report does not carry the %q scope row: %+v", scope, payload.Personas)
		}
	}
	if scopes["home"] {
		t.Errorf("a plain init wrote the home: %+v", payload.Personas)
	}

	// The global run is the one that carries the home rows, and it leaves the
	// project's files alone: the door runs in every folder and the project keeps
	// only the rules and the plan.
	code, out, errOut = runDir(t, dir, []string{"init", "--json", "--global"}, "")
	if code != 0 {
		t.Fatalf("init --global --json: code = %d (stderr %q)", code, errOut)
	}
	plainFiles := payload.Files
	payload.Personas = nil
	payload.Files = nil
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, out)
	}
	homeRows := map[string]int{
		"home|~/.config/opencode/agents": 1,
		"home|~/.claude/output-styles":   1,
	}
	if len(payload.Personas) != len(homeRows) {
		t.Fatalf("--global personas = %+v, want only the home rows", payload.Personas)
	}
	for _, row := range payload.Personas {
		key := row.Scope + "|" + row.Dir
		count, ok := homeRows[key]
		if !ok {
			t.Errorf("unexpected row %s = %+v", key, row)
			continue
		}
		if row.Written+row.Unchanged != count || len(row.Files) != count {
			t.Errorf("row %s = %+v, want %d files owned", key, row, count)
		}
	}
	if !slicesContains(plainFiles, "AGENTS.md") {
		t.Errorf("files = %v, want it to name AGENTS.md", plainFiles)
	}
	if len(payload.Skills) == 0 {
		t.Errorf("the report lost the skills: %s", out)
	}
}

// slicesContains is the membership test the JSON payload needs; the standard
// library one arrived in a version this module does not pin for tests.
func slicesContains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
