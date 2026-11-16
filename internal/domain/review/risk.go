package review

import "strings"

// riskLens is the lens each declared risk asks for. It is the table `commits`
// publishes, moved to where the gate can read it: a risk that names no lens is a
// sentence, and this project already knows which question each one answers (0301).
var riskLens = map[string]string{
	"data":        Reliability,
	"contract":    Reliability,
	"no-test":     Reliability,
	"security":    Risk,
	"concurrency": Resilience,
	"delivery":    Resilience,
}

// LensForRisk is the lens a commit's declared risk asks for, and whether the name is
// one this project knows. An unknown name asks for nothing: the commit wrote a word
// the table does not carry, and inventing a lens for it would demand a review nobody
// asked for.
func LensForRisk(name string) (string, bool) {
	lens, ok := riskLens[strings.ToLower(strings.TrimSpace(name))]
	return lens, ok
}

// LensesForRisk is the lens demand of a set of declared risks, in the order the
// commits first said them and without repeats. A risk the table does not carry is
// skipped in silence: the ready note already names a declaration it cannot read.
func LensesForRisk(names []string) []string {
	var lenses []string
	for _, name := range names {
		lens, ok := LensForRisk(name)
		if !ok || containsLens(lenses, lens) {
			continue
		}
		lenses = append(lenses, lens)
	}
	return lenses
}

// containsLens reports whether the list already carries this lens.
func containsLens(lenses []string, lens string) bool {
	for _, have := range lenses {
		if have == lens {
			return true
		}
	}
	return false
}

// AsksForVerifier reports whether a risk the commits declared asks for an
// independent verifier. The published table answers yes in its third column for
// every row, so a name it carries is a name that owes one; the day a row says no,
// this is where the column lands. A name the table does not carry asks for nobody:
// the ready note already names the declaration it cannot read (0315).
func AsksForVerifier(name string) bool {
	_, ok := LensForRisk(name)
	return ok
}
