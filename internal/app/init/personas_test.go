package init_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/persona"
)

// TestInitWritesBothPersonasForEveryClient: `odachi init` leaves each client that
// has a native mechanism with both personas, so the human can pick one with the
// client's own switch instead of learning a TOML.
func TestInitWritesBothPersonasForEveryClient(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	got := run(files)

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}

	wrote := persona.All()
	if len(wrote) == 0 {
		t.Fatalf("persona.All() vino vacío: no hay nada que escribir")
	}
	for _, file := range wrote {
		content, ok := files.get(file.Path)
		if !ok {
			t.Errorf("%s no se escribió; events: %v", file.Path, files.events)
			continue
		}
		if content != file.Content {
			t.Errorf("%s no tiene el cuerpo de la plantilla:\n%q\nvs\n%q", file.Path, content, file.Content)
		}
		if !slices.Contains(got.Files, file.Path) {
			t.Errorf("Files = %v, want it to name %s", got.Files, file.Path)
		}
	}

	if len(got.PersonaDests) != len(persona.Clients()) {
		t.Fatalf("PersonaDests = %v, want one row per client (%d)", got.PersonaDests, len(persona.Clients()))
	}
	joined := strings.Join(got.Lines, "\n")
	for _, client := range persona.Clients() {
		if !strings.Contains(joined, client.Dir) {
			t.Errorf("las líneas no nombran %s:\n%s", client.Dir, joined)
		}
		if !strings.Contains(joined, client.Switch) {
			t.Errorf("las líneas no dicen cómo se elige la persona en %s (%q):\n%s", client.Agent, client.Switch, joined)
		}
	}
}

// TestInitPersonaRerunWritesNothing: the files are generated, so a second run
// rewrites nothing and does not back up a file it just wrote.
func TestInitPersonaRerunWritesNothing(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
	if code := run(files).Code; code != 0 {
		t.Fatalf("first run code = %d", code)
	}

	files.events = nil
	got := run(files)

	if got.Code != 0 {
		t.Fatalf("second run code = %d (message %q)", got.Code, got.Message)
	}
	if len(files.events) != 0 {
		t.Errorf("second run touched %v, want no writes and no backups", files.events)
	}
	if len(got.Files) != 0 {
		t.Errorf("Files = %v, want none on a rerun that wrote nothing", got.Files)
	}
	for _, dest := range got.PersonaDests {
		if dest.Written != 0 || dest.Kept != len(persona.Names()) {
			t.Errorf("PersonaDests[%s] = %+v, want every persona kept", dest.Dir, dest)
		}
	}
	joined := strings.Join(got.Lines, "\n")
	if !strings.Contains(joined, "sin cambios") {
		t.Errorf("las líneas no dicen que no cambió nada:\n%s", joined)
	}
	if !strings.Contains(got.Message, "personas") {
		t.Errorf("Message = %q, want it to say the personas were already there", got.Message)
	}
}

// TestInitBacksUpAHandEditedPersona: the file says it is generated, so a hand
// edit is not thrown away silently either.
func TestInitBacksUpAHandEditedPersona(t *testing.T) {
	files := newFakeFiles(map[string]string{
		".git/HEAD":                         "ref: refs/heads/main\n",
		".opencode/agents/odachi-novato.md": "mi versión a mano\n",
	})

	got := run(files)

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	if !slices.Contains(files.events, "backup .opencode/agents/odachi-novato.md") {
		t.Errorf("events = %v, want a backup of the hand-edited persona", files.events)
	}
	if indexOf(files.events, "backup .opencode/agents/odachi-novato.md") > indexOf(files.events, "write .opencode/agents/odachi-novato.md") {
		t.Errorf("events = %v, want the backup before the write", files.events)
	}
	content, _ := files.get(".opencode/agents/odachi-novato.md")
	body, err := persona.Body(persona.Novato)
	if err != nil {
		t.Fatalf("Body(novato): %v", err)
	}
	if !strings.HasSuffix(content, body) {
		t.Errorf("la persona no se regeneró:\n%q", content)
	}
}
