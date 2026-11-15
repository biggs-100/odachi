package skills_test

import (
	"strings"
	"testing"
)

// TestUICraftRoutesItsDepth: every reference is reachable through a row of the
// decision gates and listed in ## References, and the two the audit found missing
// carry the questions the page used to answer nowhere: the markup and the keyboard,
// and what belongs together. The helper is the one both doors use, so this skill
// cannot drift into a reference nobody routes (0309).
func TestUICraftRoutesItsDepth(t *testing.T) {
	skill := mustSkill(t, "ui-craft")
	body := string(skill.Main())

	assertEveryReferenceIsRouted(t, skill, body)

	for _, want := range []string{
		"references/accessible-and-adaptive.md",
		"references/grouping-and-emphasis.md",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("ui-craft does not route %q", want)
		}
	}

	// The two new pages answer their own question and send the rest to their owner,
	// because a second copy of a rule is the defect this catalog keeps paying for.
	for path, want := range map[string][]string{
		"references/accessible-and-adaptive.md": {"the keyboard", "prefers-reduced-motion", "states-and-copy.md"},
		"references/grouping-and-emphasis.md":   {"Proximity", "the-numbers.md", "states-and-copy.md"},
		"references/ai-native.md":               {"shows its work", "generative UI", "irreversible", "undo"},
	} {
		var content string
		for _, file := range skill.Files {
			if file.Path == path {
				content = strings.Join(strings.Fields(string(file.Content)), " ")
			}
		}
		if content == "" {
			t.Fatalf("ui-craft does not carry %s", path)
		}
		for _, phrase := range want {
			if !strings.Contains(strings.ToLower(content), strings.ToLower(phrase)) {
				t.Errorf("%s does not say %q", path, phrase)
			}
		}
	}
}
