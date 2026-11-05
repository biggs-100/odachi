package skills_test

import (
	"slices"
	"strings"
	"testing"

	skillsapp "github.com/biggs-100/odachi/internal/app/skills"
)

// referenceSkillText is a skill that declares the files it carries: the output
// contract the grammar asks for, and the inventory.
func referenceSkillText(name string, paths ...string) string {
	var out strings.Builder
	out.WriteString("---\nname: " + name + "\ndescription: Trigger: trigger de " + name + "\n---\n\n")
	out.WriteString("## When\n\ncuando\n\n## Rules\n\nregla\n\n## Never\n\nnunca\n\n")
	out.WriteString("## Output\n\nthe file and its tradeoff.\n\n")
	out.WriteString("## References\n\n")
	for _, path := range paths {
		out.WriteString("- `" + path + "` — what it holds.\n")
	}
	return out.String()
}

// TestTheLoaderReadsTheReferencesFolder: the check sees the files the folder
// carries, not only the SKILL.md, and a coherent folder is green.
func TestTheLoaderReadsTheReferencesFolder(t *testing.T) {
	files := newFakeFiles(map[string]string{
		"skills/router/SKILL.md":              referenceSkillText("router", "references/numbers.md"),
		"skills/router/references/numbers.md": "the scale\n",
		"SKILLS.md":                           registryOf("router"),
	})

	got := skillsapp.Run(skillsapp.Options{Dir: dir, Check: true}, files)

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (lines %q)", got.Code, got.Lines)
	}
	if len(got.Skills) != 1 {
		t.Fatalf("Skills = %d, want 1", len(got.Skills))
	}
	if want := []string{"references/numbers.md"}; !slices.Equal(got.Skills[0].References, want) {
		t.Errorf("References = %q, want %q", got.Skills[0].References, want)
	}
}

// TestCheckReportsAReferenceTheFolderDoesNotHave: a path in the body is a claim,
// and the check answers with the file nobody wrote.
func TestCheckReportsAReferenceTheFolderDoesNotHave(t *testing.T) {
	files := newFakeFiles(map[string]string{
		"skills/router/SKILL.md": referenceSkillText("router", "references/numbers.md"),
		"SKILLS.md":              registryOf("router"),
	})

	got := skillsapp.Run(skillsapp.Options{Dir: dir, Check: true}, files)

	if got.Code != 1 {
		t.Fatalf("Code = %d, want 1 (message %q)", got.Code, got.Message)
	}
	joined := strings.Join(got.Lines, "\n")
	if !strings.Contains(joined, "references/numbers.md") || !strings.Contains(joined, "does not exist") {
		t.Errorf("lines %q do not name the missing reference", joined)
	}
}

// TestAReferencesEntryThatIsNotAFileIsReported: references/ holds files, and a
// folder there is a shape nobody reads.
func TestAReferencesEntryThatIsNotAFileIsReported(t *testing.T) {
	cases := []struct {
		name string
		seed map[string]string
		file string
	}{
		{
			name: "a folder",
			seed: map[string]string{
				"skills/router/SKILL.md":                   skillText("router", "trigger de router"),
				"skills/router/references/nested/notes.md": "x\n",
			},
			file: "skills/router/references/nested",
		},
		{
			name: "a file named references",
			seed: map[string]string{
				"skills/router/SKILL.md":   skillText("router", "trigger de router"),
				"skills/router/references": "not a folder\n",
			},
			file: "skills/router/references",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.seed["SKILLS.md"] = registryOf("router")
			got := skillsapp.Run(skillsapp.Options{Dir: dir, Check: true}, newFakeFiles(tc.seed))

			if got.Code != 1 {
				t.Fatalf("Code = %d, want 1 (message %q)", got.Code, got.Message)
			}
			joined := strings.Join(got.Lines, "\n")
			if !strings.Contains(joined, tc.file) {
				t.Errorf("lines %q do not name %s", joined, tc.file)
			}
		})
	}
}
