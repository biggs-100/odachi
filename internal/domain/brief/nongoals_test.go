package brief_test

import (
	"testing"

	"github.com/biggz/odachi/internal/domain/brief"
)

// TestNonGoalsAreOptional: forcing an exclusion invents a rule the human never
// meant, and a false fence violation teaches the wrong lesson. A small first
// version may exclude nothing; fence.toml then explains what a rule is for.
//
// This also removes a dead end: the interview lets the human pick no exclusion,
// so requiring the section made a valid interview unable to pass check, and the
// only command the message named reproduced the same empty section.
func TestNonGoalsAreOptional(t *testing.T) {
	b := brief.Brief{
		Name:     "gastos",
		Outcome:  []string{"quiero anotar gastos"},
		DoneWhen: []string{"anoto un gasto y aparece en la lista"},
	}

	if missing := b.Missing(); len(missing) != 0 {
		t.Fatalf("Missing() = %q; \"no incluye\" es la única parte que puede quedar vacía en una primera versión", missing)
	}
}

// TestTheTwoSubstantivePartsAreStillRequired keeps the removal narrow: the two
// parts that are actually acceptance criteria stay mandatory.
func TestTheTwoSubstantivePartsAreStillRequired(t *testing.T) {
	missing := brief.Brief{}.Missing()
	if len(missing) != 2 {
		t.Fatalf("Missing() = %q; quiero exactamente \"qué tiene que existir\" y \"listo cuando\"", missing)
	}
}
