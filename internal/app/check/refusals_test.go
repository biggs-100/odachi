package check_test

import (
	"testing"

	"github.com/biggs-100/odachi/internal/refusal"
)

// baselineFile is the frozen list of refusals this package had the day the ratchet
// was written. It only shrinks: a refusal that names no way out and is not here
// fails the suite, and one that disappears fails it too, so the number in the file
// follows the debt down instead of drifting (0287, mechanism in internal/refusal).
const baselineFile = "refusals.baseline"

// TestEveryRefusalNamesAWayOutOrIsInTheBaseline is the ratchet over this package.
// A refusal is one message built for `blocked(...)`: every one either names
// something the reader can run -a command, a flag- or is frozen in the baseline
// with the rest of today's debt.
//
// What this does not prove: that the named continuation works. A sentence that
// points at a command which runs and does nothing passes here.
func TestEveryRefusalNamesAWayOutOrIsInTheBaseline(t *testing.T) {
	found, err := refusal.Walk(".")
	if err != nil {
		t.Fatalf("walking this package: %v", err)
	}
	baseline, err := refusal.ReadBaseline(baselineFile)
	if err != nil {
		t.Fatalf("reading %s: %v", baselineFile, err)
	}

	for _, row := range refusal.Grew(found.Sentences, baseline.Sentences) {
		t.Errorf("a refusal names no way out and is not in %s: %s\nname the command in its sentence, or add it to the baseline on purpose", baselineFile, row)
	}
	for _, row := range refusal.Shrunk(found.Sentences, baseline.Sentences) {
		t.Errorf("this refusal is gone; lower %s: %s", baselineFile, row)
	}

	for file, count := range found.Unreadable {
		want, ok := baseline.Counts[file]
		switch {
		case !ok:
			t.Errorf("%s has %d refusal(s) the walk cannot read and no count in %s", file, count, baselineFile)
		case count > want:
			t.Errorf("%s has %d refusal(s) the walk cannot read and %s froze %d: a message built from a value cannot be judged, and it may not grow", file, count, baselineFile, want)
		case count < want:
			t.Errorf("%s now has %d refusal(s) the walk cannot read; lower %s", file, count, baselineFile)
		}
	}
	for file := range baseline.Counts {
		if _, ok := found.Unreadable[file]; !ok {
			t.Errorf("%s no longer builds a refusal from a value; remove its count from %s", file, baselineFile)
		}
	}
}
