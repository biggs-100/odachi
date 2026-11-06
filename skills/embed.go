// Package skills is the bundled skill set: the starter skills of this
// repository compiled into the binary with //go:embed.
//
// The source of truth is the directory that holds this file — skills/<name>/
// SKILL.md, the references and the assets beside it — not a copy. `odachi pack`
// reads them from the binary, so it needs no network and no repository, and
// `odachi skills --check` still validates the same files on disk.
package skills

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"sort"
	"strings"
)

// skillPattern embeds every file of this directory, recursively. All reads only
// the skill folders, so the Go files that live here never reach a destination:
// the pattern is wide because a build must not fail for a skill that ships no
// references yet.
//
//go:embed *
var bundled embed.FS

// BootstrapName is the one skill that explains how a project starts. It is a
// different kind of thing from the others: the rest are craft a project carries,
// and this one is how an agent arrives at a folder where nothing exists yet. It
// lives in the human's home for that reason, and it is the only bundled skill
// `odachi init` keeps out of the project copy.
const BootstrapName = "bootstrap"

// SkillFile is the one file every skill carries. The agents read exactly this
// name.
const SkillFile = "SKILL.md"

// File is one file of a skill folder, named the way the folder names it:
// "SKILL.md" or "references/numbers.md", with forward slashes on every OS.
type File struct {
	Path    string
	Content []byte
}

// Skill is one starter skill: its name (its directory) and every file the folder
// carries, sorted by path. SKILL.md is always there; the references are the depth
// the skill grew without losing its sixty-line body.
type Skill struct {
	Name  string
	Files []File
}

// Main is the SKILL.md bytes: the file every client reads, and the one the
// registry names. A build without it is broken, and the caller says so.
func (s Skill) Main() []byte {
	for _, file := range s.Files {
		if file.Path == SkillFile {
			return file.Content
		}
	}
	return nil
}

// All returns every bundled skill, sorted by name. Two runs over the same binary
// produce the same order, which is what makes the summary line stable.
func All() ([]Skill, error) {
	matches, err := fs.Glob(bundled, "*/"+SkillFile)
	if err != nil {
		return nil, fmt.Errorf("no pude leer las skills incluidas: %w", err)
	}

	all := make([]Skill, 0, len(matches))
	for _, match := range matches {
		name := path.Dir(match)
		files, err := folder(bundled, name)
		if err != nil {
			return nil, err
		}
		all = append(all, Skill{Name: name, Files: files})
	}
	sort.SliceStable(all, func(i, j int) bool { return all[i].Name < all[j].Name })
	return all, nil
}

// folder reads one skill folder: the SKILL.md every skill carries, the
// references it adds, and the assets it copies. The globs are the shape
// `odachi skills --check` enforces, so a file the check refuses never ships.
func folder(fsys fs.FS, name string) ([]File, error) {
	paths := []string{name + "/" + SkillFile}
	references, err := fs.Glob(fsys, name+"/references/*.md")
	if err != nil {
		return nil, fmt.Errorf("no pude leer %s/references: %w", name, err)
	}
	paths = append(paths, references...)
	assets, err := fs.Glob(fsys, name+"/assets/*")
	if err != nil {
		return nil, fmt.Errorf("no pude leer %s/assets: %w", name, err)
	}
	paths = append(paths, assets...)

	files := make([]File, 0, len(paths))
	for _, file := range paths {
		content, err := fs.ReadFile(fsys, file)
		if err != nil {
			return nil, fmt.Errorf("no pude leer %s: %w", file, err)
		}
		files = append(files, File{Path: strings.TrimPrefix(file, name+"/"), Content: content})
	}
	slices.SortStableFunc(files, func(a, b File) int { return strings.Compare(a.Path, b.Path) })
	return files, nil
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
// project carries. The order is the same All returns, so two runs of `odachi init`
// write the same bytes in the same order.
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
