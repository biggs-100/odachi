package persona_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/persona"
)

// TestNormalizeSkillsAcceptsTheTwoValues: the key knows where the craft lives,
// and a human writes the file, so case and surrounding space do not matter.
func TestNormalizeSkillsAcceptsTheTwoValues(t *testing.T) {
	cases := map[string]string{
		"project":  persona.SkillsProject,
		"  HOME  ": persona.SkillsHome,
		"Project":  persona.SkillsProject,
		"home":     persona.SkillsHome,
		"":         "",
		"   ":      "",
	}
	for input, want := range cases {
		got, err := persona.NormalizeSkills(input)
		if err != nil {
			t.Fatalf("NormalizeSkills(%q) = %v, want %q", input, err, want)
		}
		if got != want {
			t.Errorf("NormalizeSkills(%q) = %q, want %q", input, got, want)
		}
	}
}

// TestNormalizeSkillsRefusesATypo: an unknown scope names the two choices, so a
// typo never silently puts the skills back in the project.
func TestNormalizeSkillsRefusesATypo(t *testing.T) {
	got, err := persona.NormalizeSkills("machine")
	if err == nil {
		t.Fatalf("NormalizeSkills(machine) = %q, want an error", got)
	}
	for _, want := range []string{"machine", "project", "home"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal %q does not name %q", err, want)
		}
	}
}
