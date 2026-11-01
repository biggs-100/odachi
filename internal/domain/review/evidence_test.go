package review_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/review"
)

// TestAFindingSaysWhereItsEvidenceComesFrom: the class is the reviewer's own
// account of what it did -quoted a line of the diff, or argued about it- and it is
// what decides how many refutations a block needs. A verdict written before the
// field existed reads as argued: the strict side, because a finding that does not
// say it cited anything did not cite anything.
func TestAFindingSaysWhereItsEvidenceComesFrom(t *testing.T) {
	cases := []struct {
		name     string
		evidence string
		want     string
	}{
		{name: "cited", evidence: "cited", want: review.EvidenceCited},
		{name: "argued", evidence: "argued", want: review.EvidenceArgued},
		{name: "absent, from a verdict written before the field", evidence: "", want: review.EvidenceArgued},
		{name: "spelled with spaces and capitals", evidence: "  Cited ", want: review.EvidenceCited},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			finding := review.Finding{Path: "a.go", Severity: review.Blocker, Causal: true, Why: "x", Evidence: c.evidence}
			if got := finding.EvidenceClass(); got != c.want {
				t.Errorf("EvidenceClass = %q, want %q", got, c.want)
			}
		})
	}
}

// TestTheEvidenceVocabularyIsClosed: a value that is neither is refused by name,
// the way an invented severity is. The class decides how many witnesses a block
// needs, so a third word would decide nothing.
func TestTheEvidenceVocabularyIsClosed(t *testing.T) {
	verdict := review.Verdict{Reviewer: "x", Findings: []review.Finding{
		{Path: "a.go", Severity: review.Blocker, Causal: true, Why: "y", Evidence: "vibes"},
	}}

	err := verdict.Validate()
	if err == nil {
		t.Fatal("Validate accepted an evidence class nobody defined")
	}
	for _, want := range []string{"vibes", review.EvidenceCited, review.EvidenceArgued} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal %q does not name %q", err, want)
		}
	}
}

// TestTheReviewAsksForTheEvidenceClass: the field is useless if the prompt never
// asks for it, and the reviewer cannot guess the vocabulary.
func TestTheReviewAsksForTheEvidenceClass(t *testing.T) {
	prompt := review.Prompt(promptInput())

	for _, want := range []string{
		`"evidence": "` + review.EvidenceCited + ` | ` + review.EvidenceArgued + `"`,
		"quoted from the diff",
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("the prompt does not carry %q", want)
		}
	}
}
