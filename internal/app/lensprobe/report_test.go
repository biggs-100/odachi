package lensprobe_test

import (
	"testing"

	"github.com/biggs-100/odachi/internal/app/lensprobe"
	"github.com/biggs-100/odachi/internal/domain/probe"
	"github.com/biggs-100/odachi/internal/domain/review"
)

// row builds one lens's score for a draw: how many it found, called low and
// missed, and what the findings nobody planted claimed.
func row(lens string, found, low, missed, noise, notes int) probe.Score {
	score := probe.Score{Lenses: []probe.LensScore{{
		Lens:   lens,
		Total:  found + low + missed,
		Found:  found,
		Low:    low,
		Missed: missed,
	}}}
	for i := 0; i < noise; i++ {
		score.Noise = append(score.Noise, probe.Noise{Lens: lens, Finding: review.Finding{Path: "x.go", Line: i, Severity: review.Critical}})
	}
	for i := 0; i < notes; i++ {
		score.Noise = append(score.Noise, probe.Noise{Lens: lens, Finding: review.Finding{Path: "y.go", Line: i, Severity: review.Suggestion}})
	}
	return score
}

// TestSummarizeKeepsTheMedianAndTheRange: the reading of several draws is the
// middle value and the range it moved in. The mean is not used: one draw that
// finds seven things nobody planted must not drag the centre.
func TestSummarizeKeepsTheMedianAndTheRange(t *testing.T) {
	t.Parallel()

	runs := lensprobe.Summarize([]probe.Score{
		row(review.Risk, 1, 0, 0, 4, 1),
		row(review.Risk, 0, 0, 1, 2, 3),
		row(review.Risk, 1, 0, 0, 4, 1),
	}, []probe.Tally{
		{Noise: 10, Notes: 0},
		{Noise: 1, Notes: 6},
		{Noise: 4, Notes: 2},
	})

	if runs.Count != 3 {
		t.Fatalf("Count = %d, want 3", runs.Count)
	}
	risk := runs.Lenses[0]
	if risk.Lens != review.Risk {
		t.Fatalf("row = %+v, want risk first", risk)
	}
	if risk.Found.Median != 1 || risk.Found.Min != 0 || risk.Found.Max != 1 {
		t.Errorf("found = %+v, want median 1 in 0..1", risk.Found)
	}
	if risk.Missed.Median != 0 || risk.Missed.Max != 1 {
		t.Errorf("missed = %+v, want median 0 and max 1", risk.Missed)
	}
	if runs.Noise.Median != 4 || runs.Noise.Min != 2 || runs.Noise.Max != 4 {
		t.Errorf("noise = %+v, want median 4 in 2..4", runs.Noise)
	}
	if runs.Notes.Median != 1 || runs.Notes.Min != 1 || runs.Notes.Max != 3 {
		t.Errorf("notes = %+v, want median 1 in 1..3", runs.Notes)
	}
	if runs.Clean.Median != 4 || runs.Clean.Min != 1 || runs.Clean.Max != 10 {
		t.Errorf("clean noise = %+v, want median 4 in 1..10", runs.Clean)
	}
	if runs.CleanNotes.Median != 2 || runs.CleanNotes.Min != 0 || runs.CleanNotes.Max != 6 {
		t.Errorf("clean notes = %+v, want median 2 in 0..6", runs.CleanNotes)
	}
}

// TestSummarizeAveragesTheTwoMiddleDraws: an even number of draws has no middle
// one, and the reading takes the two it has.
func TestSummarizeAveragesTheTwoMiddleDraws(t *testing.T) {
	t.Parallel()

	runs := lensprobe.Summarize([]probe.Score{
		row(review.Risk, 1, 0, 0, 1, 0),
		row(review.Risk, 0, 0, 1, 2, 0),
		row(review.Risk, 1, 0, 0, 3, 0),
		row(review.Risk, 0, 0, 1, 4, 0),
	}, []probe.Tally{
		{Noise: 0, Notes: 0},
		{Noise: 2, Notes: 2},
		{Noise: 4, Notes: 4},
		{Noise: 6, Notes: 6},
	})

	if runs.Noise.Median != 2 || runs.Noise.Min != 1 || runs.Noise.Max != 4 {
		t.Errorf("noise = %+v, want median 2 in 1..4", runs.Noise)
	}
	if runs.Clean.Median != 3 || runs.Clean.Min != 0 || runs.Clean.Max != 6 {
		t.Errorf("clean noise = %+v, want median 3 in 0..6", runs.Clean)
	}
	if runs.CleanNotes.Median != 3 || runs.CleanNotes.Min != 0 || runs.CleanNotes.Max != 6 {
		t.Errorf("clean notes = %+v, want median 3 in 0..6", runs.CleanNotes)
	}
}

// TestSummarizeWithNoDrawMeasuresNothing: a reading with no draw is empty, not a
// perfect score. The caller refuses that case, and this is the value it gets if
// it ever forgets.
func TestSummarizeWithNoDrawMeasuresNothing(t *testing.T) {
	t.Parallel()

	runs := lensprobe.Summarize(nil, nil)
	if runs.Count != 0 || len(runs.Lenses) != 0 {
		t.Errorf("runs = %+v, want nothing measured", runs)
	}
	if runs.Clean.Median != 0 || runs.Clean.Max != 0 {
		t.Errorf("clean noise = %+v, want zeros", runs.Clean)
	}
	if runs.CleanNotes.Median != 0 || runs.CleanNotes.Max != 0 {
		t.Errorf("clean notes = %+v, want zeros", runs.CleanNotes)
	}
}
