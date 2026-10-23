package gate_test

import (
	"testing"

	"github.com/biggs-100/odachi/internal/domain/gate"
)

// TestASilentSuccessIsAValidProof covers the most common real case: `go vet`,
// `tsc --noEmit` and `gofmt -l` all succeed while printing nothing. Once the
// proof file names the tree it came from, an empty body is still evidence.
// Reporting "the proof is missing" right after running one is a dead end, and a
// message may only name a command that resolves the block.
func TestASilentSuccessIsAValidProof(t *testing.T) {
	got := gate.Evaluate(gate.State{
		HasBrief:      true,
		TreeHash:      "abc",
		ProofTreeHash: "abc",
		ProofOutput:   "",
	})

	if got.Code != gate.Ready || got.Reason != gate.ReasonReady {
		t.Fatalf("proof with a header and empty output: code=%d reason=%s message=%q; it is a valid proof, not a missing one",
			got.Code, got.Reason, got.Message)
	}
}

// TestAFailedSilentRunIsAFailureNotAMissingProof keeps the other half honest: a
// command that failed without printing must still be reported as failed.
func TestAFailedSilentRunIsAFailureNotAMissingProof(t *testing.T) {
	got := gate.Evaluate(gate.State{
		HasBrief:      true,
		TreeHash:      "abc",
		ProofTreeHash: "abc",
		ProofOutput:   "",
		ProofFailed:   true,
	})

	if got.Code != gate.NeedsWork || got.Reason != gate.ReasonProofFailed {
		t.Fatalf("failed and silent proof: code=%d reason=%s; it has to say that it failed", got.Code, got.Reason)
	}
}
