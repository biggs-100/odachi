package agentstoml_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/agentstoml"
	"github.com/biggs-100/odachi/internal/domain/persona"
)

// TestDecodeReadsTheSelection: the key the file carried before the project
// choice existed, normalized by the domain.
func TestDecodeReadsTheSelection(t *testing.T) {
	cases := map[string]string{
		"agents = \"opencode\"\n": persona.SelectionOpenCode,
		"agents = \"BOTH\"\n":     persona.SelectionBoth,
	}

	for data, want := range cases {
		got, err := agentstoml.Decode([]byte(data))
		if err != nil {
			t.Fatalf("Decode(%q) = %v", data, err)
		}
		if got.Agents != want {
			t.Errorf("Decode(%q).Agents = %q, want %q", data, got.Agents, want)
		}
	}
}

// TestDecodeReadsTheProjectChoice: the second key says whether the project
// carries its agent files or keeps them in the home.
func TestDecodeReadsTheProjectChoice(t *testing.T) {
	cases := map[string]string{
		"agents = \"opencode\"\nproject_agents = \"none\"\n": persona.ProjectAgentsNone,
		"agents = \"both\"\nproject_agents = \" WRITE \"\n":  persona.ProjectAgentsWrite,
		"agents = \"opencode\"\n":                            "",
	}

	for data, want := range cases {
		got, err := agentstoml.Decode([]byte(data))
		if err != nil {
			t.Fatalf("Decode(%q) = %v", data, err)
		}
		if got.ProjectAgents != want {
			t.Errorf("Decode(%q).ProjectAgents = %q, want %q", data, got.ProjectAgents, want)
		}
	}
}

// TestDecodeReadsTheSkillsScope: the third key says where the craft lives, and
// its absence is the answer every project had before the key existed.
func TestDecodeReadsTheSkillsScope(t *testing.T) {
	cases := map[string]string{
		"agents = \"opencode\"\nskills = \"home\"\n":      persona.SkillsHome,
		"agents = \"opencode\"\nskills = \" PROJECT \"\n": persona.SkillsProject,
		"agents = \"opencode\"\n":                         "",
	}

	for data, want := range cases {
		got, err := agentstoml.Decode([]byte(data))
		if err != nil {
			t.Fatalf("Decode(%q) = %v", data, err)
		}
		if got.Skills != want {
			t.Errorf("Decode(%q).Skills = %q, want %q", data, got.Skills, want)
		}
	}
}

// TestDecodeRefusesATypo: unknown keys and unknown values are errors, because a
// store the tool half-understands would silently pick a client or silently put
// the agent back in the tree.
func TestDecodeRefusesATypo(t *testing.T) {
	cases := map[string]string{
		"unknown key":        "agents = \"opencode\"\nrequired = true\n",
		"bad value":          "agents = \"maybe\"\n",
		"empty value":        "agents = \"\"\n",
		"malformed":          "agents = \"opencode\n",
		"bad project choice": "agents = \"opencode\"\nproject_agents = \"maybe\"\n",
		"bad skills scope":   "agents = \"opencode\"\nskills = \"machine\"\n",
	}

	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			if got, err := agentstoml.Decode([]byte(data)); err == nil {
				t.Fatalf("Decode = %+v, want an error", got)
			}
		})
	}
}

// TestDecodeNamesTheUnknownKey: the human must know which line to remove.
func TestDecodeNamesTheUnknownKey(t *testing.T) {
	_, err := agentstoml.Decode([]byte("agents = \"opencode\"\nrequired = true\n"))
	if err == nil {
		t.Fatal("Decode accepted an unknown key")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("error = %q, want the unknown key named", err)
	}
}

// TestRenderWritesACanonicalStore: the comment says the choice disables writing
// a client, not deleting what is already there, and the value is normalized.
func TestRenderWritesACanonicalStore(t *testing.T) {
	data, err := agentstoml.Render(persona.Store{Agents: " OPENCODE "})
	if err != nil {
		t.Fatalf("Render(OPENCODE) = %v", err)
	}
	rendered := string(data)
	for _, want := range []string{"does not delete it", `agents = "opencode"`} {
		if !strings.Contains(rendered, want) {
			t.Errorf("Render = %q, want %q", rendered, want)
		}
	}

	decoded, err := agentstoml.Decode(data)
	if err != nil {
		t.Fatalf("Decode(Render(opencode)) = %v", err)
	}
	if decoded.Agents != persona.SelectionOpenCode {
		t.Errorf("Decode(Render(opencode)).Agents = %q, want %q", decoded.Agents, persona.SelectionOpenCode)
	}
}

// TestRenderWritesTheProjectChoiceOnlyWhenChosen: a project that never chose the
// place gains no new line, so a plain run does not rewrite its store.
func TestRenderWritesTheProjectChoiceOnlyWhenChosen(t *testing.T) {
	without, err := agentstoml.Render(persona.Store{Agents: persona.SelectionOpenCode})
	if err != nil {
		t.Fatalf("Render(no choice) = %v", err)
	}
	if strings.Contains(string(without), "project_agents") {
		t.Errorf("Render wrote the project key for a project that never chose:\n%s", without)
	}

	with, err := agentstoml.Render(persona.Store{Agents: persona.SelectionOpenCode, ProjectAgents: persona.ProjectAgentsNone})
	if err != nil {
		t.Fatalf("Render(none) = %v", err)
	}
	if !strings.Contains(string(with), `project_agents = "none"`) {
		t.Errorf("Render = %q, want project_agents = \"none\"", with)
	}

	round, err := agentstoml.Decode(with)
	if err != nil {
		t.Fatalf("Decode(Render(none)) = %v", err)
	}
	if round != (persona.Store{Agents: persona.SelectionOpenCode, ProjectAgents: persona.ProjectAgentsNone}) {
		t.Errorf("Decode(Render(none)) = %+v, want the same store back", round)
	}
}

// TestRenderWritesTheSkillsScopeOnlyWhenChosen: the same rule as the project
// choice. A store that never named the scope keeps its bytes, and a store that
// did round-trips the answer the next plain init has to honour.
func TestRenderWritesTheSkillsScopeOnlyWhenChosen(t *testing.T) {
	without, err := agentstoml.Render(persona.Store{Agents: persona.SelectionBoth})
	if err != nil {
		t.Fatalf("Render(no scope) = %v", err)
	}
	if strings.Contains(string(without), "skills") {
		t.Errorf("Render wrote the skills key for a project that never chose:\n%s", without)
	}

	with, err := agentstoml.Render(persona.Store{Agents: persona.SelectionBoth, Skills: persona.SkillsHome})
	if err != nil {
		t.Fatalf("Render(home) = %v", err)
	}
	if !strings.Contains(string(with), `skills = "home"`) {
		t.Errorf("Render = %q, want skills = \"home\"", with)
	}

	round, err := agentstoml.Decode(with)
	if err != nil {
		t.Fatalf("Decode(Render(home)) = %v", err)
	}
	if round != (persona.Store{Agents: persona.SelectionBoth, Skills: persona.SkillsHome}) {
		t.Errorf("Decode(Render(home)) = %+v, want the same store back", round)
	}
}
