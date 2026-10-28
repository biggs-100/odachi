package init_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	initapp "github.com/biggs-100/odachi/internal/app/init"
	"github.com/biggs-100/odachi/internal/domain/persona"
)

// fakeAgentSelection is the selection port in memory: the stored choice, when
// there is one, and the canonical bytes the app writes. It never touches disk.
type fakeAgentSelection struct {
	stored string
	err    error
	// targetErr is the failure of the store lookup itself: Written can succeed
	// and the write still have nowhere to go.
	targetErr error
}

func (s *fakeAgentSelection) Written(string) (string, error) { return s.stored, s.err }

// Target is the file the fake choice lives in: inside the project, because this
// test is about the choice and not about where the store hangs.
func (s *fakeAgentSelection) Target(dir string) (string, string, error) {
	if s.targetErr != nil {
		return "", "", s.targetErr
	}
	if s.err != nil {
		return "", "", s.err
	}
	return filepath.Join(dir, "agents.toml"), "agents.toml", nil
}

func (s *fakeAgentSelection) Render(selection string) ([]byte, error) {
	return []byte("agents = \"" + selection + "\"\n"), nil
}

// runAgents performs one init with the selection and, when the test wants it, a
// home. An empty home keeps the global half out of the test. The door in the
// home is opt-in, so the tests that compare clients ask for it with Global: this
// half is about which client receives the door, not about when it is installed.
func runAgents(files *fakeFiles, requested string, store *fakeAgentSelection, home string) initapp.Result {
	opts := initapp.Options{Dir: dir, Agents: requested, AgentSelection: store, Global: true}
	if home != "" {
		opts.Home = fakeHome{dir: home}
	}
	return initapp.Run(opts, files, &fakeSkills{})
}

// TestInitAgentsOpenCodeOnlyWritesOpenCode: the flag filters the project mirror
// and the home door, and a Claude file that already exists is left alone. The
// choice is recorded so a later plain init keeps it.
func TestInitAgentsOpenCodeOnlyWritesOpenCode(t *testing.T) {
	home := "hogar"
	files := newFakeFiles(map[string]string{
		".git/HEAD":                       "ref: refs/heads/main\n",
		".claude/output-styles/odachi.md": "editado a mano\n",
		".claude/agents/odachi-worker.md": "editado a mano\n",
	})
	store := &fakeAgentSelection{}

	got := runAgents(files, persona.SelectionOpenCode, store, home)

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	for _, file := range persona.SubagentFilesFor(persona.SelectionOpenCode) {
		if _, ok := files.get(file.Path); !ok {
			t.Errorf("%s was not written", file.Path)
		}
	}
	if _, ok := files.get(".opencode/agents/odachi.md"); !ok {
		t.Error(".opencode/agents/odachi.md was not written")
	}
	if content, _ := files.get(".claude/output-styles/odachi.md"); content != "editado a mano\n" {
		t.Errorf("init touched the Claude agent: %q", content)
	}
	if content, _ := files.get(".claude/agents/odachi-worker.md"); content != "editado a mano\n" {
		t.Errorf("init touched the Claude subagent: %q", content)
	}
	for _, event := range files.events {
		if strings.Contains(event, ".claude/") {
			t.Errorf("init wrote or backed up a Claude file: %s", event)
		}
	}

	storeContent, ok := files.get(persona.SelectionPath)
	if !ok {
		t.Fatalf("%s was not written", persona.SelectionPath)
	}
	if !strings.Contains(storeContent, `agents = "opencode"`) {
		t.Errorf("%s = %q, want agents = \"opencode\"", persona.SelectionPath, storeContent)
	}

	door := filepath.Join(home, filepath.FromSlash(".config/opencode/agents/odachi.md"))
	if _, ok := files.files[door]; !ok {
		t.Errorf("the OpenCode door was not installed in the home")
	}
	claudeDoor := filepath.Join(home, filepath.FromSlash(".claude/output-styles/odachi.md"))
	if _, ok := files.files[claudeDoor]; ok {
		t.Errorf("the Claude door was installed although the choice was opencode")
	}
	if joined := strings.Join(got.Lines, "\n"); strings.Contains(joined, ".claude") {
		t.Errorf("the lines name .claude although it was not written:\n%s", joined)
	}
}

// TestInitAgentsClaudeOnlyWritesClaude is the same decision on the other side:
// the client that was not chosen is not written and its existing files stay.
func TestInitAgentsClaudeOnlyWritesClaude(t *testing.T) {
	home := "hogar"
	files := newFakeFiles(map[string]string{
		".git/HEAD":                  "ref: refs/heads/main\n",
		".opencode/agents/odachi.md": "editado a mano\n",
	})
	store := &fakeAgentSelection{}

	got := runAgents(files, persona.SelectionClaude, store, home)

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	if _, ok := files.get(".claude/output-styles/odachi.md"); !ok {
		t.Error(".claude/output-styles/odachi.md was not written")
	}
	for _, file := range persona.SubagentFilesFor(persona.SelectionClaude) {
		if _, ok := files.get(file.Path); !ok {
			t.Errorf("%s was not written", file.Path)
		}
	}
	if content, _ := files.get(".opencode/agents/odachi.md"); content != "editado a mano\n" {
		t.Errorf("init touched the OpenCode agent: %q", content)
	}
	for _, event := range files.events {
		if strings.Contains(event, ".opencode/") {
			t.Errorf("init wrote or backed up an OpenCode file: %s", event)
		}
	}
	if storeContent, _ := files.get(persona.SelectionPath); !strings.Contains(storeContent, `agents = "claude"`) {
		t.Errorf("%s = %q, want agents = \"claude\"", persona.SelectionPath, storeContent)
	}
}

// TestInitAgentsRerunKeepsTheStoredChoice: the flag is sticky. A plain init
// after `--agents opencode` must not bring Claude back.
func TestInitAgentsRerunKeepsTheStoredChoice(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
	store := &fakeAgentSelection{}

	if code := runAgents(files, persona.SelectionOpenCode, store, "").Code; code != 0 {
		t.Fatalf("first run code = %d", code)
	}
	store.stored = persona.SelectionOpenCode
	files.events = nil

	got := runAgents(files, "", store, "")

	if got.Code != 0 {
		t.Fatalf("second run code = %d (message %q)", got.Code, got.Message)
	}
	if len(files.events) != 0 {
		t.Errorf("the second run touched %v, want no writes", files.events)
	}
	if len(got.Files) != 0 {
		t.Errorf("Files = %v, want none on a rerun that wrote nothing", got.Files)
	}
	if _, ok := files.get(".claude/output-styles/odachi.md"); ok {
		t.Error("the second run wrote the Claude agent again")
	}
	if joined := strings.Join(got.Lines, "\n"); !strings.Contains(joined, "unchanged") {
		t.Errorf("the second run does not say the choice was already stored:\n%s", joined)
	}
}

// TestInitAgentsRefusesABrokenStore: a file the tool cannot read is named, and
// nothing is written, because guessing a client is exactly what the switch
// exists to prevent.
func TestInitAgentsRefusesABrokenStore(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
	store := &fakeAgentSelection{err: errors.New("agents: invalid TOML")}

	got := runAgents(files, "", store, "")

	if got.Code != 1 {
		t.Fatalf("Code = %d, want 1 (message %q)", got.Code, got.Message)
	}
	joined := strings.Join(got.Lines, "\n")
	for _, want := range []string{persona.SelectionPath, "--agents", "odachi init"} {
		if !strings.Contains(joined, want) {
			t.Errorf("the lines do not name %q:\n%s", want, joined)
		}
	}
	if _, ok := files.get("AGENTS.md"); ok {
		t.Error("init wrote AGENTS.md before resolving the selection")
	}
}

// TestInitAgentsRefusesAnInvalidFlag: a typo names the three values instead of
// choosing one.
func TestInitAgentsRefusesAnInvalidFlag(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	got := runAgents(files, "invented", &fakeAgentSelection{}, "")

	if got.Code != 1 {
		t.Fatalf("Code = %d, want 1 (message %q)", got.Code, got.Message)
	}
	joined := strings.Join(got.Lines, "\n")
	for _, want := range []string{"both", "opencode", "claude"} {
		if !strings.Contains(joined, want) {
			t.Errorf("the lines do not name %q:\n%s", want, joined)
		}
	}
	if _, ok := files.get("AGENTS.md"); ok {
		t.Error("init wrote AGENTS.md with an invalid selection")
	}
}

// TestInitAgentsExplicitBothIsStored: choosing the default out loud is still a
// decision, so it is recorded and the next plain run keeps it.
func TestInitAgentsExplicitBothIsStored(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	got := runAgents(files, persona.SelectionBoth, &fakeAgentSelection{}, "")

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	content, ok := files.get(persona.SelectionPath)
	if !ok {
		t.Fatalf("%s was not written", persona.SelectionPath)
	}
	if !strings.Contains(content, `agents = "both"`) {
		t.Errorf("%s = %q, want agents = \"both\"", persona.SelectionPath, content)
	}
	for _, client := range persona.Clients() {
		if _, ok := files.get(filepath.ToSlash(filepath.Join(client.Dir, persona.FileName))); !ok {
			t.Errorf("%s was not written", client.Dir)
		}
	}
}

// TestInitAgentsNoChoiceWritesNoStore: a project that never chose keeps the
// default and gains no new file. The absence is the default, not a missing
// write.
func TestInitAgentsNoChoiceWritesNoStore(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	got := run(files)

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	if _, ok := files.get(persona.SelectionPath); ok {
		t.Errorf("init wrote %s without anyone choosing", persona.SelectionPath)
	}
	for _, client := range persona.Clients() {
		if _, ok := files.get(filepath.ToSlash(filepath.Join(client.Dir, persona.FileName))); !ok {
			t.Errorf("%s was not written: the default is still both", client.Dir)
		}
	}
}
