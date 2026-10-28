package review_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/review"
)

// TestAReviewOverTheSizeLimitStartsNoReviewer: a model that stalls at a prompt
// size does not answer late, it does not answer at all, and odachi knows the size
// before it spends anything. The refusal names both numbers and the two ways out,
// and no reviewer runs.
func TestAReviewOverTheSizeLimitStartsNoReviewer(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.output = `{"findings":[]}`

	result := p.run(review.Options{Exec: "revisor --model {model}", MaxPromptKB: 1})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 (%s)", result.Code, result.Message)
	}
	if len(p.shell.commands) != 0 {
		t.Fatalf("commands = %v, want none: the reviewer must not be started", p.shell.commands)
	}
	for _, want := range []string{"KB", "max_prompt_kb", "no reviewer was started"} {
		if !strings.Contains(strings.Join(result.Lines, "\n"), want) {
			t.Errorf("the refusal does not carry %q:\n%s", want, strings.Join(result.Lines, "\n"))
		}
	}
}

// TestAReviewUnderTheSizeLimitRuns: the limit is the project's, and a prompt that
// fits it is the normal case.
func TestAReviewUnderTheSizeLimitRuns(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

	if result := p.run(review.Options{Exec: "revisor --model {model}", MaxPromptKB: 1024}); result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
}

// TestThePrintedFormStillPrintsAndNamesTheSize: printing a prompt spends nothing,
// so a size over the limit cannot refuse the human their own debugging tool. It
// says the size where the run's notes go, so the human learns before pasting it
// into a reviewer that will stall.
func TestThePrintedFormStillPrintsAndNamesTheSize(t *testing.T) {
	p := newProject()

	result := p.run(review.Options{MaxPromptKB: 1})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0: the printed form spends nothing (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Prompt, "odachi review") {
		t.Error("the prompt was not printed")
	}
	if !strings.Contains(strings.Join(result.Lines, "\n"), "KB") {
		t.Errorf("the run does not name the size:\n%s", strings.Join(result.Lines, "\n"))
	}
}

// TestNoLimitMeansNoRefusal: a project that never wrote the key reviews as it
// always did. A default limit invented here would refuse prompts the human never
// agreed to refuse.
func TestNoLimitMeansNoRefusal(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

	if result := p.run(review.Options{Exec: "revisor --model {model}"}); result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
}
