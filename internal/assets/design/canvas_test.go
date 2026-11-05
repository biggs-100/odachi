package design

import (
	"strings"
	"testing"
)

// The canvas can only name a gap it can see: the page reads the manifest's
// screenshot and calls the artboard that has none by name (0176).
func TestTheCanvasNamesAnArtboardWithoutAScreenshot(t *testing.T) {
	page := string(Canvas())
	for _, want := range []string{"artboard.screenshot", "no screenshot"} {
		if !strings.Contains(page, want) {
			t.Errorf("the canvas does not carry %q", want)
		}
	}
}

// A second open moves the tab that is already there: the page listens for the
// server's show event, updates the URL in place, and re-renders.
func TestTheCanvasMovesOnAShowEvent(t *testing.T) {
	page := string(Canvas())
	for _, want := range []string{"show:", "searchParams.set(\"design\""} {
		if !strings.Contains(page, want) {
			t.Errorf("the canvas does not carry %q", want)
		}
	}
}
