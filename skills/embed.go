// Package skills is the bundled skill set: the starter skills of this
// repository compiled into the binary with //go:embed.
//
// The source of truth is the directory that holds this file — skills/<name>/
// SKILL.md — not a copy. `odachi pack` reads them from the binary, so it needs
// no network and no repository, and `odachi skills --check` still validates the
// same files on disk.
package skills

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

// skillPattern is one SKILL.md per skill directory, and nothing else: the
// versioned Go file that lives next to them is never part of the payload.
//
//go:embed */SKILL.md
var bundled embed.FS

// Skill is one starter skill: its name (its directory) and its exact bytes.
type Skill struct {
	Name    string
	Content []byte
}

// All returns every bundled skill, sorted by name. Two runs over the same binary
// produce the same order, which is what makes the summary line stable.
func All() ([]Skill, error) {
	matches, err := fs.Glob(bundled, "*/SKILL.md")
	if err != nil {
		return nil, fmt.Errorf("no pude leer las skills incluidas: %w", err)
	}

	all := make([]Skill, 0, len(matches))
	for _, match := range matches {
		content, err := bundled.ReadFile(match)
		if err != nil {
			return nil, fmt.Errorf("no pude leer %s: %w", match, err)
		}
		all = append(all, Skill{Name: path.Dir(match), Content: content})
	}
	sort.SliceStable(all, func(i, j int) bool { return all[i].Name < all[j].Name })
	return all, nil
}

// Names returns the bundled skill names, in the order All returns them.
func Names() []string {
	all, err := All()
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(all))
	for _, s := range all {
		names = append(names, strings.TrimSpace(s.Name))
	}
	return names
}
