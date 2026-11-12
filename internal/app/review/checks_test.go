package review_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/review"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
	storeentry "github.com/biggs-100/odachi/internal/domain/store"
)

// lintCheck is the one tool most of these tests declare.
var lintCheck = domainreview.Check{
	Name:     "lint",
	Command:  "golangci-lint run ./...",
	Severity: domainreview.Warning,
}

// approveChecks is the clone's agreement, written where the store keeps it: the
// declaration travels with the repository and the agreement to run it does not.
func (p *project) approveChecks(checks []domainreview.Check) {
	approval := domainreview.NewCheckApproval(checks, "biggs", "2026-11-12T09:00:00Z")
	p.files.files[filepath.Join(store, storeentry.Checks)] = domainreview.RenderCheckApproval(approval)
}

// TestTheDeclaredCheckRunsOnTheCandidate: the tool runs in the project the prompt
// names, before the reviewer is paid, and its words travel into the prompt.
func TestTheDeclaredCheckRunsOnTheCandidate(t *testing.T) {
	p := newProject()
	p.approveChecks([]domainreview.Check{lintCheck})
	p.shell.script = []fakeAnswer{
		{output: "a.go:1:1: something is wrong\n", exitCode: 1},
		{output: `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`},
	}

	result := p.run(review.Options{Checks: []domainreview.Check{lintCheck}})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	if len(p.shell.commands) != 1 {
		t.Fatalf("commands = %v, want the check to have run once", p.shell.commands)
	}
	if p.shell.commands[0] != lintCheck.Command {
		t.Errorf("the first command was %q, want the declared check", p.shell.commands[0])
	}
	if len(p.shell.dirs) == 0 || p.shell.dirs[0] != dir {
		t.Errorf("the check ran in %v, want the candidate at %q", p.shell.dirs, dir)
	}
	if !strings.Contains(result.Prompt, "## The checks that ran") {
		t.Fatalf("the prompt does not carry the checks:\n%s", result.Prompt)
	}
	for _, want := range []string{"- lint: failed (exit 1)", "a.go:1:1: something is wrong"} {
		if !strings.Contains(result.Prompt, want) {
			t.Errorf("the prompt does not carry %q", want)
		}
	}
}

// TestACheckThisCloneNeverApprovedDoesNotRun: the declaration travels with the
// repository and the agreement to execute it does not, so a cloned review.toml
// cannot run anything. The prompt says what was skipped, because a review that
// stayed quiet would report on a tree nobody measured.
func TestACheckThisCloneNeverApprovedDoesNotRun(t *testing.T) {
	p := newProject()

	result := p.run(review.Options{Checks: []domainreview.Check{lintCheck}})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if len(p.shell.commands) != 0 {
		t.Errorf("commands = %v, want nothing to run before an approval", p.shell.commands)
	}
	for _, want := range []string{
		"## The checks that did not run",
		"this clone has not approved them",
		"odachi review checks approve",
	} {
		if !strings.Contains(result.Prompt, want) {
			t.Errorf("the prompt does not carry %q:\n%s", want, result.Prompt)
		}
	}
}

// TestAListTheProjectChangedIsNotApproved: the approval covers the exact list, so
// a command that gained a flag is a tool nobody agreed to run.
func TestAListTheProjectChangedIsNotApproved(t *testing.T) {
	p := newProject()
	p.approveChecks([]domainreview.Check{lintCheck})
	changed := domainreview.Check{
		Name:     "lint",
		Command:  "golangci-lint run --fix ./...",
		Severity: domainreview.Warning,
	}

	result := p.run(review.Options{Checks: []domainreview.Check{changed}})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if len(p.shell.commands) != 0 {
		t.Errorf("commands = %v, want nothing to run for a list that changed", p.shell.commands)
	}
	if !strings.Contains(result.Prompt, "approved a different list") {
		t.Errorf("the prompt does not say the approval covers another list:\n%s", result.Prompt)
	}
}

// TestAPassingCheckSaysSoWithoutOutput: a tool that found nothing is one line, and
// the reviewer reads it as the part of the tree it does not have to inspect.
func TestAPassingCheckSaysSoWithoutOutput(t *testing.T) {
	p := newProject()
	p.approveChecks([]domainreview.Check{lintCheck})

	result := p.run(review.Options{Checks: []domainreview.Check{lintCheck}})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Prompt, "- lint: passed") {
		t.Errorf("the prompt does not say the check passed:\n%s", result.Prompt)
	}
}

// TestACheckThatCannotStartIsARowNotARefusal: a tool nobody installed is the
// project's mistake, and it must be readable as a row instead of hiding the
// reviewer's answer behind a crash.
func TestACheckThatCannotStartIsARowNotARefusal(t *testing.T) {
	p := newProject()
	p.approveChecks([]domainreview.Check{lintCheck})
	p.shell.script = []fakeAnswer{
		{err: errors.New("executable file not found in PATH")},
		{output: `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`},
	}

	result := p.run(review.Options{Checks: []domainreview.Check{lintCheck}})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	for _, want := range []string{"- lint: could not run", "executable file not found in PATH"} {
		if !strings.Contains(result.Prompt, want) {
			t.Errorf("the prompt does not carry %q:\n%s", want, result.Prompt)
		}
	}
}

// TestACheckThatTimesOutSaysHowLongItWaited: a hung tool is not a pass, and the
// prompt names the ceiling so the human can raise it or drop the tool.
func TestACheckThatTimesOutSaysHowLongItWaited(t *testing.T) {
	p := newProject()
	p.approveChecks([]domainreview.Check{lintCheck})
	p.shell.script = []fakeAnswer{
		{timedOut: true},
		{output: `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`},
	}

	result := p.run(review.Options{
		Checks:       []domainreview.Check{lintCheck},
		CheckTimeout: 45,
	})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Prompt, "- lint: timed out · stopped after 45 s") {
		t.Errorf("the prompt does not name the timed-out check:\n%s", result.Prompt)
	}
	if p.shell.timeout != 45 {
		t.Errorf("the check ran with a %ds ceiling, want the project's 45", p.shell.timeout)
	}
}

// TestTheChecksRunOnceHoweverManyGroupsPay: two models read two prompts about one
// candidate, and the tools that measured it ran once. A tool paid twice for one
// tree is money thrown at the same answer.
func TestTheChecksRunOnceHoweverManyGroupsPay(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.approveChecks([]domainreview.Check{lintCheck})
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

	result := p.run(review.Options{
		Exec:       "revisor --model {model}",
		Lens:       "reliability,risk",
		Because:    "the test splits the run by model",
		LensModels: map[string]string{"risk": "sharp/model"},
		Checks:     []domainreview.Check{lintCheck},
	})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	runs := 0
	for _, command := range p.shell.commands {
		if command == lintCheck.Command {
			runs++
		}
	}
	if runs != 1 {
		t.Errorf("the check ran %d times over one candidate, want once: %v", runs, p.shell.commands)
	}
}

// TestAProjectThatDeclaresNoChecksRunsNothing: the absence of the table is the
// absence of the feature, and the reviewer's command is the only thing executed.
func TestAProjectThatDeclaresNoChecksRunsNothing(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

	result := p.run(review.Options{Exec: "revisor --model {model}"})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if len(p.shell.commands) != 1 || p.shell.commands[0] == lintCheck.Command {
		t.Errorf("commands = %v, want only the reviewer", p.shell.commands)
	}
	if strings.Contains(result.Prompt, "## The checks that ran") {
		t.Errorf("the prompt claims checks nobody declared:\n%s", result.Prompt)
	}
}
