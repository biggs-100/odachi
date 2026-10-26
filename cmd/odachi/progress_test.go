package main

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// headSHA is the commit the next change starts from.
func headSHA(t *testing.T, dir string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", dir, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatalf("git rev-parse HEAD: %v", err)
	}
	return strings.TrimSpace(string(out))
}

// featureAtBase is a repository where the feature has started: the task document
// names its base, its one step is still open, and the review switch is on.
func featureAtBase(t *testing.T) (dir, base string) {
	t.Helper()
	dir = t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)
	writeFile(t, dir, ".odachi/review.toml", "mode = \"on\"\n")
	commitAll(t, dir, "the project")
	base = headSHA(t, dir)

	writeFile(t, dir, "odd/tasks/0001-turnos.md", taskdoc.TaskDoc{
		Title:        "turnos",
		Base:         base,
		Requirements: []taskdoc.Requirement{{ID: "R1", Text: "reservar turnos"}},
		DoneWhen:     []string{"reservo un turno"},
		Steps: []taskdoc.Step{{
			ID: "T1", Text: "reservar un turno", Requirement: "R1",
		}},
		History: []string{`L1 · "reservar turnos"`},
	}.Render())
	commitAll(t, dir, "the plan")
	return dir, base
}

// TestProgressJSONIsTheContract is the shape another program reads: the report is
// not only lines. The names here are the contract, so a renderer in another
// language can depend on them without parsing English.
func TestProgressJSONIsTheContract(t *testing.T) {
	t.Parallel()
	dir, _ := featureAtBase(t)

	code, out, errOut := runDir(t, dir, []string{"progress", "--json"}, "")
	if code != 0 {
		t.Fatalf("progress --json: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}

	var got struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		State   *struct {
			Number        int    `json:"number"`
			Slug          string `json:"slug"`
			OpenQuestions int    `json:"open_questions"`
			Steps         []struct {
				ID     string `json:"id"`
				Text   string `json:"text"`
				Done   bool   `json:"done"`
				Commit string `json:"commit"`
			} `json:"steps"`
			Proof struct {
				State string `json:"state"`
				Path  string `json:"path"`
			} `json:"proof"`
			Review struct {
				State string `json:"state"`
				Why   string `json:"why"`
				Base  string `json:"base"`
			} `json:"review"`
		} `json:"state"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("the machine output is not JSON: %v\n%s", err, out)
	}
	if got.Code != 0 || got.State == nil {
		t.Fatalf("output = %+v, want a state and code 0", got)
	}
	state := got.State
	if state.Number != 1 || state.Slug != "turnos" {
		t.Errorf("feature = %04d-%s, want 0001-turnos", state.Number, state.Slug)
	}
	if len(state.Steps) != 1 || state.Steps[0].ID != "T1" || state.Steps[0].Done {
		t.Errorf("steps = %+v, want one pending T1", state.Steps)
	}
	if state.Proof.State != "none" {
		t.Errorf("proof = %+v, want none yet", state.Proof)
	}
	if state.Review.State != "not_due" || state.Review.Why != "feature_open" {
		t.Errorf("review = %+v, want not due because the feature is open", state.Review)
	}

	// The key names are the contract too: a field that vanishes from the object is
	// as broken as one that changes value.
	var keys struct {
		State map[string]json.RawMessage `json:"state"`
	}
	if err := json.Unmarshal([]byte(out), &keys); err != nil {
		t.Fatalf("the machine output is not JSON: %v", err)
	}
	for _, key := range []string{"number", "slug", "steps", "open_questions", "proof", "review"} {
		if _, ok := keys.State[key]; !ok {
			t.Errorf("the state does not carry %q: %s", key, out)
		}
	}
	if !strings.Contains(got.Message, "0 of 1") {
		t.Errorf("message = %q, want the same headline the human reads", got.Message)
	}
}

// TestProgressAndCheckAgreeAboutTheReview is the wiring test: the report asks the
// gate's own rule, connected the way main connects it. A stubbed answer in a unit
// test cannot catch a wrong base, and a wrong base is exactly what made the two
// verbs disagree on a released build.
func TestProgressAndCheckAgreeAboutTheReview(t *testing.T) {
	t.Parallel()
	dir, base := featureAtBase(t)

	// The close: the document that marks the last step done, not committed yet.
	// Both verbs owe the verdict of the feature from here.
	writeFile(t, dir, "odd/tasks/0001-turnos.md", taskdoc.TaskDoc{
		Title:        "turnos",
		Base:         base,
		Requirements: []taskdoc.Requirement{{ID: "R1", Text: "reservar turnos"}},
		DoneWhen:     []string{"reservo un turno"},
		Steps: []taskdoc.Step{{
			ID: "T1", Text: "reservar un turno", Requirement: "R1",
			Done: true, Commit: "abc1234",
		}},
		History: []string{`L1 · "reservar turnos"`},
	}.Render())

	code, out, errOut := runDir(t, dir, []string{"progress"}, "")
	if code != 0 {
		t.Fatalf("progress: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if !strings.Contains(out, "review: owed") || !strings.Contains(out, base) {
		t.Fatalf("progress does not owe the verdict of the close:\n%s", out)
	}
	if code, _, errOut := runDir(t, dir, []string{"check", "--run", "echo ok"}, ""); code == 0 {
		t.Fatalf("check was ready at the close: %s", errOut)
	}

	// The close lands and the release tail follows: documents only. The gate
	// exempts it, and the report has to say the same. With an empty base every
	// file looks like code, which is the disagreement this test exists to catch.
	commitAll(t, dir, "the close")
	writeFile(t, dir, "CHANGELOG.md", "## v0.0.1\n\nThe first release.\n")
	commitAll(t, dir, "release")

	code, out, errOut = runDir(t, dir, []string{"progress"}, "")
	if code != 0 {
		t.Fatalf("progress after the release: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if !strings.Contains(out, "review: not due") {
		t.Fatalf("progress owes a verdict on a documents-only tail:\n%s", out)
	}
	if code, _, errOut := runDir(t, dir, []string{"check", "--run", "echo ok"}, ""); code != 0 {
		t.Fatalf("check asked for the verdict on a documents-only tail (code %d): %s", code, errOut)
	}
}
