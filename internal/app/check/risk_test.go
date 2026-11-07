package check_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/check"
	"github.com/biggs-100/odachi/internal/domain/gate"
)

// riskBase is the base a feature that carries code starts from: the note measures
// the diff the review will judge, so the document has to name one.
const riskBase = "9a211be952e250bfc7adc0acdbe496c88143ff1f"

// fullPath is a path the classifier calls full: auth decides who reads what.
const fullPath = "internal/domain/auth/session.go"

// focusedPath is ordinary code: it asks for one lens, and the note is not about it.
const focusedPath = "internal/app/check/check.go"

// risktFiles is the closed feature of the shared fixtures with the base its diff
// starts from.
func riskFiles() map[string]string {
	files := readyFiles()
	files["odd/tasks/0007-turnos.md"] = reviewableDoc(riskBase, true)
	return files
}

// TestTheReadyOutputNamesAFullChangeThatDeclaresNoRisk: the commits say `none` and
// the paths say auth, so the reviewer reads the contrast; the gate stays ready,
// because the note is not a verdict.
func TestTheReadyOutputNamesAFullChangeThatDeclaresNoRisk(t *testing.T) {
	git := &fakeGit{
		tree:     currentTree,
		between:  []string{fullPath},
		messages: []string{"add the session\n\nRisk: none\n"},
	}

	got := run(t, newFakeFiles(riskFiles()), git, check.Options{})

	if got.Code != int(gate.Ready) {
		t.Fatalf("Code = %d, want ready (message %q)", got.Code, got.Message)
	}
	joined := strings.Join(got.Lines, "\n")
	for _, want := range []string{"classifies this change as full", fullPath, "Risk: none", "name the risk"} {
		if !strings.Contains(joined, want) {
			t.Errorf("the ready output does not say %q:\n%s", want, joined)
		}
	}
}

// TestAFullChangeWithNoRiskLineIsNamed: a missing line is silence, and silence is
// the same claim of harmlessness as `none`; the note says which one happened so
// the reviewer knows what to fix.
func TestAFullChangeWithNoRiskLineIsNamed(t *testing.T) {
	git := &fakeGit{
		tree:     currentTree,
		between:  []string{fullPath},
		messages: []string{"add the session\n"},
	}

	got := run(t, newFakeFiles(riskFiles()), git, check.Options{})

	if got.Code != int(gate.Ready) {
		t.Fatalf("Code = %d, want ready (message %q)", got.Code, got.Message)
	}
	joined := strings.Join(got.Lines, "\n")
	for _, want := range []string{"classifies this change as full", fullPath, "no commit declares a Risk: line"} {
		if !strings.Contains(joined, want) {
			t.Errorf("the ready output does not say %q:\n%s", want, joined)
		}
	}
}

// TestANamedRiskKeepsTheReadyOutputSilent: the human answered the question the
// paths asked, and a note that keeps firing after the answer is noise.
func TestANamedRiskKeepsTheReadyOutputSilent(t *testing.T) {
	git := &fakeGit{
		tree:     currentTree,
		between:  []string{fullPath},
		messages: []string{"add the session\n\nRisk: security\n"},
	}

	got := run(t, newFakeFiles(riskFiles()), git, check.Options{})

	if got.Code != int(gate.Ready) {
		t.Fatalf("Code = %d, want ready (message %q)", got.Code, got.Message)
	}
	if joined := strings.Join(got.Lines, "\n"); strings.Contains(joined, "classifies this change") {
		t.Fatalf("a named risk still produced the note:\n%s", joined)
	}
}

// TestAFocusedChangeKeepsTheReadyOutputSilent: ordinary code asks for one lens,
// and the Risk: line is about the four.
func TestAFocusedChangeKeepsTheReadyOutputSilent(t *testing.T) {
	git := &fakeGit{
		tree:     currentTree,
		between:  []string{focusedPath},
		messages: []string{"add the note\n\nRisk: none\n"},
	}

	got := run(t, newFakeFiles(riskFiles()), git, check.Options{})

	if got.Code != int(gate.Ready) {
		t.Fatalf("Code = %d, want ready (message %q)", got.Code, got.Message)
	}
	if joined := strings.Join(got.Lines, "\n"); strings.Contains(joined, "classifies this change") {
		t.Fatalf("a focused change produced the note:\n%s", joined)
	}
}

// TestAnUnreadableRiskLineIsNamed: a declaration with no name answered nothing,
// and the note must not let the broken line pass as the answer.
func TestAnUnreadableRiskLineIsNamed(t *testing.T) {
	git := &fakeGit{
		tree:     currentTree,
		between:  []string{fullPath},
		messages: []string{"add the session\n\nRisk:   \n"},
	}

	got := run(t, newFakeFiles(riskFiles()), git, check.Options{})

	if got.Code != int(gate.Ready) {
		t.Fatalf("Code = %d, want ready (message %q)", got.Code, got.Message)
	}
	joined := strings.Join(got.Lines, "\n")
	for _, want := range []string{"classifies this change as full", "cannot be read"} {
		if !strings.Contains(joined, want) {
			t.Errorf("the ready output does not say %q:\n%s", want, joined)
		}
	}
}
