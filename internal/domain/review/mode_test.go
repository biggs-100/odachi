package review_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/review"
)

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
