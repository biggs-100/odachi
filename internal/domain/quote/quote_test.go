package quote

import (
	"strings"
	"testing"
)

const briefMD = `# the brief

## What must exist
- light in the room

## 0314 · el requisito

## What must exist
- un requisito que no es la cita del humano no pasa la puerta

## Done when
- el check lo prueba

## 0315 · otra

## What must exist
- otra cosa
`

const docMD = `# el-requisito
Base: abc1234

## Requirements (verbatim, from the user)
R1 · "un requisito que no es la cita del humano no pasa la puerta"

## Steps
- [ ] T1 · algo   R1 · route:inline
`

// TestBlockTakesOnlyItsFeature: the brief accumulates one block per feature, and
// a comparison against the whole file would read the first block's words for
// every feature and call a good one wrong.
func TestBlockTakesOnlyItsFeature(t *testing.T) {
	body := Block(briefMD, "0314")
	if !strings.Contains(body, "no pasa la puerta") {
		t.Errorf("Block did not carry the feature's bullet:\n%s", body)
	}
	if strings.Contains(body, "otra cosa") {
		t.Errorf("Block read past its own feature into the next one:\n%s", body)
	}
	if got := Block(briefMD, "0999"); got != "" {
		t.Errorf("Block(0999) = %q, want nothing", got)
	}
}

// TestCompareHoldsWhenTheWordsAreTheBriefs: the tool writes the same sentence
// twice, so a feature nobody touched compares clean.
func TestCompareHoldsWhenTheWordsAreTheBriefs(t *testing.T) {
	if got := Compare(briefMD, docMD, "0314"); len(got) != 0 {
		t.Fatalf("Compare = %+v, want no mismatch", got)
	}
}

// TestCompareNamesTheRewrite: an agent that answers a requirement change by
// rewriting R1 is the defect this check exists for, and the mismatch carries both
// sentences so the fix is a copy.
func TestCompareNamesTheRewrite(t *testing.T) {
	rewritten := strings.Replace(docMD, "un requisito que no es la cita del humano no pasa la puerta", "el check compara requisitos", 1)
	got := Compare(briefMD, rewritten, "0314")
	if len(got) != 1 {
		t.Fatalf("Compare = %+v, want one mismatch", got)
	}
	if got[0].ID != "R1" || !strings.Contains(got[0].Brief, "no pasa la puerta") || !strings.Contains(got[0].Doc, "compara requisitos") {
		t.Errorf("mismatch = %+v, want both sentences", got[0])
	}
}

// TestCompareJudgesNothingItCannotCompare: a feature with no block and a light
// feature never wrote the sentence twice, and a gate that refuses them refuses
// its own past.
func TestCompareJudgesNothingItCannotCompare(t *testing.T) {
	if got := Compare(briefMD, docMD, "0999"); got != nil {
		t.Errorf("Compare with no block = %+v, want nothing", got)
	}
	light := "# una-linea\nBase: abc1234\nLight: yes\n\n## Requirements (verbatim, from the user)\nR1 · \"cualquier cosa\"\n"
	if got := Compare(briefMD, light, "0314"); got != nil {
		t.Errorf("Compare on a light feature = %+v, want nothing", got)
	}
}

// TestCompareNamesARequirementTheBriefNeverHad: a second R with no bullet beside
// it is an invented requirement, and the mismatch says which one.
func TestCompareNamesARequirementTheBriefNeverHad(t *testing.T) {
	extra := strings.Replace(docMD, "## Steps", "R2 · \"una cosa que el humano no pidio\"\n\n## Steps", 1)
	got := Compare(briefMD, extra, "0314")
	if len(got) != 1 || got[0].ID != "R2" || got[0].Brief != "" {
		t.Fatalf("Compare = %+v, want R2 with no brief bullet", got)
	}
}

// TestSameSentenceIgnoresSpacingOnly: the brief wraps a long bullet and the tool
// joins it back, so spacing is not a rewrite; a changed word is.
func TestSameSentenceIgnoresSpacingOnly(t *testing.T) {
	if !sameSentence("una cosa\n  que existe", "una cosa que existe") {
		t.Error("sameSentence refused the same sentence wrapped")
	}
	if sameSentence("una cosa que existe", "una cosa que no existe") {
		t.Error("sameSentence accepted a rewritten sentence")
	}
}
