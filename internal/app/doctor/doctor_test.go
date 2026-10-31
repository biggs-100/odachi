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
		return false, errors.New("permission denied")
	}
	if _, ok := f.files[path]; ok {
		return true, nil
	}
	return f.dirs[path], nil
}

func (f *fakeFiles) IsDir(path string) (bool, error) {
	if path == f.errOn {
		return false, errors.New("permission denied")
	}
	return f.dirs[path], nil
}

func (f *fakeFiles) Read(path string) ([]byte, error) {
	if path == f.errOn {
		return nil, errors.New("permission denied")
	}
	f.reads = append(f.reads, path)
	content, ok := f.files[path]
	if !ok {
		return nil, errors.New("does not exist")
	}
	return []byte(content), nil
}

func (f *fakeFiles) List(dir string) ([]string, error) {
	if dir == f.errOn {
		return nil, errors.New("permission denied")
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

// fakeLookup is the PATH port: the answer to "can a shell run odachi?".
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
	return "#!/bin/sh\n" + domaindoctor.HookMarker + " no commit without a proof for this tree\n" +
		"ODACHI='" + exe + "'\n[ -x \"$ODACHI\" ] || ODACHI=odachi\n\"$ODACHI\" check --quiet || exit 1\n"
}

// configured is a project that went through `git init`, `odachi init`,
// `odachi level --to <profile>` and `odachi skills --write`: the state doctor must
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
	for _, file := range persona.SubagentFiles() {
		files.put(file.Path, file.Content)
	}
	for _, file := range persona.GlobalFiles() {
		files.putHome(file.Path, file.Content)
	}
	// The default init leaves bootstrap in the home; it is the one skill every
	// project gets there without a project copy.
	files.putHome(".agents/skills/bootstrap/SKILL.md", skill.Template("bootstrap"))
	files.put("brief.md", briefComplete)
	files.put("odd/profile.toml", "name = \"novato\"\n")
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

// fakeAgentSelection is the selection port: the choice the project stored, or
// the error a malformed file gives. It only reads.
type fakeAgentSelection struct {
	stored  string
	project string
	err     error
}

func (s fakeAgentSelection) Written(string) (persona.Store, error) {
	return persona.Store{Agents: s.stored, ProjectAgents: s.project}, s.err
}

// run is one doctor over the given world.
func run(files *fakeFiles, lookup *fakeLookup, profiles *fakeProfiles, exe string) Result {
	return runSelected(files, lookup, profiles, exe, fakeAgentSelection{})
}

// runSelected is the same run with a stored client selection.
func runSelected(files *fakeFiles, lookup *fakeLookup, profiles *fakeProfiles, exe string, agents fakeAgentSelection) Result {
	return Run(Options{Dir: root, Home: homeRoot, Version: "0.5.0", Exe: exe, Agents: agents}, files, lookup, profiles)
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

// TestVersionIsTheFirstLine: the human asks "what did I install?" first, so the
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
			detail: "the agent cannot run the rules",
		},
		{
			name: "there is no .git",
			mutate: func(f *fakeFiles, _ *fakeLookup, _ *fakeProfiles) {
				delete(f.dirs, at(".git"))
			},
			check:  domaindoctor.CheckGit,
			fix:    "git init",
			detail: "no hook is possible",
		},
		{
			name: "AGENTS.md is gone",
			mutate: func(f *fakeFiles, _ *fakeLookup, _ *fakeProfiles) {
				delete(f.files, at("AGENTS.md"))
			},
			check:  domaindoctor.CheckAgents,
			fix:    "odachi init",
			detail: "the agent does not read the rules",
		},
		{
			name: "AGENTS.md has no block",
			mutate: func(f *fakeFiles, _ *fakeLookup, _ *fakeProfiles) {
				f.put("AGENTS.md", "# my project\n")
			},
			check:  domaindoctor.CheckAgents,
			fix:    "odachi init",
			detail: "does not have the odachi block",
		},
		{
			name: "the hook is gone",
			mutate: func(f *fakeFiles, _ *fakeLookup, _ *fakeProfiles) {
				delete(f.files, at(".git/hooks/pre-commit"))
			},
			check:  domaindoctor.CheckHook,
			fix:    "odachi init",
			detail: "there is no local gate",
		},
		{
			name: "the hook belongs to somebody else",
			mutate: func(f *fakeFiles, _ *fakeLookup, _ *fakeProfiles) {
				f.put(".git/hooks/pre-commit", "#!/bin/sh\nnpm test\n")
			},
			check:  domaindoctor.CheckHook,
			fix:    "odachi init",
			detail: "is not odachi's",
		},
		{
			name: "the hook names a binary that is gone",
			mutate: func(f *fakeFiles, _ *fakeLookup, _ *fakeProfiles) {
				f.put(".git/hooks/pre-commit", hookOf("/home/ana/bin/odachi-viejo"))
			},
			check:  domaindoctor.CheckHookExe,
			fix:    "odachi init",
			detail: "no longer exists",
		},
		{
			name: "the profile does not resolve",
			mutate: func(_ *fakeFiles, _ *fakeLookup, p *fakeProfiles) {
				p.err = errors.New(`el name "inventado" no existe`)
			},
			check:  domaindoctor.CheckProfile,
			fix:    "odachi doctor",
			detail: "does not resolve",
		},
		{
			name: "there is no brief",
			mutate: func(f *fakeFiles, _ *fakeLookup, _ *fakeProfiles) {
				delete(f.files, at("brief.md"))
			},
			check:  domaindoctor.CheckBrief,
			fix:    "odachi start",
			detail: "without a vision there is nothing to build",
		},
		{
			name: "the brief is incomplete",
			mutate: func(f *fakeFiles, _ *fakeLookup, _ *fakeProfiles) {
				f.put("brief.md", briefIncomplete)
			},
			check:  domaindoctor.CheckBrief,
			fix:    "odachi start --force",
			detail: "done when",
		},
		{
			name: "SKILLS.md drifted",
			mutate: func(f *fakeFiles, _ *fakeLookup, _ *fakeProfiles) {
				f.put(skill.RegistryFile, "<!-- generado por odachi skills --write; no lo edites a mano -->\n")
			},
			check:  domaindoctor.CheckSkills,
			fix:    "odachi skills --write",
			detail: "the catalogue drifted",
		},
		{
			name: "proof/ is not ignored",
			mutate: func(f *fakeFiles, _ *fakeLookup, _ *fakeProfiles) {
				f.put("proof/0001-turnos.txt", "# tree abc\n# command true\n")
			},
			check:  domaindoctor.CheckIgnored,
			fix:    "odachi init",
			detail: "every proof would invalidate the next one",
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
		put("proof/.gitignore", "*\n!.gitignore\n")

	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")
	if result.Code != CodeOK {
		t.Fatalf("must be clean:\n%s", strings.Join(result.Lines, "\n"))
	}
	got := checkOf(t, result, domaindoctor.CheckIgnored)
	if !strings.Contains(got.Detail, "proof") {
		t.Fatalf("the line must say what it found protected: %s", domaindoctor.Line(got))
	}
}

// TestAnUnreadableForeignDocumentIsBrokenAndNotProse: a document the doctor cannot
// open is not a document with no plan shape, and saying "nothing is ignored" about a
// file it never read is the false claim this feature exists to remove.
func TestAnUnreadableForeignDocumentIsBrokenAndNotProse(t *testing.T) {
	files := configured().
		put("proof/.gitignore", "*\n!.gitignore\n")
	files.put("odd/tasks/notas.md", "# Notas\n")
	files.errOn = at("odd/tasks/notas.md")

	got := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	check := checkOf(t, got, domaindoctor.CheckTaskDocs)
	if check.Status != domaindoctor.StatusProblem {
		t.Fatalf("a document that cannot be read is not prose: %s", domaindoctor.Line(check))
	}
	if !strings.Contains(check.Detail, "odd/tasks/notas.md") {
		t.Errorf("the refusal does not name the file: %s", domaindoctor.Line(check))
	}
	if !strings.Contains(check.Fix, "check the permissions of odd/tasks") {
		t.Errorf("the refusal does not say what to fix: %s", domaindoctor.Line(check))
	}
}

// TestAForeignProseDocumentIsNamedAndNotAWarning: a note in the folder of the plan
// is in the wrong folder, and odachi says it does not read it without calling the
// install unfinished -nothing was lost by skipping a document with no plan shape.
func TestAForeignProseDocumentIsNamedAndNotAWarning(t *testing.T) {
	files := configured().
		put("proof/.gitignore", "*\n!.gitignore\n")
	files.put("odd/tasks/notas.md", "# Notas\n\nlo que aprendimos esta semana\n")
	files.put("odd/tasks/0001-a-real-feature.md", "# a-real-feature\n")

	got := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")
	if got.Code != CodeOK {
		t.Fatalf("a foreign document must not fail the run, got %d:\n%s", got.Code, strings.Join(got.Lines, "\n"))
	}
	check := checkOf(t, got, domaindoctor.CheckTaskDocs)
	if check.Status != domaindoctor.StatusOK {
		t.Fatalf("prose is named, not warned about: %s", domaindoctor.Line(check))
	}
	if !strings.Contains(check.Detail, "notas.md") || !strings.Contains(check.Detail, "no plan shape") {
		t.Errorf("the line does not say what it skipped and why: %s", domaindoctor.Line(check))
	}
	if strings.Contains(check.Detail, "a-real-feature") {
		t.Errorf("the line counts a document odachi reads: %s", domaindoctor.Line(check))
	}
}

// TestAForeignPlanIsAWarning: a document that carries a plan's shape and a name
// odachi cannot read is a feature that would be silently ignored, and that is the
// one case worth a warning. A note beside it is still named: a mixed folder is the
// ordinary state, and the line must not hide what it skipped.
func TestAForeignPlanIsAWarning(t *testing.T) {
	files := configured().
		put("proof/.gitignore", "*\n!.gitignore\n")
	files.put("odd/tasks/qonta-einvoicing.md", "# qonta-einvoicing\nBase: abc1234\n\n- [ ] T1 · emitir facturas\n")
	files.put("odd/tasks/notas.md", "# Notas\n\nlo que aprendimos esta semana\n")
	files.put("odd/tasks/0001-a-real-feature.md", "# a-real-feature\n")

	got := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")
	if got.Code != CodeOK {
		t.Fatalf("a foreign plan must not fail the run, got %d:\n%s", got.Code, strings.Join(got.Lines, "\n"))
	}
	check := checkOf(t, got, domaindoctor.CheckTaskDocs)
	if check.Status != domaindoctor.StatusWarning {
		t.Fatalf("a plan odachi cannot read is a warning: %s", domaindoctor.Line(check))
	}
	if !strings.Contains(check.Detail, "qonta-einvoicing.md") || !strings.Contains(check.Detail, "1 document") {
		t.Errorf("the line does not name the plan: %s", domaindoctor.Line(check))
	}
	if !strings.Contains(check.Detail, "notas.md") {
		t.Errorf("the warning hides the note beside the plan: %s", domaindoctor.Line(check))
	}
	if strings.Contains(check.Detail, "a-real-feature") {
		t.Errorf("the line counts a document odachi reads: %s", domaindoctor.Line(check))
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
	if !strings.Contains(got.Detail, "skills/: none") {
		t.Fatalf("check 9 must say the project has no skills of its own: %s", domaindoctor.Line(got))
	}
}

// TestCheckNineAnswersBothSkillFacts is the debt this check closes: a project
// that just ran `odachi init` has no skills/ registry and nine skills installed
// in .agents/skills. Answering only the first read "none: there is no skills/" to a
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
		"skills/: 1 skill, " + skill.RegistryFile + " up to date",
		".agents/skills: 2 installed",
		"compared: 1 match in skills/",
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
	for _, want := range []string{"do not match skills/", "acceptance", "odachi init"} {
		if !strings.Contains(got.Detail, want) {
			t.Fatalf("check 9 must say %q: %q", want, got.Detail)
		}
	}
}

// TestInstalledHomeSkillThatDoesNotMatchTheSourceIsAWarning: --skills home puts
// the payload in the home, so the same drift is possible there. Check 9 names
// it and the command that refreshes it.
func TestInstalledHomeSkillThatDoesNotMatchTheSourceIsAWarning(t *testing.T) {
	files := configured().putHome(".agents/skills/acceptance/SKILL.md", "viejo\n")

	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	got := checkOf(t, result, domaindoctor.CheckSkills)
	if got.Status != domaindoctor.StatusWarning {
		t.Fatalf("a drifted home skill must be a warning: %s", domaindoctor.Line(got))
	}
	for _, want := range []string{"~/.agents/skills", "do not match skills/", "acceptance", "odachi init --skills home"} {
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
	if !strings.Contains(got.Detail, ".agents/skills: none") {
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
	for _, want := range []string{"senior", "odd/profile.toml says novato", "the agent talks differently"} {
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
	for _, want := range []string{rules.Section, "odd/profile.toml says novato"} {
		if !strings.Contains(got.Detail, want) {
			t.Fatalf("the detail must say %q, got %q", want, got.Detail)
		}
	}
	if !strings.Contains(got.Fix, "odachi init") {
		t.Fatalf("the fix must name odachi init, got %q", got.Fix)
	}
}

// TestNoProfileFileIsNotAProblem: with no odd/profile.toml there is no claim
// to compare against, and the block carries the voice init wrote. Failing here
// would tell a person who just ran init that their green project is broken.
func TestNoProfileFileIsNotAProblem(t *testing.T) {
	files := configured()
	delete(files.files, at("odd/profile.toml"))
	profiles := &fakeProfiles{name: "novato", inferred: true}

	result := run(files, cleanLookup(), profiles, "/home/ana/bin/odachi")

	got := checkOf(t, result, domaindoctor.CheckVoice)
	if got.Status != domaindoctor.StatusOK {
		t.Fatalf("check 11 must be ok with no profile file: %s", domaindoctor.Line(got))
	}
	for _, want := range []string{"there is no profile to compare", "novato"} {
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
	for _, want := range []string{"novato", "odd/profile.toml"} {
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
	for _, want := range []string{"v0.10.0", ".opencode/agents", ".claude/output-styles", "you can delete them"} {
		if !strings.Contains(got.Detail, want) {
			t.Fatalf("the line must say %q, got %q", want, got.Detail)
		}
	}
}

// TestAMissingSubagentIsSaid: a project can have the primary agent and lose the
// worker or the verifier. The parent cannot delegate to a file that is not
// there; check 11 names it and the command that installs it.
func TestAMissingSubagentIsSaid(t *testing.T) {
	files := configured()
	delete(files.files, at(".opencode/agents/odachi-worker.md"))

	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	if result.Code != CodeOK {
		t.Fatalf("a missing subagent does not break the project: %s", strings.Join(result.Lines, " / "))
	}
	got := checkOf(t, result, domaindoctor.CheckVoice)
	for _, want := range []string{"are missing in", "odachi-worker", "odachi init"} {
		if !strings.Contains(got.Detail, want) {
			t.Fatalf("the line must say %q: %q", want, got.Detail)
		}
	}
}

// TestAStaleSubagentIsSaid: the worker and the verifier are generated too; a
// hand edit is a drift, not a new contract.
func TestAStaleSubagentIsSaid(t *testing.T) {
	files := configured().put(".opencode/agents/odachi-verify.md", "viejo\n")

	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	if result.Code != CodeOK {
		t.Fatalf("an edited subagent does not break the project: %s", strings.Join(result.Lines, " / "))
	}
	got := checkOf(t, result, domaindoctor.CheckVoice)
	for _, want := range []string{"do not match what was generated", "odachi-verify", "odachi init"} {
		if !strings.Contains(got.Detail, want) {
			t.Fatalf("the line must say %q: %q", want, got.Detail)
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
		t.Fatalf("an edited mirror does not break the project: %s", strings.Join(result.Lines, " / "))
	}
	got := checkOf(t, result, domaindoctor.CheckVoice)
	for _, want := range []string{"do not match what was generated", ".opencode/agents/odachi.md", "odachi init"} {
		if !strings.Contains(got.Detail, want) {
			t.Fatalf("the line must say %q: %q", want, got.Detail)
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
		t.Fatalf("a missing agent does not break the project: %s", strings.Join(result.Lines, " / "))
	}
	got := checkOf(t, result, domaindoctor.CheckVoice)
	for _, want := range []string{"are missing in", ".opencode/agents/odachi.md", "odachi init"} {
		if !strings.Contains(got.Detail, want) {
			t.Fatalf("the line must say %q: %q", want, got.Detail)
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
		t.Fatalf("a missing door does not break the project: %s", strings.Join(result.Lines, " / "))
	}
	got := checkOf(t, result, domaindoctor.CheckVoice)
	for _, want := range []string{"global door", "is not in", ".config/opencode/agents/odachi.md", "odachi init --global"} {
		if !strings.Contains(got.Detail, want) {
			t.Fatalf("the line must say %q: %q", want, got.Detail)
		}
	}
}

// TestAGlobalDoorThatIsStaleIsSaid: a hand-edited door is the same class as a
// hand-edited mirror: generated content that no longer matches this binary.
func TestAGlobalDoorThatIsStaleIsSaid(t *testing.T) {
	files := configured().putHome(".config/opencode/agents/odachi.md", "viejo\n")

	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	if result.Code != CodeOK {
		t.Fatalf("an edited door does not break the project: %s", strings.Join(result.Lines, " / "))
	}
	got := checkOf(t, result, domaindoctor.CheckVoice)
	for _, want := range []string{"global door", "does not match this binary", ".config/opencode/agents/odachi.md", "odachi init"} {
		if !strings.Contains(got.Detail, want) {
			t.Fatalf("the line must say %q: %q", want, got.Detail)
		}
	}
}

// TestDoctorOnlyChecksTheChosenClient: the project chose OpenCode, so Claude is
// neither missing nor stale. The line says which choice it obeyed.
func TestDoctorOnlyChecksTheChosenClient(t *testing.T) {
	files := configured()
	body, err := rules.Body(persona.Novato)
	if err != nil {
		t.Fatalf("rules.Body(novato): %v", err)
	}
	claude, _ := persona.FilesFor(persona.SelectionClaude, body)
	delete(files.files, at(claude.Path))
	for _, file := range persona.SubagentFilesFor(persona.SelectionClaude) {
		delete(files.files, at(file.Path))
	}
	door, _ := persona.GlobalFilesFor(persona.SelectionClaude)
	delete(files.files, atHome(door.Path))

	result := runSelected(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi",
		fakeAgentSelection{stored: persona.SelectionOpenCode})

	if result.Code != CodeOK {
		t.Fatalf("the selection does not break the project: %s", strings.Join(result.Lines, " / "))
	}
	got := checkOf(t, result, domaindoctor.CheckVoice)
	if got.Status != domaindoctor.StatusOK {
		t.Fatalf("check 11 must be ok with the chosen client complete: %s", domaindoctor.Line(got))
	}
	for _, want := range []string{"I only check opencode", persona.SelectionPath} {
		if !strings.Contains(got.Detail, want) {
			t.Errorf("the line must say %q: %q", want, got.Detail)
		}
	}
	if strings.Contains(got.Detail, ".claude") {
		t.Errorf("the line names .claude although it was not the chosen client: %q", got.Detail)
	}
}

// TestDoctorNamesTheChosenClientWhenMissing: the filter is not a way to hide a
// missing file: the client that was chosen is still checked and named.
func TestDoctorNamesTheChosenClientWhenMissing(t *testing.T) {
	files := configured()
	delete(files.files, at(".opencode/agents/odachi.md"))

	result := runSelected(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi",
		fakeAgentSelection{stored: persona.SelectionOpenCode})

	if result.Code != CodeOK {
		t.Fatalf("a missing agent does not break the project: %s", strings.Join(result.Lines, " / "))
	}
	got := checkOf(t, result, domaindoctor.CheckVoice)
	for _, want := range []string{"are missing in", ".opencode/agents/odachi.md", "odachi init"} {
		if !strings.Contains(got.Detail, want) {
			t.Errorf("the line must say %q: %q", want, got.Detail)
		}
	}
	if strings.Contains(got.Detail, ".claude") {
		t.Errorf("the line names .claude although it was not the chosen client: %q", got.Detail)
	}
}

// TestDoctorNamesABrokenSelection: a store the tool cannot read is said, and the
// persona checks are skipped instead of guessing a client.
func TestDoctorNamesABrokenSelection(t *testing.T) {
	result := runSelected(configured(), cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi",
		fakeAgentSelection{err: errors.New("agents: invalid TOML")})

	if result.Code != CodeOK {
		t.Fatalf("a broken selection does not break the project: %s", strings.Join(result.Lines, " / "))
	}
	got := checkOf(t, result, domaindoctor.CheckVoice)
	if got.Status != domaindoctor.StatusOK {
		t.Fatalf("check 11 must stay ok with a note: %s", domaindoctor.Line(got))
	}
	for _, want := range []string{persona.SelectionPath, "odachi init"} {
		if !strings.Contains(got.Detail, want) {
			t.Errorf("the line must say %q: %q", want, got.Detail)
		}
	}
	if strings.Contains(got.Detail, "matches the block") {
		t.Errorf("the line checked the agents by guessing a client: %q", got.Detail)
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
	if !strings.Contains(got.Detail, "global door is up to date") {
		t.Fatalf("the line does not say the door is up to date: %q", got.Detail)
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
	if got.Status != domaindoctor.StatusProblem || !strings.Contains(got.Detail, "permission denied") {
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
		t.Fatalf("doctor counts %s and init installs into %s", installedSkillsDir, row.Dest)
	}
}

// TestATaskListThatCannotBeReadIsBroken: another shape of task document is a
// decision and stays a warning; a directory the tool cannot read is a broken
// install. "no foreign documents" and "I could not look" are different answers,
// and the second one is not green.
func TestATaskListThatCannotBeReadIsBroken(t *testing.T) {
	files := configured().
		put("proof/.gitignore", "*\n!.gitignore\n")
	files.errOn = at("odd/tasks")

	got := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	if got.Code == CodeOK {
		t.Fatalf("an unreadable task list must not be clean:\n%s", strings.Join(got.Lines, "\n"))
	}
	if got.Problems != 1 {
		t.Errorf("problems = %d, want exactly the unreadable list", got.Problems)
	}
	check := checkOf(t, got, domaindoctor.CheckTaskDocs)
	if check.Status != domaindoctor.StatusProblem {
		t.Fatalf("an unreadable task list is a problem: %s", domaindoctor.Line(check))
	}
	if !strings.Contains(check.Detail, "could not read odd/tasks") {
		t.Errorf("the line does not say what could not be read: %s", domaindoctor.Line(check))
	}
	if !strings.Contains(check.Fix, "check the permissions of odd/tasks") {
		t.Errorf("the repair is not named: %s", domaindoctor.Line(check))
	}
}

// TestDoctorNamesAGlobalSubagentThatDoesNotMatch: the door is global and so are
// the agents it invokes; a stale one is named with the command that fixes it.
func TestDoctorNamesAGlobalSubagentThatDoesNotMatch(t *testing.T) {
	files := configured()
	for _, client := range persona.Clients() {
		for _, file := range persona.GlobalSubagentFilesFor(client.Agent) {
			files.putHome(file.Path, file.Content)
		}
	}

	clean := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")
	if got := checkOf(t, clean, domaindoctor.CheckVoice); !strings.Contains(got.Detail, "global subagents are up to date") {
		t.Fatalf("a seeded home is not reported as clean: %q", got.Detail)
	}

	stale := persona.GlobalSubagentFilesFor("opencode")[0]
	files.putHome(stale.Path, "stale\n")

	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	got := checkOf(t, result, domaindoctor.CheckVoice)
	for _, want := range []string{"global subagents do not match this binary", "~/" + stale.Path, "odachi init --global"} {
		if !strings.Contains(got.Detail, want) {
			t.Errorf("the line must say %q: %q", want, got.Detail)
		}
	}
}

// TestDoctorDoesNotOverwriteARecordedWrite: the store says this project writes
// its agents, so a missing mirror is the mirror's problem. Sending the human to
// --no-project-agents would record the opposite of what the store already says.
func TestDoctorDoesNotOverwriteARecordedWrite(t *testing.T) {
	files := configured()
	body, err := rules.Body(persona.Novato)
	if err != nil {
		t.Fatalf("rules.Body(novato): %v", err)
	}
	for _, file := range persona.Files(body) {
		delete(files.files, at(file.Path))
	}
	for _, file := range persona.SubagentFiles() {
		delete(files.files, at(file.Path))
	}
	for _, client := range persona.Clients() {
		for _, file := range persona.GlobalSubagentFilesFor(client.Agent) {
			files.putHome(file.Path, file.Content)
		}
	}

	result := runSelected(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi",
		fakeAgentSelection{stored: persona.SelectionBoth, project: persona.ProjectAgentsWrite})

	got := checkOf(t, result, domaindoctor.CheckVoice)
	if strings.Contains(got.Detail, "--no-project-agents") {
		t.Fatalf("the line records the opposite of the store: %q", got.Detail)
	}
	for _, want := range []string{"project_agents = write", "run odachi init"} {
		if !strings.Contains(got.Detail, want) {
			t.Errorf("the line must say %q: %q", want, got.Detail)
		}
	}
}

// TestDoctorNamesLeftoverMirrorsTheStoreTurnedOff: the store says the agent lives
// in the home, and an older mirror is still in the tree. "run odachi init" is a
// no-op that skips those files, so the line names what actually resolves it.
func TestDoctorNamesLeftoverMirrorsTheStoreTurnedOff(t *testing.T) {
	files := configured()
	files.files[at(".opencode/agents/odachi.md")] = "an old voice\n"

	result := runSelected(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi",
		fakeAgentSelection{stored: persona.SelectionBoth, project: persona.ProjectAgentsNone})

	got := checkOf(t, result, domaindoctor.CheckVoice)
	for _, want := range []string{"leftover", "--agents " + persona.SelectionBoth} {
		if !strings.Contains(got.Detail, want) {
			t.Errorf("the line must say %q: %q", want, got.Detail)
		}
	}
	if strings.Contains(got.Detail, "run odachi init\"") {
		t.Errorf("the line sends the human to a command that skips these files: %q", got.Detail)
	}
}

// TestDoctorNamesTheStoredSelectionAsTheWayBack: the leftover line points at the
// client selection the store already has. "opencode" written into the message
// would rewrite a project that chose both.
func TestDoctorNamesTheStoredSelectionAsTheWayBack(t *testing.T) {
	files := configured()
	files.files[at(".opencode/agents/odachi.md")] = "an old voice\n"

	result := runSelected(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi",
		fakeAgentSelection{stored: persona.SelectionBoth, project: persona.ProjectAgentsNone})

	got := checkOf(t, result, domaindoctor.CheckVoice)
	if !strings.Contains(got.Detail, "--agents "+persona.SelectionBoth) {
		t.Fatalf("the line does not name the stored selection: %q", got.Detail)
	}
	if strings.Contains(got.Detail, "--agents "+persona.SelectionOpenCode) {
		t.Errorf("the line names one client for a project that chose both: %q", got.Detail)
	}
}

// TestDoctorCallsTheCleanModeADecision: with the door and its subagents in the
// home and the store saying so, a project without agent files is the clean mode,
// not a defect - and the line names the store, the reason the files are absent.
func TestDoctorCallsTheCleanModeADecision(t *testing.T) {
	files := configured()
	body, err := rules.Body(persona.Novato)
	if err != nil {
		t.Fatalf("rules.Body(novato): %v", err)
	}
	for _, file := range persona.Files(body) {
		delete(files.files, at(file.Path))
	}
	for _, file := range persona.SubagentFiles() {
		delete(files.files, at(file.Path))
	}
	for _, client := range persona.Clients() {
		for _, file := range persona.GlobalSubagentFilesFor(client.Agent) {
			files.putHome(file.Path, file.Content)
		}
	}

	result := runSelected(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi",
		fakeAgentSelection{stored: persona.SelectionBoth, project: persona.ProjectAgentsNone})

	got := checkOf(t, result, domaindoctor.CheckVoice)
	if !strings.Contains(got.Detail, "the project has no agent files") {
		t.Fatalf("the clean mode is not named: %q", got.Detail)
	}
	if !strings.Contains(got.Detail, persona.SelectionPath) {
		t.Errorf("the line does not name the store that says so: %q", got.Detail)
	}
	if strings.Contains(got.Detail, "are missing in") {
		t.Errorf("a decision is reported as a defect: %q", got.Detail)
	}
}

// TestDoctorNamesAProjectWhoseStoreSaysNothing: the mirrors are gone and nothing
// recorded where they went. The clean mode is not a decision the tool can assume:
// the line names the command that records it, instead of blessing the absence.
func TestDoctorNamesAProjectWhoseStoreSaysNothing(t *testing.T) {
	files := configured()
	body, err := rules.Body(persona.Novato)
	if err != nil {
		t.Fatalf("rules.Body(novato): %v", err)
	}
	for _, file := range persona.Files(body) {
		delete(files.files, at(file.Path))
	}
	for _, file := range persona.SubagentFiles() {
		delete(files.files, at(file.Path))
	}
	for _, client := range persona.Clients() {
		for _, file := range persona.GlobalSubagentFilesFor(client.Agent) {
			files.putHome(file.Path, file.Content)
		}
	}

	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	got := checkOf(t, result, domaindoctor.CheckVoice)
	for _, want := range []string{persona.SelectionPath, "odachi init --global to record it", "odachi init --agents"} {
		if !strings.Contains(got.Detail, want) {
			t.Errorf("the line must say %q: %q", want, got.Detail)
		}
	}
	// The advice has to be a command this binary registers: --no-project-agents
	// was removed with the global install, and a human following it got an
	// invalid flag instead of a fix.
	if strings.Contains(got.Detail, "--no-project-agents") {
		t.Errorf("the line names a flag that no longer exists: %q", got.Detail)
	}
}

// TestDoctorPointsToTheDoorWhenTheStoreIsHomeOnly: the store says none and the
// home carries nothing, so no agent runs. The fix is the door, not the mirrors:
// "run odachi init" would put back the files the person chose to leave behind.
func TestDoctorPointsToTheDoorWhenTheStoreIsHomeOnly(t *testing.T) {
	files := configured()
	body, err := rules.Body(persona.Novato)
	if err != nil {
		t.Fatalf("rules.Body(novato): %v", err)
	}
	for _, file := range persona.Files(body) {
		delete(files.files, at(file.Path))
	}
	for _, file := range persona.SubagentFiles() {
		delete(files.files, at(file.Path))
	}

	result := runSelected(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi",
		fakeAgentSelection{stored: persona.SelectionBoth, project: persona.ProjectAgentsNone})

	got := checkOf(t, result, domaindoctor.CheckVoice)
	if !strings.Contains(got.Detail, "project_agents = none") {
		t.Fatalf("the line does not name the stored choice: %q", got.Detail)
	}
	if !strings.Contains(got.Detail, "odachi init --global") {
		t.Errorf("the line does not name the door: %q", got.Detail)
	}
}

// TestDoctorSaysTheMirrorsAreLeftoversWhenTheStoreIsHomeOnly: the files match the
// block and the store says the agent lives in the home. Saying only "they match"
// would bless files nobody reads; the line names the store and what the files are,
// which is the state right after `init --global --no-project-agents` over a project
// whose mirrors were current.
func TestDoctorSaysTheMirrorsAreLeftoversWhenTheStoreIsHomeOnly(t *testing.T) {
	files := configured()
	for _, client := range persona.Clients() {
		for _, file := range persona.GlobalSubagentFilesFor(client.Agent) {
			files.putHome(file.Path, file.Content)
		}
	}

	result := runSelected(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi",
		fakeAgentSelection{stored: persona.SelectionBoth, project: persona.ProjectAgentsNone})

	got := checkOf(t, result, domaindoctor.CheckVoice)
	for _, want := range []string{"matches the block", "project_agents = none", "leftovers nobody reads"} {
		if !strings.Contains(got.Detail, want) {
			t.Errorf("the line must say %q: %q", want, got.Detail)
		}
	}
}

// TestDoctorWarnsWhenTheBlockIsOlder: presence is not currency. A project that upgraded
// odachi keeps the rules it adopted with, and a rule that arrived later never reaches its
// agent - the folder rule of 0110 is exactly that case. It is a warning, not a blocker:
// the old block still works.
func TestDoctorWarnsWhenTheBlockIsOlder(t *testing.T) {
	files := configured()
	older := strings.Replace(blockOf(persona.Novato), "One more, about folders.", "One more, about nothing at all.", 1)
	if older == blockOf(persona.Novato) {
		t.Fatal("the fixture did not change the block")
	}
	files.put("AGENTS.md", older)

	got := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	if got.Warnings != 1 {
		t.Fatalf("warnings = %d, want 1: %v", got.Warnings, got.Lines)
	}
	lines := strings.Join(got.Lines, " | ")
	if !strings.Contains(lines, "older block") || !strings.Contains(lines, "odachi init") {
		t.Errorf("the warning does not say what is old nor how to fix it: %s", lines)
	}
	if got.Problems != 0 {
		t.Errorf("an older block blocked the run: problems = %d", got.Problems)
	}
}

// TestOnlyLineEndingsDoNotMakeABlockStale: a checkout is not an older rule. CRLF and a
// trailing space are what git does to a file; 0031 wrote the same rule for the brief.
// This check is the one under test: check 11 reads the voice another way and has its own
// row for the CRLF case.
func TestOnlyLineEndingsDoNotMakeABlockStale(t *testing.T) {
	files := configured()
	files.put("AGENTS.md", strings.ReplaceAll(blockOf(persona.Novato), "\n", "\r\n")+"   ")

	got := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	if check := checkOf(t, got, domaindoctor.CheckAgents); check.Status != domaindoctor.StatusOK {
		t.Fatalf("a checkout made the block stale: %s", check.Detail)
	}
}

// TestTheBlockIsComparedWithoutAProfileFile: a project that ran init and never ran start
// has no odd/profile.toml, and the port answers an inferred profile -the name init writes
// the block with-. Skipping the comparison there would leave an older block passing in
// silence exactly where a project is most likely to carry one.
func TestTheBlockIsComparedWithoutAProfileFile(t *testing.T) {
	files := configured()
	delete(files.files, at("odd/profile.toml"))
	files.put("AGENTS.md", strings.Replace(blockOf(persona.Novato), "One more, about folders.", "One more, about nothing at all.", 1))

	// The port is what answers "inferred": no file, and the name it would infer.
	got := run(files, cleanLookup(), &fakeProfiles{name: persona.Novato, inferred: true}, "/home/ana/bin/odachi")

	if check := checkOf(t, got, domaindoctor.CheckAgents); check.Status != domaindoctor.StatusWarning {
		t.Fatalf("an older block passed without a profile file: %s %s", check.Status, check.Detail)
	}
}

// TestAHomeScopeBlockIsCurrent: an adopted project writes the block for the
// machine's copy, and the doctor cannot know which scope the project chose without
// repeating the inference that wrote it. A home block is current, and the check
// says so instead of advising an init that would write the same bytes again.
func TestAHomeScopeBlockIsCurrent(t *testing.T) {
	files := configured()
	home, err := rules.BlockFor(persona.Novato, persona.SkillsHome)
	if err != nil {
		t.Fatalf("BlockFor(home): %v", err)
	}
	if home == blockOf(persona.Novato) {
		t.Fatal("the two scopes write the same block: the case cannot be distinguished")
	}
	files.put("AGENTS.md", home)

	got := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	if check := checkOf(t, got, domaindoctor.CheckAgents); check.Status != domaindoctor.StatusOK {
		t.Fatalf("a home-scope block was called older: %s", check.Detail)
	}
}
