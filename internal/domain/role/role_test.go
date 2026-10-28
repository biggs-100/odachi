package role_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/role"
)

// complete is a role in the canonical shape: frontmatter plus the eight
// sections, each with the answer a senior would write.
const complete = `---
name: marketing
description: Trigger: when the work is marketing, growth or positioning. A senior marketer who asks before spending.
---

## Who it is

Responsible for the demand the product earns, not for the ads it buys.

## What it asks before starting

- Who is the audience, in one sentence?
- What is the baseline today, and where does the number come from?

## How it works

1. Baseline first.
2. One channel at a time.
3. Measure before scaling.

## What it refuses

- A campaign with no baseline.
- A vanity metric presented as a result.

## How it knows it worked

- The number moved, and the source of the number is written down.

## What it does not know

- It does not price a product and it does not sign a contract.

## Where this comes from

- Three interviews with heads of growth, 2026-10.

## Probe

Task: a launch with no baseline. Expected: it asks for the baseline and refuses to plan.
`

func parse(text string) role.Role {
	return role.Parse("odd/roles/marketing.md", text)
}

func joined(problems []role.Problem) string {
	parts := make([]string, 0, len(problems))
	for _, problem := range problems {
		parts = append(parts, problem.File+": "+problem.Reason)
	}
	return strings.Join(parts, "\n")
}

// TestParseReadsTheEightSections: the source is the only thing a human edits, so
// the parser has to give back every section by its heading.
func TestParseReadsTheEightSections(t *testing.T) {
	got := parse(complete)

	if !got.HasFrontmatter {
		t.Fatal("HasFrontmatter = false, want true")
	}
	if got.Name != "marketing" {
		t.Errorf("Name = %q, want marketing", got.Name)
	}
	if !strings.HasPrefix(got.Description, role.TriggerPrefix) {
		t.Errorf("Description = %q, want the Trigger prefix", got.Description)
	}
	if !strings.Contains(got.Trigger, "marketing") {
		t.Errorf("Trigger = %q, want the trigger words without the prefix", got.Trigger)
	}
	for _, section := range role.Sections {
		if strings.TrimSpace(got.Bodies[section]) == "" {
			t.Errorf("section %q came back empty", section)
		}
	}
	if problems := got.Problems("marketing"); len(problems) != 0 {
		t.Errorf("a complete role has problems: %s", joined(problems))
	}
}

// TestParseWithoutFrontmatterIsOneReason: a file that never declares itself is
// not guessed at.
func TestParseWithoutFrontmatterIsOneReason(t *testing.T) {
	got := parse("## Who it is\n\nsomething\n")

	if got.HasFrontmatter {
		t.Fatal("HasFrontmatter = true for a file without ---")
	}
	problems := got.Problems("marketing")
	if len(problems) != 1 {
		t.Fatalf("problems = %d (%s), want exactly one about the frontmatter", len(problems), joined(problems))
	}
	if !strings.Contains(problems[0].Reason, "frontmatter") {
		t.Errorf("Reason = %q, want it to name the frontmatter", problems[0].Reason)
	}
}

// TestEverySectionIsRequired: the eight sections are the contract, and a role
// that skips one is refused by name.
func TestEverySectionIsRequired(t *testing.T) {
	for _, section := range role.Sections {
		t.Run(section, func(t *testing.T) {
			without := removeSection(t, complete, section)

			problems := parse(without).Problems("marketing")

			if !strings.Contains(joined(problems), `"`+section+`" section is missing or empty`) {
				t.Errorf("removing %q was not reported: %s", section, joined(problems))
			}
		})
	}
}

// removeSection drops one `## heading` block and everything under it.
func removeSection(t *testing.T, text, section string) string {
	t.Helper()
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))
	skipping := false
	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			skipping = strings.TrimSpace(strings.TrimPrefix(line, "## ")) == section
		}
		if !skipping {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}

// TestTODOAndTheNameAreRefused: a role ships complete, and its name has to be
// the file it lives in.
func TestTODOAndTheNameAreRefused(t *testing.T) {
	withTODO := strings.Replace(complete, "Baseline first.", "TODO baseline first.", 1)
	if problems := parse(withTODO).Problems("marketing"); !strings.Contains(joined(problems), "TODO") {
		t.Errorf("a TODO was not reported: %s", joined(problems))
	}

	renamed := strings.Replace(complete, "name: marketing", "name: ventas", 1)
	if problems := parse(renamed).Problems("marketing"); !strings.Contains(joined(problems), "does not match the file") {
		t.Errorf("a name that is not the file was not reported: %s", joined(problems))
	}
}

// TestTheTemplatePassesItsOwnCheck: `role new` writes this file, so a human's
// first read cannot be a violation.
func TestTheTemplatePassesItsOwnCheck(t *testing.T) {
	got := role.Parse("odd/roles/marketing.md", role.Template("marketing"))

	if problems := got.Problems("marketing"); len(problems) != 0 {
		t.Fatalf("the template does not pass its own check:\n%s\n---\n%s", joined(problems), role.Template("marketing"))
	}
	for _, section := range role.Sections {
		if strings.TrimSpace(got.Bodies[section]) == "" {
			t.Errorf("the template leaves %q empty", section)
		}
	}
}

// TestValidSlug: the slug is the file, the agent and the skill, so it has one
// shape.
func TestValidSlug(t *testing.T) {
	for _, slug := range []string{"Marketing", "with space", "ab", "-dash"} {
		if role.ValidSlug(slug) {
			t.Errorf("ValidSlug(%q) = true, want false", slug)
		}
	}
	for _, slug := range []string{"marketing", "civil-servant-2", "growth"} {
		if !role.ValidSlug(slug) {
			t.Errorf("ValidSlug(%q) = false, want true", slug)
		}
	}
}

// TestANameIsRequired: the file has to declare its own name, and the parser does
// not fill it in from the slug: a source that never says who it is gets refused
// instead of passing on a name nobody wrote.
func TestANameIsRequired(t *testing.T) {
	withoutName := strings.Replace(complete, "name: marketing\n", "", 1)

	problems := parse(withoutName).Problems("marketing")

	if !strings.Contains(joined(problems), "does not say name") {
		t.Errorf("a source with no name was accepted: %s", joined(problems))
	}
}
