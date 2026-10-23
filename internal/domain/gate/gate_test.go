package gate_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/gate"
)

// A ready state is the base every blocked state is derived from, so a test
// case that wants to isolate one block only has to break that one thing.
func ready() gate.State {
	return gate.State{
		HasBrief:      true,
		TreeHash:      "4a1c9f",
		ProofTreeHash: "4a1c9f",
		ProofOutput:   "ok github.com/biggs-100/odachi/internal/domain/gate",
	}
}

func TestEvaluationOrder(t *testing.T) {
	noBrief := gate.State{Violations: []string{"pagos"}, OpenQuestions: 3, TreeHash: "4a1c9f"}

	incompleteBrief := ready()
	incompleteBrief.BriefMissing = []string{"vision", "non-goals"}
	incompleteBrief.Violations = []string{"pagos"}
	incompleteBrief.OpenQuestions = 1

	outOfScope := ready()
	outOfScope.Violations = []string{"pagos: internal/payments/stripe.go", "pagos: github.com/stripe/stripe-go/v76"}
	outOfScope.OpenQuestions = 1

	openQuestion := ready()
	openQuestion.OpenQuestions = 1

	missingProof := ready()
	missingProof.ProofTreeHash = ""
	missingProof.ProofOutput = ""

	staleProof := ready()
	staleProof.ProofTreeHash = "9f2b7c"

	failedProof := ready()
	failedProof.ProofFailed = true
	failedProof.ProofLastLine = "FAIL github.com/biggs-100/odachi/internal/domain/gate 0.02s"
	failedProof.ProofCommand = "go test ./..."

	tests := []struct {
		name    string
		state   gate.State
		code    gate.Code
		reason  string
		message string
		next    string
	}{
		{
			name:    "no brief",
			state:   noBrief,
			code:    gate.NeedsDecision,
			reason:  gate.ReasonNoBrief,
			message: "there is no brief: we do not know yet what must exist",
			next:    "odachi start",
		},
		{
			name:    "brief incomplete",
			state:   incompleteBrief,
			code:    gate.NeedsDecision,
			reason:  gate.ReasonBriefIncomplete,
			message: "the brief is incomplete: missing vision, non-goals",
			next:    "odachi start",
		},
		{
			name:    "out of scope",
			state:   outOfScope,
			code:    gate.NeedsDecision,
			reason:  gate.ReasonOutOfScope,
			message: "out of scope: pagos: internal/payments/stripe.go; pagos: github.com/stripe/stripe-go/v76",
			next:    "odachi ask \"<your decision>\"",
		},
		{
			name:    "open question",
			state:   openQuestion,
			code:    gate.NeedsDecision,
			reason:  gate.ReasonOpenQuestion,
			message: "1 unanswered question",
			next:    "odachi ask \"<your decision>\"",
		},
		{
			name:    "proof missing",
			state:   missingProof,
			code:    gate.NeedsWork,
			reason:  gate.ReasonProofMissing,
			message: "the proof for this change is missing",
			next:    "odachi check --proof <file>",
		},
		{
			name:    "proof stale",
			state:   staleProof,
			code:    gate.NeedsWork,
			reason:  gate.ReasonProofStale,
			message: "the proof is from another state of the code",
			next:    "odachi check --proof <file>",
		},
		{
			name:    "proof failed",
			state:   failedProof,
			code:    gate.NeedsWork,
			reason:  gate.ReasonProofFailed,
			message: "the proof failed: FAIL github.com/biggs-100/odachi/internal/domain/gate 0.02s",
			next:    `odachi check --run "go test ./..."`,
		},
		{
			name:    "ready",
			state:   ready(),
			code:    gate.Ready,
			reason:  gate.ReasonReady,
			message: "ready: there is proof for this tree",
			next:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := gate.Evaluate(tc.state)
			if got.Code != tc.code {
				t.Errorf("Code = %d, want %d (0 ready, 1 needs work, 2 needs decision)", got.Code, tc.code)
			}
			if got.Reason != tc.reason {
				t.Errorf("Reason = %q, want %q", got.Reason, tc.reason)
			}
			if got.Message != tc.message {
				t.Errorf("Message = %q, want %q", got.Message, tc.message)
			}
			if got.Next != tc.next {
				t.Errorf("Next = %q, want %q", got.Next, tc.next)
			}
			if strings.Contains(got.Message, "\n") {
				t.Errorf("Message must be one line, got %q", got.Message)
			}
		})
	}
}

func TestEvaluationOrderIsFirstBlockWins(t *testing.T) {
	// Everything is broken at once: the first reason in the documented order
	// must win, so the human is never sent to fix two things at the same time.
	state := gate.State{
		BriefMissing:  []string{"vision"},
		Violations:    []string{"pagos"},
		OpenQuestions: 2,
		TreeHash:      "4a1c9f",
	}

	got := gate.Evaluate(state)
	if got.Reason != gate.ReasonNoBrief {
		t.Fatalf("Reason = %q, want %q (no brief comes first)", got.Reason, gate.ReasonNoBrief)
	}

	state.HasBrief = true
	if got := gate.Evaluate(state); got.Reason != gate.ReasonBriefIncomplete {
		t.Fatalf("Reason = %q, want %q (brief incomplete comes second)", got.Reason, gate.ReasonBriefIncomplete)
	}

	state.BriefMissing = nil
	if got := gate.Evaluate(state); got.Reason != gate.ReasonOutOfScope {
		t.Fatalf("Reason = %q, want %q (out of scope comes third)", got.Reason, gate.ReasonOutOfScope)
	}

	state.Violations = nil
	if got := gate.Evaluate(state); got.Reason != gate.ReasonOpenQuestion {
		t.Fatalf("Reason = %q, want %q (open question comes fourth)", got.Reason, gate.ReasonOpenQuestion)
	}

	state.OpenQuestions = 0
	if got := gate.Evaluate(state); got.Reason != gate.ReasonProofMissing {
		t.Fatalf("Reason = %q, want %q (proof missing comes fifth)", got.Reason, gate.ReasonProofMissing)
	}

	state.ProofOutput = "ok"
	state.ProofTreeHash = "9f2b7c"
	if got := gate.Evaluate(state); got.Reason != gate.ReasonProofStale {
		t.Fatalf("Reason = %q, want %q (proof stale comes sixth)", got.Reason, gate.ReasonProofStale)
	}

	state.ProofTreeHash = state.TreeHash
	state.ProofFailed = true
	state.ProofLastLine = "exit status 1"
	state.ProofCommand = "go test ./..."
	if got := gate.Evaluate(state); got.Reason != gate.ReasonProofFailed {
		t.Fatalf("Reason = %q, want %q (proof failed comes seventh)", got.Reason, gate.ReasonProofFailed)
	}

	state.ProofFailed = false
	if got := gate.Evaluate(state); got.Reason != gate.ReasonReady {
		t.Fatalf("Reason = %q, want %q (ready is last)", got.Reason, gate.ReasonReady)
	}
}

func TestEveryBlockedResultNamesAContinuation(t *testing.T) {
	blocked := []struct {
		name  string
		state gate.State
	}{
		{"no brief", gate.State{}},
		{"brief incomplete", gate.State{HasBrief: true, BriefMissing: []string{"vision"}}},
		{"out of scope", gate.State{HasBrief: true, Violations: []string{"pagos"}}},
		{"open question", gate.State{HasBrief: true, OpenQuestions: 2}},
		{"proof missing", gate.State{HasBrief: true, TreeHash: "4a1c9f"}},
		{"proof stale", func() gate.State { s := ready(); s.ProofTreeHash = "otro"; return s }()},
		{"proof failed", func() gate.State { s := ready(); s.ProofFailed = true; s.ProofCommand = "go test ./..."; return s }()},
	}

	for _, tc := range blocked {
		t.Run(tc.name, func(t *testing.T) {
			got := gate.Evaluate(tc.state)
			if got.Code == gate.Ready {
				t.Fatalf("state %q evaluated as ready, want blocked", tc.name)
			}
			if got.Next == "" {
				t.Fatalf("blocked result %q has no Next: every block must name a command that resolves it", tc.name)
			}
			if !strings.HasPrefix(got.Next, "odachi ") {
				t.Fatalf("Next = %q, want a command starting with %q", got.Next, "odachi ")
			}
		})
	}
}

func TestReadyHasNoContinuation(t *testing.T) {
	got := gate.Evaluate(ready())
	if got.Code != gate.Ready {
		t.Fatalf("Code = %d, want %d", got.Code, gate.Ready)
	}
	if got.Next != "" {
		t.Fatalf("Next = %q, want empty for a ready result", got.Next)
	}
}

func TestOpenQuestionPluralisation(t *testing.T) {
	one := ready()
	one.OpenQuestions = 1

	two := ready()
	two.OpenQuestions = 2

	if got := gate.Evaluate(one).Message; got != "1 unanswered question" {
		t.Fatalf("1 question: Message = %q, want %q", got, "1 unanswered question")
	}
	if got := gate.Evaluate(two).Message; got != "2 unanswered questions" {
		t.Fatalf("2 questions: Message = %q, want %q", got, "2 unanswered questions")
	}
}

// A failed proof names the exact command that produced it, quoted, so the
// human can paste it back and rerun the same test on the same tree.
func TestProofFailedNamesTheSameCommand(t *testing.T) {
	state := ready()
	state.ProofFailed = true
	state.ProofLastLine = "2 tests fallaron"
	state.ProofCommand = `go test ./... -run "TestGate"`

	got := gate.Evaluate(state)

	if got.Code != gate.NeedsWork {
		t.Errorf("Code = %d, want %d (a failing proof is the agent's work)", got.Code, gate.NeedsWork)
	}
	want := `odachi check --run "go test ./... -run \"TestGate\""`
	if got.Next != want {
		t.Errorf("Next = %q, want %q", got.Next, want)
	}
	if !strings.Contains(got.Message, "2 tests fallaron") {
		t.Errorf("Message = %q does not carry the last output line", got.Message)
	}
}

// A command that fails without printing anything still has to say something a
// human can act on, and still has to name a command: no dead ends.
func TestProofFailedWithoutOutputOrCommandStillReadsAsASentence(t *testing.T) {
	state := ready()
	state.ProofFailed = true

	got := gate.Evaluate(state)

	if got.Message != "the proof failed: no output" {
		t.Errorf("Message = %q, want %q", got.Message, "the proof failed: no output")
	}
	if got.Next != `odachi check --run "<command>"` {
		t.Errorf("Next = %q, want the generic --run continuation", got.Next)
	}
}

// A timeout is a failed proof with its own sentence: the human needs to know
// how long the tool waited, not which line was last.
func TestProofTimedOutSaysHowLongItWaited(t *testing.T) {
	state := ready()
	state.ProofTimedOut = true
	state.ProofTimeoutSeconds = 7
	state.ProofCommand = "go test ./..."

	got := gate.Evaluate(state)

	if got.Code != gate.NeedsWork || got.Reason != gate.ReasonProofFailed {
		t.Fatalf("Code/Reason = %d/%q, want %d/%q", got.Code, got.Reason, gate.NeedsWork, gate.ReasonProofFailed)
	}
	if got.Message != "the proof did not finish in 7s" {
		t.Errorf("Message = %q, want %q", got.Message, "the proof did not finish in 7s")
	}
	if got.Next != `odachi check --run "go test ./..." --timeout 14` {
		t.Errorf("Next = %q, want the same command with twice the wait", got.Next)
	}
}

// TestProofTimedOutRaisesTheTimeout: the next command cannot be the same one
// that just timed out; a 120s wait suggests 240s, not another 120s.
func TestProofTimedOutRaisesTheTimeout(t *testing.T) {
	state := ready()
	state.ProofTimedOut = true
	state.ProofTimeoutSeconds = 120
	state.ProofCommand = "go test ./..."

	got := gate.Evaluate(state)

	want := `odachi check --run "go test ./..." --timeout 240`
	if got.Next != want {
		t.Errorf("Next = %q, want %q", got.Next, want)
	}
}

// A command that failed after printing nothing is a failed proof, not a missing
// one: the evidence was produced, it just says the command did not work.
func TestFailedProofWithNoOutputIsStillAFailedProof(t *testing.T) {
	state := ready()
	state.ProofOutput = ""
	state.ProofFailed = true
	state.ProofCommand = "true"

	got := gate.Evaluate(state)

	if got.Reason != gate.ReasonProofFailed {
		t.Fatalf("Reason = %q, want %q (a silent failure is a failure)", got.Reason, gate.ReasonProofFailed)
	}
}

func TestViolationsAreListed(t *testing.T) {
	state := ready()
	state.Violations = []string{
		"pagos: internal/payments/stripe.go",
		"pagos: github.com/stripe/stripe-go/v76",
	}

	message := gate.Evaluate(state).Message
	for _, violation := range state.Violations {
		if !strings.Contains(message, violation) {
			t.Errorf("Message %q does not mention violation %q", message, violation)
		}
	}
}

func TestCodeConstantsMatchTheCheckContract(t *testing.T) {
	if gate.Ready != 0 {
		t.Errorf("Ready = %d, want 0", gate.Ready)
	}
	if gate.NeedsWork != 1 {
		t.Errorf("NeedsWork = %d, want 1", gate.NeedsWork)
	}
	if gate.NeedsDecision != 2 {
		t.Errorf("NeedsDecision = %d, want 2", gate.NeedsDecision)
	}
}
