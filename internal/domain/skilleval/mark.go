// Package skilleval carries the shape of one skill's exam: the arms that are run,
// the behaviours an event stream can decide, and the record a run leaves. It is
// the second instrument after the lens probe, and it asks a different question:
// the probe asks whether a reader with a candidate is better than one without it,
// and this asks whether a skill changes the work an agent does.
package skilleval

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Mark is one skill taken out of the home for a run: where it lives, where it is
// while the arms run, and the journal that says so. A run that dies before it
// restores leaves the mark behind, and the next run puts the folder back before
// it measures anything: a measurement that eats a skill is worse than no
// measurement.
type Mark struct {
	Skill string `json:"skill"`
	From  string `json:"from"`
	To    string `json:"to"`
}

// ErrNothingToTake is the answer when the named skill is not installed where the
// mark says it should be. The caller names the path; this only says which.
var ErrNothingToTake = errors.New("the skill is not installed there")

// ParseJournal reads the marks a previous run left, newest last. An empty journal
// is an empty list, not an error: a clone that never ran an exam carries nothing.
func ParseJournal(data []byte) ([]Mark, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var marks []Mark
	if err := json.Unmarshal(data, &marks); err != nil {
		return nil, fmt.Errorf("the journal is not a list of marks: %w", err)
	}
	for _, mark := range marks {
		if mark.Skill == "" || mark.From == "" || mark.To == "" {
			return nil, fmt.Errorf("the journal carries a mark with no skill, no origin or no destination: %+v", mark)
		}
	}
	return marks, nil
}

// RenderJournal writes the marks as the journal reads them. It is the only writer:
// a mark that is not on disk is a skill nobody will put back.
func RenderJournal(marks []Mark) ([]byte, error) {
	if len(marks) == 0 {
		return nil, nil
	}
	data, err := json.MarshalIndent(marks, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("the journal cannot be written: %w", err)
	}
	return append(data, '\n'), nil
}
