package skills_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/skills"
)

// TestUICraftIsARouter: the first skill that carries depth keeps its body and
// ships four references, and every one is named where the agent will look for it.
func TestUICraftIsARouter(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}

	var router skills.Skill
	for _, s := range all {
		if s.Name == "ui-craft" {
			router = s
		}
	}
	if router.Name == "" {
		t.Fatal("ui-craft is not in the payload")
	}

	want := []string{
		"SKILL.md",
		"references/browse-and-compare.md",
		"references/motion.md",
		"references/states-and-copy.md",
		"references/task-and-content.md",
		"references/the-numbers.md",
		"references/the-read.md",
		"references/the-tells.md",
	}
	got := make([]string, 0, len(router.Files))
	for _, file := range router.Files {
		got = append(got, file.Path)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("ui-craft carries %v, want %v", got, want)
	}

	main := string(router.Main())
	for _, reference := range want[1:] {
		if !strings.Contains(main, "`"+reference+"`") {
			t.Errorf("the body does not name %s", reference)
		}
	}
	for _, section := range []string{"## Decision gates", "## Output", "## References"} {
		if !strings.Contains(main, section) {
			t.Errorf("the router has no %q section", section)
		}
	}
}
