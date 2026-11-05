package design_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/design"
)

const handoffsJSON = `{
  "version": 1,
  "chosen": "02-b.html",
  "handoffs": {
    "02-b.html": {"status": "implemented", "sentAt": "t1", "updatedAt": "t2", "feature": "odd/tasks/pricing.md", "previewUrl": "http://localhost:3000"}
  }
}`

// The file round-trips and keeps the decision: one chosen artboard, one state, one
// preview that the canvas can frame.
func TestHandoffsRoundTrip(t *testing.T) {
	parsed, err := design.ParseHandoffs([]byte(handoffsJSON))
	if err != nil {
		t.Fatalf("ParseHandoffs: %v", err)
	}
	if parsed.Chosen != "02-b.html" {
		t.Errorf("Chosen = %q", parsed.Chosen)
	}
	entry := parsed.Handoffs["02-b.html"]
	if entry.Status != design.HandoffImplemented || entry.PreviewURL != "http://localhost:3000" || entry.Feature != "odd/tasks/pricing.md" {
		t.Errorf("entry = %+v", entry)
	}

	rendered := design.RenderHandoffs(parsed)
	again, err := design.ParseHandoffs(rendered)
	if err != nil {
		t.Fatalf("ParseHandoffs(RenderHandoffs()): %v", err)
	}
	if again.Chosen != parsed.Chosen || again.Handoffs["02-b.html"].Status != parsed.Handoffs["02-b.html"].Status {
		t.Errorf("the file does not round-trip:\n%s", rendered)
	}
}

// The three states are the only states: a hand-edited file that invents one is
// refused by name instead of drawn as a badge nobody knows.
func TestAHandoffWithAnUnknownStateIsRefused(t *testing.T) {
	_, err := design.ParseHandoffs([]byte(`{"version":1,"handoffs":{"01-a.html":{"status":"done"}}}`))
	if err == nil || !strings.Contains(err.Error(), "done") {
		t.Fatalf("err = %v, want the unknown state named", err)
	}
}

// The preview is the running app on this machine: a public URL would make the canvas
// a browser for the internet, and a query or a credential is not the app's address.
func TestOnlyALoopbackPreviewIsAccepted(t *testing.T) {
	for _, raw := range []string{"http://localhost:3000", "http://127.0.0.1:8080/app", "https://[::1]:5173"} {
		if _, err := design.LoopbackURL(raw); err != nil {
			t.Errorf("LoopbackURL(%q): %v", raw, err)
		}
	}
	for _, raw := range []string{"http://example.com", "ftp://localhost", "http://user:pass@localhost:3000", "http://localhost:3000/?x=1"} {
		if _, err := design.LoopbackURL(raw); err == nil {
			t.Errorf("LoopbackURL(%q) was accepted", raw)
		}
	}
}
