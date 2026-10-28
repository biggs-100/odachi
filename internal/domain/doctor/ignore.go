package doctor

import "strings"

// IgnoredBy reports whether the repository's .gitignore content keeps dir out of
// the tree. dir is a project-relative directory, spelled without slashes:
// "proof".
//
// It reads the file a person writes, so it understands the shapes a person
// writes for a generated folder and nothing more: the exact path, with or without
// a leading slash and a trailing slash, its contents (`proof/*`), and the blanket
// `*` that ignores every entry at the root. Comments are not rules and a negation
// keeps the path in the tree. The review store is not a caller: it lives in the
// clone, outside the worktree.
//
// This is one of the two mechanisms doctor accepts. The other is the legacy
// per-directory `.gitignore` this tool used to write, which a project that never
// ran `odachi init` still has.
func IgnoredBy(gitignore, dir string) bool {
	dir = strings.Trim(strings.TrimSpace(dir), "/")
	if dir == "" {
		return false
	}

	for _, raw := range strings.Split(gitignore, "\n") {
		rule := strings.TrimSpace(strings.TrimSuffix(raw, "\r"))
		if rule == "" || strings.HasPrefix(rule, "#") || strings.HasPrefix(rule, "!") {
			continue
		}
		if ignoresDirectory(rule, dir) {
			return true
		}
	}
	return false
}

// ignoresDirectory reads one gitignore pattern as the directory rule it is.
func ignoresDirectory(rule, dir string) bool {
	pattern := strings.TrimSuffix(rule, "/")
	if pattern == "*" || pattern == "/*" {
		// A blanket ignore at the root covers every entry, the evidence included.
		return true
	}
	pattern = strings.TrimPrefix(strings.TrimSuffix(pattern, "/*"), "/")
	return pattern == dir
}
