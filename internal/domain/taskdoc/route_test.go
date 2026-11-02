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

// TestAddStepDeclaresTheRoute: the writer takes the executor with the text, so a
// step is born saying who runs it instead of being patched afterwards.
func TestAddStepDeclaresTheRoute(t *testing.T) {
	doc, err := canonical(t).AddStepBy("escribir el parser", "R1", taskdoc.Route{
		Kind:  taskdoc.RouteDelegated,
		Agent: "odachi-worker",
	})
	if err != nil {
		t.Fatalf("AddStepBy: %v", err)
	}

	step := doc.Steps[len(doc.Steps)-1]
	if step.Route != taskdoc.RouteDelegated || step.Agent != "odachi-worker" {
		t.Errorf("Route/Agent = %q/%q, want delegated/odachi-worker", step.Route, step.Agent)
	}
	again := taskdoc.Parse(doc.Render())
	if last := again.Steps[len(again.Steps)-1]; last.Route != step.Route || last.Agent != step.Agent {
		t.Errorf("the canonical form lost the route: %+v", last)
	}
}

// TestAddStepKeepsAnUndeclaredRoute: the plain form still writes a step with no
// executor, because a document written before the field existed says nothing.
func TestAddStepKeepsAnUndeclaredRoute(t *testing.T) {
	doc, err := canonical(t).AddStep("escribir el parser", "R1")
	if err != nil {
		t.Fatalf("AddStep: %v", err)
	}
	step := doc.Steps[len(doc.Steps)-1]
	if step.Route != "" || step.Agent != "" {
		t.Errorf("Route/Agent = %q/%q, want both empty", step.Route, step.Agent)
	}
}

// TestAddStepRefusesARouteThatContradictsItsAgent: the pair is decided before the
// document is written, and the message says what to add or drop.
func TestAddStepRefusesARouteThatContradictsItsAgent(t *testing.T) {
	cases := []struct {
		name  string
		route taskdoc.Route
		want  string
	}{
		{
			name:  "delegated without an agent",
			route: taskdoc.Route{Kind: taskdoc.RouteDelegated},
			want:  "agent",
		},
		{
			name:  "inline with an agent",
			route: taskdoc.Route{Kind: taskdoc.RouteInline, Agent: "odachi-worker"},
			want:  "inline",
		},
		{
			name:  "an agent with no route",
			route: taskdoc.Route{Agent: "odachi-worker"},
			want:  "route",
		},
		{
			name:  "an unknown route",
			route: taskdoc.Route{Kind: "quien-sea"},
			want:  "delegated",
		},
		{
			name:  "an agent the line cannot carry",
			route: taskdoc.Route{Kind: taskdoc.RouteDelegated, Agent: "odachi/worker"},
			want:  "agent",
		},
		{
			name:  "an agent with a space",
			route: taskdoc.Route{Kind: taskdoc.RouteDelegated, Agent: "odachi worker"},
			want:  "agent",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			before := canonical(t)
			after, err := before.AddStepBy("escribir el parser", "R1", c.route)
			if err == nil {
				t.Fatalf("AddStepBy(%+v) wrote a step it should have refused", c.route)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("message %q does not name %q", err, c.want)
			}
			if len(after.Steps) != len(before.Steps) {
				t.Errorf("a refused step changed the document: %d steps, want %d", len(after.Steps), len(before.Steps))
			}
		})
	}
}

// TestACommitWithNoRequirementKeepsItsMark: the commit-only tail is the old
// spelling of a step closed without a link, and a sha whose letters run into
// digits is not a requirement id. The mark must survive the round trip.
func TestACommitWithNoRequirementKeepsItsMark(t *testing.T) {
	md := `# demo

## Steps
- [x] T1 · cerrar la puerta   · cafe123
`
	step := taskdoc.Parse(md).Steps[0]
	if step.Requirement != "" {
		t.Errorf("Requirement = %q, want empty: the mark is not a requirement", step.Requirement)
	}
	if step.Commit != "cafe123" {
		t.Errorf("Commit = %q, want cafe123", step.Commit)
	}

	doc := taskdoc.TaskDoc{
		Title: "demo",
		Steps: []taskdoc.Step{{ID: "T1", Text: "cerrar la puerta", Commit: "cafe123", Done: true}},
	}
	again := taskdoc.Parse(doc.Render())
	if !reflect.DeepEqual(again, doc) {
		t.Fatalf("render then parse changed the document:\ngot  %+v\nwant %+v\nrendered:\n%s", again, doc, doc.Render())
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
