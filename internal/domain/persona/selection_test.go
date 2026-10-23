package persona_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/persona"
)

// TestNormalizeSelectionAcceptsTheThreeValues: the flag and the file speak the
// same three words, and a human writing them by hand may use any case or
// surrounding space.
func TestNormalizeSelectionAcceptsTheThreeValues(t *testing.T) {
	cases := map[string]string{
		"both":        persona.SelectionBoth,
		"  OPENCODE ": persona.SelectionOpenCode,
		"Claude":      persona.SelectionClaude,
	}

	for input, want := range cases {
		got, err := persona.NormalizeSelection(input)
		if err != nil {
			t.Fatalf("NormalizeSelection(%q) = %v", input, err)
		}
		if got != want {
			t.Errorf("NormalizeSelection(%q) = %q, want %q", input, got, want)
		}
	}
}

// TestNormalizeSelectionRefusesAGuess: a typo must name the three values
// instead of silently choosing a client nobody asked for.
func TestNormalizeSelectionRefusesAGuess(t *testing.T) {
	for _, input := range []string{"", "opencode,claude", "all", "todos", "open code"} {
		if got, err := persona.NormalizeSelection(input); err == nil {
			t.Errorf("NormalizeSelection(%q) = %q, want an error", input, got)
		}
	}
}

// TestResolveSelectionDefaultsToBoth: no file and no flag is the behavior every
// existing project already has, so the zero value cannot mean "none".
func TestResolveSelectionDefaultsToBoth(t *testing.T) {
	name, clients, err := persona.ResolveSelection("")
	if err != nil {
		t.Fatalf("ResolveSelection(\"\") = %v", err)
	}
	if name != persona.SelectionBoth {
		t.Errorf("name = %q, want %q", name, persona.SelectionBoth)
	}
	if len(clients) != len(persona.Clients()) {
		t.Errorf("clients = %d, want %d", len(clients), len(persona.Clients()))
	}
}

// TestResolveSelectionReturnsOnlyTheChosenClient: the selection filters the same
// table the rest of the tool reads, so a new client cannot appear in one place
// and not the other.
func TestResolveSelectionReturnsOnlyTheChosenClient(t *testing.T) {
	cases := map[string]string{
		persona.SelectionOpenCode: ".opencode/agents",
		persona.SelectionClaude:   ".claude/output-styles",
	}

	for selection, wantDir := range cases {
		name, clients, err := persona.ResolveSelection(selection)
		if err != nil {
			t.Fatalf("ResolveSelection(%q) = %v", selection, err)
		}
		if name != selection {
			t.Errorf("name = %q, want %q", name, selection)
		}
		if len(clients) != 1 {
			t.Fatalf("ResolveSelection(%q) = %d clients, want 1", selection, len(clients))
		}
		if clients[0].Dir != wantDir {
			t.Errorf("ResolveSelection(%q)[0].Dir = %q, want %q", selection, clients[0].Dir, wantDir)
		}
	}
}

// TestFilesForSelectionFiltersTheMirror: the refresh verbs read the same table
// through the selection, so a client that was not chosen keeps its bytes.
func TestFilesForSelectionFiltersTheMirror(t *testing.T) {
	files, err := persona.FilesForSelection("bloque\n", persona.SelectionClaude)
	if err != nil {
		t.Fatalf("FilesForSelection(claude) = %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("FilesForSelection(claude) = %d files, want 1", len(files))
	}
	if files[0].Path != ".claude/output-styles/odachi.md" {
		t.Errorf("path = %q, want the Claude mirror", files[0].Path)
	}

	both, err := persona.FilesForSelection("bloque\n", "")
	if err != nil {
		t.Fatalf("FilesForSelection(\"\") = %v", err)
	}
	if len(both) != len(persona.Files("bloque\n")) {
		t.Errorf("FilesForSelection(\"\") = %d files, want both", len(both))
	}
}

// TestResolveSelectionNamesTheFile: the error a human reads must point at the
// file they can fix, not at a flag they have to remember.
func TestResolveSelectionNamesTheFile(t *testing.T) {
	_, _, err := persona.ResolveSelection("inventado")
	if err == nil {
		t.Fatal("ResolveSelection accepted an unknown client")
	}
	for _, want := range []string{"both", "opencode", "claude"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %q, want %q", err, want)
		}
	}
}
