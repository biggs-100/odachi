package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/gate"
	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// briefWithBlock is what `odachi start` writes: the same sentence twice, once in
// the feature's block of the brief and once in the R1 line of the task document.
// The two copies are what the check compares (0314).
const briefWithBlock = checkBrief + `
## 0001 · turnos

## What must exist
- reservar turnos
`

// rewrittenDoc is the closing document with R1 in the agent's words instead of
// the human's: the defect the quote check exists for.
func rewrittenDoc() string {
	return taskdoc.TaskDoc{
		Title:        "turnos",
		Requirements: []taskdoc.Requirement{{ID: "R1", Text: "reservar turnos y cancelarlos"}},
		DoneWhen:     []string{"reservo un turno"},
		Steps:        []taskdoc.Step{{ID: "T1", Text: "un paso", Requirement: "R1", Done: true}},
		History:      []string{`L1 · "reservar turnos"`},
	}.Render()
}

// TestCheckBlocksARequirementTheAgentRewrote: the plan says the human asked for
// something they did not, and the refusal carries both sentences so the fix is a
// copy instead of a hunt through two files.
func TestCheckBlocksARequirementTheAgentRewrote(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)
	writeFile(t, dir, "brief.md", briefWithBlock)
	closeFeature(t, dir)
	writeFile(t, dir, "odd/tasks/0001-turnos.md", rewrittenDoc())

	code, out, errOut := runDir(t, dir, []string{"check", "--json"}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1 (stderr: %q)", code, errOut)
	}

	var payload struct {
		Code    int    `json:"code"`
		Reason  string `json:"reason"`
		Message string `json:"message"`
		Next    string `json:"next"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, out)
	}
	if payload.Reason != gate.ReasonQuoteMismatch {
		t.Fatalf("reason = %q, want %q: %s", payload.Reason, gate.ReasonQuoteMismatch, payload.Message)
	}
	for _, want := range []string{`the brief says "reservar turnos"`, `the task doc says "reservar turnos y cancelarlos"`} {
		if !strings.Contains(payload.Message, want) {
			t.Errorf("the refusal does not carry %q: %s", want, payload.Message)
		}
	}
	if !strings.HasPrefix(payload.Next, "odachi ") {
		t.Errorf("next = %q, want the command that records the human's change", payload.Next)
	}
}

// TestCheckLetsTheQuotesWordsThrough: the same closing document, unrewritten,
// falls through to the next reason. Without this the check could refuse every
// feature and the first test would still pass.
func TestCheckLetsTheQuotesWordsThrough(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)
	writeFile(t, dir, "brief.md", briefWithBlock)
	closeFeature(t, dir)

	code, out, _ := runDir(t, dir, []string{"check", "--json"}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1: the proof is still missing", code)
	}
	var payload struct {
		Reason  string `json:"reason"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, out)
	}
	if payload.Reason == gate.ReasonQuoteMismatch {
		t.Fatalf("the check refused the human's own words: %s", payload.Message)
	}
}
