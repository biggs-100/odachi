package review_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/review"
)

// arguedAndCited are the same defect told two ways: one argued without quoting
// anything, one quoted from the diff.
func arguedAndCited() (argued, cited review.Finding) {
	argued = review.Finding{Path: "internal/auth/session.go", Line: 12, Severity: review.Blocker, Causal: true, Why: "sin verificar la firma"}
	cited = argued
	cited.Evidence = review.EvidenceCited
	return argued, cited
}

// refutedBy is one refuter's answer that kills F1 with a citation.
func refutedBy(why string) review.Refutation {
	return review.Refutation{Refutations: []review.Judgment{{ID: "F1", Verdict: review.Refuted, Why: why}}}
}

// TestAnArguedBlockNeedsTwoCitedRefutations: the finding that argues needs two
// witnesses, because the claim nobody can point at is the one a single reviewer
// can remove by being confident. One citation is not enough, and a `stands` never
// counts as one.
func TestAnArguedBlockNeedsTwoCitedRefutations(t *testing.T) {
	argued, _ := arguedAndCited()
	findings := []review.Finding{argued}

	one := review.Answers{refutedBy("line 12 already verifies the signature")}
	if got := review.DeriveAfter(findings, one); got != review.ChangesRequested {
		t.Errorf("one citation lifted an argued block: DeriveAfter = %q", got)
	}

	stand := review.Refutation{Refutations: []review.Judgment{{ID: "F1", Verdict: review.Stands, Why: "I could not cite it"}}}
	if got := review.DeriveAfter(findings, review.Answers{refutedBy("x"), stand}); got != review.ChangesRequested {
		t.Errorf("a second answer that stands lifted the block: DeriveAfter = %q", got)
	}

	two := review.Answers{refutedBy("line 12 already verifies the signature"), refutedBy("the call site passes a signed token")}
	if got := review.DeriveAfter(findings, two); got != review.Approved {
		t.Errorf("two citations did not lift an argued block: DeriveAfter = %q", got)
	}

	if _, blocking := review.BlockingAfter(findings, two); blocking {
		t.Error("BlockingAfter still names a finding two witnesses refuted")
	}
	if _, blocking := review.BlockingAfter(findings, one); !blocking {
		t.Error("BlockingAfter let a finding fall with one witness")
	}
}

// TestACitedBlockFallsWithOneRefutation: when the reviewer quoted the line, one
// counter-citation is enough -the evidence is checkable at a glance, and making
// the human pay for a second witness would punish the reviewer that did the work.
func TestACitedBlockFallsWithOneRefutation(t *testing.T) {
	_, cited := arguedAndCited()
	findings := []review.Finding{cited}

	one := review.Answers{refutedBy("line 12 is inside a helper the diff also adds")}
	if got := review.DeriveAfter(findings, one); got != review.Approved {
		t.Errorf("a cited block survived one citation: DeriveAfter = %q", got)
	}
}

// TestAnswersCountPerFinding: two witnesses about F1 say nothing about F2. A tree
// with two blocking findings needs the votes on each one.
func TestAnswersCountPerFinding(t *testing.T) {
	argued, _ := arguedAndCited()
	second := argued
	second.Path = "internal/auth/token.go"
	findings := []review.Finding{argued, second}

	answers := review.Answers{
		{Refutations: []review.Judgment{{ID: "F1", Verdict: review.Refuted, Why: "line 12"}}},
		{Refutations: []review.Judgment{{ID: "F1", Verdict: review.Refuted, Why: "call site"}}},
	}

	finding, blocking := review.BlockingAfter(findings, answers)
	if !blocking {
		t.Fatal("BlockingAfter found nothing: F2 has no witness")
	}
	if finding.Path != "internal/auth/token.go" {
		t.Errorf("the block named %s, want the finding nobody refuted", finding.Path)
	}
}

// TestTheRefutationRoundsAreNamed: the first answer keeps the name it always had,
// so an old refutation is still read; the next two are numbered, and a fourth is
// refused: three witnesses is the ceiling.
func TestTheRefutationRoundsAreNamed(t *testing.T) {
	tree := "4a1c9f2e4a1c9f2e4a1c9f2e4a1c9f2e4a1c9f2e"

	name, err := review.RefutationName(tree, 1)
	if err != nil {
		t.Fatalf("round 1: %v", err)
	}
	if !strings.HasSuffix(name, ".refuted.json") || strings.Contains(name, ".refuted.1") {
		t.Errorf("round 1 = %q, want the legacy name", name)
	}

	for _, round := range []int{2, 3} {
		name, err := review.RefutationName(tree, round)
		if err != nil {
			t.Fatalf("round %d: %v", round, err)
		}
		if !strings.Contains(name, ".refuted."+string(rune('0'+round))+".json") {
			t.Errorf("round %d = %q, want a numbered file", round, name)
		}
	}

	for _, round := range []int{0, 4} {
		if _, err := review.RefutationName(tree, round); err == nil {
			t.Errorf("round %d was accepted; a tree takes at most %d", round, review.MaxWitnesses)
		}
	}
}
