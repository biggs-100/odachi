package taskdoc

import (
	"regexp"
	"strings"
)

// This file owns the half of the waiver that is about one row instead of the whole
// feature. `Review: waived` says this feature lands without a verdict; a row waiver
// says the opposite -a verdict exists and it did ask for changes- and answers
// exactly one finding, by the fingerprint the verdict files it under.
//
// The record lives in the history because the history is what travels with the
// repository and what a later reader reads. The verdict lives in the clone's store,
// so a decision stored there would not travel with the evidence it judges.

// waivedRowPattern is the line `odachi review waive --row` writes: the fingerprint
// of the row the human answered, in the same hex the verdict prints.
var waivedRowPattern = regexp.MustCompile(`waived the row ([0-9a-f]{16})`)

// WaivedRows is the fingerprint of every blocking row this document answered, in
// the order the human wrote them. A line that waves at a row without naming one is
// prose: it waives nothing.
func (d TaskDoc) WaivedRows() []string {
	rows := make([]string, 0, len(d.History))
	for _, line := range d.History {
		match := waivedRowPattern.FindStringSubmatch(strings.ToLower(line))
		if match == nil {
			continue
		}
		if containsRow(rows, match[1]) {
			continue
		}
		rows = append(rows, match[1])
	}
	return rows
}

// WaivedRow reports whether this document answered this exact row. The fingerprint
// is the whole of the identity: another row, another tree or changed bytes are
// another decision, and the gate asks again.
func (d TaskDoc) WaivedRow(fingerprint string) bool {
	wanted := strings.ToLower(strings.TrimSpace(fingerprint))
	if wanted == "" {
		return false
	}
	return containsRow(d.WaivedRows(), wanted)
}

// containsRow is the linear lookup these lists deserve: a task document holds a
// handful of decisions, not thousands.
func containsRow(rows []string, wanted string) bool {
	for _, row := range rows {
		if row == wanted {
			return true
		}
	}
	return false
}
