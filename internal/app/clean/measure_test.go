package clean_test

import (
	"errors"
	"testing"

	"github.com/biggs-100/odachi/internal/app/clean"
)

// hashA is a tree name, long enough to be one and fixed so two runs agree.
const hashA = "4a1c9f7d2e3b5a6c8d9e0f1a2b3c4d5e6f7a8b9c"

// TestMeasureAnswersTheReportTotals: a close names the same numbers the report
// prints, or the two answers drift the moment one of them learns a class.
func TestMeasureAnswersTheReportTotals(t *testing.T) {
	f := newFake(map[string]string{
		".git/odachi/reviews/" + hashA + ".json":      "{}",
		".git/odachi/reviews/" + hashA + ".prompt.md": "question",
		".git/odachi/scratch/plan.txt":                "scratch",
		".git/odachi/backups/brief.md.bak-1":          "one",
		".git/odachi/backups/brief.md.bak-2":          "two",
		"odd/tasks/0001-a.md":                         "the plan",
		"proof/0001-a.txt":                            "the proof",
	})

	opts := clean.Options{Dir: dir, Store: store, KeepVerdicts: 0, KeepPrompts: 0, LimitsFrom: "odd/review.toml"}
	report := clean.Run(opts, f, nil).Report
	growth, problem := clean.Measure(opts, f)

	if problem != nil {
		t.Fatalf("Measure refused a healthy store: %v", problem)
	}
	if growth.Files != report.TotalFiles || growth.Bytes != report.TotalBytes {
		t.Errorf("Measure = %d files/%d B, the report says %d/%d",
			growth.Files, growth.Bytes, report.TotalFiles, report.TotalBytes)
	}
	if growth.DoomedFiles != report.DoomedFiles || growth.DoomedBytes != report.DoomedBytes {
		t.Errorf("Measure = %d doomed files/%d B, the report says %d/%d",
			growth.DoomedFiles, growth.DoomedBytes, report.DoomedFiles, report.DoomedBytes)
	}
}

// TestMeasureCountsWhatItCouldNotRead: a store this verb cannot see is not an
// empty store. The numbers it could read are named, and the count of paths it
// could not is carried so the close can say why `--apply` frees nothing.
func TestMeasureCountsWhatItCouldNotRead(t *testing.T) {
	f := newFake(nil).brokenDir(".git/odachi/reviews", errors.New("permission denied"))

	growth, problem := clean.Measure(clean.Options{Dir: dir, Store: store, KeepVerdicts: 10}, f)

	if problem != nil {
		t.Fatalf("Measure refused a store it could partly read: %v", problem)
	}
	if growth.Unreadable != 1 {
		t.Errorf("Unreadable = %d, want the one path it could not read", growth.Unreadable)
	}
}

// TestMeasureRefusesAFolderWithNoStore: a measurement of nothing is not zero
// bytes, and the close is told so instead of being handed a made-up number.
func TestMeasureRefusesAFolderWithNoStore(t *testing.T) {
	if _, problem := clean.Measure(clean.Options{Dir: dir}, newFake(nil)); problem == nil {
		t.Fatal("Measure answered a folder with no store")
	}
}

// TestTheGrowthLineNamesTheSwitch: the sentence a close prints says what the
// store holds and what the human's command would free, and it never claims a
// cleanup was made.
func TestTheGrowthLineNamesTheSwitch(t *testing.T) {
	cases := []struct {
		name   string
		growth clean.Growth
		want   string
	}{
		{
			name:   "bytes to free",
			growth: clean.Growth{Files: 1352, Bytes: 36595200, DoomedFiles: 581, DoomedBytes: 35651584},
			want:   "34.9 MB in 1352 files; odachi clean --apply would free 34.0 MB in 581 files",
		},
		{
			name:   "nothing to free",
			growth: clean.Growth{Files: 1, Bytes: 12},
			want:   "12 B in 1 file; odachi clean --apply would free nothing",
		},
		{
			name:   "a path it cannot read blocks the switch",
			growth: clean.Growth{Files: 1352, Bytes: 36595200, Unreadable: 3},
			want:   "34.9 MB in 1352 files; odachi clean --apply cannot free anything: 3 paths could not be read",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.growth.Line(); got != tc.want {
				t.Errorf("Line() = %q, want %q", got, tc.want)
			}
		})
	}
}
