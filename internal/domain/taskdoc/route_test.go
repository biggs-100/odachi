package taskdoc_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// TestParseReadsTheRouteAndTheAgent: a delegated step names the agent that runs
// it, so the document says who did the work instead of a chat nobody keeps.
func TestParseReadsTheRouteAndTheAgent(t *testing.T) {
	md := `# demo

## Steps
- [ ] T1 · escribir el parser   R1 · route:delegated/odachi-worker
`
	got := taskdoc.Parse(md)
	if len(got.Steps) != 1 {
		t.Fatalf("len(Steps) = %d, want 1: the line was read as text", len(got.Steps))
	}
	step := got.Steps[0]
	if step.Route != "delegated" {
		t.Errorf("Route = %q, want delegated", step.Route)
	}
	if step.Agent != "odachi-worker" {
		t.Errorf("Agent = %q, want odachi-worker", step.Agent)
	}
	if step.Commit != "" {
		t.Errorf("Commit = %q, want empty: the route is not a commit", step.Commit)
	}
	if step.Requirement != "R1" {
		t.Errorf("Requirement = %q, want R1", step.Requirement)
	}
}

// TestParseReadsAnInlineRoute: the parent that does the work says so, and names
// no agent, because there is none.
func TestParseReadsAnInlineRoute(t *testing.T) {
	md := `# demo

## Steps
- [ ] T1 · escribir el parser   R1 · route:inline
`
	step := taskdoc.Parse(md).Steps[0]
	if step.Route != "inline" {
		t.Errorf("Route = %q, want inline", step.Route)
	}
	if step.Agent != "" {
		t.Errorf("Agent = %q, want empty: an inline step has no other executor", step.Agent)
	}
}

// TestParseKeepsTheRouteBesideTheCommit: a done step carries two facts, who ran
// it and which bytes it produced, and neither hides the other.
func TestParseKeepsTheRouteBesideTheCommit(t *testing.T) {
	md := `# demo

## Steps
- [x] T1 · escribir el parser   R1 · route:inline · tree:4a1c9f
`
	step := taskdoc.Parse(md).Steps[0]
	if !step.Done {
		t.Error("Done = false, want true")
	}
	if step.Route != "inline" {
		t.Errorf("Route = %q, want inline", step.Route)
	}
	if step.Commit != "tree:4a1c9f" {
		t.Errorf("Commit = %q, want tree:4a1c9f", step.Commit)
	}
}

// TestARouteWithoutARequirementIsStillRead: a step may exist before it is linked
// to a requirement, and the route is not part of the link.
func TestARouteWithoutARequirementIsStillRead(t *testing.T) {
	md := `# demo

## Steps
- [ ] T1 · escribir el parser   · route:delegated/odachi-worker
`
	step := taskdoc.Parse(md).Steps[0]
	if step.Route != "delegated" || step.Agent != "odachi-worker" {
		t.Errorf("Route/Agent = %q/%q, want delegated/odachi-worker", step.Route, step.Agent)
	}
	if step.Requirement != "" {
		t.Errorf("Requirement = %q, want empty", step.Requirement)
	}
}

// TestAStepWithoutARouteIsNotDeclared: the field is declared, never inferred. A
// document written before this change does not suddenly say the parent ran it.
func TestAStepWithoutARouteIsNotDeclared(t *testing.T) {
	md := `# demo

## Steps
- [x] T1 · escribir el parser   R1 · tree:4a1c9f
`
	step := taskdoc.Parse(md).Steps[0]
	if step.Route != "" || step.Agent != "" {
		t.Errorf("Route/Agent = %q/%q, want both empty: no route means not declared", step.Route, step.Agent)
	}
}

// TestRenderKeepsTheRouteAndTheAgent: the canonical form is the only writer, so
// what Parse read has to come back unchanged.
func TestRenderKeepsTheRouteAndTheAgent(t *testing.T) {
	doc := taskdoc.TaskDoc{
		Title:        "demo",
		Requirements: []taskdoc.Requirement{{ID: "R1", Text: "que exista"}},
		Steps: []taskdoc.Step{
			{ID: "T1", Text: "escribir el parser", Requirement: "R1", Route: "inline"},
			{ID: "T2", Text: "probar el parser", Requirement: "R1", Route: "delegated", Agent: "odachi-worker", Done: true, Commit: "tree:4a1c9f"},
			{ID: "T3", Text: "sin requisito todavía", Route: "delegated", Agent: "odachi-worker"},
		},
	}

	again := taskdoc.Parse(doc.Render())
	if !reflect.DeepEqual(again, doc) {
		t.Fatalf("render then parse changed the document:\ngot  %+v\nwant %+v\nrendered:\n%s", again, doc, doc.Render())
	}
	if lost := taskdoc.Unaccounted(doc.Render()); len(lost) != 0 {
		t.Errorf("the write guard flags the canonical route lines: %v", lost)
	}
}

// TestAnUnknownRouteWordIsStillRead: Parse is total. A hand-written route the
// tool does not know is kept, not thrown away as prose, so `plan` can say what
// it accepts instead of rewriting the line silently.
func TestAnUnknownRouteWordIsStillRead(t *testing.T) {
	md := `# demo

## Steps
- [ ] T1 · escribir el parser   R1 · route:quien-sea
`
	step := taskdoc.Parse(md).Steps[0]
	if step.Route != "quien-sea" {
		t.Errorf("Route = %q, want quien-sea", step.Route)
	}
	if strings.Contains(step.Text, "route:") {
		t.Errorf("the route stayed inside the text: %q", step.Text)
	}
}
