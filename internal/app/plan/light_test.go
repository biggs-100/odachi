package plan_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/plan"
	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// lightSeed is a one-line feature: one quoted requirement and the step that
// carries the same words.
func lightSeed(t *testing.T) *fakeFiles {
	t.Helper()
	doc := taskdoc.TaskDoc{
		Title:        "saludo",
		Light:        true,
		Requirements: []taskdoc.Requirement{{ID: "R1", Text: "el saludo dice la versión"}},
	}
	grown, err := doc.AddStep("el saludo dice la versión", "R1")
	if err != nil {
		t.Fatalf("build the light document: %v", err)
	}
	return newFakeFiles(map[string]string{taskPath: grown.Render()})
}

// TestPlanRefusesASecondStepOnALightFeature: the lane is one line. A second step
// is the signal that the change does not fit it, and the way out is a brief, not
// a document that grows while the line still says it is light.
func TestPlanRefusesASecondStepOnALightFeature(t *testing.T) {
	files := lightSeed(t)
	before := files.content(t)

	message := assertBlocked(t, plan.Run(plan.Options{Dir: "raiz", Text: "otro paso"}, files))

	for _, want := range []string{"light", "start"} {
		if !strings.Contains(message, want) {
			t.Errorf("message %q does not name %s", message, want)
		}
	}
	if after := files.content(t); after != before {
		t.Errorf("a refused step changed the document:\n%s", after)
	}
}

// TestPlanStillClosesALightFeature: the refusal is about opening a second step.
// Closing the only one is how the lane ships.
func TestPlanStillClosesALightFeature(t *testing.T) {
	files := lightSeed(t)

	res := plan.Run(plan.Options{Dir: "raiz", Done: "T1", Commit: "cafe123"}, files)

	if res.Code != plan.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, plan.CodeOK, res.Message)
	}
	step := taskdoc.Parse(files.content(t)).Steps[0]
	if !step.Done || step.Commit != "cafe123" {
		t.Errorf("T1 = %+v, want it closed with its mark", step)
	}
}
