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
	// skills is the third key: where the project said the craft lives, empty
	// when it never chose.
	skills string
	err    error
	// targetErr is the failure of the store lookup itself: Written can succeed
	// and the write still have nowhere to go.
	targetErr error
}

func (s *fakeAgentSelection) Written(string) (persona.Store, error) {
	return persona.Store{Agents: s.stored, ProjectAgents: s.project, Skills: s.skills}, s.err
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
	if store.Skills != "" {
		text += "skills = \"" + store.Skills + "\"\n"
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

// TestInitNoProjectAgentsIsStored was the flag's test: the option is gone with
// the project half, and the store's key is pinned by the test below as inert.

// A stored project_agents from an older release survives the one branch that
// could drop it: a run that rewrites the store. The explicit scope makes the run
// write, and the key must come back with it.
func TestInitLeavesAStoredProjectAgentsKeyAlone(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
	store := &fakeAgentSelection{stored: persona.SelectionOpenCode, project: persona.ProjectAgentsWrite}

	got := initapp.Run(initapp.Options{
		Dir:            dir,
		AgentSelection: store,
		Global:         true,
		Skills:         persona.SkillsProject,
	}, files, &fakeSkills{})

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	content, ok := files.get(persona.SelectionPath)
	if !ok {
		t.Fatalf("the run did not write the store: events %v", files.events)
	}
	if !strings.Contains(content, `project_agents = "write"`) {
		t.Errorf("%s = %q, want the stored key carried through the rewrite", persona.SelectionPath, content)
	}
	if !strings.Contains(content, `skills = "project"`) {
		t.Errorf("%s = %q, want the scope the run was asked for", persona.SelectionPath, content)
	}
	if _, ok := files.get(".opencode/agents/odachi.md"); ok {
		t.Errorf("init wrote a project agent file")
	}
	if !strings.Contains(strings.Join(got.Lines, "\n"), "not written") {
		t.Errorf("the report does not say the project gets no agent:\n%s", strings.Join(got.Lines, "\n"))
	}
}

// TestInitKeepsTheStoredClientsForTheDoor: the client selection the store already
// has is the one the door is installed for. Naming a fixed client would rewrite a
// store that says both, which is the choice the store exists to protect.
func TestInitKeepsTheStoredClientsForTheDoor(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
	store := &fakeAgentSelection{stored: persona.SelectionBoth, project: persona.ProjectAgentsNone}

	got := runAgents(files, "", store, "hogar")

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	joined := strings.Join(got.Lines, "\n")
	if !strings.Contains(joined, "the agent lives only in the home") {
		t.Errorf("the lines do not say where the agent lives:\n%s", joined)
	}
	if strings.Contains(joined, "--agents "+persona.SelectionBoth) {
		t.Errorf("the lines offer a flag that no longer puts the mirrors back:\n%s", joined)
	}
}
