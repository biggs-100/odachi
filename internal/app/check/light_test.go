package check_test

import (
	"testing"

	"github.com/biggs-100/odachi/internal/app/check"
	"github.com/biggs-100/odachi/internal/domain/gate"
	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// lightDoc is a one-line feature: the requirement quoted, one step with the same
// words, and the lane declared in the preamble.
func lightDoc(done bool) string {
	step := taskdoc.Step{ID: "T1", Text: "el saludo dice la versión", Requirement: "R1"}
	if done {
		step.Done = true
		step.Commit = "abc1234"
	}
	return taskdoc.TaskDoc{
		Title:        "saludo",
		Light:        true,
		Requirements: []taskdoc.Requirement{{ID: "R1", Text: "el saludo dice la versión"}},
		Steps:        []taskdoc.Step{step},
	}.Render()
}

// TestALightFeatureIsNotAskedForABrief: the lane's point is that the one line is
// the vision, so the brief is the one question the gate does not ask. Everything
// else is asked exactly as before.
func TestALightFeatureIsNotAskedForABrief(t *testing.T) {
	files := readyFiles()
	delete(files, "brief.md")
	files["odd/tasks/0007-turnos.md"] = lightDoc(false)

	got := run(t, newFakeFiles(files), &fakeGit{tree: currentTree, changed: []string{"a.go"}}, check.Options{})

	if got.Reason == gate.ReasonNoBrief || got.Reason == gate.ReasonBriefIncomplete {
		t.Fatalf("a light feature was asked for a brief: %d/%q (%s)", got.Code, got.Reason, got.Message)
	}

	// The same project without the lane line is asked for it, so the exception is
	// the artifact's, not the tool's.
	plain := readyFiles()
	delete(plain, "brief.md")
	if got := run(t, newFakeFiles(plain), &fakeGit{tree: currentTree, changed: []string{"a.go"}}, check.Options{}); got.Reason != gate.ReasonNoBrief {
		t.Fatalf("a normal feature without a brief = %d/%q, want no_brief", got.Code, got.Reason)
	}
}

// TestALightFeatureStillOwesItsEvidence: the lane skips the brief and nothing
// else. A light feature with an open question, or with a fence violation, is
// refused like any other.
func TestALightFeatureStillOwesItsEvidence(t *testing.T) {
	files := readyFiles()
	delete(files, "brief.md")
	doc := taskdoc.Parse(lightDoc(false)).AppendHistory(`L1 · PREGUNTA P1 · "¿el saludo va en mayúsculas?"`)
	files["odd/tasks/0007-turnos.md"] = doc.Render()

	got := run(t, newFakeFiles(files), &fakeGit{tree: currentTree, changed: []string{"a.go"}}, check.Options{})

	if got.Reason != gate.ReasonOpenQuestion {
		t.Fatalf("a light feature with an open question = %d/%q, want the question", got.Code, got.Reason)
	}

	// A light feature at its close still owes the proof, exactly as any other.
	closed := readyFiles()
	delete(closed, "brief.md")
	delete(closed, "proof/0007-turnos.txt")
	closed["odd/tasks/0007-turnos.md"] = lightDoc(true)
	if got := run(t, newFakeFiles(closed), &fakeGit{tree: currentTree, changed: []string{"a.go"}}, check.Options{}); got.Reason != gate.ReasonProofMissing {
		t.Fatalf("a light feature closed without a proof = %d/%q, want proof_missing", got.Code, got.Reason)
	}
}
