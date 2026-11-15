package skills_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/skills"
)

// TestBootstrapIsTheFirstDoor: a project's shape is no longer bootstrap's to
// carry. The first door routes the two that carry it -the shared one and the stack
// one- and restates none of the rules `drive`, `tests` and `commits` own: a second
// owner for a rule is the defect this catalog keeps paying for (0304, 0307).
func TestBootstrapIsTheFirstDoor(t *testing.T) {
	bootstrap, ok := skills.Bootstrap()
	if !ok {
		t.Fatal("the payload has no bootstrap skill")
	}
	body := string(bootstrap.Main())
	gates := sectionBetween(body, "## Decision gates", "## Output")
	if gates == "" {
		t.Fatal("bootstrap has no decision gates: the doors it sends the reader to are unreachable")
	}

	for _, want := range []string{"arch-project", "arch-tauri"} {
		if !strings.Contains(gates, want) {
			t.Errorf("bootstrap does not route %q from its decision gates", want)
		}
	}

	// The shape left through a door of its own: a reference here would be the
	// second owner of the same material, which is what 0304 refused and 0307 ends.
	for _, file := range bootstrap.Files {
		if strings.HasPrefix(file.Path, "references/") {
			t.Errorf("bootstrap still carries %q: the shape of the code belongs to arch-project", file.Path)
		}
	}

	// The delegation is the whole point: a rule restated here would have two owners.
	for _, owned := range []string{
		"An assertion counts only after",
		"Verification is layered, in fixed order",
		"Every non-trivial decision gets an ADR",
		"conventional commits",
	} {
		if strings.Contains(body, owned) {
			t.Errorf("bootstrap restates a rule another skill owns: %q", owned)
		}
	}
	if !strings.Contains(body, "Read them there") {
		t.Error("bootstrap does not send the reader to the owner")
	}
}

// TestTheCatalogDoesNotCarryTheOldName: the foundation is not a second door. A skill
// called ai-project-foundation would collide with bootstrap's trigger, which is the
// situation its own rules call a defect (0304).
func TestTheCatalogDoesNotCarryTheOldName(t *testing.T) {
	for _, name := range skills.Names() {
		if name == "ai-project-foundation" || name == "tauri-project-foundation" {
			t.Errorf("the payload still carries %q", name)
		}
	}
}

// sectionBetween is the text from one heading to the next, or empty when either is
// missing: a routing table is what sits between its heading and the one after it.
func sectionBetween(body, from, to string) string {
	start := strings.Index(body, from)
	if start < 0 {
		return ""
	}
	rest := body[start+len(from):]
	end := strings.Index(rest, to)
	if end < 0 {
		return rest
	}
	return rest[:end]
}
