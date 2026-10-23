package review

import (
	"fmt"
	"strings"
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
		if isGeneratedStore(path) {
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

// isGeneratedStore reports whether the path belongs to a store this tool writes.
// Storing a verdict creates .odachi/reviews/.gitignore and running a proof creates
// proof/.gitignore; both are plumbing, and counting them as changes turned a
// documentation-only candidate into a focused review the second time it was asked.
// Everything else under .odachi/ is a human decision (the fence, the profile) and
// stays in the candidate.
func isGeneratedStore(path string) bool {
	lower := normalizePath(path)
	return strings.HasPrefix(lower, Dir+"/") || strings.HasPrefix(lower, "proof/")
}

// touchesSignal reports whether any segment of the path starts with a signal.
func touchesSignal(path string) bool {
	lower := normalizePath(path)
	for _, segment := range strings.Split(lower, "/") {
		for _, signal := range signals {
			if strings.HasPrefix(segment, signal) {
				return true
			}
		}
	}
	return false
}

// isDocumentation reports whether the path is a document: Markdown, plain text, a
// README, or anything inside a docs directory. A `.go` file inside `docs/` is
// code, so the extension decides before the directory does.
func isDocumentation(path string) bool {
	lower := normalizePath(path)
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
