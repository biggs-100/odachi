package gate_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"strconv"
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

// evaluationCase is one answer the gate gives: the state that produces it, and the
// words a human reads when it does.
type evaluationCase struct {
	name    string
	state   gate.State
	code    gate.Code
	reason  string
	message string
	next    string
	kind    gate.TransitionKind
}

// evaluationCases is the one list of the gate's answers: every reason it can return,
// the state that produces it, and the continuation that resolves it. The exact words
// and the property "every block names a command" are both read from here, because two
// hand lists that must both grow is one list that gets forgotten (0293).
var evaluationCases = buildEvaluationCases()

// buildEvaluationCases names the states once, so a case that isolates one block only
// breaks that one thing.
func buildEvaluationCases() []evaluationCase {
	noBrief := gate.State{Violations: []string{"pagos"}, OpenQuestions: 3, TreeHash: "4a1c9f"}

	incompleteBrief := ready()
	incompleteBrief.BriefMissing = []string{"vision", "non-goals"}
	incompleteBrief.Violations = []string{"pagos"}
	incompleteBrief.OpenQuestions = 1

	outOfScope := ready()
	outOfScope.Violations = []string{"pagos: internal/payments/stripe.go", "pagos: github.com/stripe/stripe-go/v76"}
	outOfScope.OpenQuestions = 1

	loosened := ready()
	loosened.FenceLoosened = []string{"monorepo scopes: the rule is gone"}
	loosened.OpenQuestions = 1

	openQuestion := ready()
	openQuestion.OpenQuestions = 1

	// A requirement nobody planned outranks the proof: planning the step is what
	// the agent can still do, and the proof is owed after it.
	unplanned := ready()
	unplanned.Unplanned = []string{"R2"}
	unplanned.ProofTreeHash = ""

	// Two of them: the sentence counts, joins and says "them". The branch is a
	// different sentence and needs its own row.
	twoUnplanned := ready()
	twoUnplanned.Unplanned = []string{"R2", "R3"}
	twoUnplanned.ProofTreeHash = ""

	quoteMismatch := ready()
	quoteMismatch.QuoteMismatch = []string{`R1: the brief says "un paso" and the task doc says "dos pasos"`}

	// A risk asked for a verifier and no record answers this tree; the answer that
	// a record exists and says the requirement does not hold is its own row, because
	// "there is no record" would be a lie about a file that is there (0315).
	verificationMissing := ready()
	verificationMissing.VerificationMissing = []string{"contract"}

	verificationRefuted := ready()
	verificationRefuted.VerificationRefuted = []string{"the gate closes a tree with no record"}

	// A close that owes a verifier measures its candidate before spending it: past
	// the slice the record would read a tree nobody can finish, and the refusal
	// carries both numbers so the human sees how far past the limit it is (0331).
	// The record is the other half of the state on purpose: the slice exists only
	// while nobody has signed the tree, and a close that already has its record is
	// not paying for a verifier at all.
	verificationSlice := ready()
	verificationSlice.VerificationLines = 401
	verificationSlice.VerificationMissing = []string{"contract"}

	missingProof := ready()
	missingProof.ProofTreeHash = ""
	missingProof.ProofOutput = ""

	staleProof := ready()
	staleProof.ProofTreeHash = "9f2b7c"

	failedProof := ready()
	failedProof.ProofFailed = true
	failedProof.ProofLastLine = "FAIL github.com/biggs-100/odachi/internal/domain/gate 0.02s"
	failedProof.ProofCommand = "go test ./..."

	undoMissed := ready()
	undoMissed.UndoMissed = true
	undoMissed.UndoCommand = "go test ./..."

	undoMissing := ready()
	undoMissing.UndoMissing = true
	undoMissing.ProofCommand = "go test ./..."

	// The mutation row names the question it asked, so the continuation can ask it
	// again instead of sending the human to a generic command (0297).
	mutationFailed := ready()
	mutationFailed.MutationFailed = true
	mutationFailed.MutationNote = "the suite stayed green with the guard taken out"
	mutationFailed.MutationCommand = "go test ./..."
	mutationFailed.MutationFile = "internal/domain/gate/gate.go"
	mutationFailed.MutationRemoved = "ReasonFuture"
	mutationFailed.MutationExpected = "the gate declares"

	return []evaluationCase{
		{
			name:    "no brief",
			kind:    gate.TransitionCollect,
			state:   noBrief,
			code:    gate.NeedsDecision,
			reason:  gate.ReasonNoBrief,
			message: "there is no brief: we do not know yet what must exist",
			next:    "odachi start",
		},
		{
			name:    "brief incomplete",
			kind:    gate.TransitionCollect,
			state:   incompleteBrief,
			code:    gate.NeedsDecision,
			reason:  gate.ReasonBriefIncomplete,
			message: "the brief is incomplete: missing vision, non-goals",
			next:    "odachi start",
		},
		{
			name:    "out of scope",
			kind:    gate.TransitionCollect,
			state:   outOfScope,
			code:    gate.NeedsDecision,
			reason:  gate.ReasonOutOfScope,
			message: "out of scope: pagos: internal/payments/stripe.go; pagos: github.com/stripe/stripe-go/v76",
			next:    "odachi ask \"<your decision>\"",
		},
		{
			name:    "fence loosened",
			kind:    gate.TransitionCollect,
			state:   loosened,
			code:    gate.NeedsDecision,
			reason:  gate.ReasonFenceLoosened,
			message: "the fence was loosened without new non-goals in the brief: monorepo scopes: the rule is gone; the fence may only change with the brief's non-goals. Restore it, or write the non-goals first",
			next:    "odachi ask \"<your decision>\"",
		},
		{
			name:    "open question",
			kind:    gate.TransitionCollect,
			state:   openQuestion,
			code:    gate.NeedsDecision,
			reason:  gate.ReasonOpenQuestion,
			message: "1 unanswered question",
			next:    "odachi ask \"<your decision>\"",
		},
		{
			name:    "a requirement no step names",
			kind:    gate.TransitionExecute,
			state:   unplanned,
			code:    gate.NeedsWork,
			reason:  gate.ReasonUnplanned,
			message: "the feature declares R2 and no step names it: a requirement nobody planned is a requirement nobody built",
			next:    "odachi plan \"<text>\" --req R2",
		},
		{
			name:    "two requirements no step names",
			kind:    gate.TransitionExecute,
			state:   twoUnplanned,
			code:    gate.NeedsWork,
			reason:  gate.ReasonUnplanned,
			message: "the feature declares R2, R3 and no step names them: a requirement nobody planned is a requirement nobody built",
			next:    "odachi plan \"<text>\" --req R2",
		},
		{
			name:  "a requirement that is not the brief's words",
			kind:  gate.TransitionExecute,
			state: quoteMismatch,
			code:  gate.NeedsWork,
			// The two sentences travel with the row: the fix is a copy, and a
			// reader who has to open both files to find out what to copy is being
			// sent to do the check's job.
			reason:  gate.ReasonQuoteMismatch,
			message: `the plan does not quote the human: R1: the brief says "un paso" and the task doc says "dos pasos"; copy the brief's words into the requirement, or record the change with odachi ask --change`,
			next:    `odachi ask --change "<the human's words>" --req R1`,
		}, {
			name:    "proof missing",
			kind:    gate.TransitionExecute,
			state:   missingProof,
			code:    gate.NeedsWork,
			reason:  gate.ReasonProofMissing,
			message: "the proof for this change is missing",
			next:    "odachi check --proof <file>",
		},
		{
			name:    "proof stale",
			kind:    gate.TransitionExecute,
			state:   staleProof,
			code:    gate.NeedsWork,
			reason:  gate.ReasonProofStale,
			message: "the proof is from another state of the code",
			next:    "odachi check --proof <file>",
		},
		{
			name:    "proof failed",
			kind:    gate.TransitionExecute,
			state:   failedProof,
			code:    gate.NeedsWork,
			reason:  gate.ReasonProofFailed,
			message: "the proof failed: FAIL github.com/biggs-100/odachi/internal/domain/gate 0.02s",
			next:    `odachi check --run "go test ./..."`,
		},
		{
			name:    "the suite does not notice the undo",
			kind:    gate.TransitionExecute,
			state:   undoMissed,
			code:    gate.NeedsWork,
			reason:  gate.ReasonUndoMissed,
			message: "the tests do not notice the change being undone: the suite stayed green without it",
			next:    `odachi check --run "go test ./..."`,
		},
		{
			name:    "the row was never taken",
			kind:    gate.TransitionExecute,
			state:   undoMissing,
			code:    gate.NeedsWork,
			reason:  gate.ReasonUndoMissing,
			message: "the row that reverts the change was not taken for this tree, and the close owes it",
			next:    `odachi check --run "go test ./..."`,
		},
		{
			name:    "the mutation row failed",
			kind:    gate.TransitionExecute,
			state:   mutationFailed,
			code:    gate.NeedsWork,
			reason:  gate.ReasonMutationFailed,
			message: "the mutation row failed: the suite stayed green with the guard taken out",
			next:    `odachi check --run "go test ./..." --mutate "internal/domain/gate/gate.go" --remove "ReasonFuture" --expect "the gate declares"`,
		},
		{
			name:    "a close that owes a verifier and passes the slice",
			kind:    gate.TransitionExecute,
			state:   verificationSlice,
			code:    gate.NeedsWork,
			reason:  gate.ReasonVerifySlice,
			message: "the candidate is 401 authored lines and the limit for a close that owes a verifier is 400: the verifier would be handed a tree it cannot finish; the rest of the work belongs to a feature of its own, or the human accepts the overage with \"Slice: <lines> - <reason>\" in the task doc",
			next:    "odachi start",
		},
		{
			name:    "a risk with no verification",
			kind:    gate.TransitionExecute,
			state:   verificationMissing,
			code:    gate.NeedsWork,
			reason:  gate.ReasonVerifyMissing,
			message: "the commits declared a risk and this tree has no verification it can read: contract",
			next:    `odachi verify --exec "<command>"`,
		},
		{
			name:    "the verifier refuted the tree",
			kind:    gate.TransitionExecute,
			state:   verificationRefuted,
			code:    gate.NeedsWork,
			reason:  gate.ReasonVerifyRefuted,
			message: "the verifier refuted this tree: the gate closes a tree with no record",
			next:    `odachi verify --exec "<command>"`,
		},
		{
			name:    "ready",
			kind:    gate.TransitionStop,
			state:   ready(),
			code:    gate.Ready,
			reason:  gate.ReasonReady,
			message: "ready: there is proof for this tree",
			next:    "",
		},
	}
}

// TestEvaluationOrder is the whole table: the state, the words, and the continuation
// every block names. The last one is the property this file used to keep in a second
// list of seven (0293).
func TestEvaluationOrder(t *testing.T) {
	for _, tc := range evaluationCases {
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
			// The typed transition is the source of the printed next: line. Its kind
			// comes from the code and its command is the same continuation Next holds,
			// so the two can never drift (0348).
			if got.Transition.Kind != tc.kind {
				t.Errorf("Transition.Kind = %q, want %q", got.Transition.Kind, tc.kind)
			}
			if got.Transition.Command != tc.next {
				t.Errorf("Transition.Command = %q, want %q", got.Transition.Command, tc.next)
			}
			if tc.code == gate.Ready {
				if got.Transition.Kind != gate.TransitionStop {
					t.Errorf("Transition.Kind = %q, want %q for a ready result", got.Transition.Kind, gate.TransitionStop)
				}
				if got.Transition.Command != "" {
					t.Errorf("Transition.Command = %q, want empty for a ready result", got.Transition.Command)
				}
			} else if got.Transition.Command == "" {
				t.Errorf("Transition.Command is empty for blocked reason %q: every refusal names what resolves it", tc.reason)
			}
			// One property, judged for every reason the table holds: a block that
			// leaves the reader nowhere to go is the defect the continuation exists
			// to prevent, and ready is the one answer that resolves itself (0293).
			if tc.code == gate.Ready {
				if got.Next != "" {
					t.Errorf("Next = %q, want empty for a ready result", got.Next)
				}
			} else if !strings.HasPrefix(got.Next, "odachi ") {
				t.Errorf("Next = %q, want a command starting with %q: every block names what resolves it", got.Next, "odachi ")
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
		FenceLoosened: []string{"monorepo scopes: the rule is gone"},
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
	if got := gate.Evaluate(state); got.Reason != gate.ReasonFenceLoosened {
		t.Fatalf("Reason = %q, want %q (a loosened fence comes after out of scope)", got.Reason, gate.ReasonFenceLoosened)
	}

	state.FenceLoosened = nil
	if got := gate.Evaluate(state); got.Reason != gate.ReasonOpenQuestion {
		t.Fatalf("Reason = %q, want %q (open question comes after the fence)", got.Reason, gate.ReasonOpenQuestion)
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

// TestAPaperworkTreeKeepsTheProofItHas: after a feature closed, a tree that only
// moved documents is not a tree the suite has to run on again - a release note
// cannot fail it - and the proof that judged the code still speaks for it. The
// lane never rescues a missing proof, and a proof this run watched fail still
// blocks: only staleness over documents is paperwork.
func TestAPaperworkTreeKeepsTheProofItHas(t *testing.T) {
	paper := gate.State{HasBrief: true, TreeHash: "tree", ProofTreeHash: "code", ProofPaperwork: true}
	got := gate.Evaluate(paper)
	if got.Code != gate.Ready {
		t.Fatalf("a paperwork tree = %d/%q, want ready", got.Code, got.Reason)
	}
	if !strings.Contains(got.Message, "only documents moved since") {
		t.Errorf("message = %q, want the documents named as the reason", got.Message)
	}

	// A tree with no proof at all is not paperwork: the lane answers staleness.
	missing := gate.State{HasBrief: true, TreeHash: "tree", ProofPaperwork: true}
	if got := gate.Evaluate(missing); got.Reason != gate.ReasonProofMissing {
		t.Errorf("a missing proof with the lane open = %d/%q, want proof_missing", got.Code, got.Reason)
	}

	// And a red command is not a formality of the paperwork either.
	failed := gate.State{HasBrief: true, TreeHash: "tree", ProofTreeHash: "code", ProofPaperwork: true,
		ProofFailed: true, ProofLastLine: "FAIL github.com/x"}
	if got := gate.Evaluate(failed); got.Reason != gate.ReasonProofFailed {
		t.Errorf("a failed proof with the lane open = %d/%q, want proof_failed", got.Code, got.Reason)
	}
}

// TestAProofIsOwedAtTheClose: an open feature does not owe a proof for every tree it
// builds - the tree that closes it does, and that is the one a pull request and CI
// judge. What the gate must keep is the failure it just saw: a red command is not a
// formality of the closing, and silence about it would be the gate lying.
func TestAProofIsOwedAtTheClose(t *testing.T) {
	open := gate.State{HasBrief: true, TreeHash: "tree", ProofDeferred: true}
	got := gate.Evaluate(open)
	if got.Code != gate.Ready {
		t.Fatalf("an open feature without a proof = %d/%q, want ready", got.Code, got.Reason)
	}
	if !strings.Contains(got.Message, "owed when this feature closes") {
		t.Errorf("message = %q, want the close named", got.Message)
	}

	// A proof from another tree is not a block while the feature is open: it is the
	// next step's business.
	open.ProofTreeHash = "another"
	if got := gate.Evaluate(open); got.Code != gate.Ready {
		t.Errorf("an open feature with a stale proof = %d/%q, want ready", got.Code, got.Reason)
	}

	// The closing tree owes it again.
	closed := gate.State{HasBrief: true, TreeHash: "tree"}
	if got := gate.Evaluate(closed); got.Code != gate.NeedsWork || got.Reason != gate.ReasonProofMissing {
		t.Errorf("the closing tree without a proof = %d/%q, want proof_missing", got.Code, got.Reason)
	}

	// And a proof this run watched fail blocks even mid-feature.
	failed := gate.State{HasBrief: true, TreeHash: "tree", ProofTreeHash: "tree", ProofDeferred: true, ProofFailed: true, ProofLastLine: "FAIL github.com/x"}
	if got := gate.Evaluate(failed); got.Code != gate.NeedsWork || got.Reason != gate.ReasonProofFailed {
		t.Errorf("a failed proof mid-feature = %d/%q, want proof_failed", got.Code, got.Reason)
	}
}

// TestTheSliceIsInclusive: 400 authored lines are a candidate a verifier can
// finish, and the limit is the slice, not a wall one line early. 401 is the next
// feature, and the refusal names both numbers (0331).
func TestTheSliceIsInclusive(t *testing.T) {
	at := ready()
	at.VerificationLines = gate.SliceLines
	if got := gate.Evaluate(at); got.Code != gate.Ready {
		t.Errorf("a candidate of %d lines = %d/%q, want ready", gate.SliceLines, got.Code, got.Reason)
	}

	over := ready()
	over.VerificationLines = gate.SliceLines + 1
	// The close owes the verifier and no record answers this tree: that is the
	// state the slice exists for, since a close that already has its record is not
	// spending one (0331).
	over.VerificationMissing = []string{"contract"}
	got := gate.Evaluate(over)
	if got.Code != gate.NeedsWork || got.Reason != gate.ReasonVerifySlice {
		t.Fatalf("a candidate of %d lines = %d/%q, want the slice", gate.SliceLines+1, got.Code, got.Reason)
	}
	for _, want := range []string{strconv.Itoa(gate.SliceLines + 1), strconv.Itoa(gate.SliceLines)} {
		if !strings.Contains(got.Message, want) {
			t.Errorf("the refusal does not name %q: %s", want, got.Message)
		}
	}
	if got.Next != "odachi start" {
		t.Errorf("Next = %q, want the rest of the work started as a feature of its own", got.Next)
	}
}

// TestAnAcceptedOverageNamesTheNumberTheDocumentAccepts: the human's decision to
// let this close carry more than the slice is read from the task document, and the
// gate stops answering with the size it refuses. What the close still owes is the
// record the verifier signs, and the row that asks for it names the number the
// human accepted and the sentence that pays for it (0340).
func TestAnAcceptedOverageNamesTheNumberTheDocumentAccepts(t *testing.T) {
	over := ready()
	over.VerificationLines = gate.SliceLines + 38
	over.VerificationMissing = []string{"contract"}
	over.SliceAccepted = gate.SliceLines + 38
	over.SliceReason = "the ADR and the tests are the change"

	got := gate.Evaluate(over)

	if got.Reason != gate.ReasonVerifyMissing {
		t.Fatalf("an accepted overage answered %q, want %q: the size is settled and the record is what is owed (%s)", got.Reason, gate.ReasonVerifyMissing, got.Message)
	}
	if !strings.Contains(got.Message, strconv.Itoa(gate.SliceLines+38)) {
		t.Errorf("the row does not name the number the human accepted: %s", got.Message)
	}
	if !strings.Contains(got.Message, "the ADR and the tests are the change") {
		t.Errorf("the row does not name the reason that paid for it: %s", got.Message)
	}
}

// TestAnAcceptanceBelowTheCandidateStillRefuses: an acceptance is a number, and a
// number below what the tree authors accepts nothing. The refusal names both, so a
// reader sees which one moved (0340).
func TestAnAcceptanceBelowTheCandidateStillRefuses(t *testing.T) {
	over := ready()
	over.VerificationLines = gate.SliceLines + 38
	over.VerificationMissing = []string{"contract"}
	over.SliceAccepted = gate.SliceLines + 18
	over.SliceReason = "the ADR and the tests are the change"

	got := gate.Evaluate(over)

	if got.Reason != gate.ReasonVerifySlice {
		t.Fatalf("an acceptance below the candidate answered %q, want the slice: the decision does not reach this tree (%s)", got.Reason, got.Message)
	}
	for _, want := range []string{strconv.Itoa(gate.SliceLines + 38), strconv.Itoa(gate.SliceLines + 18)} {
		if !strings.Contains(got.Message, want) {
			t.Errorf("the refusal does not name %q: %s", want, got.Message)
		}
	}
}

// TestAnAcceptanceWithoutAReasonStillRefuses: the reason is what makes the decision
// legible to the person who reads the close, so a covered number with no reason
// leaves the slice standing. Otherwise the line would be written by whoever wanted
// through, and the limit would die of a silent sentence (0340).
func TestAnAcceptanceWithoutAReasonStillRefuses(t *testing.T) {
	over := ready()
	over.VerificationLines = gate.SliceLines + 38
	over.VerificationMissing = []string{"contract"}
	over.SliceAccepted = gate.SliceLines + 38

	if got := gate.Evaluate(over); got.Reason != gate.ReasonVerifySlice {
		t.Fatalf("an acceptance with no reason answered %q, want the slice (%s)", got.Reason, got.Message)
	}
}

// TestTheSliceOnlyFiresWhileTheRecordIsMissing: the slice exists to keep a
// verifier from being spent on a tree nobody can finish. Once the record is in the
// store, nobody is spending anything: a 401-line close that carries the verifier's
// own signature on this tree closes, and a refuted record keeps answering with its
// own sentence instead of the slice (0331).
func TestTheSliceOnlyFiresWhileTheRecordIsMissing(t *testing.T) {
	noRecord := ready()
	noRecord.VerificationLines = gate.SliceLines + 1
	noRecord.VerificationMissing = []string{"contract"}
	if got := gate.Evaluate(noRecord); got.Reason != gate.ReasonVerifySlice {
		t.Fatalf("401 lines with no record = %d/%q, want the slice", got.Code, got.Reason)
	}

	signed := ready()
	signed.VerificationLines = gate.SliceLines + 1
	if got := gate.Evaluate(signed); got.Code != gate.Ready || got.Reason != gate.ReasonReady {
		t.Errorf("401 lines with the record for this tree = %d/%q, want ready: the verifier already signed it", got.Code, got.Reason)
	}

	refuted := ready()
	refuted.VerificationLines = gate.SliceLines + 1
	refuted.VerificationRefuted = []string{"the gate closes a tree with no record"}
	got := gate.Evaluate(refuted)
	if got.Reason != gate.ReasonVerifyRefuted {
		t.Errorf("401 lines with a refuted record = %d/%q, want the refutation: a refuted record is not a missing one", got.Code, got.Reason)
	}
	if strings.Contains(got.Message, "authored lines") {
		t.Errorf("the refutation was answered with the slice: %s", got.Message)
	}
}

// TestARecordThatAnswersNothingKeepsItsOwnDoor: the slice exists for the close
// that has no record at all. A record that exists and answers nothing here -it
// speaks for another tree, or nobody can read it- keeps the row that names it and
// the `odachi verify` that writes the one for this tree: `odachi start` cannot,
// because the work is already here (0331).
func TestARecordThatAnswersNothingKeepsItsOwnDoor(t *testing.T) {
	other := ready()
	other.VerificationLines = gate.SliceLines + 1
	other.VerificationMissing = []string{"contract"}
	other.VerificationRecordFound = true

	got := gate.Evaluate(other)
	if got.Code != gate.NeedsWork || got.Reason != gate.ReasonVerifyMissing {
		t.Fatalf("a %d-line close whose record answers another tree = %d/%q, want verification_missing",
			gate.SliceLines+1, got.Code, got.Reason)
	}
	if got.Next != `odachi verify --exec "<command>"` {
		t.Errorf("Next = %q, want the command that writes the record for this tree", got.Next)
	}
	if strings.Contains(got.Message, "authored lines") {
		t.Errorf("the slice answered a close whose record exists: %s", got.Message)
	}
}

// TestCaseTableCoversEveryReason pins the table to the gate's own public list:
// a reason Reasons() names and no row produces would ship with no typed
// transition ever being checked (0348).
func TestCaseTableCoversEveryReason(t *testing.T) {
	held := make(map[string]bool, len(evaluationCases))
	for _, tc := range evaluationCases {
		held[tc.reason] = true
	}
	for _, reason := range gate.Reasons() {
		if !held[reason] {
			t.Errorf("Reasons() names %q and no row in evaluationCases produces it: a reason without a row ships unjudged", reason)
		}
	}
}

// TestEveryReasonInTheSourceHasItsRow: the table is the only list, so the source is the
// other half of the promise -one half that drifts is the defect this file exists to
// catch-. A reason declared in gate.go with no row would ship unjudged, and a row naming
// a reason the gate no longer returns would keep a promise nobody keeps (0293).
func TestEveryReasonInTheSourceHasItsRow(t *testing.T) {
	declared := reasonConstants(t)
	held := make(map[string]bool, len(evaluationCases))
	for _, tc := range evaluationCases {
		held[tc.reason] = true
	}

	for _, reason := range declared {
		if !held[reason] {
			t.Errorf("the gate declares %q and no row in evaluationCases produces it: add the state, its words and its continuation", reason)
		}
	}
	for reason := range held {
		if !slices.Contains(declared, reason) {
			t.Errorf("evaluationCases names %q, which gate.go does not declare", reason)
		}
	}
}

// reasonConstants reads the reasons the gate declares from its own source. Go cannot be
// asked for a package's constants, and a list copied by hand is the second list this
// feature exists to remove (0293).
func reasonConstants(t *testing.T) []string {
	t.Helper()

	file, err := parser.ParseFile(token.NewFileSet(), "gate.go", nil, 0)
	if err != nil {
		t.Fatalf("parsing gate.go: %v", err)
	}
	var reasons []string
	ast.Inspect(file, func(n ast.Node) bool {
		spec, ok := n.(*ast.ValueSpec)
		if !ok {
			return true
		}
		for i, name := range spec.Names {
			if !strings.HasPrefix(name.Name, "Reason") || i >= len(spec.Values) {
				continue
			}
			lit, ok := spec.Values[i].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				continue
			}
			if value, err := strconv.Unquote(lit.Value); err == nil {
				reasons = append(reasons, value)
			}
		}
		return true
	})
	slices.Sort(reasons)
	return reasons
}
