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
// pattern (`{skill,skills}/**/SKILL.md`) is relative to the project.
//
// A scope is a default, not a cage. `odachi pack --scope` is the human's
// override: the same agent installed under the other root, which is how one copy
// of the skills serves every project instead of one run per project. A few
// agents keep the same relative path in both roots (`~/.claude/skills` is
// `./.claude/skills`); the ones whose documentation puts the other root's config
// somewhere else say so in overrideDests.
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

// Standard is the cross-client location the Agent Skills standard describes: a
// client that odachi has never heard of can still find these skills, because the
// convention belongs to the standard and not to one program. It is a pseudo
// agent — it names a destination, not an installed client — so it is offered by
// name, refused by nobody, and never detected.
const Standard = "standard"

// portable is the row behind Standard: the project copy lives in `.agents/skills`
// and the home copy in `~/.agents/skills`, the two roots the standard's clients
// scan besides their own directory.
var portable = Agent{Name: Standard, Dest: ".agents/skills", Scope: ScopeProject}

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
	"pi": "Pi does not use a skills directory",
}

// overrideDests holds the one thing a scope override can change besides the
// root: the relative path. OpenCode documents its global config dir as
// `~/.config/opencode` (XDG_CONFIG_HOME), so forcing it into the home installs
// into `.config/opencode/skills` and not into a `~/.opencode/skills` its docs
// never mention. Every other agent keeps its own relative path in both roots,
// which is why the table stores one Dest and only this map stores an exception.
// XDG_CONFIG_HOME itself is not read here: this package has no environment, and
// the human who moved their config dir has --dest for exactly that.
var overrideDests = map[string]map[Scope]string{
	"opencode": {ScopeHome: ".config/opencode/skills"},
}

// Supported returns the verified table, safe to keep: a caller that reorders or
// trims the result cannot change what this package resolves.
func Supported() []Agent { return append([]Agent(nil), supported...) }

// Names returns the supported names, in table order: the agents that can be
// detected. The portable location is not one of them — it is not a program, and
// detecting it would serve an agent nobody asked for.
func Names() []string {
	names := make([]string, 0, len(supported))
	for _, a := range supported {
		names = append(names, a.Name)
	}
	return names
}

// List renders every name a human can type, the portable one included, for the
// one-line message every refusal carries and for `-h`. Two spellings of the same
// list would drift apart.
func List() string { return strings.Join(append(Names(), Standard), ", ") }

// Resolve returns the agent for name. The lookup trims spaces and ignores
// capitals, because a human types both; it never guesses a name that is not in
// the table. Every error names the supported agents, so a blocked run always
// shows the way out.
func Resolve(name string) (Agent, error) {
	key := strings.ToLower(strings.TrimSpace(name))
	if key == "" {
		return Agent{}, fmt.Errorf("the agent name is missing; the supported ones are: %s (for example odachi pack --agents claude,codex, or odachi pack --agents standard for the folder every client reads)", List())
	}
	if a, ok := find(key); ok {
		return a, nil
	}
	if reason, ok := refusals[key]; ok {
		return Agent{}, fmt.Errorf("%s: %s; the supported ones are: %s: %s", key, reason, List(), escape)
	}
	return Agent{}, fmt.Errorf("unknown agent: %s; the supported ones are: %s: %s", key, List(), escape)
}

// escape is the tail every refusal carries: the two ways to install for a client
// this table does not know, today or later. A refusal that only says "no" is a
// dead end, and a command whose reason to exist is serving any client cannot
// afford one.
const escape = "if your client is not in the table, use --agents standard (the folder every client reads: .agents/skills in the project, ~/.agents/skills in the home) or --dest <folder> (the path your client reads)"

// Labels parses --agents when --dest supplies the directory. The names stop being
// table lookups and become the label a human reads in the summary, so a client
// this package has never heard of is served as long as the human knows its path.
// Repeats collapse and order is kept, the same way ResolveAll does it.
func Labels(list string) []string {
	labels := make([]string, 0, strings.Count(list, ",")+1)
	for _, raw := range strings.Split(list, ",") {
		name := strings.ToLower(strings.TrimSpace(raw))
		if name == "" || slices.Contains(labels, name) {
			continue
		}
		labels = append(labels, name)
	}
	return labels
}

// ParseScope reads the --scope flag. Empty is the table's own scope, which is the
// default and not an error; anything else must be one of the two roots, and the
// refusal names the flag and its two values.
func ParseScope(value string) (Scope, error) {
	switch scope := Scope(strings.ToLower(strings.TrimSpace(value))); scope {
	case "":
		return "", nil
	case ScopeHome, ScopeProject:
		return scope, nil
	default:
		return "", fmt.Errorf("invalid scope: %s; use --scope project or --scope home (for example odachi pack --agents opencode --scope home)", strings.TrimSpace(value))
	}
}

// In returns the same agent installed under scope instead of its own default
// root. The name still decides the path and the scope only decides which root
// that path is joined to — except for the rows overrideDests knows about, whose
// other root keeps the config dir their own docs document. An empty scope is the
// table's answer, so the row comes back untouched.
func (a Agent) In(scope Scope) Agent {
	if scope == "" || scope == a.Scope {
		return a
	}
	out := a
	out.Scope = scope
	if dest, ok := overrideDests[a.Name][scope]; ok {
		out.Dest = dest
	}
	return out
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
// agent is installed here, which is what "each detected agent" means: a home
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

// find returns the table row for an already-normalized name. The portable row is
// looked up first because it lives beside the table and not in it: the table is
// what detection scans, and the portable location is never detected.
func find(name string) (Agent, bool) {
	if name == portable.Name {
		return portable, true
	}
	for _, a := range supported {
		if a.Name == name {
			return a, true
		}
	}
	return Agent{}, false
}
