package verify

import (
	verifydomain "github.com/biggs-100/odachi/internal/domain/verify"
)

// The canonical runtime receipt lives in the domain beside the record it seals.
// This package keeps its historical names as aliases, so a caller of the app
// layer and a caller of the domain see the same type, the same validation and
// the same bytes.
type (
	// SealedRuntimeReceipt is the tool-derived runtime fact of one managed run.
	SealedRuntimeReceipt = verifydomain.Receipt
	// Outcome is the terminal fact a sealed runtime receipt carries.
	Outcome = verifydomain.Outcome
)

// OutcomeCompleted says the actor ran to a zero exit over a stable snapshot under
// one bound grant and session.
const OutcomeCompleted = verifydomain.OutcomeCompleted

// ParseRuntimeReceipt reads one receipt and refuses anything that is not a sealed
// fact.
func ParseRuntimeReceipt(data []byte) (SealedRuntimeReceipt, error) {
	return verifydomain.ParseReceipt(data)
}

// RenderRuntimeReceipt writes the canonical receipt: indented JSON with a
// trailing newline.
func RenderRuntimeReceipt(receipt SealedRuntimeReceipt) ([]byte, error) {
	return verifydomain.RenderReceipt(receipt)
}
