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

// BootstrapName is the one skill that explains how a project starts. It is a
// different kind of thing from the others: the rest are craft a project carries,
// and this one is how an agent arrives at a folder where nothing exists yet. It
// lives in the human's home for that reason, and it is the only bundled skill
// `odachi init` keeps out of the project copy.
const BootstrapName = "bootstrap"

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

// Bootstrap returns the entry-point skill. ok is false when a payload was built
// without it, which can only happen on a broken build: the caller says so instead
// of installing nothing and reporting success.
func Bootstrap() (Skill, bool) {
	all, err := All()
	if err != nil {
		return Skill{}, false
	}
	for _, s := range all {
		if s.Name == BootstrapName {
			return s, true
		}
	}
	return Skill{}, false
}

// Curriculum returns every bundled skill except the entry point: the craft a
// project carries. The order is the same All returns, so two runs of `odachi
// init` write the same bytes in the same order.
func Curriculum() ([]Skill, error) {
	all, err := All()
	if err != nil {
		return nil, err
	}
	craft := make([]Skill, 0, len(all))
	for _, s := range all {
		if s.Name == BootstrapName {
			continue
		}
		craft = append(craft, s)
	}
	return craft, nil
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
