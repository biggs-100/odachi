package init_test

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	initapp "github.com/biggs-100/odachi/internal/app/init"
	domainopencode "github.com/biggs-100/odachi/internal/domain/opencode"
	"github.com/biggs-100/odachi/internal/domain/persona"
)

// homeFile is where a file in the fake home lives.
func homeFile(home, rel string) string {
	return filepath.Join(home, filepath.FromSlash(rel))
}

// TestSetupLeavesTheFolderAloneAndTheHomeReady: the installer runs setup from
// anywhere. The global half arrives; the folder it ran in gains nothing.
func TestSetupLeavesTheFolderAloneAndTheHomeReady(t *testing.T) {
	home := "hogar"
	files := newFakeFiles(nil)
	// OpenCode is on this machine: its config directory exists.
	files.files[homeFile(home, ".config/opencode/opencode.json")] = []byte("{}")
	skills := &fakeSkills{}

	got := initapp.RunSetup(initapp.Options{Home: fakeHome{dir: home}}, files, skills)

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	if !strings.Contains(got.Message, "ready") {
		t.Errorf("message = %q, want the human told the machine is ready", got.Message)
	}
	if _, ok := files.files[homeFile(home, ".config/opencode/agents/odachi.md")]; !ok {
		t.Error("the home door was not installed")
	}
	config := files.files[homeFile(home, ".config/opencode/opencode.json")]
	var doc map[string]any
	if err := json.Unmarshal(domainopencode.StripComments(config), &doc); err != nil {
		t.Fatalf("the config is not valid: %v\n%s", err, config)
	}
	if doc["default_agent"] != persona.AgentName {
		t.Errorf("default_agent = %v, want %q", doc["default_agent"], persona.AgentName)
	}
	if skills.calls != 1 {
		t.Errorf("skills installed %d times, want 1", skills.calls)
	}

	for path := range files.files {
		if !strings.HasPrefix(filepath.Clean(path), filepath.Clean(home)) {
			t.Errorf("setup wrote outside the home: %s", path)
		}
	}
}

// TestSetupWithNoClientInstallsTheSkillAndSaysSo: a machine without OpenCode or
// Claude gets the knowledge and an honest line, not a folder nobody uses.
func TestSetupWithNoClientInstallsTheSkillAndSaysSo(t *testing.T) {
	files := newFakeFiles(nil)
	skills := &fakeSkills{}

	got := initapp.RunSetup(initapp.Options{Home: fakeHome{dir: "hogar"}}, files, skills)

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	if skills.calls != 1 {
		t.Errorf("skills installed %d times, want 1", skills.calls)
	}
	if joined := strings.Join(got.Lines, "\n"); !strings.Contains(joined, "no agent client found") {
		t.Errorf("the lines do not say no client was found:\n%s", joined)
	}
	for path := range files.files {
		if strings.Contains(path, "agents") {
			t.Errorf("a door was installed for a client that is not here: %s", path)
		}
	}
}

// TestSetupHonoursAnExplicitClient: --agents is how the human says "install this
// one even if the folder is not there yet".
func TestSetupHonoursAnExplicitClient(t *testing.T) {
	home := "hogar"
	files := newFakeFiles(nil)

	got := initapp.RunSetup(initapp.Options{Home: fakeHome{dir: home}, Agents: persona.SelectionOpenCode}, files, &fakeSkills{})

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	if _, ok := files.files[homeFile(home, ".config/opencode/agents/odachi.md")]; !ok {
		t.Error("the explicit client did not get its door")
	}
	if _, ok := files.files[homeFile(home, ".config/opencode/opencode.json")]; !ok {
		t.Error("the explicit client did not get its default agent")
	}
}

// TestSetupIsIdempotent: the installer may run it twice. The second run says
// nothing changed and rewrites nothing.
func TestSetupIsIdempotent(t *testing.T) {
	home := "hogar"
	files := newFakeFiles(nil)
	files.files[homeFile(home, ".config/opencode/opencode.json")] = []byte("{}")
	opts := initapp.Options{Home: fakeHome{dir: home}}

	if got := initapp.RunSetup(opts, files, &fakeSkills{}); got.Code != 0 {
		t.Fatalf("first setup: Code = %d", got.Code)
	}
	door := string(files.files[homeFile(home, ".config/opencode/agents/odachi.md")])
	config := string(files.files[homeFile(home, ".config/opencode/opencode.json")])

	second := initapp.RunSetup(opts, files, &fakeSkills{})

	if string(files.files[homeFile(home, ".config/opencode/agents/odachi.md")]) != door {
		t.Error("the second setup rewrote the door")
	}
	if string(files.files[homeFile(home, ".config/opencode/opencode.json")]) != config {
		t.Error("the second setup rewrote the config")
	}
	if joined := strings.Join(second.Lines, "\n"); !strings.Contains(joined, "already the default") {
		t.Errorf("the second setup does not say the default stands:\n%s", joined)
	}
}

// TestSetupDoesNotCallAnUnreadableClientAbsent: a home whose client folders cannot
// be read is a run that does not know, and "not installed here" plus "install
// OpenCode" is a claim it cannot make. The error is named; the advice is not given.
func TestSetupDoesNotCallAnUnreadableClientAbsent(t *testing.T) {
	home := "hogar"
	files := newFakeFiles(nil)
	// Both clients are here, and neither folder can be read: zero clients found,
	// which is exactly the state the old run mistook for an empty machine.
	files.existsErr = map[string]error{
		homeFile(home, ".claude"):          errors.New("permission denied"),
		homeFile(home, ".config/opencode"): errors.New("permission denied"),
	}

	got := initapp.RunSetup(initapp.Options{Home: fakeHome{dir: home}}, files, &fakeSkills{})

	lines := strings.Join(got.Lines, "\n")
	if !strings.Contains(lines, "could not check") || !strings.Contains(lines, "permission denied") {
		t.Errorf("the run does not name the read that failed:\n%s", lines)
	}
	if strings.Contains(lines, "not installed here") {
		t.Errorf("a folder it could not read is reported as absent:\n%s", lines)
	}
	if strings.Contains(lines, "no agent client found") {
		t.Errorf("the run tells the human to install a client it could not check:\n%s", lines)
	}
}

// TestASetDefaultAgentIsAChangedFile: setting default_agent rewrites the config,
// and a run that omits the file from its changed list disagrees with the disk it
// just wrote. The 0121 review found this branch unpinned.
func TestASetDefaultAgentIsAChangedFile(t *testing.T) {
	home := "hogar"
	files := newFakeFiles(nil)
	files.files[homeFile(home, ".config/opencode/opencode.json")] = []byte("{}")

	got := initapp.RunSetup(initapp.Options{Home: fakeHome{dir: home}}, files, &fakeSkills{})

	want := "~/.config/opencode/opencode.json"
	found := false
	for _, file := range got.Files {
		if file == want {
			found = true
		}
	}
	if !found {
		t.Errorf("Files = %v, want %q among them", got.Files, want)
	}
	if strings.Contains(got.Message, "unchanged") {
		t.Errorf("message = %q, want a run that wrote a file", got.Message)
	}
}
