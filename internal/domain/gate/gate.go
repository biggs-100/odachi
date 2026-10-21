// Package gate answers one question: can this work be called done? It is pure
// and ordered, so the first unanswered thing is always the one the human sees.
package gate

import (
	"fmt"
	"strconv"
	"strings"
)

// Code is the exit code of `odachi check`.
type Code int

const (
	// Ready means proof exists for the exact tree.
	Ready Code = 0
	// NeedsWork means the agent can act without asking the human anything.
	NeedsWork Code = 1
	// NeedsDecision means only the human can unblock this.
	NeedsDecision Code = 2
)

// Reasons are stable identifiers; the Message is what a human reads.
const (
	ReasonReady           = "ready"
	ReasonNoBrief         = "no_brief"
	ReasonBriefIncomplete = "brief_incomplete"
	ReasonOutOfScope      = "out_of_scope"
	ReasonOpenQuestion    = "open_question"
	ReasonProofMissing    = "proof_missing"
	ReasonProofStale      = "proof_stale"
	ReasonProofFailed     = "proof_failed"
)

// Commands that resolve a block. Every blocked result names one.
const (
	nextStart = "odachi start"
	// nextAsk carries the shape of the answer. A bare `odachi ask` prints usage:
	// naming it would be a dead end for a human too, and the bench measures exactly
	// that.
	nextAsk        = "odachi ask \"<tu decisión>\""
	nextCheckProof = "odachi check --proof <archivo>"
)

// State is everything Evaluate needs, gathered by the caller. The gate itself
// reads no files and runs no commands.
type State struct {
	HasBrief     bool
	BriefMissing []string

	TreeHash      string
	ProofTreeHash string
	ProofOutput   string

	// ProofFailed says the command that produced the proof exited non-zero. The
	// gate cannot see an exit code, so the caller that ran the command reports
	// it, together with the last line it printed and the command itself: the
	// continuation has to name the exact command, not a generic one.
	ProofFailed   bool
	ProofLastLine string
	ProofCommand  string

	// ProofTimedOut is the other way a proof command fails: it never finished.
	// The sentence says how long the tool waited, which is what the human has to
	// raise or what they have to fix.
	ProofTimedOut       bool
	ProofTimeoutSeconds int

	OpenQuestions int
	Violations    []string
}

// Result is the verdict plus what to do next.
type Result struct {
	Code    Code
	Reason  string
	Message string
	Next    string
}

// Evaluate applies the documented order: no brief -> brief incomplete -> out of
// scope -> open question -> proof missing -> proof stale -> proof failed ->
// ready. The order decides the Code too: a block only the human can clear is
// NeedsDecision, a block the agent can clear is NeedsWork.
func Evaluate(s State) Result {
	switch {
	case !s.HasBrief:
		return Result{NeedsDecision, ReasonNoBrief, "no hay brief: todavía no sabemos qué tiene que existir", nextStart}
	case len(s.BriefMissing) > 0:
		return Result{NeedsDecision, ReasonBriefIncomplete, "el brief está incompleto: falta " + strings.Join(s.BriefMissing, ", "), nextStart}
	case len(s.Violations) > 0:
		return Result{NeedsDecision, ReasonOutOfScope, "fuera de alcance: " + strings.Join(s.Violations, "; "), nextAsk}
	case s.OpenQuestions > 0:
		return Result{NeedsDecision, ReasonOpenQuestion, openQuestions(s.OpenQuestions), nextAsk}
	// A proof exists as soon as it names the tree it came from. An empty body is
	// still a proof: `go vet`, `tsc --noEmit` and `gofmt -l` all succeed while
	// printing nothing, and telling a human "falta la prueba" right after they
	// ran one is a dead end that names a command which cannot resolve it.
	case s.ProofTreeHash == "":
		return Result{NeedsWork, ReasonProofMissing, "falta la prueba de este cambio", nextCheckProof}
	case s.ProofTreeHash != s.TreeHash:
		return Result{NeedsWork, ReasonProofStale, "la prueba es de otro estado del código", nextCheckProof}
	case s.ProofTimedOut:
		return Result{NeedsWork, ReasonProofFailed, "la prueba no terminó en " + strconv.Itoa(s.ProofTimeoutSeconds) + "s", nextRun(s.ProofCommand)}
	case s.ProofFailed:
		return Result{NeedsWork, ReasonProofFailed, "la prueba falló: " + proofLastLine(s.ProofLastLine), nextRun(s.ProofCommand)}
	default:
		return Result{Ready, ReasonReady, "listo: hay prueba para este árbol", ""}
	}
}

// proofLastLine says what the failing command last printed. A command that
// failed in silence is still a failure, and the message never ends in a colon
// with nothing after it.
func proofLastLine(line string) string {
	line = strings.Join(strings.Fields(line), " ")
	if line == "" {
		return "sin salida"
	}
	return line
}

// nextRun names the command that produced the failing proof, quoted so a
// human can paste it back exactly. Without a known command the continuation
// keeps the placeholder the other blocked results use.
func nextRun(command string) string {
	command = strings.TrimSpace(command)
	if command == "" {
		return `odachi check --run "<comando>"`
	}
	return `odachi check --run "` + strings.ReplaceAll(command, `"`, `\"`) + `"`
}

// openQuestions keeps the sentence readable in Spanish for one and for many.
func openQuestions(n int) string {
	if n == 1 {
		return "hay 1 pregunta sin contestar"
	}
	return fmt.Sprintf("hay %d preguntas sin contestar", n)
}
