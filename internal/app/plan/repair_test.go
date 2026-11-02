package plan_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/plan"
	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// unmarked is a project whose only step is done and carries no mark: the state
// 0163's T2 was left in when the worktree was dirty.
func unmarked(t *testing.T) *fakeFiles {
	t.Helper()
	doc := taskdoc.TaskDoc{
		Title:        "reserva-de-turno",
		Requirements: []taskdoc.Requirement{{ID: "R1", Text: "reservar turnos"}},
		DoneWhen:     []string{"Reservo un turno y aparece en la agenda (R1)"},
		Steps: []taskdoc.Step{
			{ID: "T1", Text: "horario libre u ocupado", Requirement: "R1", Done: true},
		},
		History: []string{`L1 · "reservar turnos"`},
	}
	return newFakeFiles(map[string]string{taskPath: doc.Render()})
}

// TestPlanRecordsAMarkAFinishedStepWasMissing: a record nobody can repair is a
// hole. The step stays done and gets the bytes it produced.
func TestPlanRecordsAMarkAFinishedStepWasMissing(t *testing.T) {
	files := unmarked(t)

	res := plan.Run(plan.Options{Dir: "raiz", Done: "T1", Commit: "cafe123"}, files)

	if res.Code != plan.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, plan.CodeOK, res.Message)
	}
	if !strings.Contains(res.Message, "cafe123") {
		t.Errorf("message %q does not name the mark it recorded", res.Message)
	}
	step := taskdoc.Parse(files.content(t)).Steps[0]
	if !step.Done {
		t.Error("Done = false, want true: repairing a mark does not reopen the step")
	}
	if step.Commit != "cafe123" {
		t.Errorf("Commit = %q, want cafe123", step.Commit)
	}
	if len(files.wrote) != 1 {
		t.Errorf("wrote %v, want one write", files.wrote)
	}
}

// TestPlanSaysHowToRecordAMarkItCouldNot: with no --commit the tool records
// nothing and names the flag, because silence is the dead end this feature
// closes.
func TestPlanSaysHowToRecordAMarkItCouldNot(t *testing.T) {
	files := unmarked(t)

	res := plan.Run(plan.Options{Dir: "raiz", Done: "T1"}, files)

	if res.Code != plan.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, plan.CodeOK, res.Message)
	}
	if !strings.Contains(res.Message, "--commit") {
		t.Errorf("message %q does not name --commit", res.Message)
	}
	if len(files.wrote) != 0 {
		t.Errorf("a step with no mark and no commit wrote %v", files.wrote)
	}
}

// TestPlanKeepsAMarkedStepUntouched: a good mark is the memory of the work, and
// re-running the verb does not rewrite it.
func TestPlanKeepsAMarkedStepUntouched(t *testing.T) {
	files := project()

	res := plan.Run(plan.Options{Dir: "raiz", Done: "T1", Commit: "other"}, files)

	if res.Code != plan.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, plan.CodeOK, res.Message)
	}
	if res.Message != "step T1 was already done" {
		t.Errorf("Message = %q, want the old sentence", res.Message)
	}
	if len(files.wrote) != 0 {
		t.Errorf("a marked step wrote %v", files.wrote)
	}
	if got := taskdoc.Parse(files.content(t)).Steps[0].Commit; got != "4a1c9f" {
		t.Errorf("Commit = %q, want the mark that was already there", got)
	}
}
