package change_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/change"
)

// A change with no document names itself in the trailer of its commit: the line
// is the vision, so it must survive the prose around it untouched.
func TestTheChangeIsNamedByItsTrailer(t *testing.T) {
	messages := []string{
		"restore the button\n\nThe wait state was lost in the rewrite.\n\nOdachi-Change: bring back the wait state on the pay button\n",
	}

	doc, ok, err := change.Parse(messages)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !ok {
		t.Fatalf("the trailer was not read")
	}
	if doc.Line != "bring back the wait state on the pay button" {
		t.Fatalf("line = %q", doc.Line)
	}
}

// A message with no trailer is the past: any project that never writes one keeps
// the answer it always had.
func TestAMessageWithNoTrailerIsNotAChange(t *testing.T) {
	doc, ok, err := change.Parse([]string{"the release body comes from the changelog\n"})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if ok || doc.Line != "" {
		t.Fatalf("a message with no trailer named a change: %+v", doc)
	}
}

// The same line twice is one decision written twice; two different lines are two
// changes in one range, and a reader cannot choose between them.
func TestTheSameLineTwiceIsOneChange(t *testing.T) {
	doc, ok, err := change.Parse([]string{
		"first\n\nOdachi-Change: one line\n",
		"second\n\nOdachi-Change: one line\n",
	})
	if err != nil || !ok || doc.Line != "one line" {
		t.Fatalf("the same line twice = %+v, %v, %v", doc, ok, err)
	}
}

func TestATrailerWithoutALineIsRefused(t *testing.T) {
	_, _, err := change.Parse([]string{"something\n\nOdachi-Change:   \n"})
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("error = %v, want the empty trailer named", err)
	}
}

func TestTwoDifferentTrailersAreRefusedByName(t *testing.T) {
	_, _, err := change.Parse([]string{
		"first\n\nOdachi-Change: one line\n",
		"second\n\nOdachi-Change: another line\n",
	})
	if err == nil {
		t.Fatalf("two different trailers parsed")
	}
	for _, want := range []string{"one line", "another line"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not name %q: %v", want, err)
		}
	}
}
