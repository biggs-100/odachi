package skills_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/skills"
)

// TestArchProjectIsTheSecondDoor: the shared half of a project's architecture has
// a door of its own, one level under `bootstrap` and one above `arch-<stack>`. The
// rules are pinned because they are the value: a door that lost the guard test or
// the twelve decisions would still read like a foundation (0307).
func TestArchProjectIsTheSecondDoor(t *testing.T) {
	body, ok := archProject(t)
	if !ok {
		t.Fatal("the payload has no arch-project skill: a project's shape has no door of its own")
	}

	for _, want := range []string{
		"bootstrap",
		"arch-tauri",
		"Dependencies point inward",
		"One owner per business rule",
		"exactly one domain place",
		"enforced mechanically",
		"an unanswered line is a decision taken by accident",
		"not evidence",
		"references/layout.md",
		"references/stack-commands.md",
		"references/go-htmx.md",
		"assets/adr-template.md",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("arch-project does not say %q", want)
		}
	}
}

// TestArchProjectRoutesEveryReference: a reference is reachable only through its
// row of the decision gates, and a file the folder holds without a bullet is a file
// nobody reads. Both directions are checked, because the grammar refuses either the
// dead link or the orphan (0307).
func TestArchProjectRoutesEveryReference(t *testing.T) {
	body, ok := archProject(t)
	if !ok {
		t.Fatal("the payload has no arch-project skill")
	}
	skill := mustSkill(t, "arch-project")

	gates := sectionBetween(body, "## Decision gates", "## Output")
	listed := sectionBetween(body, "## References", "## Assets")
	if gates == "" {
		t.Fatal("arch-project has no decision gates: its references are unreachable")
	}

	references := 0
	for _, file := range skill.Files {
		if !strings.HasPrefix(file.Path, "references/") {
			continue
		}
		references++
		if !strings.Contains(gates, "`"+file.Path+"`") {
			t.Errorf("%s is not a row of ## Decision gates", file.Path)
		}
		if !strings.Contains(listed, "`"+file.Path+"`") {
			t.Errorf("%s is not a bullet of ## References", file.Path)
		}
	}
	if references == 0 {
		t.Error("arch-project carries no references: a door with no depth is a paragraph")
	}

	// An asset is a file the skill hands over, not text it reads: it is inventoried
	// the same way and never routed from the gates.
	assets := ""
	if start := strings.Index(body, "## Assets"); start >= 0 {
		assets = body[start:]
	}
	for _, file := range skill.Files {
		if !strings.HasPrefix(file.Path, "assets/") {
			continue
		}
		if !strings.Contains(assets, "`"+file.Path+"`") {
			t.Errorf("%s is not a bullet of ## Assets", file.Path)
		}
	}
}

// archProject is the body of the shared door, and whether the payload has it.
func archProject(t *testing.T) (string, bool) {
	t.Helper()
	for _, s := range mustAll(t) {
		if s.Name == "arch-project" {
			return string(s.Main()), true
		}
	}
	return "", false
}

// mustSkill finds one bundled skill by name.
func mustSkill(t *testing.T, name string) skills.Skill {
	t.Helper()
	for _, s := range mustAll(t) {
		if s.Name == name {
			return s
		}
	}
	t.Fatalf("the payload has no %s skill", name)
	return skills.Skill{}
}

// mustAll reads the bundled payload or stops the test: a payload that cannot be
// read proves nothing about what it carries.
func mustAll(t *testing.T) []skills.Skill {
	t.Helper()
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	return all
}
