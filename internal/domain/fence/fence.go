// Package fence evaluates the "not" rules a human wrote in fence.toml: things
// the project agreed not to touch. It is pure: no filesystem, no git, no
// network, so the same rules give the same answer on every machine.
package fence

import "strings"

// Rule is one "not" entry: a human-readable name plus the paths and
// dependencies it forbids.
type Rule struct {
	Not   string
	Paths []string
	Deps  []string
}

// Violation is one offending path or dependency, attributed to the rule that
// forbids it.
type Violation struct {
	Rule string
	What string
	Kind string
}

const (
	// KindPath marks a violation caused by a changed file path.
	KindPath = "path"
	// KindDep marks a violation caused by an added dependency.
	KindDep = "dep"
	// KindRule marks a rule of the base fence the candidate does not have.
	KindRule = "rule"
)

// Loosening is one piece of the base policy the candidate did not keep: a rule
// that disappeared, or a path or a dependency that rule no longer forbids.
type Loosening struct {
	Rule string
	// What is the path or the dependency that disappeared; empty when the rule
	// itself is gone.
	What string
	Kind string
}

// Loosened compares the fence of the base ref with the candidate's. Every rule
// the base had must survive under the same name with at least the paths and
// dependencies it had; adding rules or entries is stricter and never reported.
// The comparison is exact on purpose: a glob is a string here, and a changed
// string is a changed policy, not a smarter one.
func Loosened(base, candidate []Rule) []Loosening {
	paths := make(map[string]map[string]bool, len(candidate))
	deps := make(map[string]map[string]bool, len(candidate))
	for _, rule := range candidate {
		if paths[rule.Not] == nil {
			paths[rule.Not] = make(map[string]bool, len(rule.Paths))
		}
		if deps[rule.Not] == nil {
			deps[rule.Not] = make(map[string]bool, len(rule.Deps))
		}
		for _, path := range rule.Paths {
			paths[rule.Not][path] = true
		}
		for _, dep := range rule.Deps {
			deps[rule.Not][dep] = true
		}
	}

	var loosened []Loosening
	for _, rule := range base {
		kept, ok := paths[rule.Not]
		if !ok {
			loosened = append(loosened, Loosening{Rule: rule.Not, Kind: KindRule})
			continue
		}
		for _, path := range rule.Paths {
			if !kept[path] {
				loosened = append(loosened, Loosening{Rule: rule.Not, What: path, Kind: KindPath})
			}
		}
		for _, dep := range rule.Deps {
			if !deps[rule.Not][dep] {
				loosened = append(loosened, Loosening{Rule: rule.Not, What: dep, Kind: KindDep})
			}
		}
	}
	return loosened
}

// Check reports one violation per rule per offending path, and one per rule per
// offending dependency. Nil or empty input produces an empty slice: a fence with
// nothing to check is not a violation.
func Check(rules []Rule, changedPaths, addedDeps []string) []Violation {
	violations := []Violation{}

	for _, rule := range rules {
		for _, path := range changedPaths {
			if matchesAnyPath(rule.Paths, path) {
				violations = append(violations, Violation{Rule: rule.Not, What: path, Kind: KindPath})
			}
		}
		for _, dep := range addedDeps {
			if matchesAnyDep(rule.Deps, dep) {
				violations = append(violations, Violation{Rule: rule.Not, What: dep, Kind: KindDep})
			}
		}
	}

	return violations
}

// matchesAnyPath reports whether any pattern forbids the path.
func matchesAnyPath(patterns []string, path string) bool {
	for _, pattern := range patterns {
		if matchPath(pattern, path) {
			return true
		}
	}
	return false
}

// matchesAnyDep reports whether any dependency pattern names a segment of the
// dependency.
func matchesAnyDep(patterns []string, dep string) bool {
	segments := strings.Split(strings.ToLower(dep), "/")
	for _, pattern := range patterns {
		for _, segment := range segments {
			if segment == strings.ToLower(pattern) {
				return true
			}
		}
	}
	return false
}

// matchPath matches case-insensitively. A pattern without "/" matches any
// single path segment ("payment*" matches "internal/payments/x.go"). A pattern
// with "/" matches the whole path, where "**" matches zero or more segments.
func matchPath(pattern, path string) bool {
	pattern = strings.ToLower(pattern)
	path = strings.ToLower(path)

	if !strings.Contains(pattern, "/") {
		for _, segment := range strings.Split(path, "/") {
			if matchSegment(pattern, segment) {
				return true
			}
		}
		return false
	}

	patternSegments := strings.Split(pattern, "/")
	pathSegments := strings.Split(path, "/")
	return matchSegments(patternSegments, pathSegments)
}

// matchSegments matches path segments against pattern segments, where a "**"
// pattern segment absorbs zero or more path segments.
func matchSegments(pattern, path []string) bool {
	if len(pattern) == 0 {
		return len(path) == 0
	}
	if pattern[0] == "**" {
		for i := 0; i <= len(path); i++ {
			if matchSegments(pattern[1:], path[i:]) {
				return true
			}
		}
		return false
	}
	if len(path) == 0 {
		return false
	}
	if !matchSegment(pattern[0], path[0]) {
		return false
	}
	return matchSegments(pattern[1:], path[1:])
}

// matchSegment matches one segment against a pattern where "*" matches any run
// of characters. Both arguments are already lowercased.
func matchSegment(pattern, segment string) bool {
	p, s := 0, 0
	star, resume := -1, 0

	for s < len(segment) {
		switch {
		case p < len(pattern) && pattern[p] == segment[s]:
			p++
			s++
		case p < len(pattern) && pattern[p] == '*':
			star = p
			resume = s
			p++
		case star >= 0:
			p = star + 1
			resume++
			s = resume
		default:
			return false
		}
	}

	for p < len(pattern) && pattern[p] == '*' {
		p++
	}
	return p == len(pattern)
}
