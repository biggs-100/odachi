package skills_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/skills"
)

// TestDesignDeclaresItsOutput: the delivery skill says what it hands back, and
// the formats of the folder it writes live in one reference so the body keeps
// its sixty-line budget.
func TestDesignDeclaresItsOutput(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}

	var design skills.Skill
	for _, s := range all {
		if s.Name == "design" {
			design = s
		}
	}
	if design.Name == "" {
		t.Fatal("design is not in the payload")
	}

	want := []string{"SKILL.md", "references/files.md"}
	got := make([]string, 0, len(design.Files))
	for _, file := range design.Files {
		got = append(got, file.Path)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("design carries %v, want %v", got, want)
	}

	main := string(design.Main())
	for _, section := range []string{"## Output", "## References", "`references/files.md`"} {
		if !strings.Contains(main, section) {
			t.Errorf("the design skill does not say %q", section)
		}
	}
	files := string(design.Files[1].Content)
	for _, want := range []string{"design.json", "comments.json", "screenshot"} {
		if !strings.Contains(files, want) {
			t.Errorf("references/files.md does not say %q", want)
		}
	}
}
