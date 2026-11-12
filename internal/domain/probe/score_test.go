package probe_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/probe"
	"github.com/biggs-100/odachi/internal/domain/review"
)

func key() probe.Manifest {
	return probe.Manifest{
		Window: 3,
		Defects: []probe.Defect{
			{ID: "risk-sql", Path: "server.go", Line: 30, Lens: review.Risk, Severity: review.Critical, Why: "the id is concatenated"},
			{ID: "readability-name", Path: "limits.go", Line: 11, Lens: review.Readability, Severity: review.Suggestion, Why: "a name that says nothing"},
		},
	}
}

func finding(path string, line int, severity review.Severity, causal bool) review.Finding {
	return review.Finding{Path: path, Line: line, Severity: severity, Causal: causal}
}

// TestScoreCountsACatchAtThePlace: the lens that names the defect, within the
// window, at or above the floor, is what the probe counts as found. Everything
// else is a different number because it is a different failure.
func TestScoreCountsACatchAtThePlace(t *testing.T) {
	t.Parallel()

	score := probe.ScoreOne(key(), map[string]review.Verdict{
		review.Risk: {
			Findings: []review.Finding{finding("server.go", 31, review.Critical, true)},
		},
		review.Readability: {
			Findings: []review.Finding{finding("limits.go", 9, review.Warning, true)},
		},
	})

	risk := score.Lenses[0]
	if risk.Lens != review.Risk || risk.Found != 1 || risk.Missed != 0 || risk.Low != 0 {
		t.Errorf("risk row = %+v, want 1 found", risk)
	}
	readability := score.Lenses[1]
	if readability.Lens != review.Readability || readability.Found != 1 {
		t.Errorf("readability row = %+v, want 1 found above its floor", readability)
	}
	if len(score.Noise) != 0 {
		t.Errorf("noise = %+v, want none", score.Noise)
	}
	if got, want := score.Summary(), "risk 1/1, readability 1/1 · noise 0, notes 0"; got != want {
		t.Errorf("Summary = %q, want %q", got, want)
	}
}

// TestScoreSeparatesASoftCallFromSilence: the right place under the floor is a
// miss of judgment, not a miss of reading, and the report says which. A finding
// outside the window is silence even when the file is right.
func TestScoreSeparatesASoftCallFromSilence(t *testing.T) {
	t.Parallel()

	score := probe.ScoreOne(key(), map[string]review.Verdict{
		review.Risk: {
			Findings: []review.Finding{finding("server.go", 30, review.Warning, true)},
		},
		review.Readability: {
			Findings: []review.Finding{finding("limits.go", 20, review.Suggestion, true)},
		},
	})

	risk := score.Lenses[0]
	if risk.Low != 1 || risk.Found != 0 || risk.Missed != 0 {
		t.Errorf("risk row = %+v, want one low call", risk)
	}
	readability := score.Lenses[1]
	if readability.Missed != 1 || len(readability.MissedIDs) != 1 || readability.MissedIDs[0] != "readability-name" {
		t.Errorf("readability row = %+v, want the defect reported as missed", readability)
	}
}

// TestScoreCallsEverythingNobodyPlantedNoise: a finding at a place the key does
// not name, on another path, or about bytes the candidate did not cause, is
// noise. Noise is what a lens costs: it is counted once, with its author.
func TestScoreCallsEverythingNobodyPlantedNoise(t *testing.T) {
	t.Parallel()

	score := probe.ScoreOne(key(), map[string]review.Verdict{
		review.Risk: {
			Findings: []review.Finding{
				finding("server.go", 30, review.Critical, true),
				finding("cache.go", 32, review.Warning, true),
				finding("server.go", 45, review.Suggestion, false),
			},
		},
		review.Reliability: {
			Findings: []review.Finding{finding("limits_test.go", 8, review.Warning, true)},
		},
	})

	if score.Lenses[0].Found != 1 {
		t.Errorf("risk row = %+v, want the planted defect found", score.Lenses[0])
	}
	if len(score.Noise) != 3 {
		t.Fatalf("noise = %+v, want three findings nobody planted", score.Noise)
	}
	for _, item := range score.Noise {
		if item.Finding.Path == "server.go" && item.Finding.Line == 30 {
			t.Errorf("noise includes the planted defect: %+v", item)
		}
	}
	if score.Noise[0].Lens == "" {
		t.Error("a noise row does not name the lens that said it")
	}
}

// TestTallyReadsASuggestionAsANoteAndAClaimAsNoise: on the arm with nothing
// planted, a suggestion says the code could be better and a warning or above
// claims a defect. Counted as one number, the reviewer that improves the fixture
// is scored as noisy and the one that says nothing is scored as precise (0281),
// and the comparison between models measures a style instead of an error.
func TestTallyReadsASuggestionAsANoteAndAClaimAsNoise(t *testing.T) {
	t.Parallel()

	score := probe.ScoreOne(key(), map[string]review.Verdict{
		review.Risk: {
			Findings: []review.Finding{
				finding("server.go", 30, review.Critical, true),
				finding("cache.go", 12, review.Suggestion, true),
				finding("cache.go", 20, review.Warning, true),
				finding("cache.go", 28, review.Blocker, true),
			},
		},
	})

	tally := score.Tally()
	if tally.Notes != 1 {
		t.Errorf("notes = %d, want 1: a suggestion is an improvement, not an invented defect", tally.Notes)
	}
	if tally.Noise != 2 {
		t.Errorf("noise = %d, want 2: a warning and a blocker claim defects nobody planted", tally.Noise)
	}
}

// TestTallyOfReadsTheArmWithNothingPlanted: the clean arm has no key, so what a
// reviewer says there is read by what it claims and nothing else. One function
// reads both arms, because two readings of the same vocabulary would disagree.
func TestTallyOfReadsTheArmWithNothingPlanted(t *testing.T) {
	t.Parallel()

	got := probe.TallyOf([]probe.Noise{
		{Finding: finding("server.go", 12, review.Suggestion, true)},
		{Finding: finding("cache.go", 8, review.Suggestion, true)},
		{Finding: finding("limits.go", 4, review.Critical, true)},
	})
	if got.Notes != 2 || got.Noise != 1 {
		t.Errorf("tally = %+v, want two notes and one claim", got)
	}
}

// TestScoreCreditsADefectFoundByAnotherLensWithoutCallingItFound: the key names
// the lens that owes the defect. A different lens finding it is a real catch by
// somebody and a miss for the row under test, never noise: the probe measures
// lenses, not the team.
func TestScoreCreditsADefectFoundByAnotherLensWithoutCallingItFound(t *testing.T) {
	t.Parallel()

	score := probe.ScoreOne(key(), map[string]review.Verdict{
		review.Readability: {
			Findings: []review.Finding{finding("server.go", 30, review.Critical, true)},
		},
	})

	risk := score.Lenses[0]
	if risk.Missed != 1 || risk.Found != 0 {
		t.Errorf("risk row = %+v, want the SQL defect reported as missed by risk", risk)
	}
	if len(score.Noise) != 0 {
		t.Errorf("noise = %+v, want none: another lens found a real defect", score.Noise)
	}
}

// TestScoreCountsAFindingAtThePlaceEvenWhenItSaysPreExisting: the key counts the
// place. A reviewer that names the planted line and calls it pre-existing judged
// the sentence, not the location, and one sentence is never both a miss and
// noise.
func TestScoreCountsAFindingAtThePlaceEvenWhenItSaysPreExisting(t *testing.T) {
	t.Parallel()

	score := probe.ScoreOne(key(), map[string]review.Verdict{
		review.Risk: {Findings: []review.Finding{finding("server.go", 30, review.Critical, false)}},
	})

	if score.Lenses[0].Found != 1 {
		t.Errorf("risk row = %+v, want the place counted as found", score.Lenses[0])
	}
	if len(score.Noise) != 0 {
		t.Errorf("noise = %+v, want none: the finding names a planted place", score.Noise)
	}
}

// TestScoreWithNoVerdictMeasuresSilence: a lens that produced no answer missed
// every defect it owed. An absent run is not a skipped measurement.
func TestScoreWithNoVerdictMeasuresSilence(t *testing.T) {
	t.Parallel()

	score := probe.ScoreOne(key(), map[string]review.Verdict{})
	if len(score.Lenses) != 2 {
		t.Fatalf("rows = %+v, want one per lens the key names", score.Lenses)
	}
	for _, row := range score.Lenses {
		if row.Missed != row.Total || row.Total == 0 {
			t.Errorf("row = %+v, want every defect missed", row)
		}
		if !strings.Contains(score.Summary(), row.Lens+" 0/"+itoa(row.Total)) {
			t.Errorf("Summary = %q, want it to name %s", score.Summary(), row.Lens)
		}
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	out := ""
	for n > 0 {
		out = string(rune('0'+n%10)) + out
		n /= 10
	}
	return out
}
