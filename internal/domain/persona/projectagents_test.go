package persona_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/persona"
)

// TestNormalizeProjectAgentsAcceptsTheTwoValues: the key knows where the native
// agent files live, and nothing else. Matching ignores case and space because a
// human edits the store by hand.
func TestNormalizeProjectAgentsAcceptsTheTwoValues(t *testing.T) {
	cases := map[string]string{
		"none":                 persona.ProjectAgentsNone,
		"  WRITE  ":            persona.ProjectAgentsWrite,
		"None":                 persona.ProjectAgentsNone,
		"write":                persona.ProjectAgentsWrite,
		strings.Repeat(" ", 3): "",
	}

	for input, want := range cases {
		got, err := persona.NormalizeProjectAgents(input)
		if err != nil {
			t.Fatalf("NormalizeProjectAgents(%q) = %v", input, err)
		}
		if got != want {
			t.Errorf("NormalizeProjectAgents(%q) = %q, want %q", input, got, want)
		}
	}
}

// TestNormalizeProjectAgentsRefusesATypo: an unknown value names the two choices
// instead of guessing one, because guessing here decides whether a project
// carries its agent at all.
func TestNormalizeProjectAgentsRefusesATypo(t *testing.T) {
	got, err := persona.NormalizeProjectAgents("skipped")
	if err == nil {
		t.Fatalf("NormalizeProjectAgents(skipped) = %q, want an error", got)
	}
	for _, want := range []string{"write", "none", "skipped"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %q, want it to name %q", err, want)
		}
	}
}
