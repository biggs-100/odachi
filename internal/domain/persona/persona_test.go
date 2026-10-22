package persona_test

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/agent"
	"github.com/biggs-100/odachi/internal/domain/persona"
)

// splitFrontmatter separates the wrapper one client needs from the body every
// client shares. A file without frontmatter is all body.
func splitFrontmatter(t *testing.T, content string) (front, body string) {
	t.Helper()
	if !strings.HasPrefix(content, "---\n") {
		return "", content
	}
	end := strings.Index(content, "\n---\n")
	if end < 0 {
		t.Fatalf("frontmatter has no closing marker:\n%s", content)
	}
	return content[:end+len("\n---\n")], content[end+len("\n---\n"):]
}

// TestBodyIsTheSameForEveryClient is the whole point of one template: the voice
// is written once, and only the wrapper around it can differ.
func TestBodyIsTheSameForEveryClient(t *testing.T) {
	for _, name := range persona.Names() {
		want, err := persona.Body(name)
		if err != nil {
			t.Fatalf("Body(%q): %v", name, err)
		}

		for _, client := range persona.Clients() {
			files, ok := persona.Files(client.Agent)
			if !ok {
				t.Fatalf("%s tiene mecanismo nativo y Files() dijo que no", client.Agent)
			}
			var found bool
			for _, file := range files {
				if !strings.HasSuffix(file.Path, "odachi-"+name+".md") {
					continue
				}
				found = true
				_, body := splitFrontmatter(t, file.Content)
				if body != want {
					t.Errorf("%s/%s: el cuerpo difiere del de Body(%q):\n%q\nvs\n%q", client.Agent, name, name, body, want)
				}
			}
			if !found {
				t.Errorf("%s no recibió la persona %q: %v", client.Agent, name, files)
			}
		}
	}
}

// TestEveryFileStartsWithTheGeneratedNoteAndTheNonDivergence: the first line
// says the file is generated and that a hand edit is lost work; the second says
// which source of truth owns the ceremony.
func TestEveryFileStartsWithTheGeneratedNoteAndTheNonDivergence(t *testing.T) {
	files := persona.All()
	if len(files) == 0 {
		t.Fatalf("All() vino vacío: no hay ninguna persona que escribir")
	}

	for _, file := range files {
		_, body := splitFrontmatter(t, file.Content)
		lines := strings.Split(body, "\n")
		if len(lines) < 2 {
			t.Fatalf("%s: no tiene las dos primeras líneas:\n%s", file.Path, body)
		}

		first := strings.ToLower(lines[0])
		for _, want := range []string{"generado", "odachi init", "no lo edites"} {
			if !strings.Contains(first, want) {
				t.Errorf("%s: la primera línea no dice %q: %q", file.Path, want, lines[0])
			}
		}

		second := lines[1]
		for _, want := range []string{".odachi/profile.toml", "odachi level --to"} {
			if !strings.Contains(second, want) {
				t.Errorf("%s: la segunda línea no nombra %q: %q", file.Path, want, second)
			}
		}
		if !strings.Contains(second, "cómo hablás") {
			t.Errorf("%s: la segunda línea no dice que este archivo es solo cómo hablás: %q", file.Path, second)
		}
	}
}

// TestNovatoTeachesAndSeniorGoesToThePoint: the two personas are two voices, not
// two spellings of the same text.
func TestNovatoTeachesAndSeniorGoesToThePoint(t *testing.T) {
	novato, err := persona.Body(persona.Novato)
	if err != nil {
		t.Fatalf("Body(novato): %v", err)
	}
	senior, err := persona.Body(persona.Senior)
	if err != nil {
		t.Fatalf("Body(senior): %v", err)
	}

	for _, want := range []string{"aprenda el oficio", "odachi ask --craft", "odachi summary", "preguntá de vuelta", "sermón"} {
		if !strings.Contains(novato, want) {
			t.Errorf("la persona novato no contiene %q:\n%s", want, novato)
		}
	}
	for _, want := range []string{"salvo que la pidan", "breve", "prueba"} {
		if !strings.Contains(senior, want) {
			t.Errorf("la persona senior no contiene %q:\n%s", want, senior)
		}
	}
	if strings.Contains(senior, "odachi ask --craft") {
		t.Errorf("la persona senior repite la ceremonia del novato, que la decide el perfil:\n%s", senior)
	}
}

// TestUnknownPersonaIsRefusedNamingTheTwo: nothing is guessed for a name that is
// not in the table.
func TestUnknownPersonaIsRefusedNamingTheTwo(t *testing.T) {
	_, err := persona.Body("principiante")
	if err == nil {
		t.Fatalf("Body(principiante) devolvió una persona que no existe")
	}
	for _, want := range []string{"principiante", persona.Novato, persona.Senior} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("el error no nombra %q: %v", want, err)
		}
	}
}

// TestUnknownClientGetsNothing: a client odachi does not know gets no file and no
// guess.
func TestUnknownClientGetsNothing(t *testing.T) {
	for _, client := range []string{"neovim", "emacs", "", "  "} {
		files, ok := persona.Files(client)
		if ok {
			t.Errorf("Files(%q) dijo que sí", client)
		}
		if len(files) != 0 {
			t.Errorf("Files(%q) = %v, want none", client, files)
		}
	}
}

// TestClientsAreTheAgentsTheSkillsTableKnows keeps one list of client names in
// the repo: the persona table hangs off the agent table instead of inventing
// spellings of its own.
func TestClientsAreTheAgentsTheSkillsTableKnows(t *testing.T) {
	clients := persona.Clients()
	if len(clients) == 0 {
		t.Fatalf("Clients() vino vacío")
	}

	for _, client := range clients {
		if client.Agent == "ci" {
			t.Errorf("ci tiene persona: en CI nadie elige un agente")
		}
		if _, err := agent.Resolve(client.Agent); err != nil {
			t.Errorf("el cliente %q no está en la tabla de agentes: %v", client.Agent, err)
		}
		if client.Dir == "" || strings.HasPrefix(client.Dir, "/") || strings.Contains(client.Dir, "\\") {
			t.Errorf("%s: Dir = %q, want a project-relative directory with forward slashes", client.Agent, client.Dir)
		}
		if client.Switch == "" {
			t.Errorf("%s: Switch está vacío: el humano no sabría con qué tecla elige la persona", client.Agent)
		}
	}
}

// TestOpenCodePersonaIsPrimary: OpenCode only offers it in its own selector when
// the frontmatter says primary.
func TestOpenCodePersonaIsPrimary(t *testing.T) {
	files, ok := persona.Files("opencode")
	if !ok {
		t.Fatalf("opencode no tiene persona")
	}
	for _, file := range files {
		front, _ := splitFrontmatter(t, file.Content)
		for _, want := range []string{"description:", "mode: primary"} {
			if !strings.Contains(front, want) {
				t.Errorf("%s: el frontmatter no tiene %q:\n%s", file.Path, want, front)
			}
		}
		if !strings.HasPrefix(file.Path, ".opencode/agents/") {
			t.Errorf("%s: OpenCode lee los agentes del proyecto de .opencode/agents/", file.Path)
		}
	}
}

// TestClaudePersonaKeepsTheCodingInstructions: an output style replaces the
// system prompt unless it says to keep the coding instructions, and the persona
// is about how the agent talks, never about dropping its tools.
func TestClaudePersonaKeepsTheCodingInstructions(t *testing.T) {
	files, ok := persona.Files("claude")
	if !ok {
		t.Fatalf("claude no tiene persona")
	}
	for _, file := range files {
		front, _ := splitFrontmatter(t, file.Content)
		// Claude Code shows `name` in its selector, so it has to be the style
		// name a human types, not the file name with its extension.
		style := strings.TrimSuffix(filepath.Base(file.Path), ".md")
		declared := false
		for _, line := range strings.Split(front, "\n") {
			if strings.TrimSpace(line) == "name: "+style {
				declared = true
			}
		}
		if !declared {
			t.Errorf("%s: el frontmatter no declara \"name: %s\":\n%s", file.Path, style, front)
		}
		for _, want := range []string{"description:", "keep-coding-instructions: true"} {
			if !strings.Contains(front, want) {
				t.Errorf("%s: el frontmatter no tiene %q:\n%s", file.Path, want, front)
			}
		}
		if !strings.HasPrefix(file.Path, ".claude/output-styles/") {
			t.Errorf("%s: Claude Code lee los estilos de .claude/output-styles/", file.Path)
		}
	}
}

// TestAllIsEveryClientTimesEveryPersona: what init writes, in one call.
func TestAllIsEveryClientTimesEveryPersona(t *testing.T) {
	files := persona.All()
	want := len(persona.Clients()) * len(persona.Names())
	if len(files) != want {
		t.Fatalf("All() tiene %d archivos, want %d", len(files), want)
	}

	seen := map[string]bool{}
	for _, file := range files {
		if seen[file.Path] {
			t.Errorf("%s aparece dos veces", file.Path)
		}
		seen[file.Path] = true
		if !strings.HasSuffix(file.Path, ".md") {
			t.Errorf("%s: una persona es un .md", file.Path)
		}
	}
}

// TestTheRendererDoesNotReadTheProfile: the personas do not depend on
// .odachi/profile.toml. If the two axes ever get wired together, this fails.
func TestTheRendererDoesNotReadTheProfile(t *testing.T) {
	out, err := exec.Command("go", "list", "-f", "{{join .Imports \"\\n\"}}", ".").CombinedOutput()
	if err != nil {
		t.Skipf("go list no corrió en este entorno: %v: %s", err, out)
	}
	for _, path := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if strings.Contains(path, "profile") {
			t.Errorf("el renderer de personas importa %q: el ceremonial y la voz son dos ejes y no se tocan", path)
		}
	}
}
