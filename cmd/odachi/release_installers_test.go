package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestTheReleasePublishesTheInstallers: the README teaches
// `releases/latest/download/install.ps1`, and a release that does not carry it
// answers 404. The goreleaser config is the only place that decides which files
// travel with the tag, so the test pins the two installers there.
func TestTheReleasePublishesTheInstallers(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", ".goreleaser.yaml"))
	if err != nil {
		t.Fatalf("read .goreleaser.yaml: %v", err)
	}

	text := string(data)
	for _, want := range []string{
		"release:",
		"extra_files:",
		"scripts/install.ps1",
		"scripts/install.sh",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("the release config does not carry %q: the documented install URL would 404", want)
		}
	}
}
