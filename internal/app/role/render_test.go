package role_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/role"
)

// source is one role in the canonical shape: the professional, its evidence and
// its probe.
const source = `---
name: marketing
description: Trigger: when the work is marketing or positioning. A senior marketer who asks before spending.
---

## Who it is

Responsible for the demand the product earns.

## What it asks before starting

- Who is the audience, in one sentence?
- What is the baseline today?

## How it works

1. Baseline first.
2. One channel at a time.

## What it refuses

- A campaign with no baseline.

## How it knows it worked

- The number moved and the source is written down.

## What it does not know

- It does not price a product.

## Where this comes from

- One interview with a head of growth, 2026-10.

## Probe

Task: a launch with no baseline. Expected: it asks for the baseline.
`

// withSource seeds one role source in the fake disk.
func withSource(text string) *fakeFiles {
	path := filepath.Join("raiz", role.Dir, "marketing.md")
	return newFakeFiles(map[string]string{path: text})
}

// projectDest is the destination of a project that keeps its craft in the tree:
// what every project had before the scope existed, and the default this feature
// must not move.
func projectDest() role.SkillDest {
	return role.SkillDest{Root: "raiz", Base: ".agents/skills", Label: "./.agents/skills"}
}

// homeDest is the destination of a project whose craft lives once per machine.
func homeDest() role.SkillDest {
	return role.SkillDest{Root: "hogar", Base: ".agents/skills", Label: "~/.agents/skills"}
}

// TestARoleWritesItsSkillWhereTheDestinationSays: the scope is not the role's to
// decide. A project that keeps the craft in the machine gets the skill there, and
// the project keeps no copy -the folder the adoption inference reads.
func TestARoleWritesItsSkillWhereTheDestinationSays(t *testing.T) {
	files := withSource(source)

	res := role.Render("raiz", "marketing", "THE BLOCK", []string{"opencode"}, homeDest(), files)

	if res.Code != role.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s / %v)", res.Code, role.CodeOK, res.Message, res.Lines)
	}
	if _, ok := files.files[filepath.Join("hogar", ".agents", "skills", "marketing", "SKILL.md")]; !ok {
		t.Errorf("the skill did not go to the machine; wrote %v", files.wrote)
	}
	if _, ok := files.files[filepath.Join("raiz", ".agents", "skills", "marketing", "SKILL.md")]; ok {
		t.Errorf("the project got a copy of the craft it said lives in the machine; wrote %v", files.wrote)
	}
	if _, ok := files.files[filepath.Join("raiz", ".opencode", "agents", "marketing.md")]; !ok {
		t.Errorf("the agent file did not stay in the project; wrote %v", files.wrote)
	}
	// The line a human reads names the folder the skill went to, in the shape the
	// rest of the tool prints a home path.
	if joined := strings.Join(res.Lines, "\n"); !strings.Contains(joined, "wrote ~/.agents/skills/marketing/SKILL.md") {
		t.Errorf("the run does not name the machine's copy:\n%s", joined)
	}
}

// TestListReadsTheDestinationItIsGiven: the listing compares the bytes the scope
// names, so a role rendered into the machine is up to date and not stale.
func TestListReadsTheDestinationItIsGiven(t *testing.T) {
	files := withSource(source)
	if rendered := role.Render("raiz", "marketing", "THE BLOCK", []string{"opencode"}, homeDest(), files); rendered.Code != role.CodeOK {
		t.Fatalf("render: %s (lines %v)", rendered.Message, rendered.Lines)
	}

	fresh := role.List("raiz", "THE BLOCK", []string{"opencode"}, homeDest(), files)
	if !strings.Contains(fresh.Message, "up to date") {
		t.Errorf("a role rendered into the machine is reported as %q", fresh.Message)
	}

	// The same project read against the project destination is stale: the bytes
	// are not there, and the listing says so instead of guessing.
	stale := role.List("raiz", "THE BLOCK", []string{"opencode"}, projectDest(), files)
	if !strings.Contains(stale.Message, "stale") {
		t.Errorf("the listing does not read the destination it was given: %q", stale.Message)
	}
}

// TestRenderWritesTheAgentAndTheSkill: one source, two artifacts, and the agent
// carries the block the project reads plus its role section.
func TestRenderWritesTheAgentAndTheSkill(t *testing.T) {
	files := withSource(source)

	res := role.Render("raiz", "marketing", "THE BLOCK", []string{"opencode"}, projectDest(), files)

	if res.Code != role.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s / %v)", res.Code, role.CodeOK, res.Message, res.Lines)
	}
	skill := string(files.files[filepath.FromSlash(projectDest().Path("marketing"))])
	for _, want := range []string{"## When", "## Rules", "## Never", "Baseline first", "A campaign with no baseline", "when the work is marketing"} {
		if !strings.Contains(skill, want) {
			t.Errorf("the rendered skill does not say %q:\n%s", want, skill)
		}
	}
	agent := string(files.files[filepath.Join("raiz", ".opencode", "agents", "marketing.md")])
	for _, want := range []string{"THE BLOCK", "The role you play", "mode: primary", "A senior marketer"} {
		if !strings.Contains(agent, want) {
			t.Errorf("the rendered agent does not say %q:\n%s", want, agent)
		}
	}
	if len(res.Files) != 2 {
		t.Errorf("Files = %v, want the skill and the agent", res.Files)
	}
}

// TestASecondRenderWritesNothing: the command is safe to repeat, so an unchanged
// source leaves the tree alone.
func TestASecondRenderWritesNothing(t *testing.T) {
	files := withSource(source)
	if first := role.Render("raiz", "marketing", "THE BLOCK", []string{"opencode"}, projectDest(), files); first.Code != role.CodeOK {
		t.Fatalf("first render: %s", first.Message)
	}
	before := len(files.wrote)

	second := role.Render("raiz", "marketing", "THE BLOCK", []string{"opencode"}, projectDest(), files)

	if second.Code != role.CodeOK {
		t.Fatalf("second render: %s", second.Message)
	}
	if len(second.Files) != 0 || len(files.wrote) != before {
		t.Errorf("a second render wrote %v", second.Files)
	}
	if !strings.Contains(second.Message, "already up to date") {
		t.Errorf("message %q does not say nothing changed", second.Message)
	}
}

// TestARoleThatDoesNotFitIsRefused: the cap is a contract of the renderer, not a
// suggestion, and the refusal writes nothing.
func TestARoleThatDoesNotFitIsRefused(t *testing.T) {
	refuses := strings.Repeat("- a thing a senior would not accept\n", 60)
	tooLong := strings.Replace(source, "## What it refuses\n\n- A campaign with no baseline.\n", "## What it refuses\n\n"+refuses, 1)
	files := withSource(tooLong)

	res := role.Render("raiz", "marketing", "THE BLOCK", []string{"opencode"}, projectDest(), files)

	if res.Code != role.CodeBlocked {
		t.Fatalf("Code = %d, want %d", res.Code, role.CodeBlocked)
	}
	if !strings.Contains(res.Message, "maximum is 60") {
		t.Errorf("message %q does not name the cap", res.Message)
	}
	if len(files.wrote) != 0 {
		t.Errorf("a refused render wrote %v", files.wrote)
	}
}

// TestRenderRefusesASourceThatDoesNotPassItsCheck: the problems are printed by
// file, and nothing is written.
func TestRenderRefusesASourceThatDoesNotPassItsCheck(t *testing.T) {
	files := withSource(strings.Replace(source, "## What it refuses\n\n- A campaign with no baseline.\n", "", 1))

	res := role.Render("raiz", "marketing", "THE BLOCK", []string{"opencode"}, projectDest(), files)

	if res.Code != role.CodeBlocked {
		t.Fatalf("Code = %d, want %d", res.Code, role.CodeBlocked)
	}
	if joined := strings.Join(res.Lines, "\n"); !strings.Contains(joined, `"What it refuses" section is missing`) {
		t.Errorf("the refusal does not name the missing section: %s", joined)
	}
	if len(files.wrote) != 0 {
		t.Errorf("a refused render wrote %v", files.wrote)
	}
}

// TestListSaysWhichRolesAreStale: the listing is the map a human reads before
// running the render.
func TestListSaysWhichRolesAreStale(t *testing.T) {
	files := withSource(source)

	stale := role.List("raiz", "THE BLOCK", []string{"opencode"}, projectDest(), files)
	if stale.Code != role.CodeOK {
		t.Fatalf("List: %s", stale.Message)
	}
	if !strings.Contains(stale.Message, "stale") {
		t.Errorf("a role never rendered is not reported as stale: %s", stale.Message)
	}

	if rendered := role.Render("raiz", "marketing", "THE BLOCK", []string{"opencode"}, projectDest(), files); rendered.Code != role.CodeOK {
		t.Fatalf("render: %s", rendered.Message)
	}

	fresh := role.List("raiz", "THE BLOCK", []string{"opencode"}, projectDest(), files)
	if !strings.Contains(fresh.Message, "up to date") {
		t.Errorf("a rendered role is not reported as up to date: %s", fresh.Message)
	}
}

// TestListWithoutRolesNamesTheCommand: an empty project is not an error, it is a
// next step.
func TestListWithoutRolesNamesTheCommand(t *testing.T) {
	files := newFakeFiles(nil)

	res := role.List("raiz", "THE BLOCK", []string{"opencode"}, projectDest(), files)

	if res.Code != role.CodeBlocked {
		t.Fatalf("Code = %d, want %d", res.Code, role.CodeBlocked)
	}
	if !strings.Contains(res.Message, "odachi role new") {
		t.Errorf("message %q does not name the command that starts", res.Message)
	}
}

// TestRenderRefusesAClientWithoutANativeFile: the clients that have no agent
// file are named instead of silently skipped.
func TestRenderRefusesAClientWithoutANativeFile(t *testing.T) {
	files := withSource(source)

	res := role.Render("raiz", "marketing", "THE BLOCK", []string{"codex"}, projectDest(), files)

	if res.Code != role.CodeBlocked {
		t.Fatalf("Code = %d, want %d", res.Code, role.CodeBlocked)
	}
	if !strings.Contains(res.Message, "opencode") || !strings.Contains(res.Message, "claude") {
		t.Errorf("message %q does not name the clients that have a native file: %s", res.Message, fmt.Sprint(res.Lines))
	}
}
