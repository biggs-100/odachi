package review_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/review"
)

// oneAnswer is what a fake reviewer answers for one draw: no findings, and the
// declaration every complete answer carries.
const oneAnswer = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

// TestReviewRefusesADrawCountItCannotPay: --runs is a count of answers, and a count
// outside the range is refused before the store is read or a reviewer is started.
// Eleven draws of four lenses is an experiment, and an experiment is designed
// somewhere else - the ceiling is the one `review probe` already names.
func TestReviewRefusesADrawCountItCannotPay(t *testing.T) {
	for _, runs := range []int{-1, review.MaxRuns + 1} {
		p := newProject()
		p.host.model = "host/default"
		p.shell.output = oneAnswer

		result := p.run(review.Options{Exec: "revisor --model {model}", Runs: runs})

		if result.Code != review.CodeBlocked {
			t.Fatalf("--runs %d = code %d, want blocked (%s)", runs, result.Code, result.Message)
		}
		if !strings.Contains(result.Message, strconv.Itoa(review.MaxRuns)) {
			t.Errorf("--runs %d does not name the ceiling %d: %s", runs, review.MaxRuns, result.Message)
		}
		if len(p.shell.commands) != 0 {
			t.Errorf("--runs %d started a reviewer: %v", runs, p.shell.commands)
		}
		if written := p.files.written(); len(written) != 0 {
			t.Errorf("--runs %d wrote %v, want nothing", runs, written)
		}
	}
}

// TestReviewWithoutAnAnswerHasNoDrawToCount: the bare form prints the question and
// asks nobody, so a count of answers means nothing there. The refusal names the
// flag that gives it a meaning, the way --model does.
func TestReviewWithoutAnAnswerHasNoDrawToCount(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"

	result := p.run(review.Options{Runs: 3})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want blocked (%s)", result.Code, result.Message)
	}
	for _, want := range []string{"--runs 3", "--exec"} {
		if !strings.Contains(result.Message, want) {
			t.Errorf("the refusal does not name %q: %s", want, result.Message)
		}
	}
}

// TestAZeroDrawCountIsTheCallerThatNeverSaid: the Go zero value is the run that
// chose no count, and it means one - the same rule an absent field follows in a
// stored verdict. Otherwise a stored verdict would depend on which caller asked.
func TestAZeroDrawCountIsTheCallerThatNeverSaid(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.output = oneAnswer

	result := p.run(review.Options{Exec: "revisor --model {model}"})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if len(p.shell.commands) != 1 {
		t.Errorf("reviewer calls = %v, want one", p.shell.commands)
	}
}
