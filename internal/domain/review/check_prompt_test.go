package review_test

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/biggs-100/odachi/internal/domain/review"
)

// promptWithChecks is the smallest prompt that carries the checks: the section
// lives in it and nothing else in the test needs a candidate.
func promptWithChecks(checks []review.CheckResult) string {
	return review.Prompt(review.PromptInput{
		Brief:  "the brief\n",
		Tree:   "abc1234",
		Base:   "def5678",
		Paths:  []string{"a.go"},
		Diff:   "+++ b/a.go\n",
		Checks: checks,
	})
}

// TestThePromptCarriesWhatTheChecksAnswered: the review reads what ran before it,
// so it neither repeats a failure nor claims the tree is clean.
func TestThePromptCarriesWhatTheChecksAnswered(t *testing.T) {
	prompt := promptWithChecks([]review.CheckResult{
		{Name: "lint", Command: "golangci-lint run ./...", Status: review.CheckPassed},
		{
			Name:    "vuln",
			Command: "govulncheck ./...",
			Status:  review.CheckFailed,
			Exit:    1,
			Output:  "Vulnerability #1: GO-2026-1234\n",
		},
	})

	for _, want := range []string{
		"## The checks that ran (deterministic, before any lens)",
		"- lint: passed",
		"- vuln: failed (exit 1)",
		"GO-2026-1234",
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("the prompt does not carry %q:\n%s", want, prompt)
		}
	}
}

// TestTheChecksComeBeforeTheLenses: the reading a model did not make is the one
// it must not repeat, so it travels before the specialists.
func TestTheChecksComeBeforeTheLenses(t *testing.T) {
	prompt := promptWithChecks([]review.CheckResult{
		{Name: "lint", Command: "golangci-lint run ./...", Status: review.CheckPassed},
	})

	checks := strings.Index(prompt, "## The checks that ran")
	lenses := strings.Index(prompt, "## The six rules")
	if checks < 0 || lenses < 0 {
		t.Fatalf("the prompt is missing a section: checks=%d rules=%d", checks, lenses)
	}
	if checks > lenses {
		t.Errorf("the checks section comes after the rules: the reviewer reads the candidate before the rules")
	}
}

// TestACandidateWithNoChecksPrintsNoSection: a project that declared no tools
// reads exactly the prompt this verb printed before the feature existed.
func TestACandidateWithNoChecksPrintsNoSection(t *testing.T) {
	prompt := promptWithChecks(nil)
	if strings.Contains(prompt, "## The checks that ran") {
		t.Errorf("the prompt claims checks nobody declared:\n%s", prompt)
	}
}

// TestCheckOutputIsClippedAndSaysHowMuchWasCut: a linter over a whole package can
// print thousands of lines, and a prompt that pays for them is a prompt nobody
// reads. What is cut is counted, never hidden.
func TestCheckOutputIsClippedAndSaysHowMuchWasCut(t *testing.T) {
	var printed strings.Builder
	for i := 1; i <= review.CheckOutputLines+12; i++ {
		fmt.Fprintf(&printed, "row %d\n", i)
	}

	clipped := review.ClipCheckOutput(printed.String())
	if strings.Count(clipped, "\n") > review.CheckOutputLines+1 {
		t.Errorf("ClipCheckOutput kept more than %d lines:\n%s", review.CheckOutputLines, clipped)
	}
	if !strings.Contains(clipped, "12 more lines") {
		t.Errorf("ClipCheckOutput does not say how much it cut:\n%s", clipped)
	}
	if !strings.Contains(clipped, "row 1\n") {
		t.Errorf("ClipCheckOutput dropped the first line:\n%s", clipped)
	}
}

// TestAShortOutputIsUntouched: the clip is for the noise, not for the answer.
func TestAShortOutputIsUntouched(t *testing.T) {
	output := "one line\n"
	if got := review.ClipCheckOutput(output); got != output {
		t.Errorf("ClipCheckOutput(%q) = %q, want it unchanged", output, got)
	}
}

// TestOneEnormousLineIsClippedByBytes: a tool that prints one minified line has
// spent the budget in a currency lines do not count, so the clip is charged there
// too -and it never cuts inside a rune, because a row that is not UTF-8 is a store
// the gate refuses to read-.
func TestOneEnormousLineIsClippedByBytes(t *testing.T) {
	output := strings.Repeat("ñ", review.CheckOutputBytes)

	clipped := review.ClipCheckOutput(output)
	if len(clipped) > review.CheckOutputBytes+64 {
		t.Errorf("ClipCheckOutput kept %d bytes, want at most %d plus the note", len(clipped), review.CheckOutputBytes)
	}
	if !utf8.ValidString(clipped) {
		t.Error("ClipCheckOutput cut inside a rune")
	}
	if !strings.Contains(clipped, "bytes cut") {
		t.Errorf("ClipCheckOutput does not say how much it cut:\n%s", clipped)
	}
}

// TestOnlyAPassedCheckIsPassed: every other answer -failed, timed out, or a
// command that could not start- is a row the verdict has to carry.
func TestOnlyAPassedCheckIsPassed(t *testing.T) {
	if (review.CheckResult{Status: review.CheckPassed}).Failed() {
		t.Error("a passed check is reported as failed")
	}
	for _, status := range []review.CheckStatus{
		review.CheckFailed, review.CheckTimedOut, review.CheckBroken,
	} {
		if !(review.CheckResult{Status: status}).Failed() {
			t.Errorf("a check that answered %q reads as passed", status)
		}
	}
}
