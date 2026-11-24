package check_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/check"
	"github.com/biggs-100/odachi/internal/domain/gate"
	domainverify "github.com/biggs-100/odachi/internal/domain/verify"
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

// The strict lane refuses a declaration that cannot tell the reviewer what the
// change can hurt. A valid declaration moves on to the independent verification
// and review questions; it does not pass those questions by itself.
func TestStrictRiskDeclarationRefusesInvalidLines(t *testing.T) {
	const invalid = "risk_declaration_invalid"
	tests := []struct {
		name       string
		messages   []string
		paths      []string
		wantReason string
		wantWords  []string
	}{
		{
			name:       "missing line",
			messages:   []string{"add the session\n"},
			paths:      []string{fullPath},
			wantReason: invalid,
			wantWords:  []string{"no commit declares a Risk: line", "amend the commit", "Risk: none - <reason>"},
		},
		{
			name:       "bare none",
			messages:   []string{"add the session\n\nRisk: none\n"},
			paths:      []string{fullPath},
			wantReason: invalid,
			wantWords:  []string{"written reason", "amend the commit", "Risk: none - <reason>"},
		},
		{
			name:       "unknown risk",
			messages:   []string{"add the session\n\nRisk: mystery\n"},
			paths:      []string{fullPath},
			wantReason: invalid,
			wantWords:  []string{"unknown risk", "mystery", "amend the commit"},
		},
		{
			name:       "reasoned none",
			messages:   []string{"add the session\n\nRisk: none - only the session reader changes\n"},
			paths:      []string{fullPath},
			wantReason: check.ReasonReviewMissing,
		},
		{
			name:       "valid named risk",
			messages:   []string{"add the session\n\nRisk: security\n"},
			paths:      []string{fullPath},
			wantReason: check.ReasonReviewMissing,
		},
		{
			name:       "passive none",
			messages:   []string{"write the guide\n\nRisk: none\n"},
			paths:      []string{"README.md"},
			wantReason: gate.ReasonReady,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			files := newFakeFiles(riskFiles())
			// A named risk still owes an independent verifier. Seed its record so
			// this table judges the declaration, not the next gate.
			files.files[verifyPath(currentTree)] = []byte(
				verificationFile(t, currentTree, "otro/modelo", domainverify.Verified, nil))
			git := &fakeGit{tree: currentTree, between: tc.paths, messages: tc.messages}

			got := run(t, files, git, check.Options{ReviewMode: true})

			if got.Reason != tc.wantReason {
				t.Fatalf("Reason = %q, want %q (code %d, message %q)", got.Reason, tc.wantReason, got.Code, got.Message)
			}
			if tc.wantReason == invalid {
				if got.Code != int(gate.NeedsWork) || got.Next != "odachi check" {
					t.Errorf("invalid risk = %d, next %q; want needs work and odachi check", got.Code, got.Next)
				}
				for _, want := range tc.wantWords {
					if !strings.Contains(got.Message, want) {
						t.Errorf("message %q does not name %q", got.Message, want)
					}
				}
			}
		})
	}
}

func TestStrictRiskDeclarationRefusesUnreadableMessages(t *testing.T) {
	files := newFakeFiles(riskFiles())
	git := &fakeGit{
		tree:        currentTree,
		between:     []string{fullPath},
		messagesErr: errors.New("git roto"),
	}

	got := run(t, files, git, check.Options{ReviewMode: true})

	if got.Reason != "risk_declaration_invalid" {
		t.Fatalf("Reason = %q, want risk_declaration_invalid (%s)", got.Reason, got.Message)
	}
	if !strings.Contains(got.Message, "could not read commit messages") {
		t.Errorf("message %q does not name the unreadable messages", got.Message)
	}
}

// Outside strict mode the declaration stays a note, but an unknown name and a
// bare none must be named explicitly. A written none reason answers the question.
func TestOrdinaryRiskNoteNamesInvalidDeclarations(t *testing.T) {
	tests := []struct {
		name     string
		messages []string
		want     []string
		silent   bool
	}{
		{
			name:     "unknown risk",
			messages: []string{"add the session\n\nRisk: mystery\n"},
			want:     []string{"unknown risk", "mystery"},
		},
		{
			name:     "bare none",
			messages: []string{"add the session\n\nRisk: none\n"},
			want:     []string{"written reason", "Risk: none - <reason>"},
		},
		{
			name:     "reasoned none",
			messages: []string{"add the session\n\nRisk: none - only the session reader changes\n"},
			silent:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			git := &fakeGit{tree: currentTree, between: []string{fullPath}, messages: tc.messages}
			got := run(t, newFakeFiles(riskFiles()), git, check.Options{})
			if got.Code != int(gate.Ready) {
				t.Fatalf("Code = %d, want ready (%s)", got.Code, got.Message)
			}
			joined := strings.Join(got.Lines, "\n")
			if tc.silent {
				if strings.Contains(joined, "classifies this change") {
					t.Fatalf("a reasoned none still produced the note:\n%s", joined)
				}
				return
			}
			for _, want := range tc.want {
				if !strings.Contains(joined, want) {
					t.Errorf("the ready output does not say %q:\n%s", want, joined)
				}
			}
		})
	}
}
