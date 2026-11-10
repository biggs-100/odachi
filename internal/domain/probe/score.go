package probe

import (
	"fmt"
	"sort"
	"strings"

	"github.com/biggs-100/odachi/internal/domain/review"
)

// Score is the measurement: what a reviewer found, what it missed, and what it
// reported that nobody planted.
//
// Three outcomes per defect, and they are not the same number because they do
// not teach the same thing:
//
//   - found: the lens named it, at the right place, with at least the severity
//     the key asks for, and caused by this candidate;
//   - low: the right place with a severity under the floor. The reviewer saw the
//     defect and called it harmless, which is a different failure from silence;
//   - missed: nothing at the place at all.
//
// Everything else a reviewer says is noise, counted once, with the lens that
// said it.
type Score struct {
	// Lenses is one row per lens, in the canonical order.
	Lenses []LensScore
	// Noise is every finding that matches no planted defect.
	Noise []Noise
}

// LensScore is one lens's row.
type LensScore struct {
	Lens   string
	Total  int
	Found  int
	Low    int
	Missed int
	// MissedIDs names the defects this lens never mentioned, so the report can
	// print the row a person has to read.
	MissedIDs []string
}

// Noise is one finding nobody planted.
type Noise struct {
	Lens    string
	Finding review.Finding
}

// scoreOrder is the order the rows are printed in: the same canonical order the
// review uses for its lenses.
var scoreOrder = []string{review.Risk, review.Readability, review.Reliability, review.Resilience}

// ScoreOne measures one run against the key. verdicts maps a lens to the
// verdict that lens answered: the probe runs one reviewer per lens, so the map
// says who said what, and a lens with no verdict is reported as missing every
// defect of its list rather than counted as silent.
//
// A nil entry for a lens is a run that produced no verdict at all: it missed
// those defects, which is exactly what silence measures.
func ScoreOne(manifest Manifest, verdicts map[string]review.Verdict) Score {
	rows := make([]LensScore, 0, len(scoreOrder))
	noise := make([]Noise, 0)

	findings := make([]lensFinding, 0)
	for _, lens := range scoreOrder {
		verdict, ok := verdicts[lens]
		if !ok {
			continue
		}
		for _, finding := range verdict.Findings {
			findings = append(findings, lensFinding{lens: lens, finding: finding})
		}
	}
	claimed := make([]bool, len(findings))

	for _, lens := range scoreOrder {
		row := LensScore{Lens: lens}
		for _, defect := range manifest.Defects {
			if defect.Lens != lens {
				continue
			}
			row.Total++
			matched, reached := matchedAt(findings, claimed, defect, manifest.Window)
			switch {
			case matched && reached:
				row.Found++
			case matched:
				row.Low++
			default:
				row.Missed++
				row.MissedIDs = append(row.MissedIDs, defect.ID)
			}
		}
		if row.Total > 0 {
			rows = append(rows, row)
		}
	}

	for i, item := range findings {
		if claimed[i] {
			continue
		}
		noise = append(noise, Noise{Lens: item.lens, Finding: item.finding})
	}
	sort.SliceStable(noise, func(i, j int) bool {
		if noise[i].Finding.Path != noise[j].Finding.Path {
			return noise[i].Finding.Path < noise[j].Finding.Path
		}
		return noise[i].Finding.Line < noise[j].Finding.Line
	})

	return Score{Lenses: rows, Noise: noise}
}

// lensFinding is one verdict's finding with the lens that answered it.
type lensFinding struct {
	lens    string
	finding review.Finding
}

// matchedAt reports whether one of this lens's findings sits at the defect's
// place, and whether it reached the severity floor.
//
// Every finding at that place is claimed, whatever lens said it: two findings on
// one planted line are one defect found, and a lens that is not the defect's own
// still found something real - it is a miss for the lens under test, never noise.
//
// Causal is not part of the place. A reviewer that names the right line and calls
// it pre-existing judged the sentence, not the location: the key counts the
// place, the noise counts what corresponds to no defect, and one sentence is
// never both.
func matchedAt(findings []lensFinding, claimed []bool, defect Defect, window int) (matched, reached bool) {
	for i, item := range findings {
		if claimed[i] {
			continue
		}
		if item.finding.Path != defect.Path {
			continue
		}
		if distance(item.finding.Line, defect.Line) > window {
			continue
		}
		claimed[i] = true
		if item.lens != defect.Lens {
			continue
		}
		matched = true
		if rank(item.finding.Severity) >= rank(defect.Severity) {
			reached = true
		}
	}
	return matched, reached
}

// distance is how many lines separate two line numbers, never negative.
func distance(a, b int) int {
	if a > b {
		return a - b
	}
	return b - a
}

// rank orders the closed severity vocabulary from least to most severe. The
// key's floor and the finding's severity are compared with it.
func rank(severity review.Severity) int {
	switch severity {
	case review.Blocker:
		return 3
	case review.Critical:
		return 2
	case review.Warning:
		return 1
	case review.Suggestion:
		return 0
	default:
		return -1
	}
}

// Percent is the found-over-total ratio as a whole number, for the report. A
// lens with no defects has no ratio and prints none.
func (l LensScore) Percent() int {
	if l.Total == 0 {
		return 0
	}
	return l.Found * 100 / l.Total
}

// Summary is the one line a person reads first: every lens's found/total, and
// how many findings nobody planted.
func (s Score) Summary() string {
	parts := make([]string, 0, len(s.Lenses))
	for _, row := range s.Lenses {
		parts = append(parts, fmt.Sprintf("%s %d/%d", row.Lens, row.Found, row.Total))
	}
	return fmt.Sprintf("%s · noise %d", strings.Join(parts, ", "), len(s.Noise))
}
