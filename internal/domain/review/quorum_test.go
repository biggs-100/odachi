package review_test

import (
	"testing"

	"github.com/biggs-100/odachi/internal/domain/review"
)

// TestAQuorumOfOneNamesTheModelEveryLensShares: four lenses answered by one model
// are one opinion in four costumes, and the number the human reads has to say so
// before they trust the four answers (0316).
func TestAQuorumOfOneNamesTheModelEveryLensShares(t *testing.T) {
	models := map[string]string{
		"risk":        "opencode-go/deepseek-v4.1-flash",
		"readability": "opencode-go/deepseek-v4.1-flash",
		"reliability": "opencode-go/deepseek-v4.1-flash",
		"resilience":  "opencode-go/deepseek-v4.1-flash",
	}

	distinct, shared := review.Quorum(models)

	if distinct != 1 {
		t.Errorf("Quorum = %d, want 1: the four lenses name one model", distinct)
	}
	if shared != "opencode-go/deepseek-v4.1-flash" {
		t.Errorf("shared = %q, want the model every lens names", shared)
	}
}

// TestTwoModelsAreAQuorumOfTwoAndShareNothing: the day a second vendor judges one
// lens, the number moves and there is no model to name as the only one.
func TestTwoModelsAreAQuorumOfTwoAndShareNothing(t *testing.T) {
	models := map[string]string{
		"risk":        "vendor/one",
		"readability": "vendor/two",
		"reliability": "vendor/one",
		"resilience":  "vendor/two",
	}

	distinct, shared := review.Quorum(models)

	if distinct != 2 {
		t.Errorf("Quorum = %d, want 2", distinct)
	}
	if shared != "" {
		t.Errorf("shared = %q, want nothing when the lenses disagree", shared)
	}
}

// TestTheQuorumCountsModelsAndNotSpellings: a trailing space or a different case is
// the same model, and counting it twice would report a quorum the project does not
// have.
func TestTheQuorumCountsModelsAndNotSpellings(t *testing.T) {
	models := map[string]string{
		"risk":        " vendor/one ",
		"readability": "VENDOR/ONE",
	}

	if distinct, _ := review.Quorum(models); distinct != 1 {
		t.Errorf("Quorum = %d, want 1: the two spellings are one model", distinct)
	}
}

// TestAnEmptyTableHasNoQuorum: a project with no lens table has no number to
// report, and zero is the honest answer.
func TestAnEmptyTableHasNoQuorum(t *testing.T) {
	for _, models := range []map[string]string{nil, {}} {
		distinct, shared := review.Quorum(models)
		if distinct != 0 || shared != "" {
			t.Errorf("Quorum(%v) = %d, %q; want 0 and nothing", models, distinct, shared)
		}
	}
}
