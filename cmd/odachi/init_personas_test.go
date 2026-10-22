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
		t.Fatalf("AGENTS.md no se escribió: %v", err)
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
		t.Fatalf("el bloque no lleva la voz de novato:\n%q\nvs\n%q", got, want)
	}
	if !strings.Contains(out, persona.Novato) {
		t.Errorf("la salida no dice con qué voz se escribió el bloque:\n%s", out)
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
			t.Fatalf("persona.FilesFor(%q) no existe", client.Agent)
		}
		got, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(file.Path)))
		if err != nil {
			t.Errorf("%s no se escribió: %v", file.Path, err)
			continue
		}
		if string(got) != file.Content {
			t.Errorf("%s no es el espejo del bloque:\n%q\nvs\n%q", file.Path, got, file.Content)
		}
	}
	for _, legacy := range personaLegacyFiles {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(legacy))); err == nil {
			t.Errorf("init escribió %s: el agente de esta versión es %s", legacy, persona.FileName)
		}
	}

	// The door in the home: the same command leaves odachi available in every
	// folder, not only in the project that ran init.
	home, err := homeDir()
	if err != nil {
		t.Fatalf("homeDir: %v", err)
	}
	for _, file := range persona.GlobalFiles() {
		got, err := os.ReadFile(filepath.Join(home, filepath.FromSlash(file.Path)))
		if err != nil {
			t.Errorf("%s no se instaló en el home: %v", file.Path, err)
			continue
		}
		if string(got) != file.Content {
			t.Errorf("%s no es la puerta del home:\n%q\nvs\n%q", file.Path, got, file.Content)
		}
	}
}

// TestInitUsesTheProfileTheHumanWroteDown: init cannot know the person, so the
// only thing that can say the voice is a profile already on disk.
func TestInitUsesTheProfileTheHumanWroteDown(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	writeFile(t, dir, ".odachi/profile.toml", "name = \"senior\"\n")

	if code, _, errOut := runDir(t, dir, []string{"init"}, ""); code != 0 {
		t.Fatalf("code = %d (stderr %q)", code, errOut)
	}

	want, err := rules.Block(persona.Senior)
	if err != nil {
		t.Fatalf("rules.Block(senior): %v", err)
	}
	if got := readAgents(t, dir); got != want {
		t.Fatalf("el bloque no lleva la voz del perfil:\n%q\nvs\n%q", got, want)
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
		t.Errorf("el segundo init tocó el proyecto: before %v after %v", keys(before), keys(after))
	}
	if afterHome := homeTree(t, home); strings.Join(beforeHome, ",") != strings.Join(afterHome, ",") {
		t.Errorf("el segundo init tocó el home: before %v after %v", beforeHome, afterHome)
	}
	if !strings.Contains(out, "sin cambios") {
		t.Errorf("la salida no dice que no cambió nada:\n%s", out)
	}
}

// TestInitReportsAndKeepsTheLegacyPersonaFiles: they are the human's files now.
// init names them in one line, says they are leftovers, and writes nothing near
// them.
func TestInitReportsAndKeepsTheLegacyPersonaFiles(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	legacy := ".opencode/agents/odachi-senior.md"
	writeFile(t, dir, legacy, "la versión vieja\n")

	code, out, errOut := runDir(t, dir, []string{"init"}, "")
	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout %q, stderr %q)", code, out, errOut)
	}

	for _, want := range []string{"v0.10.0", ".opencode/agents", "AGENTS.md", "borralas"} {
		if !strings.Contains(out, want) {
			t.Errorf("la salida no dice %q:\n%s", want, out)
		}
	}
	data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(legacy)))
	if err != nil {
		t.Fatalf("%s se borró: odachi no borra archivos que no escribió esta versión", legacy)
	}
	if string(data) != "la versión vieja\n" {
		t.Errorf("%s cambió: %q", legacy, data)
	}
	// One line, not a list: a human reads a report, not an inventory.
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	for _, line := range lines {
		if strings.Contains(line, "v0.10.0") && strings.Contains(line, persona.Deprecated()[0]) {
			t.Errorf("la línea de leftovers enumera archivos en vez de carpetas: %q", line)
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
	if len(payload.Personas) != 2*len(persona.Clients()) {
		t.Fatalf("personas = %+v, want one project and one home row per client", payload.Personas)
	}
	scopes := map[string]bool{}
	for _, row := range payload.Personas {
		scopes[row.Scope] = true
		if row.Written+row.Unchanged != 1 || len(row.Files) != 1 {
			t.Errorf("fila de %s/%s = %+v, want one file owned", row.Client, row.Scope, row)
			continue
		}
		if want := row.Dir + "/" + persona.FileName; row.Files[0] != want {
			t.Errorf("fila de %s/%s = %+v, want %s", row.Client, row.Scope, row, want)
		}
	}
	for _, scope := range []string{"project", "home"} {
		if !scopes[scope] {
			t.Errorf("el reporte no trae la fila de scope %q: %+v", scope, payload.Personas)
		}
	}
	if !slicesContains(payload.Files, "AGENTS.md") {
		t.Errorf("files = %v, want it to name AGENTS.md", payload.Files)
	}
	if len(payload.Skills) == 0 {
		t.Errorf("el reporte perdió las skills: %s", out)
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
