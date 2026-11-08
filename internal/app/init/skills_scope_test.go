package init_test

import (
	"strings"
	"testing"

	initapp "github.com/biggs-100/odachi/internal/app/init"
	"github.com/biggs-100/odachi/internal/domain/persona"
)

// storedSkills is a repository with a commit and a store that already chose
// where the craft lives.
func storedSkills(scope string) (*fakeFiles, *fakeAgentSelection, *fakeSkills) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
	store := &fakeAgentSelection{stored: persona.SelectionBoth, project: persona.ProjectAgentsWrite, skills: scope}
	return files, store, &fakeSkills{}
}

// TestAPlainInitAsksForTheStoredScope: the flag is a decision, not a one-time
// order. A project that put its skills in the home must not get the project copy
// rebuilt by the next plain init.
func TestAPlainInitAsksForTheStoredScope(t *testing.T) {
	files, store, skills := storedSkills(persona.SkillsHome)

	got := initapp.Run(initapp.Options{Dir: dir, AgentSelection: store}, files, skills)

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	if skills.scope != persona.SkillsHome {
		t.Errorf("the skills half was asked for %q, want %q", skills.scope, persona.SkillsHome)
	}
	if content, ok := files.get("agents.toml"); ok {
		t.Errorf("a plain init rewrote a store that already said %q:\n%s", persona.SkillsHome, content)
	}
	if !strings.Contains(strings.Join(got.Lines, "\n"), persona.SkillsHome) {
		t.Errorf("the run does not name the scope it used:\n%s", strings.Join(got.Lines, "\n"))
	}
}

// TestTheSkillsFlagIsTheChoiceTheStoreKeeps: naming a scope the store's silence
// does not already say installs it and records it, so the next plain init answers
// the same way. The flag here is the project's own copy: home is what an absent
// key means, and a file gains nothing by saying it aloud.
func TestTheSkillsFlagIsTheChoiceTheStoreKeeps(t *testing.T) {
	files, store, skills := storedSkills("")

	got := initapp.Run(initapp.Options{Dir: dir, AgentSelection: store, Skills: persona.SkillsProject}, files, skills)

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	if skills.scope != persona.SkillsProject {
		t.Errorf("the skills half was asked for %q, want %q", skills.scope, persona.SkillsProject)
	}
	content, ok := files.get("agents.toml")
	if !ok {
		t.Fatal("the choice was not stored")
	}
	if !strings.Contains(content, `skills = "`+persona.SkillsProject+`"`) {
		t.Errorf("the store does not keep the scope:\n%s", content)
	}
	if !strings.Contains(strings.Join(got.Lines, "\n"), "--skills") {
		t.Errorf("the run does not say the flag was the source:\n%s", strings.Join(got.Lines, "\n"))
	}
}

// TestAProjectThatNeverChoseReadsTheCraftFromTheMachine: absence is not a third
// value. The run answers the tool's default since 0242 -one copy per machine- and
// writes no store only to say what the silence already says.
func TestAProjectThatNeverChoseReadsTheCraftFromTheMachine(t *testing.T) {
	files, store, skills := storedSkills("")

	got := initapp.Run(initapp.Options{Dir: dir, AgentSelection: store}, files, skills)

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	if skills.scope != persona.SkillsHome {
		t.Errorf("the skills half was asked for %q, want %q", skills.scope, persona.SkillsHome)
	}
	if content, ok := files.get("agents.toml"); ok {
		t.Errorf("a project that never chose gained a store:\n%s", content)
	}
}

// TestTheStoredProjectScopeIsHonouredToo: the other direction. A project that
// brought the craft back keeps it there across a plain init.
func TestTheStoredProjectScopeIsHonouredToo(t *testing.T) {
	files, store, skills := storedSkills(persona.SkillsProject)

	initapp.Run(initapp.Options{Dir: dir, AgentSelection: store}, files, skills)

	if skills.scope != persona.SkillsProject {
		t.Errorf("the skills half was asked for %q, want %q", skills.scope, persona.SkillsProject)
	}
}
