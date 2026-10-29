package init_test

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	initapp "github.com/biggs-100/odachi/internal/app/init"
)

const dir = "/proyecto"

// fakeFiles is the Files port in memory. Every call is recorded in order so a
// test can assert backup-before-write, file modes and, above all, no write at
// all on a rerun.
type fakeFiles struct {
	files    map[string][]byte
	modes    map[string]string
	events   []string
	writeErr error
	modeErr  error
}

func newFakeFiles(seed map[string]string) *fakeFiles {
	f := &fakeFiles{files: map[string][]byte{}, modes: map[string]string{}}
	for path, content := range seed {
		f.put(path, content)
	}
	return f
}

func (f *fakeFiles) put(path, content string) {
	f.files[filepath.Join(dir, filepath.FromSlash(path))] = []byte(content)
}

func (f *fakeFiles) get(path string) (string, bool) {
	data, ok := f.files[filepath.Join(dir, filepath.FromSlash(path))]
	return string(data), ok
}

func (f *fakeFiles) Exists(path string) (bool, error) {
	if _, ok := f.files[path]; ok {
		return true, nil
	}
	// The real adapter sees directories; the double only stores files, so a
	// directory exists when something lives under it.
	prefix := path + string(filepath.Separator)
	for file := range f.files {
		if strings.HasPrefix(file, prefix) {
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeFiles) Read(path string) ([]byte, error) {
	data, ok := f.files[path]
	if !ok {
		return nil, fmt.Errorf("missing file: %s", path)
	}
	return data, nil
}

func (f *fakeFiles) Write(path string, data []byte) error {
	if f.writeErr != nil {
		return f.writeErr
	}
	f.events = append(f.events, "write "+slashed(dir, path))
	f.files[path] = data
	return nil
}

func (f *fakeFiles) Backup(path string) (string, error) {
	f.events = append(f.events, "backup "+slashed(dir, path))
	return path + ".bak-test", nil
}

func (f *fakeFiles) MakeExecutable(path string) error {
	if f.modeErr != nil {
		return f.modeErr
	}
	f.events = append(f.events, "chmod "+slashed(dir, path))
	f.modes[path] = "0755"
	return nil
}

func slashed(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}

// fakeSkills is the Skills port in memory: the half of the job `odachi pack`
// owns. init decides that two destinations get skills and phrases the report;
// this double is what a test hands back instead of touching a home.
type fakeSkills struct {
	report initapp.SkillReport
	calls  int
}

func (f *fakeSkills) Install() initapp.SkillReport {
	f.calls++
	return f.report
}

// run performs one `odachi init` with a skills half that has nothing to say.
func run(files *fakeFiles) initapp.Result {
	return runWith(files, &fakeSkills{})
}

// runWith performs one `odachi init` with the skills half the test supplies.
func runWith(files *fakeFiles, skills *fakeSkills) initapp.Result {
	return initapp.Run(initapp.Options{Dir: dir}, files, skills)
}

// A project that has never been initialised gets both artifacts, and the block
// carries the four hard rules and the six verbs verbatim.
func TestCreatesAgentsMDWithTheBlock(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	got := run(files)

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}

	content, ok := files.get("AGENTS.md")
	if !ok {
		t.Fatalf("AGENTS.md was not written; events: %v", files.events)
	}
	if !strings.HasPrefix(content, "<!-- odachi:begin -->") {
		t.Errorf("AGENTS.md starts with %q, want the begin marker", firstLine(content))
	}
	if !strings.HasSuffix(strings.TrimSpace(content), "<!-- odachi:end -->") {
		t.Errorf("AGENTS.md ends with %q, want the end marker", lastLine(content))
	}
	if strings.Count(content, "<!-- odachi:begin -->") != 1 || strings.Count(content, "<!-- odachi:end -->") != 1 {
		t.Errorf("the block appears more than once:\n%s", content)
	}
	if line := "Skills: the ones in `.agents/skills/`; inside this project, load `drive` before acting: it is the manual."; strings.Count(content, line) != 1 {
		t.Errorf("the line %q appears %d times, want exactly one:\n%s", line, strings.Count(content, line), content)
	}

	for _, want := range []string{
		"Brief before code",
		"Nothing outside the fence",
		"odachi ask",
		`odachi check --run`,
		"one commit",
		"Skills: the ones in `.agents/skills/`; inside this project, load `drive` before acting: it is the manual.",
		"odachi start",
		"odachi level",
		"odachi summary",
		"odachi check",
		"odachi init",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("AGENTS.md does not contain %q:\n%s", want, content)
		}
	}

	if !slices.Contains(got.Files, "AGENTS.md") {
		t.Errorf("Files = %v, want it to list AGENTS.md", got.Files)
	}
	if !slices.Contains(files.events, "write AGENTS.md") {
		t.Errorf("events = %v, want a write of AGENTS.md", files.events)
	}
}

// Rerunning changes nothing: same bytes, no backup, no write, no chmod.
func TestRerunIsIdempotent(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	if code := run(files).Code; code != 0 {
		t.Fatalf("first run code = %d", code)
	}
	first, _ := files.get("AGENTS.md")
	hookFirst, _ := files.get(".git/hooks/pre-commit")

	files.events = nil
	got := run(files)

	if got.Code != 0 {
		t.Fatalf("second run code = %d (message %q)", got.Code, got.Message)
	}
	if len(files.events) != 0 {
		t.Errorf("second run touched %v, want no writes, backups or chmods", files.events)
	}
	if second, _ := files.get("AGENTS.md"); second != first {
		t.Errorf("AGENTS.md changed on a rerun:\n%q\nvs\n%q", second, first)
	}
	if second, _ := files.get(".git/hooks/pre-commit"); second != hookFirst {
		t.Errorf("the hook changed on a rerun:\n%q\nvs\n%q", second, hookFirst)
	}
	if !strings.Contains(got.Message, "unchanged") {
		t.Errorf("Message = %q, want it to say nothing changed", got.Message)
	}
}

// The block is replaced in place: everything the human wrote around it stays.
func TestReplacesTheBlockInPlaceAndPreservesTheRest(t *testing.T) {
	files := newFakeFiles(map[string]string{
		".git/HEAD": "ref: refs/heads/main\n",
		"AGENTS.md": "# Mis reglas\n\nno toques production\n\n<!-- odachi:begin -->\nregla vieja\n<!-- odachi:end -->\n\ny esto va después\n",
	})

	got := run(files)

	if got.Code != 0 {
		t.Fatalf("Code = %d (message %q)", got.Code, got.Message)
	}
	content, _ := files.get("AGENTS.md")

	if strings.Contains(content, "regla vieja") {
		t.Errorf("the old block survived:\n%s", content)
	}
	for _, want := range []string{"# Mis reglas", "no toques production", "y esto va después", "Brief before code"} {
		if !strings.Contains(content, want) {
			t.Errorf("AGENTS.md lost %q:\n%s", want, content)
		}
	}
	if !strings.HasPrefix(content, "# Mis reglas") {
		t.Errorf("the block moved to the top:\n%s", content)
	}
	if !slices.Contains(files.events, "backup AGENTS.md") {
		t.Errorf("events = %v, want a backup before the rewrite", files.events)
	}
	if indexOf(files.events, "backup AGENTS.md") > indexOf(files.events, "write AGENTS.md") {
		t.Errorf("events = %v, want the backup before the write", files.events)
	}
}

// A file with no markers keeps every line and gets the block at the end.
func TestAppendsTheBlockWhenThereAreNoMarkers(t *testing.T) {
	files := newFakeFiles(map[string]string{
		".git/HEAD": "ref: refs/heads/main\n",
		"AGENTS.md": "# Mis reglas\n\nusá tabs\n",
	})

	if code := run(files).Code; code != 0 {
		t.Fatalf("code = %d", code)
	}
	content, _ := files.get("AGENTS.md")

	if !strings.HasPrefix(content, "# Mis reglas\n\nusá tabs\n") {
		t.Errorf("the existing content changed:\n%q", content)
	}
	if strings.Count(content, "<!-- odachi:begin -->") != 1 {
		t.Errorf("the block was appended more than once:\n%s", content)
	}
	if !strings.Contains(content, "Brief before code") {
		t.Errorf("the block was not appended:\n%s", content)
	}
}

// The hook is installed with the shell it needs and the command the gate
// provides, and it is made executable.
func TestInstallsTheHook(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	got := run(files)

	if got.Code != 0 {
		t.Fatalf("Code = %d (message %q)", got.Code, got.Message)
	}
	hook, ok := files.get(".git/hooks/pre-commit")
	if !ok {
		t.Fatalf("the hook was not written; events: %v", files.events)
	}
	if !strings.HasPrefix(hook, "#!/bin/sh\n") {
		t.Errorf("hook starts with %q, want the shebang", firstLine(hook))
	}
	if !strings.Contains(hook, "# odachi") {
		t.Errorf("hook has no marker comment:\n%s", hook)
	}
	if !strings.Contains(hook, "check --quiet || exit 1") {
		t.Errorf("hook does not run the gate:\n%s", hook)
	}
	if !slices.Contains(files.events, "chmod .git/hooks/pre-commit") {
		t.Errorf("events = %v, want the hook made executable", files.events)
	}
	if !slices.Contains(got.Files, ".git/hooks/pre-commit") {
		t.Errorf("Files = %v, want it to list the hook", got.Files)
	}
}

// A hook someone else wrote is never clobbered: the human gets the three lines
// to paste and a non-zero exit naming the file.
func TestForeignHookIsNotClobbered(t *testing.T) {
	foreign := "#!/bin/sh\necho hello\n"
	files := newFakeFiles(map[string]string{
		".git/HEAD":             "ref: refs/heads/main\n",
		".git/hooks/pre-commit": foreign,
	})

	got := run(files)

	if got.Code != 1 {
		t.Fatalf("Code = %d, want 1 (message %q)", got.Code, got.Message)
	}
	if !strings.Contains(got.Message, ".git/hooks/pre-commit") {
		t.Errorf("Message = %q, want it to name the file", got.Message)
	}
	if after, _ := files.get(".git/hooks/pre-commit"); after != foreign {
		t.Errorf("the foreign hook was modified:\n%q", after)
	}
	if slices.Contains(files.events, "chmod .git/hooks/pre-commit") {
		t.Errorf("events = %v, want no chmod of a foreign hook", files.events)
	}

	printed := strings.Join(got.Lines, "\n")
	for _, want := range []string{"#!/bin/sh", "# odachi", "check --quiet || exit 1"} {
		if !strings.Contains(printed, want) {
			t.Errorf("Lines = %q, want the three lines to paste (missing %q)", printed, want)
		}
	}
}

// Our own hook is kept, not refused, and not rewritten when it is already right.
func TestOurHookIsNotRefused(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
	if code := run(files).Code; code != 0 {
		t.Fatalf("first run code = %d", code)
	}

	got := run(files)

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0: our own hook is not foreign (message %q)", got.Code, got.Message)
	}
	if slices.Contains(files.events, "backup .git/hooks/pre-commit") {
		t.Errorf("events = %v, want no backup of an unchanged hook", files.events)
	}
}

// A hook written before the English migration carries the old marker. It is
// still odachi's: init backs it up and rewrites it in place, instead of asking
// the human to paste the lines again.
func TestTheSpanishHookMarkerIsStillOurs(t *testing.T) {
	old := "#!/bin/sh\n# odachi: no se commitea sin prueba de este árbol\n" +
		"ODACHI='odachi'\n[ -x \"$ODACHI\" ] || ODACHI=odachi\n\"$ODACHI\" check --quiet || exit 1\n"
	files := newFakeFiles(map[string]string{
		".git/HEAD":             "ref: refs/heads/main\n",
		".git/hooks/pre-commit": old,
	})

	got := run(files)

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0: the old marker is still odachi's (message %q)", got.Code, got.Message)
	}
	after, _ := files.get(".git/hooks/pre-commit")
	if strings.Contains(after, "no se commitea") {
		t.Errorf("the hook kept the old marker:\n%q", after)
	}
	if !strings.Contains(after, "# odachi: no commit with the gate red") {
		t.Errorf("the hook was not rewritten in English:\n%q", after)
	}
	if !slices.Contains(files.events, "backup .git/hooks/pre-commit") {
		t.Errorf("events = %v, want a backup before rewriting the hook", files.events)
	}
}

// Without a repository there is nowhere to put the hook, and the tool says so
// instead of writing half of the job.
func TestNoGitDirectoryBlocksWithoutWriting(t *testing.T) {
	files := newFakeFiles(map[string]string{"README.md": "hello\n"})

	got := run(files)

	if got.Code != 1 {
		t.Fatalf("Code = %d, want 1 (message %q)", got.Code, got.Message)
	}
	if !strings.Contains(got.Message, "git init") {
		t.Errorf("Message = %q, want it to name git init", got.Message)
	}
	if len(files.events) != 0 {
		t.Errorf("events = %v, want no writes", files.events)
	}
}

func firstLine(text string) string {
	line, _, _ := strings.Cut(text, "\n")
	return line
}

func lastLine(text string) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	return lines[len(lines)-1]
}

func indexOf(lines []string, want string) int {
	return slices.Index(lines, want)
}

// The project is not ready until the agent has the skills it reads: init writes
// the two destinations as part of the same run, and says where each one went.
func TestInitInstallsTheSkillsAfterTheProjectArtifacts(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
	skills := &fakeSkills{report: initapp.SkillReport{
		Dests: []initapp.SkillDest{
			{Label: "./.agents/skills", Installed: 8, Files: []string{".agents/skills/acceptance/SKILL.md"}},
			{Label: "~/.agents/skills", Installed: 1, Files: []string{"~/.agents/skills/bootstrap/SKILL.md"}},
		},
	}}

	got := runWith(files, skills)

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (lines %q)", got.Code, got.Lines)
	}
	if skills.calls != 1 {
		t.Fatalf("Install() was called %d times, want 1", skills.calls)
	}

	project := indexOfLine(got.Lines, "./.agents/skills: 8 installed skills")
	home := indexOfLine(got.Lines, "~/.agents/skills: 1 installed skill")
	if project < 0 || home < 0 {
		t.Fatalf("the lines do not name both destinations: %q", got.Lines)
	}
	// The skills complete the project the three artifacts started, so they are
	// reported after them and before the closing line.
	if project < indexOfLine(got.Lines, "AGENTS.md: created with the odachi block") {
		t.Errorf("the skills are reported before the project artifacts: %q", got.Lines)
	}
	if home < project {
		t.Errorf("the home is reported before the project: %q", got.Lines)
	}
	if last := got.Lines[len(got.Lines)-1]; !strings.HasPrefix(last, "next: ") {
		t.Errorf("the last line does not name the next step: %q", last)
	}
	for _, want := range []string{".agents/skills/acceptance/SKILL.md", "~/.agents/skills/bootstrap/SKILL.md"} {
		if !slices.Contains(got.Files, want) {
			t.Errorf("Files = %v, want it to list %s", got.Files, want)
		}
	}
}

// A second run writes nothing and says so: the skills half reports the same
// `unchanged` the three artifacts report, in the one message a human reads.
func TestInitRerunSaysTheSkillsWereAlreadyThere(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
	full := &fakeSkills{report: initapp.SkillReport{
		Dests: []initapp.SkillDest{
			{Label: "./.agents/skills", Installed: 8, Files: []string{".agents/skills/acceptance/SKILL.md"}},
			{Label: "~/.agents/skills", Installed: 1, Files: []string{"~/.agents/skills/bootstrap/SKILL.md"}},
		},
	}}
	if code := runWith(files, full).Code; code != 0 {
		t.Fatalf("first run code = %d", code)
	}

	files.events = nil
	quiet := &fakeSkills{report: initapp.SkillReport{
		Dests: []initapp.SkillDest{
			{Label: "./.agents/skills", Unchanged: 8},
			{Label: "~/.agents/skills", Unchanged: 1},
		},
	}}
	got := runWith(files, quiet)

	if got.Code != 0 {
		t.Fatalf("second run code = %d (message %q)", got.Code, got.Message)
	}
	if len(files.events) != 0 {
		t.Errorf("second run touched %v, want no writes", files.events)
	}
	if !strings.Contains(got.Message, "unchanged") || !strings.Contains(got.Message, "./.agents/skills") {
		t.Errorf("Message = %q, want it to say nothing changed and where", got.Message)
	}
	if len(got.Files) != 0 {
		t.Errorf("Files = %v, want none on a rerun that wrote nothing", got.Files)
	}
	if !slices.Contains(got.Lines, "./.agents/skills: 8 unchanged") {
		t.Errorf("Lines = %q, want the destination reported as unchanged", got.Lines)
	}
}

// A home that cannot be written does not fail the run that fixed the project:
// the human reads which half is pending and the command that fixes it, and the
// artifacts stay written.
func TestInitReportsABlockedHomeAndKeepsTheProject(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
	skills := &fakeSkills{report: initapp.SkillReport{
		Dests:   []initapp.SkillDest{{Label: "./.agents/skills", Installed: 8, Files: []string{".agents/skills/acceptance/SKILL.md"}}},
		Blocked: "~/.agents/skills: could not write the bootstrap skill: permission denied",
		Next:    "odachi init",
	}}

	got := runWith(files, skills)

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0: the project half finished (lines %q)", got.Code, got.Lines)
	}
	if _, ok := files.get("AGENTS.md"); !ok {
		t.Errorf("AGENTS.md was not written; events: %v", files.events)
	}
	joined := strings.Join(got.Lines, "\n")
	if !strings.Contains(joined, "~/.agents/skills: could not write the bootstrap skill") {
		t.Errorf("the lines do not say what is pending: %q", got.Lines)
	}
	if got.Next != "odachi init" {
		t.Errorf("Next = %q, want the command that fixes the block", got.Next)
	}
}

func indexOfLine(lines []string, want string) int {
	return slices.Index(lines, want)
}

// A block inside the project is a failed run: the artifacts are written, but the
// agent would work without the skills it reads. The home's block is the only one
// that never fails the run.
func TestInitFailsWhenTheProjectSkillsCannotBeWritten(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
	skills := &fakeSkills{report: initapp.SkillReport{
		Blocked: "./.agents/skills: could not write ./.agents/skills/acceptance/SKILL.md: permission denied",
		Next:    "odachi init",
		Fatal:   true,
	}}

	got := runWith(files, skills)

	if got.Code != 1 {
		t.Fatalf("Code = %d, want 1: the project was left without the skills the agent reads (lines %q)", got.Code, got.Lines)
	}
	if got.Next != "odachi init" {
		t.Errorf("Next = %q, want the command that retries it", got.Next)
	}
	if _, ok := files.get("AGENTS.md"); !ok {
		t.Errorf("AGENTS.md was not written; events: %v", files.events)
	}
}
