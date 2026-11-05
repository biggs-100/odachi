package main

import (
	"os"
	"path/filepath"
	"testing"
)

// The walk names the project's CSS and skips the folders that hold other people's
// copies: the report and the sync must see the same sources.
func TestDesignSourcesReadsTheCSSAndTheirTimes(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "app.css"), []byte(":root{--ink:#111}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "node_modules", "dep"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "node_modules", "dep", "vendor.css"), []byte(":root{--x:#000}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "notes.md"), []byte("not a source"), 0o644); err != nil {
		t.Fatal(err)
	}

	sources, err := designSources(dir)
	if err != nil {
		t.Fatalf("designSources: %v", err)
	}
	if len(sources) != 1 || sources[0].Path != "app.css" {
		t.Fatalf("sources = %+v, want only app.css", sources)
	}
	if sources[0].Modified.IsZero() {
		t.Errorf("the source has no modification time")
	}
	if sources[0].Content == "" {
		t.Errorf("the source has no content for the sync")
	}
}
