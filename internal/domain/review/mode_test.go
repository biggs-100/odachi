package review_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/review"
)

// TestNormalizeKeepIsALifetimeOrATypo: the two numbers the project decides -how
// many review records and how many prompts survive `odachi clean --apply`-. Zero
// is the absence of a choice and means the default; a negative number or one
// nobody reaches is refused instead of silently becoming something else.
func TestNormalizeKeepIsALifetimeOrATypo(t *testing.T) {
	if got, err := review.NormalizeKeepVerdicts(0); err != nil || got != review.DefaultKeepVerdicts {
		t.Errorf("NormalizeKeepVerdicts(0) = %d, %v; want the default", got, err)
	}
	if got, err := review.NormalizeKeepVerdicts(3); err != nil || got != 3 {
		t.Errorf("NormalizeKeepVerdicts(3) = %d, %v; want 3", got, err)
	}
	if got, err := review.NormalizeKeepPrompts(0); err != nil || got != 0 {
		t.Errorf("NormalizeKeepPrompts(0) = %d, %v; want 0", got, err)
	}
	for _, bad := range []int{-1, review.MaxKeep + 1} {
		if _, err := review.NormalizeKeepVerdicts(bad); err == nil {
			t.Errorf("NormalizeKeepVerdicts(%d) accepted a lifetime that cannot be one", bad)
		}
		if _, err := review.NormalizeKeepPrompts(bad); err == nil {
			t.Errorf("NormalizeKeepPrompts(%d) accepted a lifetime that cannot be one", bad)
		}
	}
}

// A model is an identifier that replaces {model} inside a command: the review
// must never run a repository's string as shell code.
func TestNormalizeModelRefusesShellCode(t *testing.T) {
	accepted := []string{"opencode-go/kimi-k3.1-pro", "provider/name#variant", "a-b_c.d:e+f@g"}
	for _, model := range accepted {
		if got, err := review.NormalizeModel(model); err != nil || got != model {
			t.Errorf("NormalizeModel(%q) = %q, %v; want it accepted", model, got, err)
		}
	}

	refused := []string{"x; rm -rf ~", "a && b", "$(whoami)", "a b", "`id`", "a|b", "a>b"}
	for _, model := range refused {
		if _, err := review.NormalizeModel(model); err == nil {
			t.Errorf("NormalizeModel(%q) accepted shell code", model)
		}
	}
}

// TestNormalizeModeAcceptsTheTwoAnswers: the switch is on or off, case and space
// included, because a human writes the file.
func TestNormalizeModeAcceptsTheTwoAnswers(t *testing.T) {
	cases := map[string]string{
		"on":     review.ModeOn,
		" ON ":   review.ModeOn,
		"off":    review.ModeOff,
		" Off\n": review.ModeOff,
	}

	for input, want := range cases {
		got, err := review.NormalizeMode(input)
		if err != nil {
			t.Fatalf("NormalizeMode(%q) = %v", input, err)
		}
		if got != want {
			t.Errorf("NormalizeMode(%q) = %q, want %q", input, got, want)
		}
	}
}

// TestNormalizeModeRefusesAThirdAnswer: a typo must not silently leave the gate
// off, so the error names both valid values.
func TestNormalizeModeRefusesAThirdAnswer(t *testing.T) {
	for _, input := range []string{"", "true", "enabled", "maybe"} {
		got, err := review.NormalizeMode(input)
		if err == nil {
			t.Errorf("NormalizeMode(%q) = %q, want a refusal", input, got)
			continue
		}
		for _, want := range []string{review.ModeOn, review.ModeOff} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("NormalizeMode(%q) error %q does not name %q", input, err, want)
			}
		}
	}
}
