package persona_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/persona"
)

// TestRoleFileCarriesTheBlockAndTheRole: the agent file of a role is the project's
// block plus the role section, under the client's own frontmatter, and its name
// and description are the role's because the selector lists them.
func TestRoleFileCarriesTheBlockAndTheRole(t *testing.T) {
	file, ok := persona.RoleFile("opencode", "marketing", "Trigger: when the work is marketing", "THE BLOCK", "## The role you play\n\nDemand.")
	if !ok {
		t.Fatal("RoleFile(opencode) does not exist")
	}
	if file.Path != ".opencode/agents/marketing.md" {
		t.Errorf("Path = %q, want the role under the client's agent directory", file.Path)
	}
	for _, want := range []string{
		`description: "Trigger: when the work is marketing"`,
		"mode: primary",
		"THE BLOCK",
		"The role you play",
		".odachi/roles/marketing.md",
	} {
		if !strings.Contains(file.Content, want) {
			t.Errorf("the agent file does not say %q:\n%s", want, file.Content)
		}
	}
}

// TestRoleFileIsOnlyForClientsWithANativeMechanism: a client without an agent
// file is refused, not skipped.
func TestRoleFileIsOnlyForClientsWithANativeMechanism(t *testing.T) {
	if _, ok := persona.RoleFile("codex", "marketing", "d", "b", "s"); ok {
		t.Error("RoleFile(codex) = ok, want a refusal")
	}
}

// TestGlobalSubagentFilesLiveInTheHome: the door is global, so the agents it
// invokes are too, in the folder each client reads from the home.
func TestGlobalSubagentFilesLiveInTheHome(t *testing.T) {
	opencode := persona.GlobalSubagentFilesFor("opencode")
	if len(opencode) != len(persona.SubagentNames()) {
		t.Fatalf("opencode has %d global subagents, want %d", len(opencode), len(persona.SubagentNames()))
	}
	for _, file := range opencode {
		if !strings.HasPrefix(file.Path, ".config/opencode/agents/") {
			t.Errorf("Path = %q, want the OpenCode global agents folder", file.Path)
		}
		if !strings.Contains(file.Content, "mode: subagent") {
			t.Errorf("%s is not an OpenCode subagent:\n%s", file.Path, file.Content)
		}
	}

	for _, file := range persona.GlobalSubagentFilesFor("claude") {
		if !strings.HasPrefix(file.Path, ".claude/agents/") {
			t.Errorf("Path = %q, want the Claude global agents folder", file.Path)
		}
	}

	if got := persona.GlobalSubagentFilesFor("codex"); got != nil {
		t.Errorf("GlobalSubagentFilesFor(codex) = %v, want none", got)
	}
}
