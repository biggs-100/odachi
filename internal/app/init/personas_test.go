package init_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	initapp "github.com/biggs-100/odachi/internal/app/init"
	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/rules"
)

// fakeProfiles is the profile port: the profile the human wrote down, or none.
// init must not infer one, so the double has no way to say "inferred".
type fakeProfiles struct {
	name string
	err  error
}

func (p fakeProfiles) Written(string) (string, error) { return p.name, p.err }

// runProfiled performs one `odachi init` with a profile already written down.
func runProfiled(files *fakeFiles, profiles fakeProfiles) initapp.Result {
	return initapp.Run(initapp.Options{Dir: dir, Profiles: profiles}, files, &fakeSkills{})
}

// TestInitWritesTheNativeAgent: one file per client, the same block the agent
// reads in AGENTS.md, and the report says how the human picks it. The old
// v0.10.0 files are not rewritten under their old names.
func TestInitWritesTheNativeAgent(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	got := run(files)

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}

	block, err := rules.Body(persona.Novato)
	if err != nil {
		t.Fatalf("rules.Body(novato): %v", err)
	}
	for _, client := range persona.Clients() {
		file, ok := persona.FilesFor(client.Agent, block)
		if !ok {
			t.Fatalf("persona.FilesFor(%q) does not exist", client.Agent)
		}
		content, ok := files.get(file.Path)
		if !ok {
			t.Errorf("%s was not written", file.Path)
			continue
		}
		if content != file.Content {
			t.Errorf("%s is not the mirror of the block:\n%q\nvs\n%q", file.Path, content, file.Content)
		}
		if !strings.Contains(strings.Join(got.Lines, "\n"), client.Dir) {
			t.Errorf("the lines do not name %s:\n%s", client.Dir, strings.Join(got.Lines, "\n"))
		}
	}
	for _, file := range persona.Deprecated() {
		if _, ok := files.get(file); ok {
			t.Errorf("%s was written again: the agent of this version is %s", file, persona.FileName)
		}
	}
}

// TestInitBacksUpAHandEditedAgent: the file is generated, so a human edit is
// lost work. init keeps it in a .bak before replacing it and says it wrote.
func TestInitBacksUpAHandEditedAgent(t *testing.T) {
	path := ".opencode/agents/odachi.md"
	files := newFakeFiles(map[string]string{
		".git/HEAD": "ref: refs/heads/main\n",
		path:        "editado a mano\n",
	})

	got := run(files)

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	backup, wrote := -1, -1
	for i, event := range files.events {
		switch {
		case event == "backup "+path:
			backup = i
		case event == "write "+path:
			wrote = i
		}
	}
	if backup < 0 || wrote < 0 || backup > wrote {
		t.Fatalf("the backup must come before the write: %v", files.events)
	}
	if content, _ := files.get(path); content == "editado a mano\n" {
		t.Errorf("the agent kept the hand edit instead of the mirror")
	}
	if !strings.Contains(strings.Join(got.Lines, "\n"), persona.FileName) {
		t.Errorf("the lines do not say which agent was written:\n%s", strings.Join(got.Lines, "\n"))
	}
}

// fakeHome is the Home port: where the human lives, or the reason there is no
// answer. init must not reach os to find it.
type fakeHome struct {
	dir string
	err error
}

func (h fakeHome) Home() (string, error) { return h.dir, h.err }

// TestInitWithoutGlobalLeavesTheHomeAlone: the door in the home is opt-in. A
// plain init does not write the home, not even a backup, and one line names the
// flag that installs it: an agent primary in every folder is the human's
// decision, not a side effect of initializing a project.
func TestInitWithoutGlobalLeavesTheHomeAlone(t *testing.T) {
	home := "hogar"
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	got := initapp.Run(initapp.Options{Dir: dir, Home: fakeHome{dir: home}}, files, &fakeSkills{})

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	for _, event := range files.events {
		if strings.Contains(event, home) {
			t.Errorf("init without --global touched the home: %s", event)
		}
	}
	for path := range files.files {
		if strings.Contains(path, home) {
			t.Errorf("init without --global left a file in the home: %s", path)
		}
	}
	if joined := strings.Join(got.Lines, "\n"); !strings.Contains(joined, "--global") {
		t.Errorf("the lines do not name the flag that installs the door:\n%s", joined)
	}
}

// TestInitGlobalInstallsTheGlobalDoor: with --global, one command leaves odachi
// available in every folder, not only in the project. The door is generated, so
// a second run keeps it and says so.
func TestInitGlobalInstallsTheGlobalDoor(t *testing.T) {
	home := "hogar"
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
	runHome := func() initapp.Result {
		return initapp.Run(initapp.Options{Dir: dir, Home: fakeHome{dir: home}, Global: true}, files, &fakeSkills{})
	}

	got := runHome()

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	for _, file := range persona.GlobalFiles() {
		abs := filepath.Join(home, filepath.FromSlash(file.Path))
		content, ok := files.files[abs]
		if !ok {
			t.Errorf("%s was not installed in the home", file.Path)
			continue
		}
		if string(content) != file.Content {
			t.Errorf("%s is not the door:\n%q\nvs\n%q", file.Path, content, file.Content)
		}
	}
	if !strings.Contains(strings.Join(got.Lines, "\n"), "~/"+persona.Clients()[0].HomeDir) {
		t.Errorf("the lines do not name the home destination:\n%s", strings.Join(got.Lines, "\n"))
	}

	files.events = nil
	again := runHome()
	if len(files.events) != 0 {
		t.Errorf("the second run touched the home: %v", files.events)
	}
	if !strings.Contains(strings.Join(again.Lines, "\n"), "unchanged") {
		t.Errorf("the second run does not say the door was already there:\n%s", strings.Join(again.Lines, "\n"))
	}
}

// TestInitKeepsTheProjectWhenTheHomeIsNotWritable: the project half is done and
// the home half is pending. Exit 0, because the project is ready; the line names
// the command that retries it.
func TestInitKeepsTheProjectWhenTheHomeIsNotWritable(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	got := initapp.Run(initapp.Options{Dir: dir, Home: fakeHome{err: errors.New("no home")}, Global: true}, files, &fakeSkills{})

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	for _, path := range []string{"AGENTS.md", ".opencode/agents/odachi.md", ".claude/output-styles/odachi.md"} {
		if _, ok := files.get(path); !ok {
			t.Errorf("%s was not written: the project was left half done", path)
		}
	}
	joined := strings.Join(got.Lines, "\n")
	for _, want := range []string{"pending:", "home", "odachi init"} {
		if !strings.Contains(joined, want) {
			t.Errorf("the lines do not say %q:\n%s", want, joined)
		}
	}
}

// TestInitWritesTheVoiceOfTheProfileAtHand: the block carries the voice of the
// profile the human accepted, and novato when there is no file yet — init runs
// before the interview and cannot know the person.
func TestInitWritesTheVoiceOfTheProfileAtHand(t *testing.T) {
	cases := []struct {
		name     string
		profiles fakeProfiles
		want     string
	}{
		{"no profile is novato", fakeProfiles{}, persona.Novato},
		{"novato written", fakeProfiles{name: "novato"}, persona.Novato},
		{"senior written", fakeProfiles{name: "senior"}, persona.Senior},
		{"ci uses the senior voice", fakeProfiles{name: "ci"}, persona.Senior},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

			got := runProfiled(files, c.profiles)

			if got.Code != 0 {
				t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
			}
			want, err := rules.Block(c.want)
			if err != nil {
				t.Fatalf("rules.Block(%q): %v", c.want, err)
			}
			content, _ := files.get("AGENTS.md")
			if content != want {
				t.Fatalf("the block does not carry the %s voice:\n%q\nvs\n%q", c.want, content, want)
			}
			if !strings.Contains(strings.Join(got.Lines, "\n"), c.want) {
				t.Errorf("the lines do not say which voice the block was written with:\n%s", strings.Join(got.Lines, "\n"))
			}
		})
	}
}

// A profile that cannot be read is not a failed install: the hook and the skills
// are what the project needs, and the voice falls back to novato with the file
// named and the action that fixes it.
func TestInitNamesABrokenProfileAndKeepsGoing(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	got := runProfiled(files, fakeProfiles{err: errors.New("profile: invalid TOML")})

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	want, err := rules.Block(persona.Novato)
	if err != nil {
		t.Fatalf("rules.Block(novato): %v", err)
	}
	if content, _ := files.get("AGENTS.md"); content != want {
		t.Errorf("the block did not come out with the novato voice:\n%q", content)
	}
	joined := strings.Join(got.Lines, "\n")
	for _, want := range []string{"odd/profile.toml", "odachi init"} {
		if !strings.Contains(joined, want) {
			t.Errorf("the lines do not name %q:\n%s", want, joined)
		}
	}
}

// A profile name odachi has no voice for is the same case: fall back, say it, and
// name the three names that have a voice.
func TestInitNamesAProfileWithoutAVoice(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	got := runProfiled(files, fakeProfiles{name: "mio"})

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	joined := strings.Join(got.Lines, "\n")
	for _, want := range []string{"mio", "novato", "senior", "ci"} {
		if !strings.Contains(joined, want) {
			t.Errorf("the lines do not name %q:\n%s", want, joined)
		}
	}
}

// TestInitReportsTheV0100PersonaFilesAndTouchesNothing: the human's disk is not
// odachi's to clean. One line says what those files are and that they can go.
func TestInitReportsTheV0100PersonaFilesAndTouchesNothing(t *testing.T) {
	leftover := ".opencode/agents/odachi-senior.md"
	files := newFakeFiles(map[string]string{
		".git/HEAD": "ref: refs/heads/main\n",
		leftover:    "la versión vieja\n",
	})

	got := run(files)

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	joined := strings.Join(got.Lines, "\n")
	for _, want := range []string{"v0.10.0", ".opencode/agents", rules.File, "delete them"} {
		if !strings.Contains(joined, want) {
			t.Errorf("the lines do not say %q:\n%s", want, joined)
		}
	}
	if content, _ := files.get(leftover); content != "la versión vieja\n" {
		t.Errorf("init touched the old file: %q", content)
	}
	for _, event := range files.events {
		if strings.Contains(event, leftover) {
			t.Errorf("init wrote or backed up the old file: %s", event)
		}
	}
}

// TestInitRerunStillWritesNothing: the block is generated, so a second run leaves
// no backup behind and says it changed nothing.
func TestInitRerunStillWritesNothing(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
	if code := run(files).Code; code != 0 {
		t.Fatalf("first run code = %d", code)
	}

	files.events = nil
	got := run(files)

	if got.Code != 0 {
		t.Fatalf("second run code = %d (message %q)", got.Code, got.Message)
	}
	if len(files.events) != 0 {
		t.Errorf("second run touched %v, want no writes and no backups", files.events)
	}
	if len(got.Files) != 0 {
		t.Errorf("Files = %v, want none on a rerun that wrote nothing", got.Files)
	}
	if !strings.Contains(got.Message, "unchanged") {
		t.Errorf("Message = %q, want it to say nothing changed", got.Message)
	}
}

// TestInitGlobalInstallsTheGlobalSubagents: the door is global, so the agents it
// invokes are too: a project that never writes .opencode keeps the worker, the
// verifier and the reviewer.
func TestInitGlobalInstallsTheGlobalSubagents(t *testing.T) {
	home := "hogar"
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	got := initapp.Run(initapp.Options{Dir: dir, Home: fakeHome{dir: home}, Global: true}, files, &fakeSkills{})

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	for _, client := range persona.Clients() {
		for _, file := range persona.GlobalSubagentFilesFor(client.Agent) {
			abs := filepath.Join(home, filepath.FromSlash(file.Path))
			content, ok := files.files[abs]
			if !ok {
				t.Errorf("%s was not installed in the home", file.Path)
				continue
			}
			if string(content) != file.Content {
				t.Errorf("%s is not the subagent:\n%q\nvs\n%q", file.Path, content, file.Content)
			}
		}
	}
	if !strings.Contains(strings.Join(got.Lines, "\n"), persona.SubagentNames()[0]) {
		t.Errorf("the report does not name the subagents:\n%s", strings.Join(got.Lines, "\n"))
	}
}

// TestInitNoProjectAgentsWritesTheHomeOnly: the clean mode. The project keeps
// the rules and the plan; the door and its subagents live in the home.
func TestInitNoProjectAgentsWritesTheHomeOnly(t *testing.T) {
	home := "hogar"
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	got := initapp.Run(initapp.Options{Dir: dir, Home: fakeHome{dir: home}, Global: true, NoProjectAgents: true}, files, &fakeSkills{})

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	body, err := rules.Body(persona.Novato)
	if err != nil {
		t.Fatalf("rules.Body(novato): %v", err)
	}
	for _, file := range persona.Files(body) {
		if _, ok := files.files[filepath.Join(dir, filepath.FromSlash(file.Path))]; ok {
			t.Errorf("%s was written in the project with --no-project-agents", file.Path)
		}
	}
	for _, client := range persona.Clients() {
		for _, file := range persona.SubagentFilesFor(client.Agent) {
			if _, ok := files.files[filepath.Join(dir, filepath.FromSlash(file.Path))]; ok {
				t.Errorf("%s was written in the project with --no-project-agents", file.Path)
			}
		}
		for _, file := range persona.GlobalSubagentFilesFor(client.Agent) {
			if _, ok := files.files[filepath.Join(home, filepath.FromSlash(file.Path))]; !ok {
				t.Errorf("%s was not written in the home", file.Path)
			}
		}
	}
	if !strings.Contains(strings.Join(got.Lines, "\n"), "not written") {
		t.Errorf("the report does not say the project agents were skipped:\n%s", strings.Join(got.Lines, "\n"))
	}
}

// TestInitNoProjectAgentsNeedsGlobal: skipping both halves would leave no agent
// at all, so the flag names what it needs.
func TestInitNoProjectAgentsNeedsGlobal(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	got := initapp.Run(initapp.Options{Dir: dir, Home: fakeHome{dir: "hogar"}, NoProjectAgents: true}, files, &fakeSkills{})

	if got.Code == 0 {
		t.Fatal("init accepted --no-project-agents without --global")
	}
	if !strings.Contains(got.Message, "--global") {
		t.Errorf("the refusal does not name --global: %q", got.Message)
	}
}

// TestTheGlobalSubagentLineHasNoEmptyClause: the subagents are not selected with
// a key, so the line must not leave the switch clause empty.
func TestTheGlobalSubagentLineHasNoEmptyClause(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	got := initapp.Run(initapp.Options{Dir: dir, Home: fakeHome{dir: "hogar"}, Global: true}, files, &fakeSkills{})

	for _, line := range got.Lines {
		if strings.Contains(line, "; ;") || strings.HasSuffix(line, "; ") {
			t.Errorf("the line leaves an empty clause: %q", line)
		}
	}
}
