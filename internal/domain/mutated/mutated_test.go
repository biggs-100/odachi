package mutated_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/mutated"
)

const tree = "4a1c9f7d2e3b5a6c8d9e0f1a2b3c4d5e6f7a8b9c"

// TestClassifyIsTheWholeRule: four answers and what each one means. Green is the
// block the row exists for; a copy that does not build is not an answer; and a red
// that does not name the expected text is a red for another cause, which is not
// evidence about this guard (0297).
func TestClassifyIsTheWholeRule(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		exit     int
		output   string
		expected string
		want     mutated.Verdict
	}{
		{"the guard fires", 1, "--- FAIL: TestTheGuard (0.00s)\n    the refusal names no way out\nFAIL", "the refusal names no way out", mutated.Noticed},
		{"the suite stays green", 0, "ok  github.com/x 0.01s", "anything", mutated.Missed},
		{"the copy does not build", 1, "# github.com/x\n[build failed]", "anything", mutated.Unknown},
		{"the run never happened", -1, "", "anything", mutated.Unknown},
		{"a red for another cause", 1, "--- FAIL: TestSomethingElse\n    want 2, got 3", "the refusal names no way out", mutated.WrongReason},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := mutated.Classify(tc.exit, tc.output, tc.expected); got != tc.want {
				t.Errorf("Classify = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestRenderAndParseRoundTrip: the file carries everything a reader needs to re-run
// the question -the tree, the command, the file, the string, what the failure had to
// say- and the output verbatim, headers included (0297).
func TestRenderAndParseRoundTrip(t *testing.T) {
	t.Parallel()

	evidence := mutated.Evidence{
		Tree:     tree,
		Command:  "go test ./...",
		File:     "internal/domain/gate/gate.go",
		Removed:  `ReasonFuture = "future"`,
		Expected: "the gate declares",
		Verdict:  mutated.Noticed,
		Note:     "removed once",
		Output:   "# a line that looks like a header stays in the quotation\nFAIL\n",
	}

	parsed, err := mutated.Parse(mutated.Render(evidence))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if parsed != evidence {
		t.Errorf("Parse = %+v, want %+v", parsed, evidence)
	}
}

// TestParseRefusesTheOtherRowsByName: a proof and an undone file are different
// questions, and a row read as another row is evidence nobody can trust (0297).
func TestParseRefusesTheOtherRowsByName(t *testing.T) {
	t.Parallel()

	for name, text := range map[string]string{
		"proof":            "# tree " + tree + "\n# command go test ./...\n# output\nok\n",
		"undone":           "# undone " + tree + "\n# command go test ./...\n# verdict noticed\n# output\nFAIL\n",
		"headers in order": "# mutated " + tree + "\n# file a.go\n# command go test ./...\n",
	} {
		if _, err := mutated.Parse(text); err == nil {
			t.Errorf("%s: Parse accepted a file that is not a mutation row", name)
		}
	}
}

// TestAParseRefusesAVerdictItDoesNotKnow: a typo in the verdict would be read as a
// pass by any reader that only asks whether it is "noticed" (0297).
func TestAParseRefusesAVerdictItDoesNotKnow(t *testing.T) {
	t.Parallel()

	text := mutated.Render(mutated.Evidence{
		Tree: tree, Command: "go test ./...", File: "a.go", Removed: "x", Expected: "y", Verdict: mutated.Verdict("fine"),
	})
	if _, err := mutated.Parse(text); err == nil {
		t.Fatal("Parse accepted a verdict nobody writes")
	}
}

// TestNoTextIsAVerdictAndNotASilence: a string that is not in the file means the
// question did not apply, and the file says so instead of leaving a row that looks
// like a pass (0297).
func TestNoTextIsAVerdictAndNotASilence(t *testing.T) {
	t.Parallel()

	text := mutated.Render(mutated.Evidence{
		Tree: tree, Command: "go test ./...", File: "a.go", Removed: "no existe", Expected: "nada",
		Verdict: mutated.NoText, Note: "the string is not in the file",
	})
	parsed, err := mutated.Parse(text)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if parsed.Verdict != mutated.NoText {
		t.Errorf("Verdict = %q, want no_text", parsed.Verdict)
	}
	if !strings.Contains(parsed.Note, "not in the file") {
		t.Errorf("Note = %q, want the cause", parsed.Note)
	}
}
