package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// craftPath is the exact SKILL.md an agent reads for a name, in the order every
// client resolves it: the project's copy first, then the one the machine keeps.
// A skill the project carries is the live one; the home copy is what a project
// with no copy of its own reads.
func craftPath(dir, name string) (string, bool) {
	for _, path := range craftPaths(dir, name) {
		st, err := os.Stat(path)
		if err == nil && !st.IsDir() {
			return path, true
		}
	}
	return "", false
}

// craftPaths lists where a name may live, in resolution order.
func craftPaths(dir, name string) []string {
	paths := []string{filepath.Join(dir, ".agents", "skills", name, "SKILL.md")}
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".agents", "skills", name, "SKILL.md"))
	}
	return paths
}

// craftNames lists the installed skill names, deduplicated by the same order:
// the project's copy wins and keeps its place in the sorted list.
func craftNames(dir string) []string {
	seen := map[string]bool{}
	names := make([]string, 0, 32)
	scan := func(root string) {
		entries, err := os.ReadDir(root)
		if err != nil {
			return
		}
		for _, entry := range entries {
			name := entry.Name()
			if !entry.IsDir() || seen[name] {
				continue
			}
			seen[name] = true
			names = append(names, name)
		}
	}
	scan(filepath.Join(dir, ".agents", "skills"))
	if home, err := os.UserHomeDir(); err == nil {
		scan(filepath.Join(home, ".agents", "skills"))
	}
	sort.Strings(names)
	return names
}

// craftList renders the names for an error that tells the reader what exists,
// so the next run is one line instead of a file listing.
func craftList(names []string) string {
	if len(names) == 0 {
		return "there are no skills installed; run odachi init or odachi pack"
	}
	return strings.Join(names, ", ")
}
