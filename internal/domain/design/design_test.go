package design_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/design"
)

// fixture is the manifest the skill writes, kept as bytes: the format is a
// contract with every client, so a test reads it the way an older writer wrote it.
const fixture = `{
  "title": "Pricing page",
  "brief": "Pricing page for a note-taking app aimed at students",
  "feature": "odd/tasks/pricing-page.md",
  "artboards": [
    { "file": "01-comparison-table.html", "title": "Comparison table", "tradeoff": "Dense and scannable; weaker emotional pull", "width": 1440, "height": 900 },
    { "file": "02-single-plan-hero.html", "title": "Single plan hero", "tradeoff": "Strong focus; hides plan differences", "width": 1440, "height": 900 }
  ]
}
`

func TestAManifestFromTheSkillParsesAndRendersBack(t *testing.T) {
	got, err := design.ParseManifest([]byte(fixture))
	if err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
	if got.Title != "Pricing page" || len(got.Artboards) != 2 {
		t.Fatalf("manifest = %+v", got)
	}
	if got.Feature != "odd/tasks/pricing-page.md" {
		t.Errorf("the feature link is lost: %q", got.Feature)
	}

	// Round trip: the canonical bytes parse to the same manifest.
	again, err := design.ParseManifest(design.RenderManifest(got))
	if err != nil {
		t.Fatalf("ParseManifest(Render): %v", err)
	}
	if again.Title != got.Title || len(again.Artboards) != len(got.Artboards) || again.Artboards[0].Tradeoff != got.Artboards[0].Tradeoff {
		t.Fatalf("round trip lost content: %+v", again)
	}
}

// A manifest that the canvas cannot draw is refused by name, so the skill learns
// what to fix instead of a design that opens empty.
func TestAManifestTheCanvasCannotDrawIsRefusedByName(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"not json", "{", "valid JSON"},
		{"no title", `{"artboards":[{"file":"a.html","tradeoff":"x","width":1,"height":1}]}`, "no title"},
		{"no artboard", `{"title":"x","artboards":[]}`, "no artboard"},
		{"a path instead of a file", `{"title":"x","artboards":[{"file":"../a.html","tradeoff":"x","width":1,"height":1}]}`, "has no path"},
		{"the same file twice", `{"title":"x","artboards":[{"file":"a.html","tradeoff":"x","width":1,"height":1},{"file":"a.html","tradeoff":"y","width":1,"height":1}]}`, "twice"},
		{"no viewport", `{"title":"x","artboards":[{"file":"a.html","tradeoff":"x"}]}`, "no viewport"},
		{"no tradeoff", `{"title":"x","artboards":[{"file":"a.html","width":1,"height":1}]}`, "tradeoff"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := design.ParseManifest([]byte(c.raw))
			if err == nil {
				t.Fatalf("a manifest that cannot be drawn parsed")
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Fatalf("error %q does not say %q", err, c.want)
			}
		})
	}
}

// commentsFixture is the sample the skill documents, verbatim: a client that has
// no command reads this file, so its shape is part of the contract.
const commentsFixture = `{
  "version": 1,
  "comments": [
    { "id": "k3x9p2ab", "file": "02-bold.html", "selector": "body > header > h1", "excerpt": "Ship faster", "comment": "Make the headline shorter.", "createdAt": "2026-10-26T10:00:00.000Z", "status": "open", "sentAt": null, "resolvedAt": null },
    { "id": "m7q2w9cd", "file": "02-bold.html", "selector": "#cta", "excerpt": "Start", "comment": "The button jumps during the hover transition.", "screenshot": ".screenshots/m7q2w9cd.png", "createdAt": "2026-10-26T10:05:00.000Z", "status": "open", "sentAt": null, "resolvedAt": null }
  ]
}
`

func TestTheCommentsFileFromTheSkillParses(t *testing.T) {
	got, err := design.ParseComments([]byte(commentsFixture))
	if err != nil {
		t.Fatalf("ParseComments: %v", err)
	}
	if len(got.Open()) != 2 {
		t.Fatalf("open comments = %d, want 2", len(got.Open()))
	}
	if got.Comments[1].Screenshot != ".screenshots/m7q2w9cd.png" {
		t.Errorf("the screenshot is lost: %+v", got.Comments[1])
	}
}

// A version this binary does not know is refused by name: accepting the future in
// silence is how a format stops being a contract.
func TestACommentsVersionThisBinaryDoesNotKnowIsRefused(t *testing.T) {
	raw := strings.Replace(commentsFixture, `"version": 1`, `"version": 9`, 1)
	_, err := design.ParseComments([]byte(raw))
	if err == nil || !strings.Contains(err.Error(), "version 9") {
		t.Fatalf("error = %v, want the version named", err)
	}
}

// Resolving keeps every entry, in order, and names the ids it did not find: the
// file is the history of what a person asked for, and an agent that resolves a typo
// must learn it was a typo.
func TestResolveKeepsTheHistoryAndNamesWhatItDidNotFind(t *testing.T) {
	got, err := design.ParseComments([]byte(commentsFixture))
	if err != nil {
		t.Fatalf("ParseComments: %v", err)
	}

	after, missing := got.Resolve([]string{"k3x9p2ab", "nope"}, "2026-10-31T00:00:00.000Z")

	if len(after.Comments) != 2 {
		t.Fatalf("resolving dropped entries: %d", len(after.Comments))
	}
	if after.Comments[0].Status != design.StatusResolved || after.Comments[0].ResolvedAt == "" {
		t.Errorf("the comment was not resolved: %+v", after.Comments[0])
	}
	if after.Comments[1].Status != design.StatusOpen {
		t.Errorf("a comment nobody named was resolved: %+v", after.Comments[1])
	}
	if len(missing) != 1 || missing[0] != "nope" {
		t.Errorf("missing = %v, want [nope]", missing)
	}
	if len(after.Open()) != 1 {
		t.Errorf("open after resolve = %d, want 1", len(after.Open()))
	}
}
