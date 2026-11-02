package plan_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/plan"
	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// TestPlanWritesTheRoute: the two flags land on the step, and the document the
// human reads says who runs it.
func TestPlanWritesTheRoute(t *testing.T) {
	files := project()

	res := plan.Run(plan.Options{
		Dir:   "raiz",
		Text:  "cancelar desde el teléfono",
		Route: "delegated",
		Agent: "odachi-worker",
	}, files)

	if res.Code != plan.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, plan.CodeOK, res.Message)
	}
	step := lastStep(t, files)
	if step.Route != "delegated" || step.Agent != "odachi-worker" {
		t.Errorf("Route/Agent = %q/%q, want delegated/odachi-worker", step.Route, step.Agent)
	}
	if got := files.content(t); got != taskdoc.Parse(got).Render() {
		t.Errorf("the document left on disk is not canonical:\n%s", got)
	}
}

// TestPlanRefusesAContradictoryRoute: a delegated step with no agent and an
// inline step with one are refused before anything is written, and the message
// says which flag to change.
func TestPlanRefusesAContradictoryRoute(t *testing.T) {
	cases := []struct {
		name  string
		route string
		agent string
		want  string
	}{
		{"delegated without an agent", "delegated", "", "--agent"},
		{"inline with an agent", "inline", "odachi-worker", "inline"},
		{"an agent with no route", "", "odachi-worker", "route"},
		{"an unknown route", "quien-sea", "", "delegated"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			files := project()

			message := assertBlocked(t, plan.Run(plan.Options{
				Dir:   "raiz",
				Text:  "cancelar desde el teléfono",
				Route: c.route,
				Agent: c.agent,
			}, files))

			if !strings.Contains(message, c.want) {
				t.Errorf("message %q does not name %q", message, c.want)
			}
			if len(files.wrote) != 0 || len(files.backups) != 0 {
				t.Errorf("a blocked plan wrote %v / backed up %v", files.wrote, files.backups)
			}
		})
	}
}

// TestPlanKeepsTheRouteOutOfTheDoneForm: closing a step is not the moment to
// declare who runs it; the route belongs to the step that is opened.
func TestPlanKeepsTheRouteOutOfTheDoneForm(t *testing.T) {
	files := project()

	message := assertBlocked(t, plan.Run(plan.Options{
		Dir:   "raiz",
		Done:  "T2",
		Route: "inline",
	}, files))

	if !strings.Contains(message, "--route") {
		t.Errorf("message %q does not name the flag", message)
	}
	if len(files.wrote) != 0 {
		t.Errorf("a blocked plan wrote %v", files.wrote)
	}
}
