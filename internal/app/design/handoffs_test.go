package design_test

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	appdesign "github.com/biggs-100/odachi/internal/app/design"
	"github.com/biggs-100/odachi/internal/domain/design"
)

const twoArtboards = `{"title":"Pricing","brief":"for students","feature":"odd/tasks/pricing.md","artboards":[{"file":"01-a.html","title":"Table","tradeoff":"dense","width":1440,"height":900},{"file":"02-b.html","title":"Hero","tradeoff":"focus","width":1440,"height":900}]}`

// The handoff is a decision and a state: the chosen artboard is written down with
// where the work stands, so the canvas and the next agent see the same thing.
func TestAHandoffRecordsTheChosenArtboardAndItsState(t *testing.T) {
	files := newFakeFiles().
		put("design/pricing/design.json", twoArtboards).
		put("odd/tasks/pricing.md", "# pricing\n\n## History\nL1 · \"start\"\n")

	got := appdesign.Run(appdesign.Options{Dir: "raiz", Action: appdesign.ActionHandoff, Slug: "pricing/02-b.html", At: "t1"}, files)
	if got.Code != appdesign.CodeOK {
		t.Fatalf("Code = %d: %s", got.Code, got.Message)
	}
	statePath := filepath.Join("raiz", "design", "pricing", "handoffs.json")
	first := readHandoffs(t, files.files[statePath])
	if first.Chosen != "02-b.html" {
		t.Errorf("Chosen = %q, want the artboard that was handed off", first.Chosen)
	}
	entry := first.Handoffs["02-b.html"]
	if entry.Status != design.HandoffSent || entry.SentAt != "t1" || entry.Feature != "odd/tasks/pricing.md" {
		t.Errorf("entry = %+v", entry)
	}

	got = appdesign.Run(appdesign.Options{
		Dir: "raiz", Action: appdesign.ActionHandoff, Slug: "pricing/02-b.html",
		Status: design.HandoffImplemented, Preview: "http://localhost:3000", Note: "done", At: "t2",
	}, files)
	if got.Code != appdesign.CodeOK {
		t.Fatalf("Code = %d: %s", got.Code, got.Message)
	}
	second := readHandoffs(t, files.files[statePath])
	entry = second.Handoffs["02-b.html"]
	if entry.Status != design.HandoffImplemented || entry.PreviewURL != "http://localhost:3000" || entry.Note != "done" {
		t.Errorf("entry = %+v", entry)
	}
	if entry.SentAt != "t1" || entry.UpdatedAt != "t2" {
		t.Errorf("the state lost its history: %+v", entry)
	}
	// The second write backs up the state file and the linked feature document; the
	// first handoff backed up the document before its log line landed.
	if len(files.backups) != 3 {
		t.Errorf("backups = %v, want the state and the feature document backed up", files.backups)
	}
}

// A state or a preview the canvas cannot draw is refused before anything is written:
// the file is the protocol, and a file that lies is worse than no file.
func TestAHandoffWithABadStateOrPreviewIsBlocked(t *testing.T) {
	cases := []struct {
		name string
		opts appdesign.Options
		want string
	}{
		{"state", appdesign.Options{Status: "done"}, "done"},
		{"preview", appdesign.Options{Preview: "http://example.com/app"}, "example.com"},
		{"query", appdesign.Options{Preview: "http://localhost:3000/?x=1"}, "query"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			files := newFakeFiles().put("design/pricing/design.json", twoArtboards)
			opts := c.opts
			opts.Dir = "raiz"
			opts.Action = appdesign.ActionHandoff
			opts.Slug = "pricing/02-b.html"
			got := appdesign.Run(opts, files)
			if got.Code != appdesign.CodeBlocked {
				t.Fatalf("Code = %d, want blocked", got.Code)
			}
			if !strings.Contains(got.Message, c.want) {
				t.Errorf("message = %q, want it to name %q", got.Message, c.want)
			}
			if len(files.writes) != 0 {
				t.Errorf("a refused handoff wrote %v", files.writes)
			}
		})
	}
}

// A design that came from a tracked feature writes the chosen artboard back into
// that document's log: the link goes both ways, and the decision is not only in the
// canvas.
func TestHandoffLogsTheChosenArtboardInItsFeature(t *testing.T) {
	files := newFakeFiles().
		put("design/pricing/design.json", twoArtboards).
		put("odd/tasks/pricing.md", "# pricing\n\n## History\nL1 · \"start\"\n")

	got := appdesign.Run(appdesign.Options{Dir: "raiz", Action: appdesign.ActionHandoff, Slug: "pricing/02-b.html", At: "t1"}, files)
	if got.Code != appdesign.CodeOK {
		t.Fatalf("Code = %d: %s", got.Code, got.Message)
	}
	doc := files.files[filepath.Join("raiz", "odd", "tasks", "pricing.md")]
	want := `L2 · "design/pricing/02-b.html chosen for this feature; the alternatives are 01-a.html"`
	if !strings.Contains(doc, want) {
		t.Errorf("the feature document does not carry the log line:\n%s", doc)
	}
	if text := strings.Join(got.Lines, "\n"); !strings.Contains(text, "logged in odd/tasks/pricing.md") {
		t.Errorf("the handoff does not say it logged: %s", text)
	}
}

// A feature link outside odd/tasks/ is refused before anything is written: the link
// is a promise about a document odachi owns, not a path a manifest can point at.
func TestHandoffRefusesAFeatureOutsideOddTasks(t *testing.T) {
	files := newFakeFiles().
		put("design/pricing/design.json", `{"title":"Pricing","brief":"b","feature":"scripts/run.md","artboards":[{"file":"01-a.html","title":"A","tradeoff":"x","width":1,"height":1}]}`)

	got := appdesign.Run(appdesign.Options{Dir: "raiz", Action: appdesign.ActionHandoff, Slug: "pricing/01-a.html"}, files)
	if got.Code != appdesign.CodeBlocked {
		t.Fatalf("Code = %d, want blocked", got.Code)
	}
	if !strings.Contains(got.Message, "scripts/run.md") {
		t.Errorf("the refusal does not name the link: %q", got.Message)
	}
	if len(files.writes) != 0 {
		t.Errorf("a refused handoff wrote %v", files.writes)
	}
}

func readHandoffs(t *testing.T, raw string) design.Handoffs {
	t.Helper()
	var h design.Handoffs
	if err := json.Unmarshal([]byte(raw), &h); err != nil {
		t.Fatalf("decode handoffs.json: %v\n%s", err, raw)
	}
	return h
}
