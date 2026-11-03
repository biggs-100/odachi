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
