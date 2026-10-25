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
	goreleaser := liveLines(readAtRepoRoot(t, ".goreleaser.yaml"))
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

	workflow := liveLines(readAtRepoRoot(t, filepath.Join(".github", "workflows", "release.yml")))
	for _, want := range []string{
		"id-token: write",
		"sigstore/cosign-installer@",
		"anchore/sbom-action/download-syft@",
	} {
		if !strings.Contains(workflow, want) {
			t.Errorf("release.yml lost %q: the release would fail at the worst moment", want)
		}
	}

	// Both actions hold the OIDC identity: a major tag can be moved to another
	// commit, and that commit would sign its own checksums with our name.
	for _, action := range []string{"sigstore/cosign-installer@", "anchore/sbom-action/download-syft@"} {
		for _, line := range strings.Split(workflow, "\n") {
			if !strings.Contains(line, action) {
				continue
			}
			pin := strings.TrimSpace(strings.SplitN(line, action, 2)[1])
			if len(strings.Fields(pin)) == 0 || len(strings.Fields(pin)[0]) != 40 {
				t.Errorf("release.yml pins %q with %q, want a 40-character commit SHA: a tag can move", action, pin)
			}
		}
	}
}

// liveLines drops comment lines: commenting out a block is the ordinary way to
// disable it, and raw bytes would stay green while the release stopped signing.
func liveLines(text string) string {
	var live strings.Builder
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		live.WriteString(line)
		live.WriteString("\n")
	}
	return live.String()
}

func readAtRepoRoot(t *testing.T, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", rel))
	if err != nil {
		t.Fatalf("could not read %s from the repository: %v", rel, err)
	}
	return string(data)
}
