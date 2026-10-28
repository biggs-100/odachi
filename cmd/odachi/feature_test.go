package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// TestStartHelpNamesTheWaiverLane: the refusal for an unattributed verdict sends a
// human to `odachi start --waive-review`, and a lane the help does not mention is a
// command nobody finds. The flag existed before this line; the help did not say so.
func TestStartHelpNamesTheWaiverLane(t *testing.T) {
	t.Parallel()

	code, out, errOut := runDir(t, t.TempDir(), []string{"start", "--help"}, "")
	if code != 0 {
		t.Fatalf("start --help: code = %d (stderr: %q)", code, errOut)
	}
	if !strings.Contains(out, "--waive-review") {
		t.Errorf("the help does not name the waiver lane:\n%s", out)
	}
}

// TestTheReviewRefusesWhatTheProjectWillNotPayFor: the limit lives in the
// project's file, and the CLI is what carries it to the app. Without this the app
// could enforce a limit no project's file ever reaches, or enforce none at all,
// and every package test would stay green.
func TestTheReviewRefusesWhatTheProjectWillNotPayFor(t *testing.T) {
	t.Parallel()
	dir := changedProject(t)
	writeFile(t, dir, "odd/review.toml", "mode = \"on\"\nmax_prompt_kb = 1\n")

	code, _, errOut := runDir(t, dir, []string{"review", "--exec", "echo ok"}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1 (stderr: %q)", code, errOut)
	}
	for _, want := range []string{"max_prompt_kb", "no reviewer was started"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("stderr %q does not carry %q", errOut, want)
		}
	}
}

// TestAFeatureCanBeChosenByItsNumber: two features live in one tree, and the
// writer acts on the one named instead of the highest. Without the flag nothing
// changes: the highest is the feature the human is working on now.
func TestAFeatureCanBeChosenByItsNumber(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)
	writeFile(t, dir, "odd/tasks/0002-otro.md", taskdoc.TaskDoc{
		Title:        "otro",
		Requirements: []taskdoc.Requirement{{ID: "R1", Text: "otro"}},
		DoneWhen:     []string{"otro"},
		History:      []string{`L1 · "otro"`},
	}.Render())
	commitAll(t, dir, "two features")

	// The writer with the flag writes the lower one.
	if code, out, errOut := runDir(t, dir, []string{"plan", "un paso", "--feature", "1"}, ""); code != 0 {
		t.Fatalf("plan --feature 1: %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if !strings.Contains(featureText(t, dir, "0001-turnos.md"), "un paso") {
		t.Error("0001 did not get the step")
	}
	if strings.Contains(featureText(t, dir, "0002-otro.md"), "un paso") {
		t.Error("0002 got a step it was not asked for")
	}

	// A question goes to the same one. The documented shape puts the question
	// first and the flags after it.
	if code, out, errOut := runDir(t, dir, []string{"ask", "¿y esto?", "--feature", "1"}, ""); code != 0 {
		t.Fatalf("ask --feature 1: %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if !strings.Contains(featureText(t, dir, "0001-turnos.md"), "¿y esto?") {
		t.Error("0001 did not get the question")
	}

	// No flag keeps today's answer: the highest.
	if code, out, errOut := runDir(t, dir, []string{"plan", "otro paso"}, ""); code != 0 {
		t.Fatalf("plan: %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if !strings.Contains(featureText(t, dir, "0002-otro.md"), "otro paso") {
		t.Error("without the flag the highest did not get the step")
	}

	// A number the tree does not have names the ones it does.
	code, _, errOut := runDir(t, dir, []string{"plan", "x", "--feature", "9"}, "")
	if code != 1 {
		t.Fatalf("plan --feature 9 = %d, want 1", code)
	}
	for _, want := range []string{"0009", "0001-turnos", "0002-otro"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("stderr %q does not name %q", errOut, want)
		}
	}

	// The gate judges the one named: 0001 holds the open question, 0002 does not.
	code, _, errOut = runDir(t, dir, []string{"check", "--feature", "1"}, "")
	if code != 2 {
		t.Fatalf("check --feature 1 = %d, want 2: its own question blocks", code)
	}
	if !strings.Contains(errOut, "¿y esto?") {
		t.Errorf("check --feature 1 does not name its question: %q", errOut)
	}
	code, _, errOut = runDir(t, dir, []string{"check", "--feature", "2"}, "")
	if code != 1 {
		t.Fatalf("check --feature 2 = %d, want 1: its proof is missing", code)
	}
	if strings.Contains(errOut, "¿y esto?") {
		t.Errorf("check --feature 2 judged 0001's question: %q", errOut)
	}

	// The gate refuses a number the tree does not have, and a negative one is not
	// a way to reach the default.
	code, _, errOut = runDir(t, dir, []string{"check", "--feature", "9"}, "")
	if code != 1 {
		t.Fatalf("check --feature 9 = %d, want 1: the number does not exist", code)
	}
	if !strings.Contains(errOut, "the ones odachi reads are") {
		t.Errorf("check --feature 9 does not name the features: %q", errOut)
	}
	if code, _, _ := runDir(t, dir, []string{"plan", "x", "--feature", "-1"}, ""); code != 1 {
		t.Errorf("plan --feature -1 = %d, want 1: it is unknown, not the default", code)
	}

	// The report describes the one named, as data too.
	code, out, errOut := runDir(t, dir, []string{"progress", "--feature", "1", "--json"}, "")
	if code != 0 {
		t.Fatalf("progress --feature 1: %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	var machine struct {
		State struct {
			Number int `json:"number"`
		} `json:"state"`
	}
	if err := json.Unmarshal([]byte(out), &machine); err != nil {
		t.Fatalf("progress --json is not JSON: %v\n%s", err, out)
	}
	if machine.State.Number != 1 {
		t.Errorf("progress --feature 1 reports feature %d, want 1", machine.State.Number)
	}
}

// featureText reads one task document the test wrote.
func featureText(t *testing.T, dir, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "odd", "tasks", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(data)
}

// TestANamedFeatureOnAProjectWithoutDocsIsRefused: a project before its first
// feature can still be verified -there is nothing to ask about-, but asking for a
// number there is refused exactly like any other unknown number. The gate is the
// last place a selector may answer nothing.
func TestANamedFeatureOnAProjectWithoutDocsIsRefused(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)
	if err := os.Remove(filepath.Join(dir, "odd", "tasks", "0001-turnos.md")); err != nil {
		t.Fatalf("remove the task doc: %v", err)
	}

	code, _, errOut := runDir(t, dir, []string{"check", "--feature", "9"}, "")

	if code != 1 {
		t.Fatalf("check --feature 9 without docs = %d, want 1 (stderr: %q)", code, errOut)
	}
	if !strings.Contains(errOut, "there is no task doc") {
		t.Errorf("stderr %q does not name the missing document", errOut)
	}
}

// TestStartWritesTheWaiverOnlyWhenAsked: the waiver is the human's decision, so
// it is written when the flag is there and absent when it is not. A document
// written without it must parse as a document without it.
func TestStartWritesTheWaiverOnlyWhenAsked(t *testing.T) {
	t.Parallel()
	base := []string{
		"start", "--force",
		"--outcome", "the waiver is recorded",
		"--audience", "cpu",
		"--done", "the waiver is written",
		"--sensitive", "none",
		"--deadline", "today",
		"--learn", "how a recorded exception stays honest",
	}

	waived := t.TempDir()
	initRepo(t, waived)
	withFlag := append(append([]string(nil), base[:2]...), append([]string{"--waive-review"}, base[2:]...)...)
	if code, out, errOut := runDir(t, waived, withFlag, ""); code != 0 {
		t.Fatalf("start --waive-review: %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if !strings.Contains(featureText(t, waived, "0001-the-waiver-is.md"), "Review: waived") {
		t.Error("the document does not carry the waiver")
	}

	plain := t.TempDir()
	initRepo(t, plain)
	if code, out, errOut := runDir(t, plain, base, ""); code != 0 {
		t.Fatalf("start: %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if strings.Contains(featureText(t, plain, "0001-the-waiver-is.md"), "Review: waived") {
		t.Error("a document nobody waived carries the waiver")
	}
}

// TestAClosedFeatureWithAWaiverIsReadyAndNamed: the waiver removes the verdict
// and nothing else. The proof is still the gate, the fence still applies, and the
// ready tree says the review was waived so nobody reads the silence as one.
func TestAClosedFeatureWithAWaiverIsReadyAndNamed(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)
	commitAll(t, dir, "the project")
	base := headSHA(t, dir)

	writeFile(t, dir, "odd/tasks/0001-turnos.md", taskdoc.TaskDoc{
		Title:        "turnos",
		Base:         base,
		ReviewWaived: true,
		Requirements: []taskdoc.Requirement{{ID: "R1", Text: "reservar turnos"}},
		DoneWhen:     []string{"reservo un turno"},
		Steps:        []taskdoc.Step{{ID: "T1", Text: "reservar un turno", Requirement: "R1", Done: true, Commit: base}},
		History:      []string{`L1 · "reservar turnos"`},
	}.Render())

	// --require-review is the strongest demand a run can make, and the recorded
	// decision outranks it: that is the lane.
	code, out, errOut := runDir(t, dir, []string{"check", "--run", "echo ok", "--require-review"}, "")
	if code != 0 {
		t.Fatalf("a waived feature is ready: %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if !strings.Contains(out, "review: waived by the human") {
		t.Errorf("the ready tree does not name the waiver: %q", out)
	}

	// The proof is still required: the waiver removes the verdict, not the evidence.
	if err := os.Remove(filepath.Join(dir, "proof", "0001-turnos.txt")); err != nil {
		t.Fatalf("remove the proof: %v", err)
	}
	if code, _, _ := runDir(t, dir, []string{"check", "--require-review"}, ""); code != 1 {
		t.Errorf("without its proof a waived feature = %d, want 1", code)
	}
}
