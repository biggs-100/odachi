package skills_test

import (
	"strings"
	"testing"

	skillsapp "github.com/biggs-100/odachi/internal/app/skills"
)

// TestTheCheckRefusesAFileTheInstallerWouldDrop: `odachi pack` writes SKILL.md,
// references/ and assets/ only, so a file at the skill's root is embedded and
// never installed. A body that names it points at a file the project does not
// have - 0252 shipped a NOTICE.md that way, and the installed skill said
// "NOTICE.md names it" about a file nobody had.
func TestTheCheckRefusesAFileTheInstallerWouldDrop(t *testing.T) {
	files := newFakeFiles(map[string]string{
		"skills/router/SKILL.md":  referenceSkillText("router"),
		"skills/router/NOTICE.md": "# Notice\n",
		"SKILLS.md":               registryOf("router"),
	})

	got := skillsapp.Run(skillsapp.Options{Dir: dir, Check: true}, files)

	if got.Code == 0 {
		t.Fatalf("Code = 0, want a refusal (lines %q)", got.Lines)
	}
	joined := strings.Join(got.Lines, "\n")
	for _, want := range []string{"skills/router/NOTICE.md", "installer", "never see it"} {
		if !strings.Contains(joined, want) {
			t.Errorf("lines %q do not name %q", joined, want)
		}
	}
}
