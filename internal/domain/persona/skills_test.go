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

// TestTheScopeIsDecidedByTheFlagThenTheStoreThenTheAdoption: one definition, in
// the order the four facts weigh. `init` and `role` ask this function, so the two
// verbs cannot answer differently about the same project.
func TestTheScopeIsDecidedByTheFlagThenTheStoreThenTheAdoption(t *testing.T) {
	cases := []struct {
		name   string
		facts  persona.SkillsScope
		scope  string
		source string
	}{
		{"the flag wins over everything", persona.SkillsScope{Stored: persona.SkillsHome, Flag: " Project ", Adopted: true}, persona.SkillsProject, persona.SkillsFromFlag},
		{"the store is the second answer", persona.SkillsScope{Stored: persona.SkillsHome}, persona.SkillsHome, persona.SkillsFromStore},
		{"a store that chose the project is not moved by the adoption", persona.SkillsScope{Stored: persona.SkillsProject, Adopted: true}, persona.SkillsProject, persona.SkillsFromStore},
		{"the adoption answers when nobody chose", persona.SkillsScope{Adopted: true}, persona.SkillsHome, persona.SkillsFromAdoption},
		{"a project that never chose keeps the project", persona.SkillsScope{}, persona.SkillsProject, persona.SkillsFromDefault},
	}
	for _, c := range cases {
		scope, source := c.facts.Resolve()
		if scope != c.scope || source != c.source {
			t.Errorf("%s: Resolve() = (%q, %q), want (%q, %q)", c.name, scope, source, c.scope, c.source)
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
