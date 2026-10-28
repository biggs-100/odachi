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
	// project is the second key the store carries: where the project said its
	// agent lives, empty when it never chose.
	project string
	err     error
	// targetErr is the failure of the store lookup itself: Written can succeed
	// and the write still have nowhere to go.
	targetErr error
}

func (s *fakeAgentSelection) Written(string) (persona.Store, error) {
	return persona.Store{Agents: s.stored, ProjectAgents: s.project}, s.err
}

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

func (s *fakeAgentSelection) Render(store persona.Store) ([]byte, error) {
	text := "agents = \"" + store.Agents + "\"\n"
	if store.ProjectAgents != "" {
		text += "project_agents = \"" + store.ProjectAgents + "\"\n"
	}
	return []byte(text), nil
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

// TestInitNoProjectAgentsIsStored: the flag is a decision, not a one-time order.
// The store records it, so the next plain run honors it instead of rebuilding the
// mirrors - the defect that opened this feature.
func TestInitNoProjectAgentsIsStored(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
	store := &fakeAgentSelection{}

	first := initapp.Run(initapp.Options{
		Dir:             dir,
		AgentSelection:  store,
		Global:          true,
		NoProjectAgents: true,
	}, files, &fakeSkills{})

	if first.Code != 0 {
		t.Fatalf("first run Code = %d, want 0 (message %q)", first.Code, first.Message)
	}
	content, ok := files.get(persona.SelectionPath)
	if !ok {
		t.Fatalf("%s was not written: the choice was not recorded", persona.SelectionPath)
	}
	if !strings.Contains(content, `project_agents = "none"`) {
		t.Errorf("%s = %q, want project_agents = \"none\"", persona.SelectionPath, content)
	}

	// The store now answers what the file says, and a plain run must respect it.
	store.project = persona.ProjectAgentsNone
	files.events = nil

	second := runAgents(files, "", store, "hogar")

	if second.Code != 0 {
		t.Fatalf("second run Code = %d, want 0 (message %q)", second.Code, second.Message)
	}
	if _, ok := files.get(".opencode/agents/odachi.md"); ok {
		t.Error("the plain run rebuilt a mirror the store had turned off")
	}
	for _, event := range files.events {
		if strings.Contains(event, "agents.toml") {
			t.Errorf("the second run rewrote the store: %s", event)
		}
	}
}

// TestInitHonorsTheStoredProjectChoice: a project that recorded that its agent
// lives in the home keeps it there on a plain run. Without this, one bare init
// silently rebuilt the mirrors the person had chosen to delete, and doctor then
// reported them as healthy.
func TestInitHonorsTheStoredProjectChoice(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
	store := &fakeAgentSelection{stored: persona.SelectionOpenCode, project: persona.ProjectAgentsNone}

	got := runAgents(files, "", store, "hogar")

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	for _, path := range []string{
		persona.SubagentFilesFor(persona.SelectionOpenCode)[0].Path,
		".opencode/agents/odachi.md",
	} {
		if _, ok := files.get(path); ok {
			t.Errorf("the run wrote %s although the store keeps the agent in the home", path)
		}
	}
	joined := strings.Join(got.Lines, "\n")
	for _, want := range []string{"home", "--agents opencode"} {
		if !strings.Contains(joined, want) {
			t.Errorf("the lines do not name %q:\n%s", want, joined)
		}
	}
}

// TestInitExplicitAgentsPutsTheMirrorsBack: the flag that names clients is the
// way out of the home-only choice, so the stored answer changes with it. The
// alternative was a second flag, and the fence of this feature forbids one.
func TestInitExplicitAgentsPutsTheMirrorsBack(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
	store := &fakeAgentSelection{stored: persona.SelectionOpenCode, project: persona.ProjectAgentsNone}

	got := runAgents(files, persona.SelectionOpenCode, store, "hogar")

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	if _, ok := files.get(".opencode/agents/odachi.md"); !ok {
		t.Error(".opencode/agents/odachi.md was not written back")
	}
	content, ok := files.get(persona.SelectionPath)
	if !ok {
		t.Fatalf("%s was not rewritten", persona.SelectionPath)
	}
	if !strings.Contains(content, `project_agents = "write"`) {
		t.Errorf("%s = %q, want project_agents = \"write\"", persona.SelectionPath, content)
	}
}

// TestInitNoChoiceWritesNoStore: a project that never chose keeps the
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
