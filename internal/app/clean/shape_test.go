package clean_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/app/clean"
)

// classNamed is the row a human reads for one class, so a test asks the same
// report the human does instead of a private field.
func classNamed(t *testing.T, report clean.Report, name string) clean.Class {
	t.Helper()
	for _, class := range report.Classes {
		if class.Name == name {
			return class
		}
	}
	t.Fatalf("the report has no class %q", name)
	return clean.Class{}
}

// TestAFolderBeyondTheWalkIsOutsideTheShape: the walk has a bottom on purpose -a
// class cannot promise to read a tree that has no end-, and a folder past it is
// the class's shape, not a failure the switch has to pay for.
func TestAFolderBeyondTheWalkIsOutsideTheShape(t *testing.T) {
	f := newFake(nil)
	f.put(".git/odachi/scratch/a/b/c/d/e/f/deep.txt", "deep", time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC))

	report := run(f).Report
	scratch := classNamed(t, report, "agent scratch")

	if len(scratch.Problems) != 0 {
		t.Errorf("a folder beyond the walk is reported as a failure: %v", scratch.Problems)
	}
	if scratch.Beyond != 1 {
		t.Errorf("Beyond = %d, want the one folder the walk did not enter", scratch.Beyond)
	}
	if !strings.Contains(strings.Join(report.Lines(), "\n"), "1 folder beyond the shape") {
		t.Errorf("the report does not say the class has a folder beyond its shape:\n%s", strings.Join(report.Lines(), "\n"))
	}
}

// TestARealReadFailureIsStillAProblem: a folder that cannot be read is not the
// shape's edge. The switch keeps refusing while one exists, and the report names
// the path.
func TestARealReadFailureIsStillAProblem(t *testing.T) {
	f := newFake(nil).brokenDir(".git/odachi/scratch", errors.New("permission denied"))

	report := run(f).Report
	scratch := classNamed(t, report, "agent scratch")

	if len(scratch.Problems) != 1 {
		t.Fatalf("Problems = %v, want the one folder it could not read", scratch.Problems)
	}
	if scratch.Beyond != 0 {
		t.Errorf("Beyond = %d, want none: an unreadable folder is not the shape's edge", scratch.Beyond)
	}
	if !strings.Contains(scratch.Problems[0], "permission denied") {
		t.Errorf("the problem does not name the cause: %q", scratch.Problems[0])
	}

	blocked := apply(f, clean.Options{Apply: true})
	if blocked.Code != clean.CodeBlocked {
		t.Fatalf("apply = %d, want blocked while a folder cannot be read", blocked.Code)
	}
}
