package review

import (
	"fmt"
	"slices"
	"strings"

	"github.com/biggs-100/odachi/internal/domain/persona"
)

// Tier is how much review a candidate asks for, decided by evidence and never by
// volume: what the changed paths touch, not how much text they carry.
type Tier string

const (
	// Passive is a candidate that only touches documentation: there is nothing
	// for a specialist lens to look at.
	Passive Tier = "passive"
	// Focused is ordinary code: one lens, the one that asks "does this hold".
	Focused Tier = "focused"
	// Full is a candidate that touches something that can hurt: security,
	// identity, money, data shape, or the machinery that ships it.
	Full Tier = "full"
)

// The four lenses. They are names, not branches of code: the prompt for each one
// is embedded content in internal/assets/review, and this package only decides
// which ones a candidate asks for.
const (
	Risk        = "risk"
	Readability = "readability"
	Reliability = "reliability"
	Resilience  = "resilience"
)

// order is the fixed order of the four lenses. It is the order `Lenses`, `Add`
// and the prompt use, so the same candidate always produces the same prompt
// bytes and two runs are comparable.
var order = []string{Risk, Readability, Reliability, Resilience}

// signals are the paths that make a candidate full: the same list the fence
// already knows, not a second taxonomy invented here. Each signal matches the
// START of a path segment, case-insensitively, so `migrat` catches `migrations/`
// and `ci` catches `ci.yml` without catching `decision.md`.
var signals = []string{
	// Identity and money: a defect here can hurt a person.
	"auth", "login", "session", "token", "password", "secret",
	"permission", "role", "payment", "billing", "migrat", "schema",
	// Public contracts: other programs and other people depend on the shape.
	"api", "contract", "openapi", "proto", "graphql", "sdk",
	// Concurrency: the defect is a race, and a race is not visible in one path.
	"concurrency", "race", "mutex", "lock", "async", "worker", "queue",
	// Delivery: the change decides what reaches a machine.
	".env", "docker", "ci", ".github", "deploy", "release", "goreleaser", ".goreleaser",
	"makefile", "go.mod", "terraform", "helm", "k8s", "pipeline",
}

// Classify decides the tier from the changed paths alone. It is pure,
// deterministic and volume-blind: the signature has no line counts and no diff,
// so a thousand lines of documentation cannot upgrade themselves and five lines
// of auth cannot downgrade themselves.
//
// The order of the two questions matters. A sensitive signal wins over the
// documentation exemption: over-reviewing a document that mentions auth costs one
// prompt, under-reviewing auth costs a defect.
func Classify(changedPaths []string) Tier {
	candidate := make([]string, 0, len(changedPaths))
	for _, path := range changedPaths {
		if isPlumbing(path) {
			continue
		}
		candidate = append(candidate, path)
	}

	if len(candidate) == 0 {
		return Passive
	}

	sensitive := false
	documentation := true
	for _, path := range candidate {
		if touchesSignal(path) {
			sensitive = true
		}
		if !isDocumentation(path) {
			documentation = false
		}
	}

	switch {
	case sensitive:
		return Full
	case documentation:
		return Passive
	default:
		return Focused
	}
}

// Lenses is the selection: passive asks for nothing, focused asks for one lens,
// full asks for the four in the fixed order. The slice is a copy, because the
// order is the product and a caller must not be able to sort it away.
func Lenses(t Tier) []string {
	switch t {
	case Focused:
		return []string{Reliability}
	case Full:
		return append([]string(nil), order...)
	default:
		return nil
	}
}

// Need is the per-file lens demand coverage asks about: for every changed path,
// the lenses its own path class requires, raised by the same content assessment
// the review selection uses, and omitting the paths whose class asks for
// nothing. The tier of one candidate is one question, but assignment is per
// file: a candidate may need reliability in one file and all four lenses in
// another, and coverage compares the verdict's rows against exactly this.
func Need(changedPaths []string, diff string) map[string][]string {
	floor := FloorFor(changedPaths, diff)

	need := make(map[string][]string, len(changedPaths))
	for _, path := range changedPaths {
		rows := floor.rowsFor(path)
		lenses := Lenses(assessClass([]string{path}, rows))
		if len(lenses) == 0 {
			continue
		}
		need[path] = lenses
	}
	return need
}

// Reason says in one line what makes this candidate ask for the review it asks
// for: the signal and the path when a path decided, or the added lines when the
// content raised ordinary code. It never says "high risk" in the abstract,
// because a reason a human cannot check is a label.
func Reason(paths []string, diff string) string {
	candidate := make([]string, 0, len(paths))
	for _, path := range paths {
		if !isPlumbing(path) {
			candidate = append(candidate, path)
		}
	}
	if len(candidate) == 0 {
		return "nothing changed"
	}

	switch Assess(paths, diff) {
	case Passive:
		return "documents only"
	case Full:
		if path, signal, ok := firstSignal(candidate); ok {
			return "touches " + signal + " (" + path + ")"
		}
		if signal, ok := firstContentSignal(FloorFor(paths, diff)); ok {
			return "added lines touch " + signal
		}
		// Unreachable while Assess reaches Full only through a path signal or
		// those two content signals; kept so a new route says something true
		// instead of panicking or inventing a reason.
		return "a signal the fence knows"
	default:
		return "ordinary code"
	}
}

// firstSignal names the first path that made the candidate full, with the
// signal it matched. The order is the one git reported, so two runs agree.
func firstSignal(paths []string) (path, signal string, ok bool) {
	for _, path := range paths {
		if signal, ok := matchedSignal(path); ok {
			return path, signal, true
		}
	}
	return "", "", false
}

// contentSignals are the two categories that raise ordinary code to full, in
// the fixed order Reason names them.
var contentSignals = []string{"concurrency", "secrets"}

// firstContentSignal names what the added lines carried when the path class
// alone would not have asked for the four lenses.
func firstContentSignal(floor Floor) (string, bool) {
	for _, row := range floor.Rows {
		for _, signal := range contentSignals {
			if slices.Contains(row.Signals, signal) {
				return signal, true
			}
		}
	}
	return "", false
}

// IsLens reports whether name is one of the four.
func IsLens(name string) bool {
	for _, lens := range order {
		if strings.EqualFold(strings.TrimSpace(name), lens) {
			return true
		}
	}
	return false
}

// Add is what --lens does: the lenses the evidence asks for, plus the ones the
// agent or the human asked for. Extra lenses add; none of them can remove what
// the evidence requires. The answer is always in the fixed order, no duplicates.
func Add(required, extra []string) []string {
	merged := make([]string, 0, len(required)+len(extra))
	merged = append(merged, required...)
	merged = append(merged, extra...)
	return inOrder(merged)
}

// ParseLenses reads the --lens value: one name, a comma list, or `all`. An
// invented name is refused, naming the four, because a typo that quietly
// reviewed nothing would be a review that never happened.
func ParseLenses(spec string) ([]string, error) {
	trimmed := strings.TrimSpace(spec)
	if trimmed == "" {
		return nil, nil
	}

	parts := strings.Split(trimmed, ",")
	names := make([]string, 0, len(parts))
	for _, part := range parts {
		name := strings.ToLower(strings.TrimSpace(part))
		switch {
		case name == "":
			continue
		case name == "all":
			names = append(names, order...)
		case IsLens(name):
			names = append(names, name)
		default:
			return nil, fmt.Errorf("the lens %q does not exist: the lenses are %s (or all)", name, strings.Join(order, ", "))
		}
	}

	if len(parts) > 1 && strings.Contains(trimmed, "all") {
		return nil, fmt.Errorf("all already is %s: do not mix it with other lenses", strings.Join(order, ", "))
	}
	return inOrder(names), nil
}

// inOrder keeps the canonical order and drops duplicates.
func inOrder(names []string) []string {
	if len(names) == 0 {
		return nil
	}

	seen := make(map[string]bool, len(names))
	for _, name := range names {
		seen[strings.ToLower(strings.TrimSpace(name))] = true
	}

	ordered := make([]string, 0, len(seen))
	for _, lens := range order {
		if seen[lens] {
			ordered = append(ordered, lens)
		}
	}
	if len(ordered) == 0 {
		return nil
	}
	return ordered
}

// isPlumbing reports whether the path is machinery odachi or its init writes,
// not work the human asked for: the review store, the proofs, and the agent
// files and skills init installs. A project that adopts odachi must not be
// charged a review for the tool's own prompt text - a skill that mentions
// secrets is not a security change to this codebase - and the tool must not
// review itself into the candidate.
//
// Everything else under .odachi/ is a human decision (the fence, the profile)
// and stays in the candidate, and so does .gitignore: an ignore rule changes the
// bytes the tree covers, so it is never plumbing.
func isPlumbing(path string) bool {
	lower := normalizePath(path)
	switch {
	case strings.HasPrefix(lower, Dir+"/"), strings.HasPrefix(lower, "proof/"):
		return true
	case lower == "agents.md":
		// The root block is instructions to the agent, not the product; a nested
		// AGENTS.md inside the tree is code the human wrote and is reviewed.
		return true
	case strings.HasPrefix(lower, ".agents/skills/"):
		// Skills are instructions to the agent. Their examples talk about money,
		// secrets and concurrency as their subject, not as this repo's change, and
		// a security lens that reads them reviews the skill's prose, not the tree.
		return true
	case odachiAgentFile(lower):
		return true
	}
	return false
}

// agentDirs are the directories whose Markdown is instructions an agent obeys,
// never documentation. An agent file grants tools and can carry a token in its
// prompt: a change there decides what the agent may do, so its Markdown is code
// and reaches a lens. The extension decides for every other directory.
var agentDirs = map[string]bool{
	".opencode/agents":      true,
	".claude/agents":        true,
	".claude/output-styles": true,
}

// odachiAgentFile reports whether the path is one of the agent files odachi
// installs: the primary mirror, the subagents and the leftovers of v0.10.0. The
// names are a closed set, not a prefix. A file the human added under the same
// directory is theirs — agent privileges and secrets live there — so
// `odachi-helper.md` is code while `odachi-worker.md` is not.
func odachiAgentFile(lower string) bool {
	dir := dirOf(lower)
	if !agentDirs[dir] {
		return false
	}
	return slices.Contains(persona.AgentNames(), baseOf(lower))
}

// dirOf is the directory part of a normalized path, without its file name.
func dirOf(path string) string {
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		return path[:i]
	}
	return ""
}

// baseOf is the file name part of a normalized path.
func baseOf(path string) string {
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		return path[i+1:]
	}
	return path
}

// touchesSignal reports whether any segment of the path starts with a signal.
func touchesSignal(path string) bool {
	_, ok := matchedSignal(path)
	return ok
}

// matchedSignal names the signal a path carries, or nothing. Case does not
// decide, and the first match in the fixed order is the one Reason reports.
func matchedSignal(path string) (string, bool) {
	lower := normalizePath(path)
	for _, segment := range strings.Split(lower, "/") {
		for _, signal := range signals {
			if strings.HasPrefix(segment, signal) {
				return signal, true
			}
		}
	}
	return "", false
}

// isDocumentation reports whether the path is a document: Markdown, plain text, a
// README, or anything inside a docs directory. A `.go` file inside `docs/` is
// code, so the extension decides before the directory does — and an agent file is
// code even when it is Markdown, because it grants tools.
func isDocumentation(path string) bool {
	lower := normalizePath(path)
	if agentDirs[dirOf(lower)] {
		return false
	}

	segments := strings.Split(lower, "/")
	base := segments[len(segments)-1]

	switch {
	case strings.HasSuffix(base, ".md"), strings.HasSuffix(base, ".txt"):
		return true
	case strings.HasPrefix(base, "readme"):
		return true
	}

	for _, segment := range segments[:len(segments)-1] {
		if segment == "docs" {
			return true
		}
	}
	return false
}

// normalizePath is the one spelling every rule here matches against: forward
// slashes and lower case, so the same tree classifies the same on Windows and
// Linux.
func normalizePath(path string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(path), "\\", "/"))
}
