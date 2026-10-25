package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The signature of a release is decided by two files no Go code reads:
// .goreleaser.yaml says what is signed, and release.yml says which identity
// signs it. Keyless signing cannot run outside Actions, so this test pins the
// contract instead of the run: remove either half and the next release still
// looks normal while publishing signed-by-nobody checksums. The real proof is
// the release run itself, not this test.
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

	workflow := readAtRepoRoot(t, filepath.Join(".github", "workflows", "release.yml"))
	for _, want := range []string{
		"id-token: write",
		"sigstore/cosign-installer",
	} {
		if !strings.Contains(workflow, want) {
			t.Errorf("release.yml lost %q: cosign would have no identity to sign with", want)
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
