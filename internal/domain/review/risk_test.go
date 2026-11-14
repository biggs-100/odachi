package review_test

import (
	"slices"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/review"
)

// TestEachRiskAsksForItsLens: the table `commits` publishes -each risk names the
// lens and the verifier it owes- becomes the table the gate reads. A risk that
// asked for nothing was a sentence: 0299 declared `contract` on three commits and
// the verdict closed with `required: none` (0301).
func TestEachRiskAsksForItsLens(t *testing.T) {
	cases := []struct {
		risk string
		lens string
	}{
		{"data", review.Reliability},
		{"contract", review.Reliability},
		{"no-test", review.Reliability},
		{"security", review.Risk},
		{"concurrency", review.Resilience},
		{"delivery", review.Resilience},
		{" Security ", review.Risk},
	}

	for _, c := range cases {
		lens, ok := review.LensForRisk(c.risk)
		if !ok || lens != c.lens {
			t.Errorf("LensForRisk(%q) = %q, %v; want %q, true", c.risk, lens, ok, c.lens)
		}
	}

	for _, unknown := range []string{"", "none", "everything"} {
		if lens, ok := review.LensForRisk(unknown); ok {
			t.Errorf("LensForRisk(%q) = %q, true; want no lens", unknown, lens)
		}
	}
}

// TestTheDeclaredLensesKeepTheirOrderAndHaveNoRepeats: the demand is a list the
// prompt and the refusal both read, so the same commits produce the same order.
func TestTheDeclaredLensesKeepTheirOrderAndHaveNoRepeats(t *testing.T) {
	cases := []struct {
		name  string
		risks []string
		want  []string
	}{
		{"nothing declared", nil, nil},
		{"one risk", []string{"security"}, []string{review.Risk}},
		{
			name:  "the order the commits came in",
			risks: []string{"contract", "security", "concurrency"},
			want:  []string{review.Reliability, review.Risk, review.Resilience},
		},
		{
			name:  "a repeated risk asks once",
			risks: []string{"security", "security", "none"},
			want:  []string{review.Risk},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := review.LensesForRisk(c.risks); !slices.Equal(got, c.want) {
				t.Errorf("LensesForRisk = %v, want %v", got, c.want)
			}
		})
	}
}
