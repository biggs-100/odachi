package doctor_test

import (
	"testing"

	"github.com/biggs-100/odachi/internal/domain/doctor"
)

// TestIgnoredByReadsTheRulesAPersonWrites: the repository's own .gitignore is
// one of the mechanisms that keeps the evidence out of the tree, and it is the
// one `odachi init` writes.
func TestIgnoredByReadsTheRulesAPersonWrites(t *testing.T) {
	block := "# odachi:begin\n/proof/\n/.odachi/reviews/\n# odachi:end\n"

	for _, dir := range []string{"proof", ".odachi/reviews"} {
		if !doctor.IgnoredBy(block, dir) {
			t.Errorf("IgnoredBy(%q) = false, want true for the block init writes", dir)
		}
	}

	cases := map[string]bool{
		"proof/\n":            true,
		"/proof\n":            true,
		"proof/*\n":           true,
		"*/\n":                true,
		"*\n!.gitignore\n":    true,
		"# proof/\n":          false, // a comment is not a rule
		"!" + "/proof/\n":     false, // a negation keeps the path in the tree
		"dist/\n.env\n":       false,
		"/proof/algo.txt\n":   false, // one file inside is not the directory
		"pruebas/\nproof-*\n": false,
	}

	for content, want := range cases {
		if got := doctor.IgnoredBy(content, "proof"); got != want {
			t.Errorf("IgnoredBy(%q, proof) = %v, want %v", content, got, want)
		}
	}
}

// TestIgnoredByIsNotFooledByAFolderName: a rule for a folder that merely starts
// with the same letters does not cover the evidence.
func TestIgnoredByIsNotFooledByAFolderName(t *testing.T) {
	if doctor.IgnoredBy("proofs/\n", "proof") {
		t.Error("IgnoredBy(proofs/, proof) = true, want false: a different folder")
	}
	if doctor.IgnoredBy(".odachi/reviews-otras/\n", ".odachi/reviews") {
		t.Error("IgnoredBy(reviews-otras, .odachi/reviews) = true, want false")
	}
}
