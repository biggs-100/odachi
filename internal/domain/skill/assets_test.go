package skill_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/skill"
)

// assetsSection renders the inventory assets ask for: one bullet per file.
func assetsSection(paths ...string) string {
	lines := []string{"## Assets", ""}
	for _, path := range paths {
		lines = append(lines, "- "+code(path)+" — what it is.")
	}
	return strings.Join(lines, "\n")
}

// withAssets parses a skill under the name router with the files its assets/
// folder holds, the way the loader fills them.
func withAssets(assets []string, sections ...string) skill.Skill {
	s := parse("router", "---\nname: router\ndescription: Trigger: cuando hay que copiar\n---\n\n"+bodyWith(sections...))
	s.Assets = assets
	return s
}

// TestASkillWithAssetsIsCoherent: an asset is a file the skill declares and the
// folder holds; the grammar accepts that pair without asking for anything else.
func TestASkillWithAssetsIsCoherent(t *testing.T) {
	s := withAssets(
		[]string{"assets/doc.md"},
		outputSection,
		assetsSection("assets/doc.md"),
	)

	if problems := s.Problems(); len(problems) != 0 {
		t.Errorf("a coherent skill with an asset has problems: %s", joined(problems))
	}
}

// TestAnAssetTheSkillNamesMustExist: a path in the body is a claim about the
// folder, and a claim without a file is refused.
func TestAnAssetTheSkillNamesMustExist(t *testing.T) {
	s := withAssets(nil, outputSection, assetsSection("assets/doc.md"))

	wants(t, s.Problems(), "does not exist")
}

// TestAFileUnderAssetsMustBeListed: a template nobody declares is a file the
// agent will never copy, so the folder and the skill must agree.
func TestAFileUnderAssetsMustBeListed(t *testing.T) {
	s := withAssets(
		[]string{"assets/doc.md"},
		outputSection,
		"## Assets\n\n- something else.",
	)

	wants(t, s.Problems(), "is not a bullet of ## Assets")
}

// TestASkillWithAssetsDeclaresThem: the folder has a file, so the body owes the
// inventory that routes it.
func TestASkillWithAssetsDeclaresThem(t *testing.T) {
	s := withAssets([]string{"assets/doc.md"}, outputSection)

	wants(t, s.Problems(), `"Assets"`)
}

// TestAnAssetMayBeAnyFile: the contract does not guess what a template is made
// of, so a schema is as valid as a markdown skeleton.
func TestAnAssetMayBeAnyFile(t *testing.T) {
	s := withAssets(
		[]string{"assets/schema.json"},
		outputSection,
		assetsSection("assets/schema.json"),
	)

	if problems := s.Problems(); len(problems) != 0 {
		t.Errorf("an asset with another extension has problems: %s", joined(problems))
	}
}
