package agent_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/agent"
)

// verified is the table Contract XII fixes, copied here on purpose: a test that
// reads the table it is checking proves nothing.
//
// Seven agents read their skills from inside the human's home; OpenCode reads
// them from inside the project (`.opencode/skills/<name>/SKILL.md`), which is
// relative to the project and therefore has no platform-dependent shape.
var verified = []struct {
	name  string
	dest  string
	scope agent.Scope
}{
	{"claude", ".claude/skills", agent.ScopeHome},
	{"codex", ".codex/skills", agent.ScopeHome},
	{"cursor", ".cursor/skills", agent.ScopeHome},
	{"gemini", ".gemini/skills", agent.ScopeHome},
	{"copilot", ".copilot/skills", agent.ScopeHome},
	{"qwen", ".qwen/skills", agent.ScopeHome},
	{"kiro", ".kiro/skills", agent.ScopeHome},
	{"opencode", ".opencode/skills", agent.ScopeProject},
}

// display is the path a human reads for one row: `~/...` inside the home,
// `./...` inside the project, never an absolute path.
func display(scope agent.Scope, dest string) string {
	if scope == agent.ScopeProject {
		return "./" + dest
	}
	return "~/" + dest
}

// TestSupportedIsTheVerifiedTable: every row is present, in this order, and
// each one resolves to its own destination and its own scope.
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
		if got[i].Scope != want.scope {
			t.Errorf("%s.Scope = %q, want %q", want.name, got[i].Scope, want.scope)
		}
	}
}

// TestResolveReturnsTheVerifiedDestination: the name a human types maps to the
// path the agent looks for, to its scope, and to the two display forms built
// from it.
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
			if got, want := a.Display(), display(want.scope, want.dest); got != want {
				t.Errorf("Display() = %q, want %q", got, want)
			}
			// The marker is the agent's config directory: the parent of the
			// skills directory, which is what `pack` looks for when it detects
			// an agent instead of being told. It is resolved against the scope's
			// root: the home for a home agent, the project for a project one.
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

// TestOpenCodeIsProjectScoped: OpenCode reads
// `.opencode/skills/<name>/SKILL.md` relative to the project. That is the path
// this command installs, so the platform-dependent global
// (`~/.config/opencode/skills`) never enters the picture.
func TestOpenCodeIsProjectScoped(t *testing.T) {
	a, err := agent.Resolve("opencode")
	if err != nil {
		t.Fatalf("Resolve(opencode): %v", err)
	}
	if a.Name != "opencode" {
		t.Errorf("Name = %q, want opencode", a.Name)
	}
	if a.Scope != agent.ScopeProject {
		t.Errorf("Scope = %q, want %q", a.Scope, agent.ScopeProject)
	}
	if a.Dest != ".opencode/skills" {
		t.Errorf("Dest = %q, want .opencode/skills", a.Dest)
	}
	if got, want := a.Display(), "./.opencode/skills"; got != want {
		t.Errorf("Display() = %q, want %q", got, want)
	}
	if strings.Contains(a.Display(), "C:") || strings.HasPrefix(a.Display(), "/") {
		t.Errorf("Display() must not be an absolute path: %q", a.Display())
	}
	if got, want := a.Marker(), ".opencode"; got != want {
		t.Errorf("Marker() = %q, want %q", got, want)
	}
}

// TestOnlyOpenCodeIsProjectScoped: the seven agents verified before this unit
// keep their home and their path. Adding a scope did not move anyone.
func TestOnlyOpenCodeIsProjectScoped(t *testing.T) {
	project := []string{}
	home := []string{}
	for _, a := range agent.Supported() {
		switch a.Scope {
		case agent.ScopeProject:
			project = append(project, a.Name)
		case agent.ScopeHome:
			home = append(home, a.Name)
			if !strings.HasPrefix(a.Display(), "~/") {
				t.Errorf("%s is home-scoped but displays %q", a.Name, a.Display())
			}
		default:
			t.Errorf("%s has an unknown scope %q", a.Name, a.Scope)
		}
	}
	if !slices.Equal(project, []string{"opencode"}) {
		t.Errorf("project-scoped agents = %v, want [opencode]", project)
	}
	if len(home) != len(verified)-1 {
		t.Errorf("home-scoped agents = %v, want the seven verified before this unit", home)
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

// TestUnknownAgentNamesTheSupportedOnes: an invented name is exit 1 material,
// and the human reads what does exist — opencode included now.
func TestUnknownAgentNamesTheSupportedOnes(t *testing.T) {
	_, err := agent.Resolve("inventado")
	if err == nil {
		t.Fatal("Resolve(inventado) succeeded; only the verified table is supported")
	}
	message := err.Error()
	if !strings.Contains(message, "inventado") {
		t.Errorf("the message does not name the rejected agent: %q", message)
	}
	if !strings.Contains(message, "opencode") {
		t.Errorf("the message does not name opencode as supported: %q", message)
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
	for _, typed := range []string{"  Claude ", "OpenCode", " opencode "} {
		a, err := agent.Resolve(typed)
		if err != nil {
			t.Fatalf("Resolve(%q): %v", typed, err)
		}
		if want := strings.ToLower(strings.TrimSpace(typed)); a.Name != want {
			t.Errorf("Resolve(%q).Name = %q, want %q", typed, a.Name, want)
		}
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

	// A home agent and a project agent in the same flag keep their own scopes:
	// one run can serve both kinds.
	mixed, err := agent.ResolveAll("opencode,claude")
	if err != nil {
		t.Fatalf("ResolveAll(opencode,claude): %v", err)
	}
	if len(mixed) != 2 || mixed[0].Scope != agent.ScopeProject || mixed[1].Scope != agent.ScopeHome {
		t.Errorf("ResolveAll(opencode,claude) = %+v, want [opencode project, claude home]", mixed)
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
