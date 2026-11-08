package taskdoc_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// baseWithRequirement is the smallest document a step can be added to.
const baseWithRequirement = `# demo

## Requirements (verbatim, from the user)
R1 · "implementalo"

## Steps

## History
`

// TestAddStepBySkillsWritesAndParsesTheCraft: the plan records the skill names in
// the step line, and the document reads them back, so the delegation knows what
// the step owes without a chat.
func TestAddStepBySkillsWritesAndParsesTheCraft(t *testing.T) {
	doc := taskdoc.Parse(baseWithRequirement)
	grown, err := doc.AddStepBySkills("implementar el parser", "R1", taskdoc.Route{Kind: taskdoc.RouteDelegated, Agent: "odachi-worker"}, []string{"tests", "drive"})
	if err != nil {
		t.Fatalf("AddStepBySkills: %v", err)
	}

	rendered := grown.Render()
	if !strings.Contains(rendered, "skills:tests,drive") {
		t.Fatalf("the rendered step does not carry the craft:\n%s", rendered)
	}

	back := taskdoc.Parse(rendered)
	if len(back.Steps) != 1 {
		t.Fatalf("len(Steps) = %d, want 1: the line was read as text", len(back.Steps))
	}
	step := back.Steps[0]
	if len(step.Skills) != 2 || step.Skills[0] != "tests" || step.Skills[1] != "drive" {
		t.Errorf("Skills = %v, want [tests drive]", step.Skills)
	}
	if step.Route != taskdoc.RouteDelegated || step.Agent != "odachi-worker" {
		t.Errorf("route = %s/%s, want delegated/odachi-worker", step.Route, step.Agent)
	}
}

// TestSkillsBelongToADelegatedStep: an inline step runs in the parent, and the
// parent already has the craft; a skill there is a declaration that means nothing.
func TestSkillsBelongToADelegatedStep(t *testing.T) {
	doc := taskdoc.Parse(baseWithRequirement)
	_, err := doc.AddStepBySkills("implementar", "R1", taskdoc.Route{Kind: taskdoc.RouteInline}, []string{"tests"})
	if err == nil {
		t.Fatal("an inline step accepted a skill")
	}
	if !strings.Contains(err.Error(), "delegated") {
		t.Errorf("error %q does not name the route that allows it", err)
	}
}

// TestSkillNamesAreSlugs: a name that is not a slug is refused where the
// document is built, so the delegation never looks for a path that cannot exist.
func TestSkillNamesAreSlugs(t *testing.T) {
	doc := taskdoc.Parse(baseWithRequirement)
	_, err := doc.AddStepBySkills("implementar", "R1", taskdoc.Route{Kind: taskdoc.RouteDelegated, Agent: "odachi-worker"}, []string{"Tests!"})
	if err == nil {
		t.Fatal("a name that is not a slug was accepted")
	}
	if !strings.Contains(err.Error(), "Tests!") {
		t.Errorf("error %q does not name the refused skill", err)
	}
}

// TestTheSameSkillIsRecordedOnce: a parent that repeats a name means one skill.
func TestTheSameSkillIsRecordedOnce(t *testing.T) {
	doc := taskdoc.Parse(baseWithRequirement)
	grown, err := doc.AddStepBySkills("implementar", "R1", taskdoc.Route{Kind: taskdoc.RouteDelegated, Agent: "odachi-worker"}, []string{"tests", "tests"})
	if err != nil {
		t.Fatalf("AddStepBySkills: %v", err)
	}
	step := taskdoc.Parse(grown.Render()).Steps[0]
	if len(step.Skills) != 1 || step.Skills[0] != "tests" {
		t.Errorf("Skills = %v, want [tests]", step.Skills)
	}
}
