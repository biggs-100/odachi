// Package skills implements `odachi skills` and `odachi skill new`: it reads
// skills/<name>/SKILL.md over the Files port, validates the set, and regenerates
// the SKILLS.md registry. The parsing and the rules live in internal/domain/skill;
// this package only orchestrates and phrases what a human reads.
package skills

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/biggs-100/odachi/internal/domain/skill"
)

// Exit codes: the same two every verb uses.
const (
	// CodeOK means the command did its job.
	CodeOK = 0
	// CodeBlocked means a violation, an existing slug or an unreadable file.
	CodeBlocked = 1
)

// The layout this command owns, relative to the project root.
const (
	Dir       = "skills"
	SkillFile = "SKILL.md"
	Registry  = skill.RegistryFile
)

// Options is one `odachi skills` or `odachi skill new` invocation.
type Options struct {
	Dir   string
	Check bool
	Write bool
	Slug  string
}

// Files is the disk port: enough to read every skill and replace the registry.
type Files interface {
	Exists(path string) (bool, error)
	List(dir string) ([]string, error)
	Read(path string) ([]byte, error)
	Write(path string, data []byte) error
}

// Result is what the CLI prints and exits with. Skills carries the parsed set so
// --json can print it without parsing twice.
type Result struct {
	Code    int
	Message string
	Lines   []string
	Files   []string
	Skills  []skill.Skill
}

// Run performs one `odachi skills`: list, --check or --write.
func Run(opts Options, files Files) Result {
	root := rootOf(opts)

	loaded, err := load(root, files)
	if err != nil {
		return failed(err.Error())
	}
	if len(loaded) == 0 {
		return noSkills()
	}

	switch {
	case opts.Write:
		return writeRegistry(root, files, loaded)
	case opts.Check:
		return check(root, files, loaded)
	default:
		return list(loaded)
	}
}

// New performs one `odachi skill new <slug>`: it validates the slug before it
// touches anything, refuses an existing skill, writes the template and brings
// the registry up to date.
func New(opts Options, files Files) Result {
	root := rootOf(opts)
	slug := opts.Slug

	if !skill.ValidSlug(slug) {
		return failed(fmt.Sprintf(
			"el slug %q no sirve: tiene que cumplir %s (minúsculas, números y guiones, empieza con una letra); elegí uno y volvé a correr odachi skill new <slug>",
			slug, skill.NamePattern))
	}

	rel := Dir + "/" + slug + "/" + SkillFile
	path := filepath.Join(root, Dir, slug, SkillFile)

	exists, err := files.Exists(path)
	if err != nil {
		return failed("no pude revisar " + rel + ": " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi skill new <slug>")
	}
	if exists {
		return failed("ya existe " + rel + ": no lo toco; elegí otro slug, o editá el que ya está y corré odachi skills --check")
	}

	if err := files.Write(path, []byte(skill.Template(slug))); err != nil {
		return failed("no pude escribir " + rel + ": " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi skill new <slug>")
	}

	loaded, err := load(root, files)
	if err != nil {
		return failed(err.Error())
	}
	registryLine, _, err := syncRegistry(root, files, loaded)
	if err != nil {
		return failed(err.Error())
	}

	lines := []string{
		rel + ": creado",
		registryLine,
		"siguiente: completá el Trigger y las tres secciones, y corré odachi skills --check",
	}
	return Result{
		Code:    CodeOK,
		Message: "listo: " + rel + " creado y SKILLS.md al día",
		Lines:   lines,
		Files:   []string{rel, Registry},
		Skills:  loaded,
	}
}

// list shows one line per skill, sorted by name: the human surface.
func list(loaded []skill.Skill) Result {
	ordered := sortedByName(loaded)
	lines := make([]string, 0, len(ordered))
	for _, s := range ordered {
		name := s.Name
		if name == "" {
			name = s.Dir
		}
		trigger := s.Trigger
		if trigger == "" {
			trigger = "(sin Trigger:)"
		}
		lines = append(lines, name+" · "+trigger+" · "+s.Path)
	}
	return Result{Code: CodeOK, Message: fmt.Sprintf("%s en %s/", count(len(ordered), "skill", "skills"), Dir), Lines: lines, Skills: ordered}
}

// check reports every violation, one Spanish line each, and the one command that
// resolves them.
func check(root string, files Files, loaded []skill.Skill) Result {
	current, hasRegistry, err := readRegistry(root, files)
	if err != nil {
		return failed(err.Error())
	}

	problems := skill.Check(loaded, current, hasRegistry)
	if len(problems) == 0 {
		return Result{
			Code:    CodeOK,
			Message: fmt.Sprintf("todo en orden: %s y SKILLS.md al día", count(len(loaded), "skill", "skills")),
			Skills:  loaded,
		}
	}

	lines := make([]string, 0, len(problems))
	for _, problem := range problems {
		lines = append(lines, problem.File+": "+problem.Reason)
	}
	return Result{
		Code:    CodeBlocked,
		Message: fmt.Sprintf("%s en las skills: corregí lo que dice cada línea", count(len(problems), "problema", "problemas")),
		Lines:   lines,
		Skills:  loaded,
	}
}

// writeRegistry regenerates SKILLS.md. Identical bytes mean no write at all, so
// the command is idempotent in bytes and in what it reports.
func writeRegistry(root string, files Files, loaded []skill.Skill) Result {
	line, written, err := syncRegistry(root, files, loaded)
	if err != nil {
		return failed(err.Error())
	}

	result := Result{Code: CodeOK, Message: line, Lines: []string{line}, Skills: loaded}
	if written {
		result.Files = []string{Registry}
	}
	return result
}

// syncRegistry writes the generated registry when its bytes differ, and reports
// the one line a human reads plus whether it wrote. It is the single writer both
// --write and `skill new` use, so the two can never disagree.
func syncRegistry(root string, files Files, loaded []skill.Skill) (line string, written bool, err error) {
	path := filepath.Join(root, Registry)
	want := skill.Registry(loaded)

	current, exists, err := readRegistry(root, files)
	if err != nil {
		return "", false, err
	}
	if exists && current == want {
		return Registry + ": sin cambios", false, nil
	}
	if err := files.Write(path, []byte(want)); err != nil {
		return "", false, fmt.Errorf("no pude escribir %s: %s; revisá los permisos y volvé a correr odachi skills --write", Registry, oneLine(err.Error()))
	}
	return fmt.Sprintf("%s: actualizado con %s", Registry, count(len(loaded), "skill", "skills")), true, nil
}

// readRegistry reads SKILLS.md when it exists. exists is false when the file is
// missing, which is a violation with its own reason, not an error.
func readRegistry(root string, files Files) (content string, exists bool, err error) {
	path := filepath.Join(root, Registry)
	exists, err = files.Exists(path)
	if err != nil {
		return "", false, fmt.Errorf("no pude revisar %s: %s; revisá los permisos y volvé a correr odachi skills --check", Registry, oneLine(err.Error()))
	}
	if !exists {
		return "", false, nil
	}
	data, err := files.Read(path)
	if err != nil {
		return "", true, fmt.Errorf("no pude leer %s: %s; revisá los permisos y volvé a correr odachi skills --check", Registry, oneLine(err.Error()))
	}
	return string(data), true, nil
}

// load reads every skills/<dir>/SKILL.md. A directory without a SKILL.md is not
// a skill, so it is skipped instead of reported.
func load(root string, files Files) ([]skill.Skill, error) {
	names, err := files.List(filepath.Join(root, Dir))
	if err != nil {
		return nil, fmt.Errorf("no pude leer %s/: %s; revisá los permisos y volvé a correr odachi skills", Dir, oneLine(err.Error()))
	}

	loaded := make([]skill.Skill, 0, len(names))
	for _, name := range names {
		rel := Dir + "/" + name + "/" + SkillFile
		path := filepath.Join(root, Dir, name, SkillFile)

		exists, err := files.Exists(path)
		if err != nil {
			return nil, fmt.Errorf("no pude revisar %s: %s; revisá los permisos y volvé a correr odachi skills", rel, oneLine(err.Error()))
		}
		if !exists {
			continue
		}

		data, err := files.Read(path)
		if err != nil {
			return nil, fmt.Errorf("no pude leer %s: %s; revisá los permisos y volvé a correr odachi skills", rel, oneLine(err.Error()))
		}
		loaded = append(loaded, skill.Parse(name, rel, string(data)))
	}
	return loaded, nil
}

// noSkills is the starting point, not a failure: it says how to create the first
// skill.
func noSkills() Result {
	message := "todavía no hay skills en " + Dir + "/: creá la primera con odachi skill new <slug>"
	return Result{Code: CodeOK, Message: message, Lines: []string{message}}
}

// failed is a blocked result with one line a human can act on.
func failed(message string) Result {
	return Result{Code: CodeBlocked, Message: message, Lines: []string{message}}
}

// rootOf defaults the project root to the current directory.
func rootOf(opts Options) string {
	if opts.Dir == "" {
		return "."
	}
	return opts.Dir
}

// sortedByName is the order the registry and the listing show.
func sortedByName(loaded []skill.Skill) []skill.Skill {
	ordered := append([]skill.Skill(nil), loaded...)
	slices.SortStableFunc(ordered, func(a, b skill.Skill) int {
		if a.Name != b.Name {
			return strings.Compare(a.Name, b.Name)
		}
		return strings.Compare(a.Path, b.Path)
	})
	return ordered
}

// count renders "1 skill" and "5 skills".
func count(n int, singular, many string) string {
	if n == 1 {
		return "1 " + singular
	}
	return fmt.Sprintf("%d %s", n, many)
}

// oneLine keeps every message to a single line, whatever an adapter said.
func oneLine(message string) string {
	return strings.Join(strings.Fields(message), " ")
}
