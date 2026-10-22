package main

import (
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/persona"
)

// The two project directories a client reads personas from, and the four files
// `odachi init` must leave behind. Pinned by name: this is the contract with each
// client's own mechanism, and a table that only agreed with itself would not
// notice a client moving its directory.
const (
	opencodePersonas = ".opencode/agents"
	claudePersonas   = ".claude/output-styles"
)

var personaFiles = []string{
	opencodePersonas + "/odachi-novato.md",
	opencodePersonas + "/odachi-senior.md",
	claudePersonas + "/odachi-novato.md",
	claudePersonas + "/odachi-senior.md",
}

// personaBody strips the frontmatter a client needs and returns the body every
// client shares.
func personaBody(t *testing.T, content string) string {
	t.Helper()
	if !strings.HasPrefix(content, "---\n") {
		t.Fatalf("el archivo no empieza con frontmatter:\n%s", content)
	}
	end := strings.Index(content, "\n---\n")
	if end < 0 {
		t.Fatalf("el frontmatter no cierra:\n%s", content)
	}
	return content[end+len("\n---\n"):]
}

// personaTree is every file under the two persona directories, keyed by slash
// path: what a rerun must leave byte for byte the same, backups included.
func personaTree(t *testing.T, dir string) map[string]string {
	t.Helper()
	all := tree(t, dir)
	out := map[string]string{}
	for path, content := range all {
		if strings.HasPrefix(path, opencodePersonas+"/") || strings.HasPrefix(path, claudePersonas+"/") {
			out[path] = content
		}
	}
	return out
}

// readPersona is one persona file from the project, or a fatal error: a missing
// file is the defect the test exists for.
func readPersona(t *testing.T, dir, path string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(path)))
	if err != nil {
		t.Fatalf("%s no se escribió: %v", path, err)
	}
	return string(data)
}

// TestInitWritesBothPersonasForEachClient: one command leaves both clients able
// to pick a persona with their own keystroke, with the same body and the wrapper
// each one needs.
func TestInitWritesBothPersonasForEachClient(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)

	code, out, errOut := runDir(t, dir, []string{"init"}, "")
	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout %q, stderr %q)", code, out, errOut)
	}

	for _, path := range personaFiles {
		content := readPersona(t, dir, path)
		name := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(path), "odachi-"), ".md")
		want, err := persona.Body(name)
		if err != nil {
			t.Fatalf("persona.Body(%q): %v", name, err)
		}
		if got := personaBody(t, content); got != want {
			t.Errorf("%s no lleva el cuerpo de la plantilla:\n%q\nvs\n%q", path, got, want)
		}
		if !strings.Contains(content, "description:") {
			t.Errorf("%s no lleva la descripción que el selector del cliente muestra:\n%s", path, content)
		}
		// Claude Code shows the declared name in its own selector, next to the
		// styles the human already has: it is the style name, not the file name.
		if strings.HasPrefix(path, claudePersonas) && !strings.Contains(content, "name: odachi-"+name+"\n") {
			t.Errorf("%s no declara \"name: odachi-%s\":\n%s", path, name, content)
		}
	}

	for _, want := range []string{opencodePersonas, claudePersonas, "Tab", "/config"} {
		if !strings.Contains(out, want) {
			t.Errorf("la salida no menciona %q:\n%s", want, out)
		}
	}
}

// TestInitPersonaRerunWritesNothing: the files are generated, so a second run
// touches nothing and leaves no backup behind.
func TestInitPersonaRerunWritesNothing(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	if code, _, errOut := runDir(t, dir, []string{"init"}, ""); code != 0 {
		t.Fatalf("first run code = %d (stderr %q)", code, errOut)
	}
	before := personaTree(t, dir)
	if len(before) != len(personaFiles) {
		t.Fatalf("el primer init dejó %d archivos de persona, want %d: %v", len(before), len(personaFiles), before)
	}

	code, out, errOut := runDir(t, dir, []string{"init"}, "")

	if code != 0 {
		t.Fatalf("second run code = %d (stdout %q, stderr %q)", code, out, errOut)
	}
	after := personaTree(t, dir)
	if !maps.Equal(before, after) {
		t.Errorf("el segundo init tocó las personas:\nbefore %v\nafter  %v", keys(before), keys(after))
	}
	for path := range after {
		if strings.Contains(path, ".bak-") {
			t.Errorf("el segundo init respaldó %s: un archivo idéntico no se reescribe", path)
		}
	}
	if !strings.Contains(out, "sin cambios") {
		t.Errorf("la salida no dice que no cambió nada:\n%s", out)
	}
}

// TestInitPersonaFilesSayTheyAreGenerated: the first line says the file is
// written by a command and how to change it; the second says which file owns the
// ceremony, so the human never has two sources of truth for the same word.
func TestInitPersonaFilesSayTheyAreGenerated(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	if code, _, errOut := runDir(t, dir, []string{"init"}, ""); code != 0 {
		t.Fatalf("code = %d (stderr %q)", code, errOut)
	}

	for _, path := range personaFiles {
		lines := strings.Split(personaBody(t, readPersona(t, dir, path)), "\n")
		if len(lines) < 2 {
			t.Fatalf("%s no tiene las dos primeras líneas", path)
		}

		first := strings.ToLower(lines[0])
		for _, want := range []string{"generado", "odachi init", "no lo edites"} {
			if !strings.Contains(first, want) {
				t.Errorf("%s: la primera línea no dice %q: %q", path, want, lines[0])
			}
		}

		for _, want := range []string{".odachi/profile.toml", "cómo hablás", "odachi level --to"} {
			if !strings.Contains(lines[1], want) {
				t.Errorf("%s: la segunda línea no dice %q: %q", path, want, lines[1])
			}
		}
	}
}

// TestLevelToDoesNotRewriteThePersonas: `level` moves the profile, which is the
// other axis. Both personas are already written and nothing about them depends on
// which profile is in force.
func TestLevelToDoesNotRewriteThePersonas(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	if code, _, errOut := runDir(t, dir, []string{"init"}, ""); code != 0 {
		t.Fatalf("init code = %d (stderr %q)", code, errOut)
	}
	before := personaTree(t, dir)

	code, out, errOut := runDir(t, dir, []string{"level", "--to", "senior"}, "")
	if code != 0 {
		t.Fatalf("level --to code = %d (stdout %q, stderr %q)", code, out, errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, ".odachi", "profile.toml")); err != nil {
		t.Fatalf("level --to no escribió el perfil: %v", err)
	}

	if after := personaTree(t, dir); !maps.Equal(before, after) {
		t.Errorf("level --to reescribió las personas:\nbefore %v\nafter  %v", keys(before), keys(after))
	}
}

// TestInitPersonasDoNotDependOnTheProfile: the two axes stay apart. Whatever the
// profile says, the persona files are the same bytes, because the persona is not
// read from the profile and never should be.
func TestInitPersonasDoNotDependOnTheProfile(t *testing.T) {
	novatoProfile, seniorProfile := t.TempDir(), t.TempDir()
	initRepo(t, novatoProfile)
	initRepo(t, seniorProfile)
	writeFile(t, novatoProfile, ".odachi/profile.toml", "name = \"novato\"\n")
	writeFile(t, seniorProfile, ".odachi/profile.toml", "name = \"senior\"\n")

	if code, _, errOut := runDir(t, novatoProfile, []string{"init"}, ""); code != 0 {
		t.Fatalf("init con perfil novato: code = %d (stderr %q)", code, errOut)
	}
	if code, _, errOut := runDir(t, seniorProfile, []string{"init"}, ""); code != 0 {
		t.Fatalf("init con perfil senior: code = %d (stderr %q)", code, errOut)
	}

	novato := personaTree(t, novatoProfile)
	senior := personaTree(t, seniorProfile)
	if !maps.Equal(novato, senior) {
		t.Errorf("las personas dependen del perfil:\nnovato %v\nsenior %v", keys(novato), keys(senior))
	}
	if len(novato) != len(personaFiles) {
		t.Errorf("las personas son %v, want %v", keys(novato), personaFiles)
	}
}

// TestInitJSONCarriesThePersonas: the machine output names the same files a human
// reads on the same run.
func TestInitJSONCarriesThePersonas(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)

	code, out, errOut := runDir(t, dir, []string{"init", "--json"}, "")
	if code != 0 {
		t.Fatalf("code = %d, want 0 (stderr %q)", code, errOut)
	}

	var payload struct {
		Files    []string `json:"files"`
		Personas []struct {
			Client    string   `json:"client"`
			Dir       string   `json:"dir"`
			Written   int      `json:"written"`
			Unchanged int      `json:"unchanged"`
			Files     []string `json:"files"`
		} `json:"personas"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, out)
	}

	if len(payload.Personas) != 2 {
		t.Fatalf("personas = %v, want one row per client", payload.Personas)
	}
	seen := map[string]bool{}
	for _, dest := range payload.Personas {
		if dest.Written != len(persona.Names()) {
			t.Errorf("%s/%s: written = %d, want %d", dest.Client, dest.Dir, dest.Written, len(persona.Names()))
		}
		for _, path := range dest.Files {
			seen[path] = true
			if !strings.Contains(strings.Join(payload.Files, " "), path) {
				t.Errorf("files = %v, want it to name %s", payload.Files, path)
			}
		}
	}
	for _, path := range personaFiles {
		if !seen[path] {
			t.Errorf("las personas no nombran %s: %v", path, seen)
		}
	}
}
