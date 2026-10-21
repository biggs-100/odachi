// Package gate answers one question: can this work be called done? It is pure
// and ordered, so the first unanswered thing is always the one the human sees.
package gate

import (
	"fmt"
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
)

// Commands that resolve a block. Every blocked result names one.
const (
	nextStart      = "odachi start"
	nextAsk        = "odachi ask"
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
// scope -> open question -> proof missing -> proof stale -> ready. The order
// decides the Code too: a block only the human can clear is NeedsDecision, a
// block the agent can clear is NeedsWork.
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
	case s.ProofTreeHash == "" || s.ProofOutput == "":
		return Result{NeedsWork, ReasonProofMissing, "falta la prueba de este cambio", nextCheckProof}
	case s.ProofTreeHash != s.TreeHash:
		return Result{NeedsWork, ReasonProofStale, "la prueba es de otro estado del código", nextCheckProof}
	default:
		return Result{Ready, ReasonReady, "listo: hay prueba para este árbol", ""}
	}
}

// openQuestions keeps the sentence readable in Spanish for one and for many.
func openQuestions(n int) string {
	if n == 1 {
		return "hay 1 pregunta sin contestar"
	}
	return fmt.Sprintf("hay %d preguntas sin contestar", n)
}
