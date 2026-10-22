package init_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/rules"
)

// The block is the only text odachi writes that an agent obeys, and it has one
// owner: internal/domain/rules, where it is pinned byte for byte. This test is
// the other half of that —init must write those exact bytes, not a second copy of
// the rules— because two copies of the rules is an agent reading one while the
// binary enforces the other.
func TestInitWritesTheBlockFromTheRenderer(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	if got := run(files); got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}

	want, err := rules.Block(persona.Novato)
	if err != nil {
		t.Fatalf("rules.Block(novato): %v", err)
	}
	got, ok := files.get("AGENTS.md")
	if !ok {
		t.Fatalf("AGENTS.md no se escribió; events: %v", files.events)
	}
	if got != want {
		t.Fatalf("el bloque escrito no es el del renderer:\n%q\nvs\n%q", got, want)
	}
	if strings.Count(got, rules.Begin) != 1 || strings.Count(got, rules.End) != 1 {
		t.Errorf("el bloque aparece más de una vez:\n%s", got)
	}
}

// The block keeps the four hard rules and the six verbs, in the order a human
// learns them. The bytes are pinned in the renderer's test; this reads the
// written file the way an agent would.
func TestTheWrittenBlockKeepsTheFourRulesAndTheSixVerbs(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
	run(files)
	content, _ := files.get("AGENTS.md")

	wantVerbs := []string{"start", "ask", "level", "summary", "check", "init"}
	verbs := []string{}
	for _, line := range strings.Split(content, "\n") {
		if !strings.HasPrefix(line, "- `odachi ") {
			continue
		}
		rest := strings.TrimPrefix(line, "- `odachi ")
		name := rest
		if i := strings.IndexAny(rest, " `"); i >= 0 {
			name = rest[:i]
		}
		verbs = append(verbs, name)
	}
	if strings.Join(verbs, ",") != strings.Join(wantVerbs, ",") {
		t.Fatalf("los verbos son %v, se esperaban %v", verbs, wantVerbs)
	}

	for _, want := range []string{
		"1. Brief antes de código",
		"2. Nada fuera del cerco",
		"3. No decir \"listo\" sin prueba",
		"4. Pedazos chicos, un commit cada uno",
		rules.Section,
	} {
		if !strings.Contains(content, want) {
			t.Errorf("falta %q en el bloque:\n%s", want, content)
		}
	}
}
