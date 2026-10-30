package pack_test

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/pack"
	"github.com/biggs-100/odachi/internal/domain/agent"
)

// bundled is the set every test installs: the five skills the binary carries,
// with content distinctive enough to tell an install from a silent skip.
func bundled() []pack.Skill {
	names := []string{"acceptance", "data", "ship", "slices", "tests"}
	skills := make([]pack.Skill, 0, len(names))
	for _, name := range names {
		skills = append(skills, pack.Skill{
			Name:    name,
			Content: []byte("---\nname: " + name + "\ndescription: Trigger: " + name + "\n---\n\n## When\n\ncuando\n"),
		})
	}
	return skills
}

// fakeFiles is the Files port in memory. It records writes and backups apart
// from reads, because "wrote nothing" is the assertion dry-run exists for.
type fakeFiles struct {
	files   map[string][]byte
	writes  []string
	backups []string
	events  []string
	// denyWrite fails every write whose path has one of these prefixes, which is
	// how a test makes one agent fail without breaking the others.
	denyWrite []string
}

func newFakeFiles(seed map[string]string) *fakeFiles {
	f := &fakeFiles{files: map[string][]byte{}}
	for path, content := range seed {
		f.files[path] = []byte(content)
	}
	return f
}

// Exists models a real filesystem: a file exists when it was stored, and a
// directory exists when any file lives under it. Detection asks about a config
// directory, which is exactly that case.
func (f *fakeFiles) Exists(path string) (bool, error) {
	if _, ok := f.files[path]; ok {
		return true, nil
	}
	prefix := path + string(filepath.Separator)
	for key := range f.files {
		if strings.HasPrefix(key, prefix) {
			return true, nil
		}
	}
	return false, nil
}

// List returns the entries of a directory: the next path segment of every file
// stored under it, which is how os.ReadDir sees directories and files.
func (f *fakeFiles) List(dir string) ([]string, error) {
	prefix := dir + string(filepath.Separator)
	var names []string
	for path := range f.files {
		if !strings.HasPrefix(path, prefix) {
			continue
		}
		rest := strings.TrimPrefix(path, prefix)
		name, _, _ := strings.Cut(rest, string(filepath.Separator))
		if !slices.Contains(names, name) {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return names, nil
}

func (f *fakeFiles) Read(path string) ([]byte, error) {
	data, ok := f.files[path]
	if !ok {
		return nil, fmt.Errorf("archivo inexistente: %s", path)
	}
	return data, nil
}

func (f *fakeFiles) Backup(path string) (string, error) {
	if _, ok := f.files[path]; !ok {
		return "", fmt.Errorf("cannot back up %s: it does not exist", path)
	}
	f.backups = append(f.backups, path)
	f.events = append(f.events, "backup "+path)
	return path + ".bak-prueba", nil
}

func (f *fakeFiles) Write(path string, data []byte) error {
	for _, prefix := range f.denyWrite {
		if strings.HasPrefix(path, prefix) {
			return errors.New("permiso denegado")
		}
	}
	f.writes = append(f.writes, path)
	f.events = append(f.events, "write "+path)
	f.files[path] = data
	return nil
}

func (f *fakeFiles) content(path string) (string, bool) {
	data, ok := f.files[path]
	return string(data), ok
}

// fakeHome is the Home port: a directory the test owns.
type fakeHome struct {
	dir string
	err error
}

func (h fakeHome) Home() (string, error) {
	if h.err != nil {
		return "", h.err
	}
	return h.dir, nil
}

// home builds a home directory string the way the app joins it.
func home() string { return filepath.Join("/", "home", "prueba") }

// projectDir is the project root a project-scoped agent resolves against: owned
// by the test, exactly like the home.
func projectDir() string { return filepath.Join("/", "work", "proyecto") }

// dest is where one agent's skills live: the root it belongs to joined with the
// table's destination. Callers pass the home for a home agent and the project
// for a project one, which is the whole point of the scope.
func dest(root, name string) string {
	a, err := agent.Resolve(name)
	if err != nil {
		panic(err)
	}
	return filepath.Join(root, filepath.FromSlash(a.Dest))
}

// target is one installed SKILL.md.
func target(root, agentName, skill string) string {
	return filepath.Join(dest(root, agentName), skill, "SKILL.md")
}

func run(files *fakeFiles, h fakeHome, opts pack.Options) pack.Result {
	if opts.Skills == nil {
		opts.Skills = bundled()
	}
	// Every run resolves the project against a root the test owns, never the
	// working directory of the suite.
	if opts.Dir == "" {
		opts.Dir = projectDir()
	}
	return pack.Run(opts, files, h)
}

// TestInstallsEverySkillOfThePayloadForOneAgent: one agent, every file, the exact bytes,
// and the one-line summary Contract VIII fixes.
func TestInstallsEverySkillOfThePayloadForOneAgent(t *testing.T) {
	root := home()
	files := newFakeFiles(nil)
	result := run(files, fakeHome{dir: root}, pack.Options{Agents: "claude"})

	if result.Code != pack.CodeOK {
		t.Fatalf("code = %d, want 0 (message %q, lines %v)", result.Code, result.Message, result.Lines)
	}
	for _, skill := range bundled() {
		got, ok := files.content(target(root, "claude", skill.Name))
		if !ok {
			t.Errorf("%s was not written", skill.Name)
			continue
		}
		if got != string(skill.Content) {
			t.Errorf("%s content = %q, want %q", skill.Name, got, skill.Content)
		}
	}
	if len(files.writes) != len(bundled()) {
		t.Errorf("writes = %d, want %d: %v", len(files.writes), len(bundled()), files.writes)
	}
	if !slices.Contains(result.Lines, "claude · 5 installed · ~/.claude/skills") {
		t.Errorf("the summary line is missing, lines: %v", result.Lines)
	}
	if result.Message == "" || strings.Contains(result.Message, "\n") {
		t.Errorf("message must be one line, got %q", result.Message)
	}
}

// TestWritesNothingOutsideTheAgentDirectories: the tool writes into the agent's
// skills directory and nowhere else — never the project, never a sibling.
func TestWritesNothingOutsideTheAgentDirectories(t *testing.T) {
	root := home()
	files := newFakeFiles(nil)
	run(files, fakeHome{dir: root}, pack.Options{Agents: "claude,codex"})

	for _, path := range append(append([]string{}, files.writes...), files.backups...) {
		if !strings.HasPrefix(path, root) {
			t.Errorf("touched a path outside the home: %s", path)
		}
		if strings.Contains(path, "SKILLS.md") {
			t.Errorf("touched the project registry: %s", path)
		}
	}
}

// TestSecondIdenticalRunWritesNothingAndBacksUpNothing: idempotent in bytes and
// in side effects. A repeated `pack` is safe to run on every pull.
func TestSecondIdenticalRunWritesNothingAndBacksUpNothing(t *testing.T) {
	root := home()
	files := newFakeFiles(nil)
	run(files, fakeHome{dir: root}, pack.Options{Agents: "claude"})

	firstWrites := len(files.writes)
	if firstWrites != len(bundled()) {
		t.Fatalf("first run wrote %d files, want %d", firstWrites, len(bundled()))
	}

	files.writes, files.backups, files.events = nil, nil, nil
	second := run(files, fakeHome{dir: root}, pack.Options{Agents: "claude"})

	if second.Code != pack.CodeOK {
		t.Fatalf("second run code = %d, want 0 (message %q)", second.Code, second.Message)
	}
	if len(files.writes) != 0 {
		t.Errorf("second run wrote %d files, want 0: %v", len(files.writes), files.writes)
	}
	if len(files.backups) != 0 {
		t.Errorf("second run made %d backups, want 0: %v", len(files.backups), files.backups)
	}
	if !slices.Contains(second.Lines, "claude · 5 unchanged · ~/.claude/skills") {
		t.Errorf("the summary line is missing or wrong, lines: %v", second.Lines)
	}
}

// TestDifferentLocalFileIsBackedUpBeforeTheOverwrite: a file the human edited is
// saved first, and the write lands after the backup, never before it.
func TestDifferentLocalFileIsBackedUpBeforeTheOverwrite(t *testing.T) {
	root := home()
	edited := target(root, "claude", "acceptance")
	files := newFakeFiles(map[string]string{edited: "editado a mano\n"})

	result := run(files, fakeHome{dir: root}, pack.Options{Agents: "claude"})

	if result.Code != pack.CodeOK {
		t.Fatalf("code = %d, want 0 (message %q)", result.Code, result.Message)
	}
	if len(files.backups) != 1 || files.backups[0] != edited {
		t.Fatalf("backups = %v, want [%s]", files.backups, edited)
	}
	backupAt := slices.Index(files.events, "backup "+edited)
	writeAt := slices.Index(files.events, "write "+edited)
	if backupAt < 0 || writeAt < 0 || backupAt > writeAt {
		t.Errorf("the backup must come before the write, events: %v", files.events)
	}
	got, _ := files.content(edited)
	if would := string(bundled()[0].Content); got != would {
		t.Errorf("the file was not replaced with the bundled bytes: %q", got)
	}
	if !slices.Contains(result.Lines, "claude · 4 installed · 1 backed up · ~/.claude/skills") {
		t.Errorf("the summary line is missing or wrong, lines: %v", result.Lines)
	}
}

// TestDryRunTouchesNothing: not one write, not one backup — asserted on the
// double, not on the printed output.
func TestDryRunTouchesNothing(t *testing.T) {
	root := home()
	edited := target(root, "claude", "acceptance")
	files := newFakeFiles(map[string]string{edited: "editado a mano\n"})

	result := run(files, fakeHome{dir: root}, pack.Options{Agents: "claude", DryRun: true})

	if result.Code != pack.CodeOK {
		t.Fatalf("code = %d, want 0 (message %q)", result.Code, result.Message)
	}
	if len(files.writes) != 0 {
		t.Errorf("dry-run wrote %d files: %v", len(files.writes), files.writes)
	}
	if len(files.backups) != 0 {
		t.Errorf("dry-run made %d backups: %v", len(files.backups), files.backups)
	}
	if got, _ := files.content(edited); got != "editado a mano\n" {
		t.Errorf("dry-run changed the local file: %q", got)
	}
	if !slices.Contains(result.Lines, "claude · 4 to install · 1 to back up · ~/.claude/skills") {
		t.Errorf("the dry-run summary is missing or wrong, lines: %v", result.Lines)
	}
	if !strings.Contains(result.Message, "dry-run") {
		t.Errorf("the message does not say it was a dry run: %q", result.Message)
	}
}

// TestDryRunFilesNameWhatWouldBeWritten: Files is documented as "what this run
// wrote, or would write", and a rehearsal of an upgraded machine is exactly the
// case where the list decides a report: a missing skill and a changed one are
// both work, and both have to appear.
func TestDryRunFilesNameWhatWouldBeWritten(t *testing.T) {
	root := home()
	edited := target(root, "claude", "acceptance")
	files := newFakeFiles(map[string]string{edited: "editado a mano\n"})

	result := run(files, fakeHome{dir: root}, pack.Options{Agents: "claude", DryRun: true})

	if result.Code != pack.CodeOK {
		t.Fatalf("code = %d, want 0 (message %q)", result.Code, result.Message)
	}
	if len(result.Dests) != 1 {
		t.Fatalf("Dests = %d, want 1", len(result.Dests))
	}
	dest := result.Dests[0]
	// Four skills: the one edited here and three that are not there yet.
	if want := dest.Installed + dest.BackedUp; len(dest.Files) != want {
		t.Fatalf("Files = %v (%d), want one per skill this run would write (%d)", dest.Files, len(dest.Files), want)
	}
	if !slices.Contains(dest.Files, "~/.claude/skills/acceptance/SKILL.md") {
		t.Errorf("the file it would replace is not named: %v", dest.Files)
	}
}

// TestOneFailureDoesNotAbortTheOthers: a denied write in one agent is reported
// with the agent and the path, and every other agent still gets its skills.
func TestOneFailureDoesNotAbortTheOthers(t *testing.T) {
	root := home()
	files := newFakeFiles(nil)
	files.denyWrite = []string{dest(root, "codex")}

	result := run(files, fakeHome{dir: root}, pack.Options{Agents: "claude,codex"})

	if result.Code != pack.CodeBlocked {
		t.Fatalf("code = %d, want %d", result.Code, pack.CodeBlocked)
	}
	if !strings.Contains(result.Message, "codex") {
		t.Errorf("the message does not name the agent: %q", result.Message)
	}
	if !strings.Contains(result.Message, "SKILL.md") {
		t.Errorf("the message does not name the path: %q", result.Message)
	}
	for _, skill := range bundled() {
		if _, ok := files.content(target(root, "claude", skill.Name)); !ok {
			t.Errorf("claude lost %s because codex failed", skill.Name)
		}
	}
	wrote := false
	for _, line := range result.Lines {
		if strings.Contains(line, "codex") && strings.Contains(line, "SKILL.md") {
			wrote = true
		}
	}
	if !wrote {
		t.Errorf("no line names the agent and the path, lines: %v", result.Lines)
	}
	if !slices.Contains(result.Lines, "claude · 5 installed · ~/.claude/skills") {
		t.Errorf("the healthy agent has no summary, lines: %v", result.Lines)
	}
}

// TestOpenCodeInstallsIntoTheProject: OpenCode reads its skills from inside the
// project, so that is where they go — and the home is not touched at all: not
// one write, not one backup, asserted on the double.
func TestOpenCodeInstallsIntoTheProject(t *testing.T) {
	homeRoot, project := home(), projectDir()
	files := newFakeFiles(nil)

	result := run(files, fakeHome{dir: homeRoot}, pack.Options{Agents: "opencode"})

	if result.Code != pack.CodeOK {
		t.Fatalf("code = %d, want 0 (message %q, lines %v)", result.Code, result.Message, result.Lines)
	}
	for _, skill := range bundled() {
		got, ok := files.content(target(project, "opencode", skill.Name))
		if !ok {
			t.Errorf("%s was not installed in the project", skill.Name)
			continue
		}
		if got != string(skill.Content) {
			t.Errorf("%s content = %q, want %q", skill.Name, got, skill.Content)
		}
	}
	if len(files.writes) != len(bundled()) {
		t.Errorf("writes = %d, want %d: %v", len(files.writes), len(bundled()), files.writes)
	}
	// Nothing in the home, and nothing outside the project either.
	for _, path := range append(append([]string{}, files.writes...), files.backups...) {
		if strings.HasPrefix(path, homeRoot) {
			t.Errorf("touched the home: %s", path)
		}
		if !strings.HasPrefix(path, project) {
			t.Errorf("touched a path outside the project: %s", path)
		}
	}
	if !slices.Contains(result.Lines, "opencode · 5 installed · ./.opencode/skills") {
		t.Errorf("the summary does not show the project-relative path, lines: %v", result.Lines)
	}
	for _, line := range append(append([]string{}, result.Lines...), result.Message) {
		if strings.Contains(line, project) || strings.Contains(line, homeRoot) {
			t.Errorf("a message prints an absolute path: %q", line)
		}
	}
}

// TestOpenCodeSecondRunIsIdenticalInTheProject: the write rule does not change
// with the scope - identical bytes are reported as `unchanged` and the file is
// not rewritten.
func TestOpenCodeSecondRunIsIdenticalInTheProject(t *testing.T) {
	files := newFakeFiles(nil)
	run(files, fakeHome{dir: home()}, pack.Options{Agents: "opencode"})
	files.writes, files.backups, files.events = nil, nil, nil

	second := run(files, fakeHome{dir: home()}, pack.Options{Agents: "opencode"})

	if second.Code != pack.CodeOK {
		t.Fatalf("code = %d, want 0 (message %q)", second.Code, second.Message)
	}
	if len(files.writes) != 0 || len(files.backups) != 0 {
		t.Errorf("the second run wrote %v and backed up %v, want nothing", files.writes, files.backups)
	}
	if !slices.Contains(second.Lines, "opencode · 5 unchanged · ./.opencode/skills") {
		t.Errorf("the second run does not report unchanged, lines: %v", second.Lines)
	}
}

// TestOpenCodeWorksWithoutTheDirectory: naming the agent is enough. `pack`
// creates what is missing; requiring `.opencode/` to exist first would make the
// one command that installs it refuse to install it.
func TestOpenCodeWorksWithoutTheDirectory(t *testing.T) {
	project := projectDir()
	files := newFakeFiles(nil)

	result := run(files, fakeHome{dir: home()}, pack.Options{Agents: "opencode"})

	if result.Code != pack.CodeOK {
		t.Fatalf("code = %d, want 0 (message %q)", result.Code, result.Message)
	}
	if _, err := files.Exists(filepath.Join(project, ".opencode")); err != nil {
		t.Fatalf("Exists: %v", err)
	}
	for _, skill := range bundled() {
		if _, ok := files.content(target(project, "opencode", skill.Name)); !ok {
			t.Errorf("%s was not installed", skill.Name)
		}
	}
}

// TestProjectAgentDoesNotNeedAHome: a project-scoped install resolves against
// the project. Asking for a home it never writes into would turn an unrelated
// problem into a blocked run.
func TestProjectAgentDoesNotNeedAHome(t *testing.T) {
	project := projectDir()
	files := newFakeFiles(nil)

	result := pack.Run(pack.Options{Agents: "opencode", Skills: bundled(), Dir: project}, files, fakeHome{err: errors.New("no home")})

	if result.Code != pack.CodeOK {
		t.Fatalf("code = %d, want 0 (message %q)", result.Code, result.Message)
	}
	for _, skill := range bundled() {
		if _, ok := files.content(target(project, "opencode", skill.Name)); !ok {
			t.Errorf("%s was not installed", skill.Name)
		}
	}
}

// TestDetectsOpenCodeInTheProject: with no --agents, an agent is detected when
// its own config directory exists in the root it belongs to — `.opencode/` in
// the project, `.codex/` in the home.
func TestDetectsOpenCodeInTheProject(t *testing.T) {
	project := projectDir()
	files := newFakeFiles(map[string]string{
		filepath.Join(project, ".opencode", "config.json"): "{}",
	})

	result := run(files, fakeHome{dir: home()}, pack.Options{})

	if result.Code != pack.CodeOK {
		t.Fatalf("code = %d, want 0 (message %q)", result.Code, result.Message)
	}
	if len(result.Dests) != 1 || !slices.Equal(result.Dests[0].Names, []string{"opencode"}) {
		t.Fatalf("dests = %+v, want only opencode", result.Dests)
	}
	if !slices.Contains(result.Lines, "opencode · 5 installed · ./.opencode/skills") {
		t.Errorf("the detected agent has no summary, lines: %v", result.Lines)
	}
	for _, path := range files.writes {
		if !strings.HasPrefix(path, dest(project, "opencode")) {
			t.Errorf("wrote outside the project agent: %s", path)
		}
	}
}

// TestOpenCodeBacksUpAModifiedFile: the write rule does not change with the
// scope either — a file the human edited is saved before it is replaced.
func TestOpenCodeBacksUpAModifiedFile(t *testing.T) {
	project := projectDir()
	edited := target(project, "opencode", "acceptance")
	files := newFakeFiles(map[string]string{edited: "editado a mano\n"})

	result := run(files, fakeHome{dir: home()}, pack.Options{Agents: "opencode"})

	if result.Code != pack.CodeOK {
		t.Fatalf("code = %d, want 0 (message %q)", result.Code, result.Message)
	}
	if len(files.backups) != 1 || files.backups[0] != edited {
		t.Fatalf("backups = %v, want [%s]", files.backups, edited)
	}
	backupAt := slices.Index(files.events, "backup "+edited)
	writeAt := slices.Index(files.events, "write "+edited)
	if backupAt < 0 || writeAt < 0 || backupAt > writeAt {
		t.Errorf("the backup must come before the write, events: %v", files.events)
	}
	if got, _ := files.content(edited); got != string(bundled()[0].Content) {
		t.Errorf("the file was not replaced with the bundled bytes: %q", got)
	}
	if !slices.Contains(result.Lines, "opencode · 4 installed · 1 backed up · ./.opencode/skills") {
		t.Errorf("the summary line is missing or wrong, lines: %v", result.Lines)
	}
}

// TestUnknownAgentWritesNothing: a bad flag blocks the whole run, so a typo
// never leaves half the skills installed.
func TestUnknownAgentWritesNothing(t *testing.T) {
	files := newFakeFiles(nil)
	result := run(files, fakeHome{dir: home()}, pack.Options{Agents: "inventado"})

	if result.Code != pack.CodeBlocked {
		t.Fatalf("code = %d, want %d", result.Code, pack.CodeBlocked)
	}
	if len(files.writes) != 0 || len(files.backups) != 0 {
		t.Errorf("a refused agent still wrote: writes %v, backups %v", files.writes, files.backups)
	}
	for _, name := range agent.Names() {
		if !strings.Contains(result.Message, name) {
			t.Errorf("the message does not name %q as supported: %q", name, result.Message)
		}
	}
}

// TestPiIsRefusedWithItsReason: Pi is refused out loud, not skipped.
func TestPiIsRefusedWithItsReason(t *testing.T) {
	files := newFakeFiles(nil)
	result := run(files, fakeHome{dir: home()}, pack.Options{Agents: "pi"})

	if result.Code != pack.CodeBlocked {
		t.Fatalf("code = %d, want %d", result.Code, pack.CodeBlocked)
	}
	if !strings.Contains(result.Message, "Pi does not use a skills directory") {
		t.Errorf("the message does not say why: %q", result.Message)
	}
	if len(files.writes) != 0 {
		t.Errorf("pi produced writes: %v", files.writes)
	}
}

// TestForeignFileIsReportedAndNotTouched: a skill odachi did not write is
// reported with its path, and its bytes stay exactly as they were.
func TestForeignFileIsReportedAndNotTouched(t *testing.T) {
	root := home()
	foreign := filepath.Join(dest(root, "claude"), "mia", "SKILL.md")
	files := newFakeFiles(map[string]string{foreign: "mi skill\n"})

	result := run(files, fakeHome{dir: root}, pack.Options{Agents: "claude"})

	if result.Code != pack.CodeOK {
		t.Fatalf("code = %d, want 0 (message %q)", result.Code, result.Message)
	}
	if got, _ := files.content(foreign); got != "mi skill\n" {
		t.Errorf("the foreign skill was modified: %q", got)
	}
	if slices.Contains(files.writes, foreign) || slices.Contains(files.backups, foreign) {
		t.Errorf("the foreign skill was written or backed up")
	}
	found := false
	for _, line := range result.Lines {
		if strings.Contains(line, "~/.claude/skills/mia/SKILL.md") && strings.Contains(line, "not touching it") {
			found = true
		}
	}
	if !found {
		t.Errorf("the foreign skill is not reported, lines: %v", result.Lines)
	}
}

// TestDetectsInstalledAgents: with no --agents, the agents whose config
// directory exists are the ones served, and the ones that are not here are left
// alone.
func TestDetectsInstalledAgents(t *testing.T) {
	root := home()
	files := newFakeFiles(map[string]string{filepath.Join(root, ".codex", "config.toml"): "x"})

	result := run(files, fakeHome{dir: root}, pack.Options{})

	if result.Code != pack.CodeOK {
		t.Fatalf("code = %d, want 0 (message %q)", result.Code, result.Message)
	}
	if len(result.Dests) != 1 || !slices.Equal(result.Dests[0].Names, []string{"codex"}) {
		t.Fatalf("dests = %+v, want only codex", result.Dests)
	}
	for _, path := range files.writes {
		if !strings.HasPrefix(path, dest(root, "codex")) {
			t.Errorf("wrote outside the detected agent: %s", path)
		}
	}
}

// TestNoAgentDetectedNamesTheFlag: an empty home is a starting point, not a
// failure: one Spanish line says how to name an agent.
func TestNoAgentDetectedNamesTheFlag(t *testing.T) {
	files := newFakeFiles(nil)
	result := run(files, fakeHome{dir: home()}, pack.Options{})

	if result.Code != pack.CodeOK {
		t.Fatalf("code = %d, want 0", result.Code)
	}
	if len(files.writes) != 0 {
		t.Errorf("wrote with no agent detected: %v", files.writes)
	}
	if !strings.Contains(result.Message, "--agents") {
		t.Errorf("the message does not name the flag that unblocks it: %q", result.Message)
	}
}

// TestNoBundledSkillsIsBlocked: an empty payload is an installation defect, and
// it says so instead of reporting a successful install of nothing.
func TestNoBundledSkillsIsBlocked(t *testing.T) {
	files := newFakeFiles(nil)
	result := pack.Run(pack.Options{Agents: "claude", Skills: []pack.Skill{}}, files, fakeHome{dir: home()})

	if result.Code != pack.CodeBlocked {
		t.Fatalf("code = %d, want %d", result.Code, pack.CodeBlocked)
	}
	if len(files.writes) != 0 {
		t.Errorf("wrote with no skills: %v", files.writes)
	}
}

// TestEverySupportedAgentResolvesItsOwnDestination: the table, not a branch of
// code, decides where each agent is served — and the scope decides which root
// that destination is joined to.
func TestEverySupportedAgentResolvesItsOwnDestination(t *testing.T) {
	homeRoot, project := home(), projectDir()
	root := func(a agent.Agent) string {
		if a.Scope == agent.ScopeProject {
			return project
		}
		return homeRoot
	}
	seed := map[string]string{}
	for _, a := range agent.Supported() {
		seed[filepath.Join(root(a), filepath.FromSlash(a.Marker()), "config")] = "x"
	}
	files := newFakeFiles(seed)

	result := run(files, fakeHome{dir: homeRoot}, pack.Options{})

	if result.Code != pack.CodeOK {
		t.Fatalf("code = %d, want 0 (message %q)", result.Code, result.Message)
	}
	if len(result.Dests) != len(agent.Supported()) {
		t.Fatalf("served %d destinations, want %d: %+v", len(result.Dests), len(agent.Supported()), result.Dests)
	}
	for _, a := range agent.Supported() {
		if !slices.Contains(result.Lines, a.Name+" · 5 installed · "+a.Display()) {
			t.Errorf("no summary for %s, lines: %v", a.Name, result.Lines)
		}
		for _, skill := range bundled() {
			if _, ok := files.content(filepath.Join(root(a), filepath.FromSlash(a.Dest), skill.Name, "SKILL.md")); !ok {
				t.Errorf("%s: %s was not installed", a.Name, skill.Name)
			}
		}
	}
}

// TestOnlyTheProjectAgentWritesIntoTheProject: a home agent still writes into
// the home and never into the project, so adding a scope moved nobody.
func TestOnlyTheProjectAgentWritesIntoTheProject(t *testing.T) {
	project := projectDir()
	files := newFakeFiles(nil)

	run(files, fakeHome{dir: home()}, pack.Options{Agents: "claude,opencode"})

	for _, path := range files.writes {
		intoProject := strings.HasPrefix(path, project)
		intoHome := strings.HasPrefix(path, home())
		if intoProject == intoHome {
			t.Errorf("path is neither in the project nor in the home: %s", path)
		}
		if intoProject && !strings.Contains(path, filepath.Join(".opencode", "skills")) {
			t.Errorf("a non-opencode agent wrote into the project: %s", path)
		}
	}
}

// TestUnresolvableHomeIsBlocked: no home, no writes, one line that says what to
// do about it.
func TestUnresolvableHomeIsBlocked(t *testing.T) {
	files := newFakeFiles(nil)
	result := pack.Run(pack.Options{Agents: "claude", Skills: bundled()}, files, fakeHome{err: errors.New("no home")})

	if result.Code != pack.CodeBlocked {
		t.Fatalf("code = %d, want %d", result.Code, pack.CodeBlocked)
	}
	if len(files.writes) != 0 {
		t.Errorf("wrote without a home: %v", files.writes)
	}
	if !strings.Contains(result.Message, "home") {
		t.Errorf("the message does not say what failed: %q", result.Message)
	}
}

// scopedTarget is one installed SKILL.md once --scope has been applied: the
// root the override names, joined with the agent's destination under it.
func scopedTarget(t *testing.T, root, name string, scope agent.Scope, skill string) string {
	t.Helper()
	a, err := agent.Resolve(name)
	if err != nil {
		t.Fatalf("Resolve(%q): %v", name, err)
	}
	return filepath.Join(root, filepath.FromSlash(a.In(scope).Dest), skill, "SKILL.md")
}

// TestScopePutsTheSameAgentInTheOtherRoot: `--scope project` installs a home
// agent inside the project and never touches the home; `--scope home` installs
// the project agent in the home, at OpenCode's documented global config dir.
func TestScopePutsTheSameAgentInTheOtherRoot(t *testing.T) {
	homeRoot, project := home(), projectDir()
	files := newFakeFiles(nil)

	result := run(files, fakeHome{dir: homeRoot}, pack.Options{Agents: "claude,opencode", Scope: "project"})

	if result.Code != pack.CodeOK {
		t.Fatalf("code = %d, want 0 (message %q, lines %v)", result.Code, result.Message, result.Lines)
	}
	for _, name := range []string{"claude", "opencode"} {
		for _, skill := range bundled() {
			if _, ok := files.content(scopedTarget(t, project, name, agent.ScopeProject, skill.Name)); !ok {
				t.Errorf("%s: %s was not installed in the project", name, skill.Name)
			}
		}
	}
	for _, path := range append(append([]string{}, files.writes...), files.backups...) {
		if strings.HasPrefix(path, homeRoot) {
			t.Errorf("--scope project touched the home: %s", path)
		}
	}
	for _, want := range []string{
		"claude · 5 installed · ./.claude/skills",
		"opencode · 5 installed · ./.opencode/skills",
	} {
		if !slices.Contains(result.Lines, want) {
			t.Errorf("missing summary %q, lines: %v", want, result.Lines)
		}
	}

	// The other root: OpenCode's home destination is its global config dir.
	homeFiles := newFakeFiles(nil)
	homeResult := run(homeFiles, fakeHome{dir: homeRoot}, pack.Options{Agents: "opencode", Scope: "home"})
	if homeResult.Code != pack.CodeOK {
		t.Fatalf("home run code = %d, want 0 (message %q)", homeResult.Code, homeResult.Message)
	}
	for _, skill := range bundled() {
		if _, ok := homeFiles.content(scopedTarget(t, homeRoot, "opencode", agent.ScopeHome, skill.Name)); !ok {
			t.Errorf("opencode: %s was not installed in the home", skill.Name)
		}
	}
	if !slices.Contains(homeResult.Lines, "opencode · 5 installed · ~/.config/opencode/skills") {
		t.Errorf("the home summary does not show the global config dir, lines: %v", homeResult.Lines)
	}
	for _, path := range append(append([]string{}, homeFiles.writes...), homeFiles.backups...) {
		if strings.HasPrefix(path, project) {
			t.Errorf("--scope home touched the project: %s", path)
		}
	}
}

// TestDestServesAClientTheTableDoesNotKnow: --dest is the escape hatch that does
// not depend on odachi predicting anything. The name is a label, the folder is
// the human's, and neither the table nor the home enters the picture.
func TestDestServesAClientTheTableDoesNotKnow(t *testing.T) {
	project := projectDir()
	files := newFakeFiles(nil)

	result := run(files, fakeHome{err: errors.New("no home")}, pack.Options{Agents: "inventado,claude", Dest: "./skills-x"})

	if result.Code != pack.CodeOK {
		t.Fatalf("code = %d, want 0 (message %q, lines %v)", result.Code, result.Message, result.Lines)
	}
	for _, skill := range bundled() {
		got, ok := files.content(filepath.Join(project, "skills-x", skill.Name, "SKILL.md"))
		if !ok {
			t.Errorf("%s was not installed in --dest", skill.Name)
			continue
		}
		if got != string(skill.Content) {
			t.Errorf("%s content = %q, want %q", skill.Name, got, skill.Content)
		}
	}
	// One copy, not one per agent: both names share the folder.
	if len(files.writes) != len(bundled()) {
		t.Errorf("writes = %d, want %d: %v", len(files.writes), len(bundled()), files.writes)
	}
	// One summary per destination, naming every agent that landed there.
	if !slices.Contains(result.Lines, "inventado, claude · 5 installed · ./skills-x") {
		t.Errorf("the shared destination has no summary, lines: %v", result.Lines)
	}

	// The write rule does not change: an identical second run writes nothing.
	files.writes, files.backups, files.events = nil, nil, nil
	second := run(files, fakeHome{err: errors.New("no home")}, pack.Options{Agents: "inventado,claude", Dest: "./skills-x"})
	if second.Code != pack.CodeOK {
		t.Fatalf("second run code = %d, want 0 (message %q)", second.Code, second.Message)
	}
	if len(files.writes) != 0 || len(files.backups) != 0 {
		t.Errorf("the second run wrote %v and backed up %v, want nothing", files.writes, files.backups)
	}
	if !slices.Contains(second.Lines, "inventado, claude · 5 unchanged · ./skills-x") {
		t.Errorf("the second run does not report unchanged, lines: %v", second.Lines)
	}
}

// TestDestAbsoluteIsUsedAsGiven: an absolute path is the path, not a path
// relative to the project, and the summary shows it as the human typed it.
func TestDestAbsoluteIsUsedAsGiven(t *testing.T) {
	// An absolute path for this platform: a temp directory the test owns. The
	// fake Files never touches the disk, so nothing is written outside it.
	outside := filepath.Join(t.TempDir(), "mi-cliente")
	files := newFakeFiles(nil)

	result := run(files, fakeHome{dir: home()}, pack.Options{Agents: "claude", Dest: outside})

	if result.Code != pack.CodeOK {
		t.Fatalf("code = %d, want 0 (message %q)", result.Code, result.Message)
	}
	for _, skill := range bundled() {
		if _, ok := files.content(filepath.Join(outside, skill.Name, "SKILL.md")); !ok {
			t.Errorf("%s was not installed in the absolute --dest", skill.Name)
		}
	}
	label := filepath.ToSlash(filepath.Clean(outside))
	if !slices.Contains(result.Lines, "claude · 5 installed · "+label) {
		t.Errorf("the summary does not show %q, lines: %v", label, result.Lines)
	}
}

// TestDestAndScopeAreRefusedTogether: two flags that both say where, refused
// with a message that says which one wins where.
func TestDestAndScopeAreRefusedTogether(t *testing.T) {
	files := newFakeFiles(nil)

	result := run(files, fakeHome{dir: home()}, pack.Options{Agents: "claude", Dest: "./skills-x", Scope: "home"})

	if result.Code != pack.CodeBlocked {
		t.Fatalf("code = %d, want %d", result.Code, pack.CodeBlocked)
	}
	for _, want := range []string{"--dest", "--scope", "choose one"} {
		if !strings.Contains(result.Message, want) {
			t.Errorf("the refusal does not name %q: %q", want, result.Message)
		}
	}
	if len(files.writes) != 0 || len(files.backups) != 0 {
		t.Errorf("a refused combination wrote: %v %v", files.writes, files.backups)
	}
}

// TestDestAndScopeNeedNamedAgents: both flags speak about the agents the human
// named, so without --agents they refuse and name the command that unblocks it.
func TestDestAndScopeNeedNamedAgents(t *testing.T) {
	cases := map[string]pack.Options{
		"--dest":  {Dest: "./skills-x"},
		"--scope": {Scope: "home"},
	}
	for name, opts := range cases {
		t.Run(name, func(t *testing.T) {
			files := newFakeFiles(nil)
			result := run(files, fakeHome{dir: home()}, opts)

			if result.Code != pack.CodeBlocked {
				t.Fatalf("code = %d, want %d (message %q)", result.Code, pack.CodeBlocked, result.Message)
			}
			if !strings.Contains(result.Message, "--agents") || !strings.Contains(result.Message, name) {
				t.Errorf("the refusal names neither the flag nor --agents: %q", result.Message)
			}
			if len(files.writes) != 0 {
				t.Errorf("wrote with no named agent: %v", files.writes)
			}
		})
	}
}

// TestInvalidScopeIsRefused: the flag is read before anything is written, and a
// typo names the two values that exist.
func TestInvalidScopeIsRefused(t *testing.T) {
	files := newFakeFiles(nil)
	result := run(files, fakeHome{dir: home()}, pack.Options{Agents: "claude", Scope: "cualquier-cosa"})

	if result.Code != pack.CodeBlocked {
		t.Fatalf("code = %d, want %d", result.Code, pack.CodeBlocked)
	}
	for _, want := range []string{"--scope", "project", "home"} {
		if !strings.Contains(result.Message, want) {
			t.Errorf("the refusal does not name %q: %q", want, result.Message)
		}
	}
	if len(files.writes) != 0 {
		t.Errorf("an invalid scope wrote: %v", files.writes)
	}
}

// TestStandardInstallsTheCrossClientLocation: one copy in a location every
// client can read, in either root, without a single marker on disk.
func TestStandardInstallsTheCrossClientLocation(t *testing.T) {
	homeRoot, project := home(), projectDir()
	files := newFakeFiles(nil)

	result := run(files, fakeHome{dir: homeRoot}, pack.Options{Agents: "standard"})

	if result.Code != pack.CodeOK {
		t.Fatalf("code = %d, want 0 (message %q, lines %v)", result.Code, result.Message, result.Lines)
	}
	for _, skill := range bundled() {
		if _, ok := files.content(filepath.Join(project, ".agents", "skills", skill.Name, "SKILL.md")); !ok {
			t.Errorf("%s was not installed in ./.agents/skills", skill.Name)
		}
	}
	for _, path := range append(append([]string{}, files.writes...), files.backups...) {
		if strings.HasPrefix(path, homeRoot) {
			t.Errorf("the project copy touched the home: %s", path)
		}
	}
	if !slices.Contains(result.Lines, "standard · 5 installed · ./.agents/skills") {
		t.Errorf("the summary does not show the cross-client path, lines: %v", result.Lines)
	}

	homeFiles := newFakeFiles(nil)
	homeResult := run(homeFiles, fakeHome{dir: homeRoot}, pack.Options{Agents: "standard", Scope: "home"})
	if homeResult.Code != pack.CodeOK {
		t.Fatalf("home run code = %d, want 0 (message %q)", homeResult.Code, homeResult.Message)
	}
	for _, skill := range bundled() {
		if _, ok := homeFiles.content(filepath.Join(homeRoot, ".agents", "skills", skill.Name, "SKILL.md")); !ok {
			t.Errorf("%s was not installed in ~/.agents/skills", skill.Name)
		}
	}
	if !slices.Contains(homeResult.Lines, "standard · 5 installed · ~/.agents/skills") {
		t.Errorf("the home summary is wrong, lines: %v", homeResult.Lines)
	}
	for _, path := range append(append([]string{}, homeFiles.writes...), homeFiles.backups...) {
		if strings.HasPrefix(path, project) {
			t.Errorf("the home copy touched the project: %s", path)
		}
	}
}

// TestStandardTravelsWithARealAgent: the portable copy and a named client in the
// same run, each to its own destination, one summary each.
func TestStandardTravelsWithARealAgent(t *testing.T) {
	files := newFakeFiles(nil)

	result := run(files, fakeHome{dir: home()}, pack.Options{Agents: "standard,opencode", Scope: "project"})

	if result.Code != pack.CodeOK {
		t.Fatalf("code = %d, want 0 (message %q)", result.Code, result.Message)
	}
	if len(files.writes) != 2*len(bundled()) {
		t.Errorf("writes = %d, want %d", len(files.writes), 2*len(bundled()))
	}
	for _, want := range []string{
		"standard · 5 installed · ./.agents/skills",
		"opencode · 5 installed · ./.opencode/skills",
	} {
		if !slices.Contains(result.Lines, want) {
			t.Errorf("missing summary %q, lines: %v", want, result.Lines)
		}
	}
}

// TestDetectionIsUnchangedAndNeverInventsStandard: `.agents/` on disk is not a
// signal, because `standard` names a location and not an installed program.
func TestDetectionIsUnchangedAndNeverInventsStandard(t *testing.T) {
	project := projectDir()
	files := newFakeFiles(map[string]string{
		filepath.Join(project, ".agents", "skills", "mia", "SKILL.md"): "mi skill\n",
	})

	result := run(files, fakeHome{dir: home()}, pack.Options{})

	if result.Code != pack.CodeOK {
		t.Fatalf("code = %d, want 0 (message %q)", result.Code, result.Message)
	}
	if len(result.Dests) != 0 {
		t.Errorf("detection served %+v, want nobody", result.Dests)
	}
	if len(files.writes) != 0 {
		t.Errorf("detection wrote %v", files.writes)
	}
	if !strings.Contains(result.Message, "--agents") {
		t.Errorf("the message does not name the flag: %q", result.Message)
	}
}

// TestUnknownAgentRefusalNamesTheWayOut: still refused, never guessed — and the
// message now carries the two escape hatches, so the future is not a dead end.
func TestUnknownAgentRefusalNamesTheWayOut(t *testing.T) {
	files := newFakeFiles(nil)
	result := run(files, fakeHome{dir: home()}, pack.Options{Agents: "inventado"})

	if result.Code != pack.CodeBlocked {
		t.Fatalf("code = %d, want %d", result.Code, pack.CodeBlocked)
	}
	for _, want := range []string{"unknown agent: inventado", "--agents standard", "--dest"} {
		if !strings.Contains(result.Message, want) {
			t.Errorf("the refusal does not say %q: %q", want, result.Message)
		}
	}
	if len(files.writes) != 0 {
		t.Errorf("a refused agent wrote: %v", files.writes)
	}
}

// TestTheFirstFailedFileKeepsItsPathAndReason: `odachi init` composes its own
// report from a pack run, so a destination that could not be written has to hand
// back the file and the reason as fields. Parsing a line back into a path and an
// error is how the two reports drift apart.
func TestTheFirstFailedFileKeepsItsPathAndReason(t *testing.T) {
	root := home()
	files := newFakeFiles(nil)
	files.denyWrite = []string{dest(root, "claude")}

	result := run(files, fakeHome{dir: root}, pack.Options{Agents: "claude"})

	if result.Code != pack.CodeBlocked {
		t.Fatalf("code = %d, want %d", result.Code, pack.CodeBlocked)
	}
	if len(result.Dests) != 1 {
		t.Fatalf("dests = %d, want 1", len(result.Dests))
	}
	failure := result.Dests[0].Failure
	if failure == nil {
		t.Fatalf("the destination did not record the failure: %+v", result.Dests[0])
	}
	if want := "~/.claude/skills/acceptance/SKILL.md"; failure.Path != want {
		t.Errorf("Path = %q, want %q", failure.Path, want)
	}
	if want := "permiso denegado"; failure.Reason != want {
		t.Errorf("Reason = %q, want %q", failure.Reason, want)
	}
	if result.Dests[0].Failed != len(bundled()) {
		t.Errorf("Failed = %d, want %d", result.Dests[0].Failed, len(bundled()))
	}
}

// projectSkill is one skill of the project's own catalog, with bytes that cannot
// be confused with the bundled ones.
func projectSkill(name string) pack.Skill {
	return pack.Skill{
		Name:    name,
		Content: []byte("---\nname: " + name + "\ndescription: Trigger: the project's own " + name + "\n---\n\n## When\n\ncuando el proyecto lo pide\n"),
	}
}

// TestTheProjectSkillsAreInstalledToo: a skill created with `odachi skill new`
// lives in skills/ and the agent reads .agents/skills, so pack is the join.
func TestTheProjectSkillsAreInstalledToo(t *testing.T) {
	root := projectDir()
	files := newFakeFiles(nil)

	result := run(files, fakeHome{dir: home()}, pack.Options{
		Agents:  "standard",
		Project: []pack.Skill{projectSkill("marketing")},
	})

	if result.Code != pack.CodeOK {
		t.Fatalf("code = %d, want 0 (message %q, lines %v)", result.Code, result.Message, result.Lines)
	}
	path := target(root, "standard", "marketing")
	if got := string(files.files[path]); got != string(projectSkill("marketing").Content) {
		t.Errorf("%s = %q, want the project's own bytes", path, got)
	}
	if !slices.ContainsFunc(result.Files, func(f string) bool {
		return strings.HasSuffix(f, ".agents/skills/marketing/SKILL.md")
	}) {
		t.Errorf("Files = %v, want it to name the project skill", result.Files)
	}
}

// TestAProjectSkillReplacesABundledOne: the bundled skill is a starter and the
// project copy is the project's decision, so the project wins and the
// replacement is named instead of hidden.
func TestAProjectSkillReplacesABundledOne(t *testing.T) {
	root := projectDir()
	files := newFakeFiles(nil)
	own := projectSkill("data")

	result := run(files, fakeHome{dir: home()}, pack.Options{Agents: "standard", Project: []pack.Skill{own}})

	if result.Code != pack.CodeOK {
		t.Fatalf("code = %d, want 0 (message %q, lines %v)", result.Code, result.Message, result.Lines)
	}
	path := target(root, "standard", "data")
	if got := string(files.files[path]); got != string(own.Content) {
		t.Errorf("%s = %q, want the project's own bytes", path, got)
	}
	if !slices.Contains(result.Lines, "project skill data replaces the bundled one") {
		t.Errorf("the replacement is not named: %v", result.Lines)
	}
}

// TestAProjectWithoutSkillsInstallsTheBundledSet: a user project that never ran
// `odachi skill new` gets exactly what it got before.
func TestAProjectWithoutSkillsInstallsTheBundledSet(t *testing.T) {
	root := projectDir()
	files := newFakeFiles(nil)

	result := run(files, fakeHome{dir: home()}, pack.Options{Agents: "standard"})

	if result.Code != pack.CodeOK {
		t.Fatalf("code = %d, want 0 (message %q, lines %v)", result.Code, result.Message, result.Lines)
	}
	for _, skill := range bundled() {
		if _, ok := files.files[target(root, "standard", skill.Name)]; !ok {
			t.Errorf("the bundled skill %q was not installed", skill.Name)
		}
	}
	for _, line := range result.Lines {
		if strings.Contains(line, "replaces the bundled one") {
			t.Errorf("a run with no project skills reported a replacement: %q", line)
		}
	}
}

// TestDryRunWithProjectSkillsWritesNothing: the plan names the project skill and
// the destination, and the disk stays as it was.
func TestDryRunWithProjectSkillsWritesNothing(t *testing.T) {
	files := newFakeFiles(nil)

	result := run(files, fakeHome{dir: home()}, pack.Options{
		Agents:  "standard",
		DryRun:  true,
		Project: []pack.Skill{projectSkill("marketing")},
	})

	if result.Code != pack.CodeOK {
		t.Fatalf("code = %d, want 0 (message %q, lines %v)", result.Code, result.Message, result.Lines)
	}
	if len(files.writes) != 0 {
		t.Errorf("a dry run wrote %v", files.writes)
	}
	if !slices.ContainsFunc(result.Lines, func(line string) bool {
		return strings.Contains(line, "would install") && strings.Contains(line, "marketing")
	}) {
		t.Errorf("the plan does not name the project skill: %v", result.Lines)
	}
}
