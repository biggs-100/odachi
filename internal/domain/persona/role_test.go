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
