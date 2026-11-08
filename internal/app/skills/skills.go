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

	"github.com/biggs-100/odachi/internal/domain/persona"
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
	// UserDir is the cross-client folder a user's skill lives in, inside the
	// project or inside the home. The project's curriculum (skills/) is a
	// different thing: the skills odachi ships and packs, not where a human's own
	// skill goes.
	UserDir = ".agents/skills"
)

// Options is one `odachi skills` or `odachi skill new` invocation.
type Options struct {
	Dir   string
	Check bool
	Write bool
	Slug  string
	// Scope is where `skill new` writes a user skill: persona.SkillsProject puts
	// it in .agents/skills of the project and persona.SkillsHome in
	// ~/.agents/skills of Home. Empty is refused: the curriculum layout was never
	// a place a user skill lives, and a default would hide the decision.
	Scope string
	// Home is the human's home, resolved by the CLI. Only the home scope uses it;
	// the app never touches os.
	Home string
}

// layout is where one run reads and writes: the folder that holds the skill
// directories, the folder that holds SKILLS.md, and how a human reads each one.
// The curriculum keeps its registry at the project root; a user's skill keeps it
// beside the skills, so the folder a human opens is the whole thing.
type layout struct {
	skillsDir     string
	registryDir   string
	label         string
	registryLabel string
	retry         string
}

// curriculumLayout is the project's own catalog: skills/ here, SKILLS.md at the
// root. It is what `skills` and `skills --check/--write` read.
func curriculumLayout(opts Options) layout {
	root := rootOf(opts)
	return layout{
		skillsDir:     filepath.Join(root, Dir),
		registryDir:   root,
		label:         Dir,
		registryLabel: Registry,
		retry:         "odachi skills",
	}
}

// userLayout is where `skill new` writes: the cross-client folder of the project
// or of the home. The second return is the refusal, empty when the scope answers.
func userLayout(opts Options) (layout, string) {
	base := filepath.Join(rootOf(opts), filepath.FromSlash(UserDir))
	switch opts.Scope {
	case persona.SkillsProject:
		return layout{
			skillsDir:     base,
			registryDir:   base,
			label:         UserDir,
			registryLabel: UserDir + "/" + Registry,
			retry:         "odachi skill new",
		}, ""
	case persona.SkillsHome:
		if strings.TrimSpace(opts.Home) == "" {
			return layout{}, "could not find your home to write ~/" + UserDir + "/<slug>: run odachi skill new <slug> --scope project, or set a home and run it again"
		}
		base = filepath.Join(opts.Home, filepath.FromSlash(UserDir))
		return layout{
			skillsDir:     base,
			registryDir:   base,
			label:         "~/" + UserDir,
			registryLabel: "~/" + UserDir + "/" + Registry,
			retry:         "odachi skill new",
		}, ""
	default:
		return layout{}, "odachi skill new needs --scope: --scope project writes " + UserDir + "/<slug> in this project, and --scope home writes ~/" + UserDir + "/<slug>"
	}
}

// Files is the disk port: enough to read every skill and replace the registry.
type Files interface {
	Exists(path string) (bool, error)
	// IsDir decides what a listed entry is. It is a question about the entry, not
	// about the error a stat returns for its children: POSIX says ENOTDIR where
	// Windows says the path does not exist, and inferring "not a skill" from
	// that difference makes the same code pass on one platform and fail on the
	// other.
	IsDir(path string) (bool, error)
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

// Run performs one `odachi skills`: list, --check or --write. It reads the
// project's curriculum, the folder odachi ships, never the user's skill folder.
func Run(opts Options, files Files) Result {
	lay := curriculumLayout(opts)

	loaded, err := load(lay, files)
	if err != nil {
		return failed(err.Error())
	}
	if len(loaded) == 0 {
		return noSkills()
	}

	switch {
	case opts.Write:
		return writeRegistry(lay, files, loaded)
	case opts.Check:
		return check(lay, files, loaded)
	default:
		return list(loaded)
	}
}

// New performs one `odachi skill new <slug> --scope project|home`: it validates
// the scope and the slug before it touches anything, refuses an existing skill,
// writes the template where the scope says and brings that folder's registry up
// to date.
func New(opts Options, files Files) Result {
	lay, problem := userLayout(opts)
	if problem != "" {
		return failed(problem)
	}
	slug := opts.Slug

	if !skill.ValidSlug(slug) {
		return failed(fmt.Sprintf(
			"the slug %q does not work: it must match %s (lowercase letters, numbers and dashes, starting with a letter); choose one and run odachi skill new <slug> --scope %s again",
			slug, skill.NamePattern, opts.Scope))
	}
	// Every failure names the command that retries with the same words, so the
	// advice cannot be refused by the same verb it names.
	lay.retry = "odachi skill new " + slug + " --scope " + opts.Scope

	rel := lay.label + "/" + slug + "/" + SkillFile
	path := filepath.Join(lay.skillsDir, slug, SkillFile)

	exists, err := files.Exists(path)
	if err != nil {
		return failed("could not check " + rel + ": " + oneLine(err.Error()) + "; check the permissions and run " + lay.retry + " again")
	}
	if exists {
		// The skill is the human's now; the registry is generated. A run that
		// finds the file repairs the registry and says so, which is the way out of
		// a run whose registry write failed after the skill landed.
		loaded, err := load(lay, files)
		if err != nil {
			return failed(err.Error())
		}
		registryLine, _, err := syncRegistry(lay, files, loaded)
		if err != nil {
			return failed(err.Error())
		}
		return failed("already exists " + rel + ": not touching it; " + registryLine)
	}

	if err := files.Write(path, []byte(skill.Template(slug))); err != nil {
		return failed("could not write " + rel + ": " + oneLine(err.Error()) + "; check the permissions and run " + lay.retry + " again")
	}

	loaded, err := load(lay, files)
	if err != nil {
		return failed(err.Error())
	}
	registryLine, _, err := syncRegistry(lay, files, loaded)
	if err != nil {
		return failed(err.Error())
	}

	lines := []string{
		rel + ": created",
		registryLine,
		"next: complete the Trigger and the three sections; the file is the source",
	}
	return Result{
		Code:    CodeOK,
		Message: "ready: " + rel + " created and " + lay.registryLabel + " up to date",
		Lines:   lines,
		Files:   []string{rel, lay.registryLabel},
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
			trigger = "(no Trigger:)"
		}
		lines = append(lines, name+" · "+trigger+" · "+s.Path)
	}
	return Result{Code: CodeOK, Message: fmt.Sprintf("%s en %s/", count(len(ordered), "skill", "skills"), Dir), Lines: lines, Skills: ordered}
}

// check reports every violation, one English line each, and the one command that
// resolves them.
func check(lay layout, files Files, loaded []skill.Skill) Result {
	current, hasRegistry, err := readRegistry(lay, files)
	if err != nil {
		return failed(err.Error())
	}

	problems := skill.Check(loaded, current, hasRegistry)
	if len(problems) == 0 {
		return Result{
			Code:    CodeOK,
			Message: fmt.Sprintf("all good: %s and SKILLS.md up to date", count(len(loaded), "skill", "skills")),
			Skills:  loaded,
		}
	}

	lines := make([]string, 0, len(problems))
	for _, problem := range problems {
		lines = append(lines, problem.File+": "+problem.Reason)
	}
	return Result{
		Code:    CodeBlocked,
		Message: fmt.Sprintf("%s in the skills: fix what each line says", count(len(problems), "problem", "problems")),
		Lines:   lines,
		Skills:  loaded,
	}
}

// writeRegistry regenerates SKILLS.md. Identical bytes mean no write at all, so
// the command is idempotent in bytes and in what it reports.
func writeRegistry(lay layout, files Files, loaded []skill.Skill) Result {
	line, written, err := syncRegistry(lay, files, loaded)
	if err != nil {
		return failed(err.Error())
	}

	result := Result{Code: CodeOK, Message: line, Lines: []string{line}, Skills: loaded}
	if written {
		result.Files = []string{lay.registryLabel}
	}
	return result
}

// syncRegistry writes the generated registry when its bytes differ, and reports
// the one line a human reads plus whether it wrote. It is the single writer both
// --write and `skill new` use, so the two can never disagree.
func syncRegistry(lay layout, files Files, loaded []skill.Skill) (line string, written bool, err error) {
	path := filepath.Join(lay.registryDir, Registry)
	want := skill.Registry(loaded)

	current, exists, err := readRegistry(lay, files)
	if err != nil {
		return "", false, err
	}
	if exists && current == want {
		return lay.registryLabel + ": unchanged", false, nil
	}
	if err := files.Write(path, []byte(want)); err != nil {
		return "", false, fmt.Errorf("could not write %s: %s; check the permissions and run %s again", lay.registryLabel, oneLine(err.Error()), lay.retry)
	}
	return fmt.Sprintf("%s: updated with %s", lay.registryLabel, count(len(loaded), "skill", "skills")), true, nil
}

// readRegistry reads SKILLS.md when it exists. exists is false when the file is
// missing, which is a violation with its own reason, not an error.
func readRegistry(lay layout, files Files) (content string, exists bool, err error) {
	path := filepath.Join(lay.registryDir, Registry)
	exists, err = files.Exists(path)
	if err != nil {
		return "", false, fmt.Errorf("could not check %s: %s; check the permissions and run %s again", lay.registryLabel, oneLine(err.Error()), lay.retry)
	}
	if !exists {
		return "", false, nil
	}
	data, err := files.Read(path)
	if err != nil {
		return "", true, fmt.Errorf("could not read %s: %s; check the permissions and run %s again", lay.registryLabel, oneLine(err.Error()), lay.retry)
	}
	return string(data), true, nil
}

// load reads every <slug>/SKILL.md under the layout's skills folder. A directory
// without a SKILL.md is not a skill, so it is skipped instead of reported.
func load(lay layout, files Files) ([]skill.Skill, error) {
	names, err := files.List(lay.skillsDir)
	if err != nil {
		return nil, fmt.Errorf("could not read %s/: %s; check the permissions and run %s again", lay.label, oneLine(err.Error()), lay.retry)
	}

	loaded := make([]skill.Skill, 0, len(names))
	for _, name := range names {
		rel := lay.label + "/" + name + "/" + SkillFile
		dirPath := filepath.Join(lay.skillsDir, name)

		isDir, err := files.IsDir(dirPath)
		if err != nil {
			return nil, fmt.Errorf("could not check %s: %s; check the permissions and run %s again", rel, oneLine(err.Error()), lay.retry)
		}
		if !isDir {
			// A file here is not a skill and not an error: skills/embed.go lives
			// next to the skill directories on purpose, and the registry itself is
			// a file in the user folder.
			continue
		}

		path := filepath.Join(dirPath, SkillFile)
		exists, err := files.Exists(path)
		if err != nil {
			return nil, fmt.Errorf("could not check %s: %s; check the permissions and run %s again", rel, oneLine(err.Error()), lay.retry)
		}
		if !exists {
			continue
		}

		data, err := files.Read(path)
		if err != nil {
			return nil, fmt.Errorf("could not read %s: %s; check the permissions and run %s again", rel, oneLine(err.Error()), lay.retry)
		}

		parsed := skill.Parse(name, rel, string(data))
		references, problems, err := referencesOf(dirPath, rel, lay.retry, files)
		if err != nil {
			return nil, err
		}
		assets, assetProblems, err := assetsOf(dirPath, rel, lay.retry, files)
		if err != nil {
			return nil, err
		}
		parsed.References = references
		parsed.Assets = assets
		parsed.FolderProblems = append(problems, assetProblems...)
		loaded = append(loaded, parsed)
	}
	return loaded, nil
}

// referencesOf reads what a skill folder holds under references/, as paths
// relative to the folder. The folder is optional; when it is not a folder, that
// is a problem of the skill, not a reason to refuse the run that lists it.
func referencesOf(dir, rel, retry string, files Files) ([]string, []skill.Problem, error) {
	refDir := filepath.Join(dir, "references")
	label := strings.TrimSuffix(rel, "/"+SkillFile) + "/references"

	exists, err := files.Exists(refDir)
	if err != nil {
		return nil, nil, fmt.Errorf("could not check %s: %s; check the permissions and run %s again", label, oneLine(err.Error()), retry)
	}
	if !exists {
		return nil, nil, nil
	}

	isDir, err := files.IsDir(refDir)
	if err != nil {
		return nil, nil, fmt.Errorf("could not check %s: %s; check the permissions and run %s again", label, oneLine(err.Error()), retry)
	}
	if !isDir {
		return nil, []skill.Problem{{
			File:   label,
			Reason: "it is a file, not a folder: a reference is a .md file inside references/",
		}}, nil
	}

	names, err := files.List(refDir)
	if err != nil {
		return nil, nil, fmt.Errorf("could not read %s/: %s; check the permissions and run %s again", label, oneLine(err.Error()), retry)
	}

	references := make([]string, 0, len(names))
	problems := []skill.Problem{}
	for _, name := range names {
		entry := filepath.Join(refDir, name)
		entryIsDir, err := files.IsDir(entry)
		if err != nil {
			return nil, nil, fmt.Errorf("could not check %s/%s: %s; check the permissions and run %s again", label, name, oneLine(err.Error()), retry)
		}
		if entryIsDir {
			problems = append(problems, skill.Problem{
				File:   label + "/" + name,
				Reason: "it is a folder: a reference is a .md file, not a tree",
			})
			continue
		}
		references = append(references, "references/"+name)
	}
	slices.Sort(references)
	return references, problems, nil
}

// assetsOf reads what a skill folder holds under assets/, as paths relative to
// the folder. An asset is a file the skill copies or uses -a template, a schema-
// so any extension is allowed and a tree is refused: a folder is not a file.
func assetsOf(dir, rel, retry string, files Files) ([]string, []skill.Problem, error) {
	assetDir := filepath.Join(dir, "assets")
	label := strings.TrimSuffix(rel, "/"+SkillFile) + "/assets"

	exists, err := files.Exists(assetDir)
	if err != nil {
		return nil, nil, fmt.Errorf("could not check %s: %s; check the permissions and run %s again", label, oneLine(err.Error()), retry)
	}
	if !exists {
		return nil, nil, nil
	}

	isDir, err := files.IsDir(assetDir)
	if err != nil {
		return nil, nil, fmt.Errorf("could not check %s: %s; check the permissions and run %s again", label, oneLine(err.Error()), retry)
	}
	if !isDir {
		return nil, []skill.Problem{{
			File:   label,
			Reason: "it is a file, not a folder: an asset is a file inside assets/",
		}}, nil
	}

	names, err := files.List(assetDir)
	if err != nil {
		return nil, nil, fmt.Errorf("could not read %s/: %s; check the permissions and run %s again", label, oneLine(err.Error()), retry)
	}

	assets := make([]string, 0, len(names))
	problems := []skill.Problem{}
	for _, name := range names {
		entry := filepath.Join(assetDir, name)
		entryIsDir, err := files.IsDir(entry)
		if err != nil {
			return nil, nil, fmt.Errorf("could not check %s/%s: %s; check the permissions and run %s again", label, name, oneLine(err.Error()), retry)
		}
		if entryIsDir {
			problems = append(problems, skill.Problem{
				File:   label + "/" + name,
				Reason: "it is a folder: an asset is a file, not a tree",
			})
			continue
		}
		assets = append(assets, "assets/"+name)
	}
	slices.Sort(assets)
	return assets, problems, nil
}

// noSkills is the starting point, not a failure: it says how to create the first
// skill.
func noSkills() Result {
	message := "there are no skills in " + Dir + "/ yet: create a user skill with odachi skill new <slug> --scope project|home"
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
