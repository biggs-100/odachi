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
// server's show event, updates the URL in place, and re-renders. The artboards are
// sandboxed, so the picker and the mode message are the only way in.
func TestTheCanvasMovesOnAShowEvent(t *testing.T) {
	page := string(Canvas())
	for _, want := range []string{"show:", "searchParams.set(\"design\"", "sandbox", "allow-scripts", "odachi:pick", "odachi:comment-mode", "Chosen", "Compare", "handoffs", "previewUrl", "Back to canvas", "odachi:pins", "odachi:blur", "Send all", "Collect", "enterFocus", "shiftKey"} {
		if !strings.Contains(page, want) {
			t.Errorf("the canvas does not carry %q", want)
		}
	}
	if strings.Contains(page, "allow-scripts allow-same-origin") || strings.Contains(page, `sandbox", "allow-same-origin`) {
		t.Errorf("the canvas grants the artboard its own origin: the sandbox is not a sandbox")
	}
}
