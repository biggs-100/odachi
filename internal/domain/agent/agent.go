// Package agent owns the table of agents `odachi pack` can install skills into:
// which name maps to which skills directory, which scope that directory lives
// in, and which names are refused and why.
//
// The table is data, not branches of code: adding an agent is one row, and
// removing a guess is one deleted row. Every destination here was verified
// against the agent's own documentation or source; a path nobody verified does
// not ship.
//
// Two scopes exist, and the distinction is the whole point of the second one:
// a `home` agent reads its skills from inside the human's home directory, and a
// `project` agent reads them from inside the project being worked on. OpenCode
// is the project one — `.opencode/skills/<name>/SKILL.md` — and its documented
// pattern (`{skill,skills}/**/SKILL.md`) is relative to the project, so the
// platform-dependent global (`~/.config/opencode/skills`) never enters the
// picture. An earlier unit refused OpenCode because its *global* path varies by
// platform; that path is not the one we install.
//
// It is pure: no filesystem, no os. The app layer resolves the home and the
// project, and this package only says where and whether.
package agent

import (
	"fmt"
	"slices"
	"strings"
)

// Scope is where an agent reads its skills from: inside the human's home, or
// inside the project being worked on. It is what decides which root the
// relative Dest is joined to, and how the path is shown to a human.
type Scope string

const (
	// ScopeHome is a skills directory inside the human's home directory.
	ScopeHome Scope = "home"
	// ScopeProject is a skills directory inside the project directory.
	ScopeProject Scope = "project"
)

// Agent is one supported agent: the name a human types, the skills directory it
// reads (relative to its scope's root, always with forward slashes) and the
// scope that root belongs to.
type Agent struct {
	Name  string
	Dest  string
	Scope Scope
}

// supported is the verified table, in the order `odachi pack` lists and
// summarizes it.
var supported = []Agent{
	{Name: "claude", Dest: ".claude/skills", Scope: ScopeHome},
	{Name: "codex", Dest: ".codex/skills", Scope: ScopeHome},
	{Name: "cursor", Dest: ".cursor/skills", Scope: ScopeHome},
	{Name: "gemini", Dest: ".gemini/skills", Scope: ScopeHome},
	{Name: "copilot", Dest: ".copilot/skills", Scope: ScopeHome},
	{Name: "qwen", Dest: ".qwen/skills", Scope: ScopeHome},
	{Name: "kiro", Dest: ".kiro/skills", Scope: ScopeHome},
	// OpenCode reads skills from `.opencode/skills/<name>/SKILL.md` in the
	// project (and, as compatible external locations, from `.claude/skills/`
	// and `.agents/skills/`). The project path is the one we install: it is
	// relative to the project, so nothing about it depends on the platform.
	{Name: "opencode", Dest: ".opencode/skills", Scope: ScopeProject},
}

// refusals are the names this command knows and will not use, each with the
// reason a human reads. They are refused out loud, never skipped: a silent skip
// leaves the agent without skills and without a clue why.
var refusals = map[string]string{
	// Pi's adapter reports an empty skills directory because Pi does not have
	// one, so there is no destination to verify and none to write.
	"pi": "Pi no usa un directorio de skills",
}

// Supported returns the verified table, safe to keep: a caller that reorders or
// trims the result cannot change what this package resolves.
func Supported() []Agent { return append([]Agent(nil), supported...) }

// Names returns the supported names, in table order.
func Names() []string {
	names := make([]string, 0, len(supported))
	for _, a := range supported {
		names = append(names, a.Name)
	}
	return names
}

// List renders the supported names for the one-line message every refusal
// carries. Two spellings of the same list would drift apart.
func List() string { return strings.Join(Names(), ", ") }

// Resolve returns the agent for name. The lookup trims spaces and ignores
// capitals, because a human types both; it never guesses a name that is not in
// the table. Every error names the supported agents, so a blocked run always
// shows the way out.
func Resolve(name string) (Agent, error) {
	key := strings.ToLower(strings.TrimSpace(name))
	if key == "" {
		return Agent{}, fmt.Errorf("falta el nombre del agente; los soportados son: %s (por ejemplo odachi pack --agents claude,codex)", List())
	}
	if a, ok := find(key); ok {
		return a, nil
	}
	if reason, ok := refusals[key]; ok {
		return Agent{}, fmt.Errorf("%s: %s; los soportados son: %s (probá odachi pack --agents claude)", key, reason, List())
	}
	return Agent{}, fmt.Errorf("agente desconocido: %s; los soportados son: %s (por ejemplo odachi pack --agents claude,codex)", key, List())
}

// ResolveAll parses the comma-separated --agents flag. An empty flag returns no
// agents and no error: it means "detect the ones that are here", which is the
// caller's job because only it can look at the disk. Repeats collapse, and one
// bad name refuses the whole list, so a typo never leaves half a run written.
func ResolveAll(list string) ([]Agent, error) {
	if strings.TrimSpace(list) == "" {
		return nil, nil
	}

	resolved := make([]Agent, 0, strings.Count(list, ",")+1)
	for _, raw := range strings.Split(list, ",") {
		name := strings.ToLower(strings.TrimSpace(raw))
		if name == "" {
			continue
		}
		a, err := Resolve(name)
		if err != nil {
			return nil, err
		}
		if slices.ContainsFunc(resolved, func(r Agent) bool { return r.Name == a.Name }) {
			continue
		}
		resolved = append(resolved, a)
	}
	return resolved, nil
}

// Marker is the agent's own config directory: the parent of its skills
// directory, relative to the same root. `pack` looks for it to decide whether an
// agent is installed here, which is what "cada agente detectado" means: a home
// agent is here when its config directory exists in the home, a project agent
// when its config directory exists in the project.
func (a Agent) Marker() string {
	if i := strings.LastIndex(a.Dest, "/"); i >= 0 {
		return a.Dest[:i]
	}
	return a.Dest
}

// Display is the path a human reads in a message or a summary: `~/...` for a
// home agent, `./...` for a project one. An absolute path is never printed: the
// home and the project belong to the human, not to the report.
func (a Agent) Display() string {
	if a.Scope == ScopeProject {
		return "./" + a.Dest
	}
	return "~/" + a.Dest
}

// find returns the table row for an already-normalized name.
func find(name string) (Agent, bool) {
	for _, a := range supported {
		if a.Name == name {
			return a, true
		}
	}
	return Agent{}, false
}
