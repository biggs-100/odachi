package doctor

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/agent"
	domaindoctor "github.com/biggs-100/odachi/internal/domain/doctor"
	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/rules"
	"github.com/biggs-100/odachi/internal/domain/skill"
)

// root is the fake project every test works on. It is not a temporary
// directory: doctor must not need one, and the port has no way to write.
const root = "proyecto"

// homeRoot is the fake home: the global door lives outside the project, and
// doctor reads it through the same read-only port.
const homeRoot = "hogar"

// fakeFiles is the disk port as data. It has no write method on purpose: the
// port doctor receives cannot write, so "doctor never writes" is a property of
// the type and not a promise in a comment.
type fakeFiles struct {
	files map[string]string
	dirs  map[string]bool
	reads []string
	// errOn is the path whose Exists/Read fails, for the permission tests.
	errOn string
}

func newFakeFiles() *fakeFiles {
	return &fakeFiles{files: map[string]string{}, dirs: map[string]bool{}}
}

// at is the fake's spelling of a project-relative path.
func at(rel string) string {
	return filepath.Join(root, filepath.FromSlash(rel))
}

// atHome is the fake's spelling of a home-relative path.
func atHome(rel string) string {
	return filepath.Join(homeRoot, filepath.FromSlash(rel))
}

func (f *fakeFiles) put(rel, content string) *fakeFiles {
	f.files[at(rel)] = content
	// Every parent directory exists once a file is in it.
	for dir := filepath.Dir(at(rel)); len(dir) > len(root); dir = filepath.Dir(dir) {
		f.dirs[dir] = true
	}
	return f
}

func (f *fakeFiles) putHome(rel, content string) *fakeFiles {
	f.files[atHome(rel)] = content
	for dir := filepath.Dir(atHome(rel)); len(dir) > len(homeRoot); dir = filepath.Dir(dir) {
		f.dirs[dir] = true
	}
	return f
}

func (f *fakeFiles) mkdir(rel string) *fakeFiles {
	f.dirs[at(rel)] = true
	return f
}

func (f *fakeFiles) Exists(path string) (bool, error) {
	if path == f.errOn {
		return false, errors.New("permiso denegado")
	}
	if _, ok := f.files[path]; ok {
		return true, nil
	}
	return f.dirs[path], nil
}

func (f *fakeFiles) IsDir(path string) (bool, error) {
	if path == f.errOn {
		return false, errors.New("permiso denegado")
	}
	return f.dirs[path], nil
}

func (f *fakeFiles) Read(path string) ([]byte, error) {
	if path == f.errOn {
		return nil, errors.New("permiso denegado")
	}
	f.reads = append(f.reads, path)
	content, ok := f.files[path]
	if !ok {
		return nil, errors.New("no existe")
	}
	return []byte(content), nil
}

func (f *fakeFiles) List(dir string) ([]string, error) {
	if dir == f.errOn {
		return nil, errors.New("permiso denegado")
	}
	names := []string{}
	for path := range f.files {
		if filepath.Dir(path) == dir {
			names = append(names, filepath.Base(path))
		}
	}
	for path := range f.dirs {
		if filepath.Dir(path) == dir {
			names = append(names, filepath.Base(path))
		}
	}
	return names, nil
}

// snapshot is what the fake held, so a test can prove the run changed nothing.
func (f *fakeFiles) snapshot() map[string]string {
	out := make(map[string]string, len(f.files))
	for path, content := range f.files {
		out[path] = content
	}
	return out
}

// fakeLookup is the PATH port: the answer to "¿un shell puede correr odachi?".
type fakeLookup struct {
	resolved string
	err      error
	asked    []string
}

func (l *fakeLookup) LookPath(name string) (string, error) {
	l.asked = append(l.asked, name)
	if l.err != nil {
		return "", l.err
	}
	return l.resolved, nil
}

// fakeProfiles is the profile port: the same resolution the CLI does before
// every verb that reads a profile.
type fakeProfiles struct {
	name     string
	inferred bool
	err      error
}

func (p *fakeProfiles) Resolve(string) (string, bool, error) {
	return p.name, p.inferred, p.err
}

// briefComplete is a brief that satisfies brief.Missing().
const briefComplete = `# turnos

## Qué tiene que existir
- reservar turnos desde el teléfono

## Listo cuando
- reservo un turno y aparece en la agenda

## Decisiones tomadas
- el turno se guarda en sqlite
`

// briefIncomplete has the outcome but no acceptance criteria.
const briefIncomplete = `# turnos

## Qué tiene que existir
- reservar turnos desde el teléfono
`

// hookOf builds the hook `odachi init` would write for exe.
func hookOf(exe string) string {
	return "#!/bin/sh\n" + domaindoctor.HookMarker + " no se commitea sin prueba de este árbol\n" +
		"ODACHI='" + exe + "'\n[ -x \"$ODACHI\" ] || ODACHI=odachi\n\"$ODACHI\" check --quiet || exit 1\n"
}

// configured is a project that went through `git init`, `odachi init`,
// `odachi level --to <perfil>` and `odachi skills --write`: the state doctor must
// call clean.
func configured() *fakeFiles {
	files := newFakeFiles()
	files.mkdir(".git")
	files.put(".git/hooks/pre-commit", hookOf("odachi"))
	// The block `odachi init` writes, with the voice of the profile in the file:
	// check 11 compares the two, so a configured project has to carry both.
	files.put("AGENTS.md", blockOf(persona.Novato))
	body, err := rules.Body(persona.Novato)
	if err != nil {
		panic(err)
	}
	for _, file := range persona.Files(body) {
		files.put(file.Path, file.Content)
	}
	for _, file := range persona.GlobalFiles() {
		files.putHome(file.Path, file.Content)
	}
	files.put("brief.md", briefComplete)
	files.put(".odachi/profile.toml", "name = \"novato\"\n")
	files.put("skills/acceptance/SKILL.md", skill.Template("acceptance"))
	files.put(skill.RegistryFile, skill.Registry([]skill.Skill{
		skill.Parse("acceptance", "skills/acceptance/SKILL.md", skill.Template("acceptance")),
	}))
	return files
}

// blockOf is the rules block for one voice.
func blockOf(voice string) string {
	block, err := rules.Block(voice)
	if err != nil {
		panic(err)
	}
	return block
}

func cleanLookup() *fakeLookup { return &fakeLookup{resolved: "/home/ana/bin/odachi"} }

func cleanProfiles() *fakeProfiles {
	return &fakeProfiles{name: "novato"}
}

// run is one doctor over the given world.
func run(files *fakeFiles, lookup *fakeLookup, profiles *fakeProfiles, exe string) Result {
	return Run(Options{Dir: root, Home: homeRoot, Version: "0.5.0", Exe: exe}, files, lookup, profiles)
}

// checkOf returns one check by number, failing when the report does not have it.
func checkOf(t *testing.T, result Result, number int) domaindoctor.Check {
	t.Helper()
	for _, c := range result.Checks {
		if c.Number == number {
			return c
		}
	}
	t.Fatalf("the report has no check %d", number)
	return domaindoctor.Check{}
}

// TestFreshProjectIsClean is the promise of the verb: a project that was
// configured reports every check green and exits 0.
func TestFreshProjectIsClean(t *testing.T) {
	files := configured()
	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	if result.Code != CodeOK {
		t.Fatalf("a configured project must exit 0, got %d:\n%s", result.Code, strings.Join(result.Lines, "\n"))
	}
	if len(result.Checks) != domaindoctor.CheckCount {
		t.Fatalf("eleven checks, got %d", len(result.Checks))
	}
	for i, c := range result.Checks {
		if c.Number != i+1 {
			t.Fatalf("check %d is at position %d", c.Number, i)
		}
		if c.Status != domaindoctor.StatusOK {
			t.Fatalf("check %d is %v in a configured project: %s", c.Number, c.Status, domaindoctor.Line(c))
		}
		if c.Title == "" {
			t.Fatalf("check %d has no title", c.Number)
		}
	}
	if len(result.Lines) != domaindoctor.CheckCount+1 {
		t.Fatalf("the report is one line per check and a summary, got %d", len(result.Lines))
	}
	if result.Problems != 0 || result.Warnings != 0 {
		t.Fatalf("problems=%d warnings=%d", result.Problems, result.Warnings)
	}
}

// TestVersionIsTheFirstLine: the human asks "¿qué instalé?" first, so the
// answer comes first.
func TestVersionIsTheFirstLine(t *testing.T) {
	result := run(configured(), cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")
	if !strings.Contains(result.Lines[0], "0.5.0") {
		t.Fatalf("line 1 must carry the version, got %q", result.Lines[0])
	}
}

// TestEmptyProjectReportsTheMissingSetup: nothing configured is exactly what
// doctor is for, and each problem names its repair. A git repository without the
// odachi artifacts is still somebody's project, so it gets every check: the
// one-situation pre-flight is only for a folder where odachi was never started.
func TestEmptyProjectReportsTheMissingSetup(t *testing.T) {
	files := newFakeFiles().mkdir(".git")
	result := run(files, &fakeLookup{err: errors.New("not found")}, &fakeProfiles{name: "senior", inferred: true}, "")

	if result.Code != CodeProblems {
		t.Fatalf("an empty project must exit 1, got %d", result.Code)
	}
	if len(result.Checks) != domaindoctor.CheckCount {
		t.Fatalf("eleven checks even when everything is missing, got %d", len(result.Checks))
	}
	for _, number := range []int{domaindoctor.CheckOnPath, domaindoctor.CheckAgents, domaindoctor.CheckHook, domaindoctor.CheckBrief} {
		if got := checkOf(t, result, number); got.Status != domaindoctor.StatusProblem {
			t.Fatalf("check %d must be a problem in an empty project: %s", number, domaindoctor.Line(got))
		}
	}
	if got := checkOf(t, result, domaindoctor.CheckHookExe); got.Status != domaindoctor.StatusOK {
		t.Fatalf("there is no hook to check, so check 6 is not a second problem: %s", domaindoctor.Line(got))
	}
}

// TestEachProblemNamesItsCommand is the contract: a problem a human cannot act
// on is noise. Every case breaks exactly one thing and asserts that check.
func TestEachProblemNamesItsCommand(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*fakeFiles, *fakeLookup, *fakeProfiles)
		check  int
		fix    string
		detail string
	}{
		{
			name:   "odachi is not on the PATH",
			mutate: func(*fakeFiles, *fakeLookup, *fakeProfiles) {},
			check:  domaindoctor.CheckOnPath,
			fix:    "/home/ana/bin",
			detail: "el agente no puede ejecutar las reglas",
		},
		{
			name: "there is no .git",
			mutate: func(f *fakeFiles, _ *fakeLookup, _ *fakeProfiles) {
				delete(f.dirs, at(".git"))
			},
			check:  domaindoctor.CheckGit,
			fix:    "git init",
			detail: "no hay hook posible",
		},
		{
			name: "AGENTS.md is gone",
			mutate: func(f *fakeFiles, _ *fakeLookup, _ *fakeProfiles) {
				delete(f.files, at("AGENTS.md"))
			},
			check:  domaindoctor.CheckAgents,
			fix:    "odachi init",
			detail: "el agente no lee las reglas",
		},
		{
			name: "AGENTS.md has no block",
			mutate: func(f *fakeFiles, _ *fakeLookup, _ *fakeProfiles) {
				f.put("AGENTS.md", "# mi proyecto\n")
			},
			check:  domaindoctor.CheckAgents,
			fix:    "odachi init",
			detail: "no tiene el bloque",
		},
		{
			name: "the hook is gone",
			mutate: func(f *fakeFiles, _ *fakeLookup, _ *fakeProfiles) {
				delete(f.files, at(".git/hooks/pre-commit"))
			},
			check:  domaindoctor.CheckHook,
			fix:    "odachi init",
			detail: "no hay gate local",
		},
		{
			name: "the hook belongs to somebody else",
			mutate: func(f *fakeFiles, _ *fakeLookup, _ *fakeProfiles) {
				f.put(".git/hooks/pre-commit", "#!/bin/sh\nnpm test\n")
			},
			check:  domaindoctor.CheckHook,
			fix:    "odachi init",
			detail: "no es de odachi",
		},
		{
			name: "the hook names a binary that is gone",
			mutate: func(f *fakeFiles, _ *fakeLookup, _ *fakeProfiles) {
				f.put(".git/hooks/pre-commit", hookOf("/home/ana/bin/odachi-viejo"))
			},
			check:  domaindoctor.CheckHookExe,
			fix:    "odachi init",
			detail: "ya no existe",
		},
		{
			name: "the profile does not resolve",
			mutate: func(_ *fakeFiles, _ *fakeLookup, p *fakeProfiles) {
				p.err = errors.New(`el name "inventado" no existe`)
			},
			check:  domaindoctor.CheckProfile,
			fix:    "odachi doctor",
			detail: "no resuelve",
		},
		{
			name: "there is no brief",
			mutate: func(f *fakeFiles, _ *fakeLookup, _ *fakeProfiles) {
				delete(f.files, at("brief.md"))
			},
			check:  domaindoctor.CheckBrief,
			fix:    "odachi start",
			detail: "sin visión no hay nada que construir",
		},
		{
			name: "the brief is incomplete",
			mutate: func(f *fakeFiles, _ *fakeLookup, _ *fakeProfiles) {
				f.put("brief.md", briefIncomplete)
			},
			check:  domaindoctor.CheckBrief,
			fix:    "odachi start --force",
			detail: "listo cuando",
		},
		{
			name: "SKILLS.md drifted",
			mutate: func(f *fakeFiles, _ *fakeLookup, _ *fakeProfiles) {
				f.put(skill.RegistryFile, "<!-- generado por odachi skills --write; no lo edites a mano -->\n")
			},
			check:  domaindoctor.CheckSkills,
			fix:    "odachi skills --write",
			detail: "el catálogo se desvió",
		},
		{
			name: "the review store is not ignored",
			mutate: func(f *fakeFiles, _ *fakeLookup, _ *fakeProfiles) {
				f.put(".odachi/reviews/abc123.json", "{}")
			},
			check:  domaindoctor.CheckIgnored,
			fix:    "odachi init",
			detail: "cada review invalidaría la prueba",
		},
		{
			name: "proof/ is not ignored",
			mutate: func(f *fakeFiles, _ *fakeLookup, _ *fakeProfiles) {
				f.put("proof/0001-turnos.txt", "# tree abc\n# command true\n")
			},
			check:  domaindoctor.CheckIgnored,
			fix:    "odachi init",
			detail: "cada prueba invalidaría la siguiente",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			files, lookup, profiles := configured(), cleanLookup(), cleanProfiles()
			if c.name == "odachi is not on the PATH" {
				lookup = &fakeLookup{err: errors.New("not found")}
			}
			c.mutate(files, lookup, profiles)

			result := run(files, lookup, profiles, "/home/ana/bin/odachi")
			if result.Code != CodeProblems {
				t.Fatalf("must exit 1:\n%s", strings.Join(result.Lines, "\n"))
			}
			got := checkOf(t, result, c.check)
			if got.Status != domaindoctor.StatusProblem {
				t.Fatalf("check %d must be a problem: %s", c.check, domaindoctor.Line(got))
			}
			if !strings.Contains(got.Detail, c.detail) {
				t.Fatalf("check %d detail = %q, want it to name %q", c.check, got.Detail, c.detail)
			}
			if !strings.Contains(got.Fix, c.fix) {
				t.Fatalf("check %d fix = %q, want it to name %q", c.check, got.Fix, c.fix)
			}
		})
	}
}

// TestEmptyProtectedDirectoryIsAWarning: nothing is broken yet, and a warning
// never fails the run. It is still worth saying, because the next write is the
// one that would invalidate a proof.
func TestEmptyProtectedDirectoryIsAWarning(t *testing.T) {
	files := configured().mkdir("proof")
	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	if result.Code != CodeOK {
		t.Fatalf("a warning must not fail the run, got %d:\n%s", result.Code, strings.Join(result.Lines, "\n"))
	}
	got := checkOf(t, result, domaindoctor.CheckIgnored)
	if got.Status != domaindoctor.StatusWarning {
		t.Fatalf("an empty unprotected directory is a warning: %s", domaindoctor.Line(got))
	}
	if result.Warnings != 1 {
		t.Fatalf("warnings = %d, want 1", result.Warnings)
	}
}

// TestProtectedDirectoriesAreClean: the two ignore files are the difference
// between a usable review loop and one that invalidates the proof every time.
func TestProtectedDirectoriesAreClean(t *testing.T) {
	files := configured().
		put(".odachi/reviews/.gitignore", "*\n")
	files.put("proof/.gitignore", "*\n!.gitignore\n")

	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")
	if result.Code != CodeOK {
		t.Fatalf("must be clean:\n%s", strings.Join(result.Lines, "\n"))
	}
	got := checkOf(t, result, domaindoctor.CheckIgnored)
	if !strings.Contains(got.Detail, ".odachi/reviews") || !strings.Contains(got.Detail, "proof") {
		t.Fatalf("the line must say what it found protected: %s", domaindoctor.Line(got))
	}
}

// TestNoSkillsIsClean: a project without skills/ has nothing to drift.
func TestNoSkillsIsClean(t *testing.T) {
	files := configured()
	delete(files.files, at(skill.RegistryFile))
	delete(files.files, at("skills/acceptance/SKILL.md"))

	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")
	if result.Code != CodeOK {
		t.Fatalf("a project without skills must be clean:\n%s", strings.Join(result.Lines, "\n"))
	}
	got := checkOf(t, result, domaindoctor.CheckSkills)
	if !strings.Contains(got.Detail, "skills/: ninguna") {
		t.Fatalf("check 9 must say the project has no skills of its own: %s", domaindoctor.Line(got))
	}
}

// TestCheckNineAnswersBothSkillFacts is the debt this check closes: a project
// that just ran `odachi init` has no skills/ registry and nine skills installed
// in .agents/skills. Answering only the first read "ninguna: no hay skills/" to a
// person who had just installed them, and the two facts are two answers.
func TestCheckNineAnswersBothSkillFacts(t *testing.T) {
	files := configured().
		put(".agents/skills/drive/SKILL.md", "x").
		put(".agents/skills/acceptance/SKILL.md", skill.Template("acceptance")).
		// A file is not a skill: only a directory with a SKILL.md in it counts.
		put(".agents/skills/README.md", "no es una skill\n")

	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")
	got := checkOf(t, result, domaindoctor.CheckSkills)

	if got.Status != domaindoctor.StatusOK {
		t.Fatalf("installed skills are not a problem: %s", domaindoctor.Line(got))
	}
	for _, want := range []string{
		"skills/: 1 skill, " + skill.RegistryFile + " al día",
		".agents/skills: 2 instaladas",
		"coinciden con skills/",
	} {
		if !strings.Contains(got.Detail, want) {
			t.Fatalf("check 9 must say %q, got %q", want, got.Detail)
		}
	}
	if strings.Contains(got.Detail, "README") {
		t.Fatalf("a file is not a skill and must not count: %q", got.Detail)
	}
}

// TestInstalledSkillThatDoesNotMatchTheSourceIsAWarning: the agent reads the
// installed copy, so a copy that differs from skills/ is the drift the two
// directories create. It is a warning with the command that refreshes it.
func TestInstalledSkillThatDoesNotMatchTheSourceIsAWarning(t *testing.T) {
	files := configured().put(".agents/skills/acceptance/SKILL.md", "viejo\n")

	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	got := checkOf(t, result, domaindoctor.CheckSkills)
	if got.Status != domaindoctor.StatusWarning {
		t.Fatalf("a drifted installed skill must be a warning: %s", domaindoctor.Line(got))
	}
	for _, want := range []string{"no coinciden con skills/", "acceptance", "odachi init"} {
		if !strings.Contains(got.Detail, want) {
			t.Fatalf("check 9 must say %q: %q", want, got.Detail)
		}
	}
}

// TestCheckNineSaysWhenNothingIsInstalled: the installed half of the line has an
// honest zero too, and it never turns into a problem: nothing is broken yet.
func TestCheckNineSaysWhenNothingIsInstalled(t *testing.T) {
	files := configured()
	delete(files.files, at(skill.RegistryFile))
	delete(files.files, at("skills/acceptance/SKILL.md"))

	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")
	got := checkOf(t, result, domaindoctor.CheckSkills)

	if got.Status != domaindoctor.StatusOK {
		t.Fatalf("nothing installed is not a problem: %s", domaindoctor.Line(got))
	}
	if !strings.Contains(got.Detail, ".agents/skills: ninguna") {
		t.Fatalf("check 9 must answer the installed half too: %q", got.Detail)
	}
}

// TestAVoiceThatDoesNotMatchTheProfileIsAProblem: the block says one thing and
// the file the binary reads says another, so the agent talks in a register the
// tool did not promise. The line names both and the one command that settles it.
func TestAVoiceThatDoesNotMatchTheProfileIsAProblem(t *testing.T) {
	files := configured().put("AGENTS.md", blockOf(persona.Senior))

	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	if result.Code != CodeProblems {
		t.Fatalf("a mismatched voice must exit 1: %s", strings.Join(result.Lines, " / "))
	}
	got := checkOf(t, result, domaindoctor.CheckVoice)
	if got.Status != domaindoctor.StatusProblem {
		t.Fatalf("check 11 must be a problem: %s", domaindoctor.Line(got))
	}
	for _, want := range []string{"senior", ".odachi/profile.toml dice novato", "el agente habla distinto"} {
		if !strings.Contains(got.Detail, want) {
			t.Fatalf("the detail must say %q, got %q", want, got.Detail)
		}
	}
	if !strings.Contains(got.Fix, "odachi level --to novato") {
		t.Fatalf("the fix must name odachi level --to novato, got %q", got.Fix)
	}
}

// TestABlockWithoutTheVoiceSectionAndAProfileIsAProblem: a project initialised
// before the voice moved into the block, with a profile on disk, has a voice that
// is nowhere. `init` is the command that writes the section.
func TestABlockWithoutTheVoiceSectionAndAProfileIsAProblem(t *testing.T) {
	old := "<!-- odachi:begin -->\n# Reglas del proyecto (odachi)\n\nreglas\n<!-- odachi:end -->\n"
	files := configured().put("AGENTS.md", old)

	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")
	got := checkOf(t, result, domaindoctor.CheckVoice)

	if got.Status != domaindoctor.StatusProblem {
		t.Fatalf("a block with no voice and a profile is a problem: %s", domaindoctor.Line(got))
	}
	for _, want := range []string{rules.Section, ".odachi/profile.toml dice novato"} {
		if !strings.Contains(got.Detail, want) {
			t.Fatalf("the detail must say %q, got %q", want, got.Detail)
		}
	}
	if !strings.Contains(got.Fix, "odachi init") {
		t.Fatalf("the fix must name odachi init, got %q", got.Fix)
	}
}

// TestNoProfileFileIsNotAProblem: with no .odachi/profile.toml there is no claim
// to compare against, and the block carries the voice init wrote. Failing here
// would tell a person who just ran init that their green project is broken.
func TestNoProfileFileIsNotAProblem(t *testing.T) {
	files := configured()
	delete(files.files, at(".odachi/profile.toml"))
	profiles := &fakeProfiles{name: "novato", inferred: true}

	result := run(files, cleanLookup(), profiles, "/home/ana/bin/odachi")

	got := checkOf(t, result, domaindoctor.CheckVoice)
	if got.Status != domaindoctor.StatusOK {
		t.Fatalf("check 11 must be ok with no profile file: %s", domaindoctor.Line(got))
	}
	for _, want := range []string{"no hay perfil que comparar", "novato"} {
		if !strings.Contains(got.Detail, want) {
			t.Fatalf("the line must say %q, got %q", want, got.Detail)
		}
	}
}

// TestAVoiceThatMatchesTheProfileIsClean: the useful half of a green line names
// the voice and the file it agrees with.
func TestAVoiceThatMatchesTheProfileIsClean(t *testing.T) {
	result := run(configured(), cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")
	got := checkOf(t, result, domaindoctor.CheckVoice)

	if got.Status != domaindoctor.StatusOK {
		t.Fatalf("check 11 must be ok in a configured project: %s", domaindoctor.Line(got))
	}
	for _, want := range []string{"novato", ".odachi/profile.toml"} {
		if !strings.Contains(got.Detail, want) {
			t.Fatalf("the line must name %q, got %q", want, got.Detail)
		}
	}
}

// TestLeftoverPersonaFilesAreNotAProblem: v0.10.0 wrote one persona file per
// client, and this version does not read them. They are leftovers, not a defect:
// the line says so, names the folders, and the run still exits 0.
func TestLeftoverPersonaFilesAreNotAProblem(t *testing.T) {
	files := configured().put(".opencode/agents/odachi-senior.md", "viejo")

	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	if result.Code != CodeOK {
		t.Fatalf("leftover files are not a problem: %s", strings.Join(result.Lines, " / "))
	}
	got := checkOf(t, result, domaindoctor.CheckVoice)
	if got.Status != domaindoctor.StatusOK {
		t.Fatalf("check 11 must be ok with leftovers: %s", domaindoctor.Line(got))
	}
	for _, want := range []string{"v0.10.0", ".opencode/agents", ".claude/output-styles", "podés borrarlas"} {
		if !strings.Contains(got.Detail, want) {
			t.Fatalf("the line must say %q, got %q", want, got.Detail)
		}
	}
}

// TestAStaleNativeAgentIsSaidAndNotAFailure: the mirror is generated, so a hand
// edit is lost work — but the block still works. The line names the file and the
// command that regenerates it, and the run does not fail.
func TestAStaleNativeAgentIsSaidAndNotAFailure(t *testing.T) {
	files := configured().put(".opencode/agents/odachi.md", "editado a mano\n")

	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	if result.Code != CodeOK {
		t.Fatalf("un espejo editado no rompe el proyecto: %s", strings.Join(result.Lines, " / "))
	}
	got := checkOf(t, result, domaindoctor.CheckVoice)
	for _, want := range []string{"no coincide", ".opencode/agents/odachi.md", "odachi init"} {
		if !strings.Contains(got.Detail, want) {
			t.Fatalf("la línea tiene que decir %q: %q", want, got.Detail)
		}
	}
}

// TestAMissingNativeAgentIsSaidAndNotAFailure: a project that never ran the new
// init still has the block; the selector is what is missing. One note, exit 0.
func TestAMissingNativeAgentIsSaidAndNotAFailure(t *testing.T) {
	files := configured()
	delete(files.files, at(".opencode/agents/odachi.md"))

	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	if result.Code != CodeOK {
		t.Fatalf("un agente ausente no rompe el proyecto: %s", strings.Join(result.Lines, " / "))
	}
	got := checkOf(t, result, domaindoctor.CheckVoice)
	for _, want := range []string{"no está", ".opencode/agents/odachi.md", "odachi init"} {
		if !strings.Contains(got.Detail, want) {
			t.Fatalf("la línea tiene que decir %q: %q", want, got.Detail)
		}
	}
}

// TestAGlobalDoorThatIsMissingIsSaid: the project can be perfect and the human
// still not find odachi in another folder. Check 11 names the missing door and
// the command that installs it, without failing the project.
func TestAGlobalDoorThatIsMissingIsSaid(t *testing.T) {
	files := configured()
	delete(files.files, atHome(".config/opencode/agents/odachi.md"))

	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	if result.Code != CodeOK {
		t.Fatalf("la puerta ausente no rompe el proyecto: %s", strings.Join(result.Lines, " / "))
	}
	got := checkOf(t, result, domaindoctor.CheckVoice)
	for _, want := range []string{"puerta global", "no está", ".config/opencode/agents/odachi.md", "odachi init"} {
		if !strings.Contains(got.Detail, want) {
			t.Fatalf("la línea tiene que decir %q: %q", want, got.Detail)
		}
	}
}

// TestAGlobalDoorThatIsStaleIsSaid: a hand-edited door is the same class as a
// hand-edited mirror: generated content that no longer matches this binary.
func TestAGlobalDoorThatIsStaleIsSaid(t *testing.T) {
	files := configured().putHome(".config/opencode/agents/odachi.md", "viejo\n")

	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	if result.Code != CodeOK {
		t.Fatalf("una puerta editada no rompe el proyecto: %s", strings.Join(result.Lines, " / "))
	}
	got := checkOf(t, result, domaindoctor.CheckVoice)
	for _, want := range []string{"puerta global", "no coincide", ".config/opencode/agents/odachi.md", "odachi init"} {
		if !strings.Contains(got.Detail, want) {
			t.Fatalf("la línea tiene que decir %q: %q", want, got.Detail)
		}
	}
}

// TestAGlobalDoorThatMatchesIsClean: a configured home is information, not
// noise: the line says the door is up to date.
func TestAGlobalDoorThatMatchesIsClean(t *testing.T) {
	result := run(configured(), cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	got := checkOf(t, result, domaindoctor.CheckVoice)
	if got.Status != domaindoctor.StatusOK {
		t.Fatalf("check 11 must be ok with a clean home: %s", domaindoctor.Line(got))
	}
	if !strings.Contains(got.Detail, "puerta global está al día") {
		t.Fatalf("la línea no dice que la puerta está al día: %q", got.Detail)
	}
}

// TestHookNamingOdachiByNameUsesThePath is the install that works twice and
// fails once: the hook says `odachi`, and a shell cannot find it.
func TestHookNamingOdachiByNameUsesThePath(t *testing.T) {
	files := configured().put(".git/hooks/pre-commit", hookOf("odachi"))
	result := run(files, &fakeLookup{err: errors.New("not found")}, cleanProfiles(), "/home/ana/bin/odachi")

	got := checkOf(t, result, domaindoctor.CheckHookExe)
	if got.Status != domaindoctor.StatusProblem {
		t.Fatalf("a hook that cannot find its binary fails on every commit: %s", domaindoctor.Line(got))
	}
	if !strings.Contains(got.Fix, "/home/ana/bin") {
		t.Fatalf("the fix must name the folder of the running binary, got %q", got.Fix)
	}
}

// TestUnreadableFileIsAProblemNotACrash: doctor reads a project it did not
// write, so a permission error is a line and never a panic.
func TestUnreadableFileIsAProblemNotACrash(t *testing.T) {
	files := configured()
	files.errOn = at("brief.md")
	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	if result.Code != CodeProblems {
		t.Fatalf("must exit 1:\n%s", strings.Join(result.Lines, "\n"))
	}
	got := checkOf(t, result, domaindoctor.CheckBrief)
	if got.Status != domaindoctor.StatusProblem || !strings.Contains(got.Detail, "permiso denegado") {
		t.Fatalf("the permission error must be the line: %s", domaindoctor.Line(got))
	}
}

// TestDoctorNeverWrites: the port has no write method, and the world it read is
// byte for byte what it was.
func TestDoctorNeverWrites(t *testing.T) {
	files := configured()
	before := files.snapshot()

	run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	after := files.snapshot()
	if len(after) != len(before) {
		t.Fatalf("doctor changed the number of files: %d -> %d", len(before), len(after))
	}
	for path, content := range before {
		if after[path] != content {
			t.Fatalf("doctor changed %s", path)
		}
	}
	if len(files.reads) == 0 {
		t.Fatalf("doctor must actually read the project")
	}
}

// TestInstalledSkillsDirIsTheStandardRow: the count doctor reports is the count
// `odachi init` wrote, so the directory has to be the one the agent table owns.
// Two spellings of the same folder is how a report starts lying.
func TestInstalledSkillsDirIsTheStandardRow(t *testing.T) {
	row, err := agent.Resolve(agent.Standard)
	if err != nil {
		t.Fatalf("agent.Resolve(%q): %v", agent.Standard, err)
	}
	if installedSkillsDir != row.Dest {
		t.Fatalf("doctor cuenta %s y init instala en %s", installedSkillsDir, row.Dest)
	}
}
