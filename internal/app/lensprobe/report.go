package lensprobe

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/biggs-100/odachi/internal/domain/probe"
	"github.com/biggs-100/odachi/internal/domain/review"
)

// Runs is one measurement made of several draws: the median of each number and
// the range it moved in.
//
// The median and not the mean, because one draw that finds seven things nobody
// planted must not drag the centre of three quiet ones. The range rides next to
// it because the range is what 0264 found: the same clean code gave ten
// findings, then one, then four, and a single number hid that.
type Runs struct {
	// Count is how many draws the measurement holds.
	Count int
	// Lenses is one row per lens the key names, in the canonical order.
	Lenses []LensRuns
	// Noise is what the findings nobody planted claimed on the seeded arm, per
	// draw: a defect at a place the key does not name.
	Noise Spread
	// Notes is what those findings only suggested, per draw.
	Notes Spread
	// Clean is the same claim read on the arm with nothing planted, per draw.
	Clean Spread
	// CleanNotes is what the clean arm's findings only suggested, per draw.
	CleanNotes Spread
}

// LensRuns is one lens's row across the draws.
type LensRuns struct {
	Lens   string
	Found  Spread
	Low    Spread
	Missed Spread
}

// Spread is a value with the range it was observed in: the median of the draws,
// and the smallest and largest one.
type Spread struct {
	Median int
	Min    int
	Max    int
}

// lensorder is the order the rows print in, the same one the score uses.
var lensorder = []string{review.Risk, review.Readability, review.Reliability, review.Resilience}

// Summarize folds the draws into one reading. A draw that does not score a lens
// counts as zero found, zero low and zero missed for it, which is what an absent
// row means. The clean arm has no key, so it arrives as a tally per draw, read by
// the same function the seeded arm's noise goes through.
func Summarize(scores []probe.Score, clean []probe.Tally) Runs {
	runs := Runs{Count: len(scores)}
	cleanNoise := make([]int, 0, len(clean))
	cleanNotes := make([]int, 0, len(clean))
	for _, tally := range clean {
		cleanNoise = append(cleanNoise, tally.Noise)
		cleanNotes = append(cleanNotes, tally.Notes)
	}
	runs.Clean = spread(cleanNoise)
	runs.CleanNotes = spread(cleanNotes)
	if len(scores) == 0 {
		return runs
	}

	noise := make([]int, 0, len(scores))
	notes := make([]int, 0, len(scores))
	for _, score := range scores {
		tally := score.Tally()
		noise = append(noise, tally.Noise)
		notes = append(notes, tally.Notes)
	}
	runs.Noise = spread(noise)
	runs.Notes = spread(notes)

	for _, name := range lensorder {
		row := LensRuns{Lens: name}
		found := make([]int, 0, len(scores))
		low := make([]int, 0, len(scores))
		missed := make([]int, 0, len(scores))
		present := false
		for _, score := range scores {
			var hit *probe.LensScore
			for i := range score.Lenses {
				if score.Lenses[i].Lens == name {
					hit = &score.Lenses[i]
					break
				}
			}
			if hit == nil {
				found = append(found, 0)
				low = append(low, 0)
				missed = append(missed, 0)
				continue
			}
			present = true
			found = append(found, hit.Found)
			low = append(low, hit.Low)
			missed = append(missed, hit.Missed)
		}
		if !present {
			continue
		}
		row.Found = spread(found)
		row.Low = spread(low)
		row.Missed = spread(missed)
		runs.Lenses = append(runs.Lenses, row)
	}
	return runs
}

// spread folds one number measured once per draw: the median and the range.
// An empty list is a zero with no range, because nothing was measured.
func spread(values []int) Spread {
	if len(values) == 0 {
		return Spread{}
	}
	sorted := append([]int(nil), values...)
	sort.Ints(sorted)
	middle := len(sorted) / 2
	out := Spread{Min: sorted[0], Max: sorted[len(sorted)-1], Median: sorted[middle]}
	if len(sorted)%2 == 0 {
		out.Median = (sorted[middle-1] + sorted[middle]) / 2
	}
	return out
}

// renderRuns prints the reading of the draws: one row per lens with the median
// of what it found and the range it moved in, then both arms, each one split by
// what the findings nobody planted claimed.
func renderRuns(runs Runs) []string {
	lines := make([]string, 0, len(runs.Lenses)+4)
	lines = append(lines, fmt.Sprintf("%-12s %13s %13s %13s", "lens", "found", "low", "missed"))
	for _, row := range runs.Lenses {
		lines = append(lines, fmt.Sprintf("%-12s %13s %13s %13s", row.Lens, span(row.Found), span(row.Low), span(row.Missed)))
	}
	lines = append(lines, fmt.Sprintf("%-12s %13s", "seeded noise", span(runs.Noise)))
	lines = append(lines, fmt.Sprintf("%-12s %13s", "seeded notes", span(runs.Notes)))
	lines = append(lines, fmt.Sprintf("%-12s %13s", "clean noise", span(runs.Clean)))
	lines = append(lines, fmt.Sprintf("%-12s %13s", "clean notes", span(runs.CleanNotes)))
	return lines
}

// span prints one number as its median and, when the draws moved, the range.
func span(s Spread) string {
	if s.Min == s.Max {
		return strconv.Itoa(s.Median)
	}
	return strconv.Itoa(s.Median) + " (" + strconv.Itoa(s.Min) + ".." + strconv.Itoa(s.Max) + ")"
}

// Summary is the one line the reading ends with: the draws, every lens's median
// with its range when it moved, and both arms split by what a finding claimed.
func (r Runs) Summary() string {
	parts := make([]string, 0, len(r.Lenses))
	for _, row := range r.Lenses {
		parts = append(parts, row.Lens+" "+strconv.Itoa(row.Found.Median)+"/"+strconv.Itoa(row.Found.Median+row.Low.Median+row.Missed.Median))
	}
	return fmt.Sprintf("runs %d · %s · seeded noise %s · seeded notes %s · clean noise %s · clean notes %s",
		r.Count, strings.Join(parts, ", "), span(r.Noise), span(r.Notes), span(r.Clean), span(r.CleanNotes))
}
