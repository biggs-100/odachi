package review_test

import (
	"strings"
	"testing"

	assets "github.com/biggs-100/odachi/internal/assets/review"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
)

// TestTheFourLensesAreBundled: the binary carries one prompt per specialist, and
// the names are the ones the domain selects. A lens the selection can pick but
// the binary cannot load is a dead end at run time.
func TestTheFourLensesAreBundled(t *testing.T) {
	names := assets.Names()

	want := domainreview.Lenses(domainreview.Full)
	if strings.Join(names, ",") != strings.Join(want, ",") {
		t.Fatalf("bundled lenses = %v, want %v in that order", names, want)
	}

	for _, name := range want {
		if _, err := assets.Lens(name); err != nil {
			t.Errorf("Lens(%q) = %v", name, err)
		}
	}
	if _, err := assets.Lens("seguridad"); err == nil {
		t.Error("Lens(seguridad) = nil error, want a refusal naming the four")
	} else if !strings.Contains(err.Error(), domainreview.Risk) {
		t.Errorf("error %q does not name the valid lenses", err)
	}
}

// TestEachLensHasValidFrontmatter: the files are content with a registry entry,
// exactly like a skill. name is the file, description is a trigger.
func TestEachLensHasValidFrontmatter(t *testing.T) {
	for _, name := range assets.Names() {
		t.Run(name, func(t *testing.T) {
			body, err := assets.Lens(name)
			if err != nil {
				t.Fatalf("Lens(%q) = %v", name, err)
			}

			if !strings.HasPrefix(body, "---\n") {
				t.Fatalf("%s does not start with frontmatter:\n%s", name, firstLines(body, 3))
			}
			front := frontmatter(t, body)
			if !strings.Contains(front, "name: "+name) {
				t.Errorf("%s frontmatter does not say name: %s", name, name)
			}
			if !strings.Contains(front, "description: Trigger: ") {
				t.Errorf("%s frontmatter description does not start with Trigger:", name)
			}
			if strings.Contains(front, "\nname: "+name+"\nname:") {
				t.Errorf("%s declares name twice", name)
			}
		})
	}
}

// TestEachLensHasBetweenTwelveAndTwentyRules is the quality bar made checkable: a
// lens with three vague adjectives is noise wearing a specialist's name.
func TestEachLensHasBetweenTwelveAndTwentyRules(t *testing.T) {
	for _, name := range assets.Names() {
		t.Run(name, func(t *testing.T) {
			body, err := assets.Lens(name)
			if err != nil {
				t.Fatalf("Lens(%q) = %v", name, err)
			}

			rules := lookRules(t, body)
			if len(rules) < 12 || len(rules) > 20 {
				t.Fatalf("%s has %d rules under ## What to look at, want between 12 and 20:\n%s", name, len(rules), strings.Join(rules, "\n"))
			}
			for _, rule := range rules {
				if len(rule) < 30 {
					t.Errorf("%s has a rule too short to be actionable: %q", name, rule)
				}
				if len(rule) > 160 {
					t.Errorf("%s has a rule too long to read: %q", name, rule)
				}
			}
		})
	}
}

// TestEachLensHasTheTwoSectionsAndTheSeverityFloor: every lens says what to look
// at, what to stay quiet about, and that a warning never blocks.
func TestEachLensHasTheTwoSectionsAndTheSeverityFloor(t *testing.T) {
	for _, name := range assets.Names() {
		t.Run(name, func(t *testing.T) {
			body, err := assets.Lens(name)
			if err != nil {
				t.Fatalf("Lens(%q) = %v", name, err)
			}

			for _, heading := range []string{"\n## What to look at\n", "\n## What NOT to report\n"} {
				if !strings.Contains(body, heading) {
					t.Errorf("%s is missing the section %q", name, strings.TrimSpace(heading))
				}
			}
			if !strings.Contains(body, "WARNING and SUGGESTION never block") {
				t.Errorf("%s does not state the severity floor", name)
			}
			notReported := section(t, body, "## What NOT to report")
			if len(lookRulesIn(notReported)) < 3 {
				t.Errorf("%s says too little about what not to report:\n%s", name, notReported)
			}
		})
	}
}

// TestEachLensInheritsTheFiveRules: the lenses extend the review, they do not
// replace it. Style stays banned in every one of them.
func TestEachLensInheritsTheFiveRules(t *testing.T) {
	for _, name := range assets.Names() {
		t.Run(name, func(t *testing.T) {
			body, err := assets.Lens(name)
			if err != nil {
				t.Fatalf("Lens(%q) = %v", name, err)
			}
			notReported := section(t, body, "## What NOT to report")
			if !strings.Contains(notReported, "style") && !strings.Contains(notReported, "preference") {
				t.Errorf("%s does not ban style and preference in ## What NOT to report", name)
			}
			if !strings.Contains(strings.ToLower(body), "what was already there") {
				t.Errorf("%s does not inherit the causality rule", name)
			}
		})
	}
}

// TestLensesDoNotInventTheirOwnSeverityVocabulary: one closed vocabulary, four
// prompts. A lens that adds a fifth word cannot be stored.
func TestLensesDoNotInventTheirOwnSeverityVocabulary(t *testing.T) {
	for _, name := range assets.Names() {
		body, err := assets.Lens(name)
		if err != nil {
			t.Fatalf("Lens(%q) = %v", name, err)
		}
		lower := strings.ToLower(body)
		for _, invented := range []string{"severity: minor", "`nit`", "`major`", "`info`"} {
			if strings.Contains(lower, invented) {
				t.Errorf("%s invents a severity %q", name, invented)
			}
		}
	}
}

// frontmatter returns the block between the two `---` lines.
func frontmatter(t *testing.T, body string) string {
	t.Helper()
	rest := strings.TrimPrefix(body, "---\n")
	end := strings.Index(rest, "\n---\n")
	if end < 0 {
		t.Fatalf("frontmatter is not closed:\n%s", firstLines(body, 6))
	}
	return rest[:end+1]
}

// section returns the text under one heading, without the heading.
func section(t *testing.T, body, heading string) string {
	t.Helper()
	start := strings.Index(body, heading)
	if start < 0 {
		t.Fatalf("no %s section", heading)
	}
	rest := body[start+len(heading):]
	if next := strings.Index(rest, "\n## "); next >= 0 {
		rest = rest[:next]
	}
	return rest
}

// lookRules is the rules under ## What to look at.
func lookRules(t *testing.T, body string) []string {
	t.Helper()
	return lookRulesIn(section(t, body, "## What to look at"))
}

// lookRulesIn counts the bullets of a section: the rules are the bullets, and
// nothing else in those sections is one.
func lookRulesIn(text string) []string {
	var rules []string
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") {
			rules = append(rules, strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")))
		}
	}
	return rules
}

// firstLines is the head of a file, for a failure message a human can read.
func firstLines(text string, n int) string {
	lines := strings.Split(text, "\n")
	if len(lines) > n {
		lines = lines[:n]
	}
	return strings.Join(lines, "\n")
}
