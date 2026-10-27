package taskdoc_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// canonical steps a test can build on: one done step with its commit, one open
// step, and the two requirements the tests link to.
func canonical(t *testing.T, steps ...taskdoc.Step) taskdoc.TaskDoc {
	t.Helper()
	base := taskdoc.TaskDoc{
		Title:        "reserva-de-turno",
		Requirements: []taskdoc.Requirement{{ID: "R1", Text: "reservar"}, {ID: "R2", Text: "ver la agenda"}},
		DoneWhen:     []string{"Reservo un turno y aparece en la agenda (R1)"},
		Steps:        steps,
		History:      []string{`L1 · "reservar"`},
	}
	return taskdoc.Parse(base.Render())
}

// TestAddStepNumbersMonotonically: the T number is max+1, so a step is never
// reused even when the document skipped one.
func TestAddStepNumbersMonotonically(t *testing.T) {
	cases := []struct {
		name  string
		steps []taskdoc.Step
		want  string
	}{
		{"no steps", nil, "T1"},
		{"one open", []taskdoc.Step{{ID: "T1", Text: "primero"}}, "T2"},
		{"one done and one open", []taskdoc.Step{
			{ID: "T1", Text: "primero", Done: true, Commit: "4a1c9f"},
			{ID: "T2", Text: "segundo"},
		}, "T3"},
		{"a gap is never reused", []taskdoc.Step{
			{ID: "T1", Text: "primero"},
			{ID: "T3", Text: "tercero"},
		}, "T4"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			doc := canonical(t, c.steps...)

			grown, err := doc.AddStep("el paso nuevo", "")
			if err != nil {
				t.Fatalf("AddStep: %v", err)
			}

			if len(grown.Steps) != len(doc.Steps)+1 {
				t.Fatalf("steps = %d, want %d", len(grown.Steps), len(doc.Steps)+1)
			}
			last := grown.Steps[len(grown.Steps)-1]
			if last.ID != c.want {
				t.Errorf("new step id = %q, want %q", last.ID, c.want)
			}
			if last.Done {
				t.Error("a new step must start open")
			}
			if last.Text != "el paso nuevo" {
				t.Errorf("text = %q, want %q", last.Text, "el paso nuevo")
			}
			if len(doc.Steps) == len(grown.Steps) {
				t.Error("AddStep mutated the receiver")
			}
		})
	}
}

// TestAddStepWithoutRequirementLeavesNoLink: a step that is not tied to an R
// renders without the tail, and the document stays canonical.
func TestAddStepWithoutRequirementLeavesNoLink(t *testing.T) {
	doc := canonical(t, taskdoc.Step{ID: "T1", Text: "primero"})

	grown, err := doc.AddStep("un paso suelto", "")
	if err != nil {
		t.Fatalf("AddStep: %v", err)
	}

	last := grown.Steps[len(grown.Steps)-1]
	if last.Requirement != "" {
		t.Errorf("Requirement = %q, want empty", last.Requirement)
	}
	if !strings.Contains(grown.Render(), "- [ ] T2 · un paso suelto\n") {
		t.Errorf("rendered document does not carry the step without a link:\n%s", grown.Render())
	}
	if round := taskdoc.Parse(grown.Render()); !reflect.DeepEqual(round, grown) {
		t.Errorf("round-trip changed the document:\n got %+v\nwant %+v", round, grown)
	}
}

// TestAddStepLinksAnExistingRequirement: the link is stored normalized so
// `--req r2` and `--req R2` are the same requirement.
func TestAddStepLinksAnExistingRequirement(t *testing.T) {
	doc := canonical(t)

	grown, err := doc.AddStep("mostrar la agenda", "r2")
	if err != nil {
		t.Fatalf("AddStep: %v", err)
	}

	last := grown.Steps[len(grown.Steps)-1]
	if last.Requirement != "R2" {
		t.Errorf("Requirement = %q, want R2", last.Requirement)
	}
	steps := taskdoc.Parse(grown.Render()).Steps
	if got := steps[len(steps)-1].Requirement; got != "R2" {
		t.Errorf("rendered step requirement = %q, want R2", got)
	}
}

// TestAddStepUnknownRequirementListsThem: a link to a requirement the document
// does not declare is refused, and the message says which ones exist.
func TestAddStepUnknownRequirementListsThem(t *testing.T) {
	doc := canonical(t)

	_, err := doc.AddStep("mostrar la agenda", "R9")

	if err == nil {
		t.Fatal("AddStep with an unknown requirement = nil error, want one")
	}
	for _, want := range []string{"R9", "R1", "R2"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not name %s", err, want)
		}
	}
	if strings.Contains(err.Error(), "\n") {
		t.Errorf("error must be one line: %q", err)
	}
}

// TestAddStepWithoutRequirementsSaysSo: the message still tells the human what
// the document has, which is nothing yet.
func TestAddStepWithoutRequirementsSaysSo(t *testing.T) {
	doc := taskdoc.Parse("# reserva-de-turno\n\n## Pasos\n")

	_, err := doc.AddStep("primer paso", "R1")

	if err == nil {
		t.Fatal("AddStep with no requirements = nil error, want one")
	}
	if !strings.Contains(err.Error(), "there are no requirements") {
		t.Errorf("error %q does not say the document has no requirements", err)
	}
}

// TestAddStepRejectsBadText: the step text is one short line, and nothing is
// appended when it is not.
func TestAddStepRejectsBadText(t *testing.T) {
	doc := canonical(t)

	cases := []struct {
		name string
		text string
	}{
		{"empty", ""},
		{"spaces only", "   "},
		{"multiline", "primera linea\nsegunda linea"},
		{"with return", "primera linea\rsegunda linea"},
		{"leading dash", "- primer paso"},
		{"attached dash", "-primer paso"},
		{"121 characters", strings.Repeat("a", 121)},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			grown, err := doc.AddStep(c.text, "")
			if err == nil {
				t.Fatalf("AddStep(%q) = nil error, want one", c.text)
			}
			if len(grown.Steps) != len(doc.Steps) {
				t.Errorf("a rejected step was appended: %+v", grown.Steps)
			}
			if strings.Contains(err.Error(), "\n") {
				t.Errorf("error must be one line: %q", err)
			}
		})
	}
}

// TestAddStepAcceptsOneHundredTwentyCharacters: the limit is inclusive and
// counts characters, not bytes, because the text is the human's words.
func TestAddStepAcceptsOneHundredTwentyCharacters(t *testing.T) {
	doc := canonical(t)

	for _, text := range []string{strings.Repeat("a", 120), strings.Repeat("á", 120)} {
		grown, err := doc.AddStep(text, "")
		if err != nil {
			t.Fatalf("AddStep(%d chars): %v", len([]rune(text)), err)
		}
		if got := grown.Steps[len(grown.Steps)-1].Text; got != text {
			t.Errorf("text = %q, want %q", got, text)
		}
	}
}

// TestAddStepKeepsTextOnOneCanonicalLine: runs of spaces become one space, so
// the rendered step parses back to the same text instead of being mistaken for
// a requirement tail.
func TestAddStepKeepsTextOnOneCanonicalLine(t *testing.T) {
	doc := canonical(t)

	grown, err := doc.AddStep("  arreglar  el  botón  ", "")
	if err != nil {
		t.Fatalf("AddStep: %v", err)
	}

	last := grown.Steps[len(grown.Steps)-1]
	if last.Text != "arreglar el botón" {
		t.Errorf("text = %q, want %q", last.Text, "arreglar el botón")
	}
	if round := taskdoc.Parse(grown.Render()); !reflect.DeepEqual(round, grown) {
		t.Errorf("round-trip changed the document:\n got %+v\nwant %+v", round, grown)
	}
}

// TestReopenUnchecksOnlyTheStepsThatCiteTheRequirement: a change to R1 reopens
// the done steps that claimed R1 and drops their marks. A step of another
// requirement, a step still open, and a step with no link are not touched, and
// the ids it reopened come back in document order.
func TestReopenUnchecksOnlyTheStepsThatCiteTheRequirement(t *testing.T) {
	doc := canonical(t,
		taskdoc.Step{ID: "T1", Text: "reservar", Requirement: "R1", Done: true, Commit: "4a1c9f"},
		taskdoc.Step{ID: "T2", Text: "ver la agenda", Requirement: "R2", Done: true, Commit: "9b2e11"},
		taskdoc.Step{ID: "T3", Text: "cancelar", Requirement: "R1"},
		taskdoc.Step{ID: "T4", Text: "sin requisito", Done: true, Commit: "77aa01"},
	)

	grown, reopened := doc.Reopen("r1")

	if want := []string{"T1"}; !reflect.DeepEqual(reopened, want) {
		t.Fatalf("reopened = %v, want %v", reopened, want)
	}
	if grown.Steps[0].Done || grown.Steps[0].Commit != "" {
		t.Errorf("T1 = %+v, want it open and with no mark", grown.Steps[0])
	}
	if !grown.Steps[1].Done || grown.Steps[1].Commit != "9b2e11" {
		t.Errorf("T2 = %+v, want it untouched: it cites R2", grown.Steps[1])
	}
	if grown.Steps[2].Done {
		t.Errorf("T3 = %+v, want it open: it was never done", grown.Steps[2])
	}
	if !grown.Steps[3].Done || grown.Steps[3].Commit != "77aa01" {
		t.Errorf("T4 = %+v, want it untouched: it cites no requirement", grown.Steps[3])
	}
	if round := taskdoc.Parse(grown.Render()); !reflect.DeepEqual(round, grown) {
		t.Errorf("round-trip changed the document:\n got %+v\nwant %+v", round, grown)
	}
}

// TestReopenWithNoRequirementReopensNothing: an empty link must not match the
// steps that carry no requirement at all.
func TestReopenWithNoRequirementReopensNothing(t *testing.T) {
	doc := canonical(t,
		taskdoc.Step{ID: "T1", Text: "sin requisito", Done: true, Commit: "77aa01"},
	)

	grown, reopened := doc.Reopen("")

	if len(reopened) != 0 {
		t.Fatalf("reopened = %v, want none", reopened)
	}
	if !grown.Steps[0].Done {
		t.Errorf("T1 = %+v, want it still done", grown.Steps[0])
	}
}
