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
			Content: []byte("---\nname: " + name + "\ndescription: Trigger: " + name + "\n---\n\n## Cuándo\n\ncuando\n"),
		})
	}
	return skills
}

// fakeFiles is the Files port in memory. It records writes and backups apart
// from reads, because "no escribió nada" is the assertion dry-run exists for.
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
		return "", fmt.Errorf("no se puede respaldar %s: no existe", path)
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

// dest is where one agent's skills live, joined the way the app joins it.
func dest(home, name string) string {
	a, err := agent.Resolve(name)
	if err != nil {
		panic(err)
	}
	return filepath.Join(home, filepath.FromSlash(a.Dest))
}

// target is one installed SKILL.md.
func target(home, agentName, skill string) string {
	return filepath.Join(dest(home, agentName), skill, "SKILL.md")
}

func run(files *fakeFiles, h fakeHome, opts pack.Options) pack.Result {
	if opts.Skills == nil {
		opts.Skills = bundled()
	}
	return pack.Run(opts, files, h)
}

// TestInstallsTheFiveSkillsForOneAgent: one agent, five files, the exact bytes,
// and the one-line summary Contract VIII fixes.
func TestInstallsTheFiveSkillsForOneAgent(t *testing.T) {
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
	if !slices.Contains(result.Lines, "claude · 5 instaladas · ~/.claude/skills") {
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
	if !slices.Contains(second.Lines, "claude · 5 sin cambios · ~/.claude/skills") {
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
	if !slices.Contains(result.Lines, "claude · 4 instaladas · 1 respaldada · ~/.claude/skills") {
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
	if !slices.Contains(result.Lines, "claude · 4 por instalar · 1 por respaldar · ~/.claude/skills") {
		t.Errorf("the dry-run summary is missing or wrong, lines: %v", result.Lines)
	}
	if !strings.Contains(result.Message, "dry-run") {
		t.Errorf("the message does not say it was a dry run: %q", result.Message)
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
	if !slices.Contains(result.Lines, "claude · 5 instaladas · ~/.claude/skills") {
		t.Errorf("the healthy agent has no summary, lines: %v", result.Lines)
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
	if !strings.Contains(result.Message, "Pi no usa un directorio de skills") {
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
		if strings.Contains(line, "~/.claude/skills/mia/SKILL.md") && strings.Contains(line, "no lo toco") {
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
	if len(result.Agents) != 1 || result.Agents[0].Name != "codex" {
		t.Fatalf("agents = %+v, want only codex", result.Agents)
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
// code, decides where each agent is served.
func TestEverySupportedAgentResolvesItsOwnDestination(t *testing.T) {
	root := home()
	seed := map[string]string{}
	for _, a := range agent.Supported() {
		seed[filepath.Join(root, filepath.FromSlash(a.Marker()), "config")] = "x"
	}
	files := newFakeFiles(seed)

	result := run(files, fakeHome{dir: root}, pack.Options{})

	if result.Code != pack.CodeOK {
		t.Fatalf("code = %d, want 0 (message %q)", result.Code, result.Message)
	}
	if len(result.Agents) != len(agent.Supported()) {
		t.Fatalf("served %d agents, want %d: %+v", len(result.Agents), len(agent.Supported()), result.Agents)
	}
	for _, a := range agent.Supported() {
		if !slices.Contains(result.Lines, a.Name+" · 5 instaladas · "+a.Display()) {
			t.Errorf("no summary for %s, lines: %v", a.Name, result.Lines)
		}
		for _, skill := range bundled() {
			if _, ok := files.content(filepath.Join(root, filepath.FromSlash(a.Dest), skill.Name, "SKILL.md")); !ok {
				t.Errorf("%s: %s was not installed", a.Name, skill.Name)
			}
		}
	}
}

// TestUnresolvableHomeIsBlocked: no home, no writes, one line that says what to
// do about it.
func TestUnresolvableHomeIsBlocked(t *testing.T) {
	files := newFakeFiles(nil)
	result := pack.Run(pack.Options{Agents: "claude", Skills: bundled()}, files, fakeHome{err: errors.New("no hay home")})

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
