package agent_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/agent"
)

// verified is the table Contract VIII fixes, copied here on purpose: a test
// that reads the table it is checking proves nothing.
var verified = []struct {
	name string
	dest string
}{
	{"claude", ".claude/skills"},
	{"codex", ".codex/skills"},
	{"cursor", ".cursor/skills"},
	{"gemini", ".gemini/skills"},
	{"copilot", ".copilot/skills"},
	{"qwen", ".qwen/skills"},
	{"kiro", ".kiro/skills"},
}

// TestSupportedIsTheVerifiedTable: every row is present, in this order, and
// each one resolves to its own home-relative destination.
func TestSupportedIsTheVerifiedTable(t *testing.T) {
	got := agent.Supported()
	if len(got) != len(verified) {
		t.Fatalf("Supported() has %d agents, want %d: %+v", len(got), len(verified), got)
	}
	for i, want := range verified {
		if got[i].Name != want.name {
			t.Errorf("Supported()[%d].Name = %q, want %q", i, got[i].Name, want.name)
		}
		if got[i].Dest != want.dest {
			t.Errorf("%s.Dest = %q, want %q", want.name, got[i].Dest, want.dest)
		}
	}
}

// TestResolveReturnsTheVerifiedDestination: the name a human types maps to the
// path the agent looks for, and to the two display forms built from it.
func TestResolveReturnsTheVerifiedDestination(t *testing.T) {
	for _, want := range verified {
		t.Run(want.name, func(t *testing.T) {
			a, err := agent.Resolve(want.name)
			if err != nil {
				t.Fatalf("Resolve(%q): %v", want.name, err)
			}
			if a.Name != want.name || a.Dest != want.dest {
				t.Fatalf("Resolve(%q) = %+v, want %+v", want.name, a, want)
			}
			if got, want := a.Display(), "~/"+want.dest; got != want {
				t.Errorf("Display() = %q, want %q", got, want)
			}
			// The marker is the agent's config directory: the parent of the
			// skills directory, which is what `pack` looks for when it detects
			// an agent instead of being told.
			if got, want := a.Marker(), "."+want.name; got != want {
				t.Errorf("Marker() = %q, want %q", got, want)
			}
		})
	}
}

// TestSupportedNamesAreAlone: the table carries no duplicates, and the listing
// a human reads names every one of them.
func TestSupportedNamesAreAlone(t *testing.T) {
	names := agent.Names()
	if len(names) != len(verified) {
		t.Fatalf("Names() = %v, want %d names", names, len(verified))
	}
	for _, want := range verified {
		if !slices.Contains(names, want.name) {
			t.Errorf("Names() = %v, missing %q", names, want.name)
		}
	}
	for i := 1; i < len(names); i++ {
		if names[i] == names[i-1] {
			t.Errorf("Names() repeats %q", names[i])
		}
	}

	list := agent.List()
	for _, want := range verified {
		if !strings.Contains(list, want.name) {
			t.Errorf("List() = %q, missing %q", list, want.name)
		}
	}
	if lines := strings.Split(list, "\n"); len(lines) != 1 {
		t.Errorf("List() must be one line, got %d: %q", len(lines), list)
	}
}

// TestPiIsRefusedWithItsReason: Pi is not silently skipped and not written
// into a guessed path. Its adapter reports no skills directory at all.
func TestPiIsRefusedWithItsReason(t *testing.T) {
	_, err := agent.Resolve("pi")
	if err == nil {
		t.Fatal("Resolve(pi) succeeded; Pi has no skills directory and must be refused")
	}
	message := err.Error()
	if !strings.Contains(message, "Pi no usa un directorio de skills") {
		t.Errorf("the message does not say why Pi is refused: %q", message)
	}
	// A refusal still names the way out: the agents that do work.
	for _, want := range verified {
		if !strings.Contains(message, want.name) {
			t.Errorf("the refusal does not name %q as supported: %q", want.name, message)
		}
	}
}

// TestOpenCodeIsOutOfScope: opencode's path depends on the platform, so this
// unit does not guess it. The message says so instead of inventing a path.
func TestOpenCodeIsOutOfScope(t *testing.T) {
	_, err := agent.Resolve("opencode")
	if err == nil {
		t.Fatal("Resolve(opencode) succeeded; its path is not verified in this unit")
	}
	message := err.Error()
	if !strings.Contains(message, "opencode") {
		t.Errorf("the message does not name opencode: %q", message)
	}
	if !strings.Contains(message, "plataforma") {
		t.Errorf("the message does not say the path depends on the platform: %q", message)
	}
	for _, want := range verified {
		if !strings.Contains(message, want.name) {
			t.Errorf("the refusal does not name %q as supported: %q", want.name, message)
		}
	}
}

// TestUnknownAgentNamesTheSupportedOnes: an invented name is exit 1 material,
// and the human reads what does exist.
func TestUnknownAgentNamesTheSupportedOnes(t *testing.T) {
	_, err := agent.Resolve("inventado")
	if err == nil {
		t.Fatal("Resolve(inventado) succeeded; only the verified table is supported")
	}
	message := err.Error()
	if !strings.Contains(message, "inventado") {
		t.Errorf("the message does not name the rejected agent: %q", message)
	}
	for _, want := range verified {
		if !strings.Contains(message, want.name) {
			t.Errorf("the message does not name %q as supported: %q", want.name, message)
		}
	}
}

// TestResolveIsForgivingAboutShape: a human types spaces, a trailing comma and
// sometimes capitals. The lookup trims and lowercases; it never guesses a name.
func TestResolveIsForgivingAboutShape(t *testing.T) {
	a, err := agent.Resolve("  Claude ")
	if err != nil {
		t.Fatalf("Resolve with spaces and capitals: %v", err)
	}
	if a.Name != "claude" {
		t.Errorf("Resolve(\"  Claude \").Name = %q, want claude", a.Name)
	}
}

// TestResolveAllParsesTheFlag: one comma-separated list, in the order the human
// wrote it, without repeats and without empty entries.
func TestResolveAllParsesTheFlag(t *testing.T) {
	got, err := agent.ResolveAll("codex, claude,,codex")
	if err != nil {
		t.Fatalf("ResolveAll: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("ResolveAll returned %d agents, want 2: %+v", len(got), got)
	}
	if got[0].Name != "codex" || got[1].Name != "claude" {
		t.Errorf("ResolveAll order = [%s %s], want [codex claude]", got[0].Name, got[1].Name)
	}

	// An empty flag is the "detect them" case, not an error.
	none, err := agent.ResolveAll("")
	if err != nil {
		t.Fatalf("ResolveAll(\"\"): %v", err)
	}
	if len(none) != 0 {
		t.Errorf("ResolveAll(\"\") = %+v, want no agents", none)
	}
}

// TestResolveAllStopsOnTheFirstBadName: the whole list is refused, so a typo in
// the flag never leaves half of a run written.
func TestResolveAllStopsOnTheFirstBadName(t *testing.T) {
	if _, err := agent.ResolveAll("claude,inventado"); err == nil {
		t.Fatal("ResolveAll accepted an invented agent")
	}
	if _, err := agent.ResolveAll("claude,pi"); err == nil {
		t.Fatal("ResolveAll accepted pi")
	}
}

// TestEmptyNameIsRefusedWithTheWayOut: `--agents ,` is a mistake, and the
// message still names the supported agents.
func TestEmptyNameIsRefusedWithTheWayOut(t *testing.T) {
	_, err := agent.Resolve("   ")
	if err == nil {
		t.Fatal("Resolve(\"   \") succeeded; an empty name is not an agent")
	}
	for _, want := range verified {
		if !strings.Contains(err.Error(), want.name) {
			t.Errorf("the message does not name %q as supported: %q", want.name, err)
		}
	}
}
