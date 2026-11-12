package review_test

import (
	"testing"

	"github.com/biggs-100/odachi/internal/domain/review"
)

// TestAbsentDrawsReadAsOne: the field says how many answers paid for each lens, and
// a verdict written before it existed was paid by exactly one. Reading silence as
// zero would turn every stored approval into a review nobody paid for.
func TestAbsentDrawsReadAsOne(t *testing.T) {
	quiet, err := review.Seal(review.Verdict{
		Reviewer: "revisor",
		Lenses:   []string{review.Reliability},
	}, tree, "HEAD", now)
	if err != nil {
		t.Fatalf("sealing a verdict with no draws: %v", err)
	}
	if got := quiet.DrawsFor(review.Reliability); got != 1 {
		t.Errorf("DrawsFor on a verdict with no draws = %d, want 1", got)
	}

	paid, err := review.Seal(review.Verdict{
		Reviewer: "revisor",
		Lenses:   []string{review.Reliability},
		Draws:    map[string]int{review.Reliability: 3},
	}, tree, "HEAD", now)
	if err != nil {
		t.Fatalf("sealing a verdict with draws: %v", err)
	}
	if got := paid.DrawsFor(review.Reliability); got != 3 {
		t.Errorf("DrawsFor = %d, want the three draws the file says", got)
	}
	if got := paid.DrawsFor(review.Risk); got != 1 {
		t.Errorf("DrawsFor on a lens the file does not count = %d, want 1", got)
	}
}

// TestADrawCountNeedsALensAndAtLeastOneAnswer: the count is stamped by the tool, so
// a key the lens vocabulary does not know and a count of zero are both refused
// instead of stored: a lens paid by no answer was never judged.
func TestADrawCountNeedsALensAndAtLeastOneAnswer(t *testing.T) {
	for name, draws := range map[string]map[string]int{
		"a lens nobody knows": {"speed": 2},
		"no answer at all":    {review.Reliability: 0},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := review.Seal(review.Verdict{
				Reviewer: "revisor",
				Lenses:   []string{review.Reliability},
				Draws:    draws,
			}, tree, "HEAD", now)
			if err == nil {
				t.Fatalf("sealing with draws %v was accepted", draws)
			}
		})
	}
}

// TestTheDrawsSurviveTheFile: the count is in the file and comes back from it. A
// field the reader drops would make every verdict look like one draw, which is the
// silence this field exists to remove.
func TestTheDrawsSurviveTheFile(t *testing.T) {
	sealed, err := review.Seal(review.Verdict{
		Reviewer: "revisor",
		Lenses:   []string{review.Reliability, review.Risk},
		Draws: map[string]int{
			review.Reliability: 3,
			review.Risk:        1,
		},
	}, tree, "HEAD", now)
	if err != nil {
		t.Fatalf("sealing: %v", err)
	}

	parsed, err := review.Parse(review.Render(sealed))
	if err != nil {
		t.Fatalf("parsing what the tool rendered: %v", err)
	}
	if parsed.DrawsFor(review.Reliability) != 3 || parsed.DrawsFor(review.Risk) != 1 {
		t.Errorf("draws = %v, want reliability 3 and risk 1", parsed.Draws)
	}
}
