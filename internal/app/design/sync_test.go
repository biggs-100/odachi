package design_test

import (
	"path/filepath"
	"strings"
	"testing"

	appdesign "github.com/biggs-100/odachi/internal/app/design"
)

// theme is the shape sync reads: custom properties and a font family, in a
// stylesheet the project owns.
const theme = `
:root { --color-ink: #1b1b1b; --color-paper: #fafafa; --space-2: 8px; --radius-card: 8px; }
body { font-family: Inter, system-ui; }
`

func TestSyncReadsTheTokensAndWritesTheSystem(t *testing.T) {
	files := newFakeFiles()

	got := appdesign.Sync([]appdesign.Source{{Path: "src/theme.css", Content: theme}}, files, "raiz")

	if got.Code != appdesign.CodeOK {
		t.Fatalf("Code = %d: %s", got.Code, got.Message)
	}
	markdown := files.files[filepath.Join("raiz", "design", "system.md")]
	for _, want := range []string{"--color-ink", "#1b1b1b", "src/theme.css", "Inter"} {
		if !strings.Contains(markdown, want) {
			t.Errorf("system.md does not say %q:\n%s", want, markdown)
		}
	}
	html := files.files[filepath.Join("raiz", "design", "system.html")]
	if !strings.Contains(html, "background:#1b1b1b") {
		t.Errorf("system.html does not draw the colour:\n%s", html)
	}
	if len(files.writes) != 2 {
		t.Errorf("writes = %v, want the two system files", files.writes)
	}
}

// A project with no source gets a refusal, never a system invented from nothing.
func TestSyncWithoutSourcesRefuses(t *testing.T) {
	files := newFakeFiles()

	got := appdesign.Sync(nil, files, "raiz")

	if got.Code != appdesign.CodeBlocked {
		t.Fatalf("Code = %d, want blocked", got.Code)
	}
	if len(files.writes) != 0 {
		t.Errorf("a refused sync wrote %v", files.writes)
	}
}

// The handoff names the artboard, its tradeoff and the alternatives it was chosen
// over: the decision stays visible in the instruction.
func TestHandoffNamesTheArtboardAndItsAlternatives(t *testing.T) {
	files := newFakeFiles().
		put("design/pricing/design.json", `{"title":"Pricing","brief":"for students","feature":"odd/tasks/pricing.md","artboards":[{"file":"01-a.html","title":"Table","tradeoff":"dense","width":1440,"height":900},{"file":"02-b.html","title":"Hero","tradeoff":"focus","width":1440,"height":900}]}`)

	got := appdesign.Run(appdesign.Options{Dir: "raiz", Action: appdesign.ActionHandoff, Slug: "pricing/02-b.html"}, files)

	if got.Code != appdesign.CodeOK {
		t.Fatalf("Code = %d: %s", got.Code, got.Message)
	}
	text := strings.Join(got.Lines, "\n")
	for _, want := range []string{"design/pricing/02-b.html", "Hero", "focus", "odd/tasks/pricing.md", "01-a.html", "the code is the proof"} {
		if !strings.Contains(text, want) {
			t.Errorf("the handoff does not say %q:\n%s", want, text)
		}
	}
}

// An artboard nobody designed is refused, and the refusal names the ones that exist.
func TestHandoffOnAnArtboardThatDoesNotExistIsBlocked(t *testing.T) {
	files := newFakeFiles().
		put("design/pricing/design.json", `{"title":"Pricing","brief":"b","artboards":[{"file":"01-a.html","title":"A","tradeoff":"x","width":1,"height":1}]}`)

	got := appdesign.Run(appdesign.Options{Dir: "raiz", Action: appdesign.ActionHandoff, Slug: "pricing/09-z.html"}, files)

	if got.Code != appdesign.CodeBlocked {
		t.Fatalf("Code = %d, want blocked", got.Code)
	}
	if !strings.Contains(got.Message, "01-a.html") {
		t.Errorf("the refusal does not name the artboards: %q", got.Message)
	}
}
