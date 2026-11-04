package clean_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/clean"
)

// TestTheReportSaysTheSwitchIsBlocked: the one line never promises bytes --apply
// will refuse to free, and the failure is counted once even when the two review
// classes share one walk.
func TestTheReportSaysTheSwitchIsBlocked(t *testing.T) {
	f := newFake(map[string]string{
		".git/odachi/reviews/aaaa.json": "{}",
		".git/odachi/scratch/plan.txt":  "plan",
	}).brokenDir(".git/odachi/reviews", errors.New("access is denied"))

	report := run(f).Report

	if !strings.Contains(report.Message, "cannot free anything") {
		t.Errorf("message = %q, want the blocked switch named", report.Message)
	}
	if strings.Contains(report.Message, "would free") {
		t.Errorf("message = %q, promises bytes the switch refuses to free", report.Message)
	}
	if !strings.Contains(report.Message, "1 path could not be read") {
		t.Errorf("message = %q, want the one path counted once", report.Message)
	}
}

// TestTheCloseAndTheReportTellOneStory: a folder beyond the shape is the class's
// shape in both answers, and a real failure blocks both with the same count. The
// rule is asked once, so the report and the close cannot disagree.
func TestTheCloseAndTheReportTellOneStory(t *testing.T) {
	deep := newFake(map[string]string{
		".git/odachi/scratch/a/b/c/d/e/f/deep.txt": "deep",
	})
	deepReport := run(deep).Report
	deepGrowth, problem := clean.Measure(clean.Options{Dir: dir, Store: store, KeepVerdicts: 10}, deep)
	if problem != nil {
		t.Fatalf("Measure: %v", problem)
	}
	if deepGrowth.Unreadable != 0 {
		t.Errorf("the close calls a folder beyond the shape unreadable: %d", deepGrowth.Unreadable)
	}
	if strings.Contains(deepReport.Message, "cannot free anything") {
		t.Errorf("the report blocks on a folder beyond the shape:\n%s", deepReport.Message)
	}

	broken := newFake(map[string]string{
		".git/odachi/reviews/aaaa.json": "{}",
	}).brokenDir(".git/odachi/reviews", errors.New("access is denied"))
	brokenReport := run(broken).Report
	brokenGrowth, problem := clean.Measure(clean.Options{Dir: dir, Store: store, KeepVerdicts: 10}, broken)
	if problem != nil {
		t.Fatalf("Measure: %v", problem)
	}
	if brokenGrowth.Unreadable != 1 {
		t.Errorf("the close counts %d unreadable paths, want the one", brokenGrowth.Unreadable)
	}
	if !strings.Contains(brokenReport.Message, "1 path could not be read") {
		t.Errorf("the report and the close disagree about the store:\n%s", brokenReport.Message)
	}
	if !strings.Contains(brokenGrowth.Line(), "1 path could not be read") {
		t.Errorf("the close line = %q, want the same count the report gives", brokenGrowth.Line())
	}
}
