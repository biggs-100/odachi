package gate_test

import (
	"testing"

	"github.com/biggs-100/odachi/internal/domain/gate"
)

// TestASilentSuccessIsAValidProof covers the most common real case: `go vet`,
// `tsc --noEmit` and `gofmt -l` all succeed while printing nothing. Once the
// proof file names the tree it came from, an empty body is still evidence.
// Reporting "falta la prueba" right after running one is a dead end, and a
// message may only name a command that resolves the block.
func TestASilentSuccessIsAValidProof(t *testing.T) {
	got := gate.Evaluate(gate.State{
		HasBrief:      true,
		TreeHash:      "abc",
		ProofTreeHash: "abc",
		ProofOutput:   "",
	})

	if got.Code != gate.Ready || got.Reason != gate.ReasonReady {
		t.Fatalf("prueba con header y salida vacía: code=%d reason=%s message=%q; es una prueba válida, no una prueba faltante",
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
		t.Fatalf("prueba fallida y silenciosa: code=%d reason=%s; tiene que decir que falló", got.Code, got.Reason)
	}
}
