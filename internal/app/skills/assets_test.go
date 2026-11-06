package skills_test

import (
	"slices"
	"strings"
	"testing"

	skillsapp "github.com/biggs-100/odachi/internal/app/skills"
)

// assetSkillText is a skill that declares the assets it carries: the output
// contract the grammar asks for, and the inventory.
func assetSkillText(name string, assets ...string) string {
	var out strings.Builder
	out.WriteString("---\nname: " + name + "\ndescription: Trigger: trigger de " + name + "\n---\n\n")
	out.WriteString("## When\n\ncuando\n\n## Rules\n\nregla\n\n## Never\n\nnunca\n\n")
	out.WriteString("## Output\n\nthe file and its tradeoff.\n\n")
	out.WriteString("## Assets\n\n")
	for _, path := range assets {
		out.WriteString("- `" + path + "` — what it is.\n")
	}
	return out.String()
}

// TestTheLoaderReadsTheAssetsFolder: the check sees the template the folder
// carries, and a coherent folder is green.
func TestTheLoaderReadsTheAssetsFolder(t *testing.T) {
	files := newFakeFiles(map[string]string{
		"skills/router/SKILL.md":           assetSkillText("router", "assets/doc.md", "assets/schema.json"),
		"skills/router/assets/doc.md":      "# template\n",
		"skills/router/assets/schema.json": "{}\n",
		"SKILLS.md":                        registryOf("router"),
	})

	got := skillsapp.Run(skillsapp.Options{Dir: dir, Check: true}, files)

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (lines %q)", got.Code, got.Lines)
	}
	if len(got.Skills) != 1 {
		t.Fatalf("Skills = %d, want 1", len(got.Skills))
	}
	if want := []string{"assets/doc.md", "assets/schema.json"}; !slices.Equal(got.Skills[0].Assets, want) {
		t.Errorf("Assets = %q, want %q", got.Skills[0].Assets, want)
	}
}

// TestTheLoaderRefusesATreeUnderAssets: an asset is a file, not a folder; a tree
// there is a shape nobody copies.
func TestTheLoaderRefusesATreeUnderAssets(t *testing.T) {
	files := newFakeFiles(map[string]string{
		"skills/router/SKILL.md":             assetSkillText("router", "assets/doc.md"),
		"skills/router/assets/nested/doc.md": "x\n",
		"SKILLS.md":                          registryOf("router"),
	})

	got := skillsapp.Run(skillsapp.Options{Dir: dir, Check: true}, files)

	if got.Code != 1 {
		t.Fatalf("Code = %d, want 1 (message %q)", got.Code, got.Message)
	}
	joined := strings.Join(got.Lines, "\n")
	if !strings.Contains(joined, "skills/router/assets/nested") || !strings.Contains(joined, "not a tree") {
		t.Errorf("lines %q do not name the tree under assets/", joined)
	}
}
