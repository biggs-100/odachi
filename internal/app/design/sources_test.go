package design_test

import (
	"strings"
	"testing"
	"time"

	appdesign "github.com/biggs-100/odachi/internal/app/design"
)

// A source that moved after the system was captured is named: a design that follows
// stale tokens follows values the project no longer has.
func TestListNamesStaleDesignSources(t *testing.T) {
	files := newFakeFiles().put("design/pricing/design.json", manifest)
	captured := time.Date(2026, 11, 4, 10, 0, 0, 0, time.UTC)
	moved := captured.Add(time.Hour)

	got := appdesign.Run(appdesign.Options{
		Dir:            "raiz",
		Action:         appdesign.ActionList,
		Sources:        []appdesign.Source{{Path: "app.css", Modified: moved}},
		SystemExists:   true,
		SystemModified: captured,
	}, files)

	text := strings.Join(got.Lines, "\n")
	if !strings.Contains(text, "older than 1 source") || !strings.Contains(text, "app.css") {
		t.Errorf("the list does not name the stale source:\n%s", text)
	}
}

// A project with sources but no system is told to sync before designing; a project
// with no sources says so and asks for nothing.
func TestListSaysWhenTheSystemIsMissing(t *testing.T) {
	files := newFakeFiles().put("design/pricing/design.json", manifest)

	got := appdesign.Run(appdesign.Options{
		Dir:     "raiz",
		Action:  appdesign.ActionList,
		Sources: []appdesign.Source{{Path: "app.css"}},
	}, files)
	if text := strings.Join(got.Lines, "\n"); !strings.Contains(text, "no design/system.md yet") {
		t.Errorf("the list does not ask for the first sync:\n%s", text)
	}

	got = appdesign.Run(appdesign.Options{Dir: "raiz", Action: appdesign.ActionList}, files)
	if text := strings.Join(got.Lines, "\n"); !strings.Contains(text, "design sources: none detected") {
		t.Errorf("the list does not say there are no sources:\n%s", text)
	}
}

// A system captured after every source is up to date, and the report says so instead
// of warning forever.
func TestListSaysTheSystemIsUpToDate(t *testing.T) {
	files := newFakeFiles().put("design/pricing/design.json", manifest)
	captured := time.Date(2026, 11, 4, 10, 0, 0, 0, time.UTC)

	got := appdesign.Run(appdesign.Options{
		Dir:            "raiz",
		Action:         appdesign.ActionList,
		Sources:        []appdesign.Source{{Path: "app.css", Modified: captured.Add(-time.Hour)}},
		SystemExists:   true,
		SystemModified: captured,
	}, files)

	if text := strings.Join(got.Lines, "\n"); !strings.Contains(text, "up to date") {
		t.Errorf("the list does not say the system is fresh:\n%s", text)
	}
}
