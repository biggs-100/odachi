package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// roleSource is a complete source, written by the test where `role new` would
// leave the skeleton.
const roleSource = `---
name: marketing
description: Trigger: when the work is marketing or positioning. A senior marketer who asks before spending.
---

## Who it is

Responsible for the demand the product earns.

## What it asks before starting

- Who is the audience, in one sentence?

## How it works

1. Baseline first.

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

// TestRoleRendersFromTheSourceEndToEnd is the CLI contract for the whole loop:
// one source, the agent file of each chosen client, and the role's skill.
func TestRoleRendersFromTheSourceEndToEnd(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	if code, _, stderr := runDir(t, dir, []string{"role", "new", "marketing"}, ""); code != 0 {
		t.Fatalf("role new: exit %d, stderr=%s", code, stderr)
	}
	sourcePath := filepath.Join(dir, ".odachi", "roles", "marketing.md")
	if err := os.WriteFile(sourcePath, []byte(roleSource), 0o644); err != nil {
		t.Fatalf("write the source: %v", err)
	}

	code, stdout, stderr := runDir(t, dir, []string{"role", "marketing"}, "")
	if code != 0 {
		t.Fatalf("role marketing: exit %d, stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, ".agents/skills/marketing/SKILL.md") {
		t.Errorf("stdout %q does not name the skill", stdout)
	}

	files := tree(t, dir)
	for _, path := range []string{
		".agents/skills/marketing/SKILL.md",
		".opencode/agents/marketing.md",
		".claude/output-styles/marketing.md",
	} {
		if _, ok := files[path]; !ok {
			t.Errorf("%s was not rendered", path)
		}
	}
	if agent := files[".opencode/agents/marketing.md"]; !strings.Contains(agent, "The role you play") {
		t.Errorf("the agent does not carry the role section:\n%s", agent)
	}

	code, stdout, stderr = runDir(t, dir, []string{"role"}, "")
	if code != 0 {
		t.Fatalf("role: exit %d, stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "marketing: up to date") {
		t.Errorf("the listing does not say the role is up to date: %q", stdout)
	}
}
