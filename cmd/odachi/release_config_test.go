package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// What a release proves is decided by two files no Go code reads:
// .goreleaser.yaml says what is signed and catalogued, and release.yml says
// which identity signs it and which tools exist. Keyless signing cannot run
// outside Actions, so this test pins the contract instead of the run: remove
// either half and the next release still looks normal while publishing
// signed-by-nobody checksums and an archive nobody can inventory. The real
// proof is the release run itself, not this test.
func TestTheReleaseIsSignedKeyless(t *testing.T) {
	goreleaser := readAtRepoRoot(t, ".goreleaser.yaml")
	for _, want := range []string{
		"signs:",
		"artifacts: checksum",
		`signature: "${artifact}.sigstore.json"`,
		`"--bundle=${signature}"`,
	} {
		if !strings.Contains(goreleaser, want) {
			t.Errorf(".goreleaser.yaml lost %q: the release would publish unsigned checksums", want)
		}
	}

	for _, want := range []string{
		"sboms:",
		"artifacts: archive",
	} {
		if !strings.Contains(goreleaser, want) {
			t.Errorf(".goreleaser.yaml lost %q: the release would ship archives nobody can inventory", want)
		}
	}

	workflow := readAtRepoRoot(t, filepath.Join(".github", "workflows", "release.yml"))
	for _, want := range []string{
		"id-token: write",
		"sigstore/cosign-installer",
		"anchore/sbom-action/download-syft",
	} {
		if !strings.Contains(workflow, want) {
			t.Errorf("release.yml lost %q: the release would fail at the worst moment", want)
		}
	}
}

func readAtRepoRoot(t *testing.T, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", rel))
	if err != nil {
		t.Fatalf("could not read %s from the repository: %v", rel, err)
	}
	return string(data)
}
