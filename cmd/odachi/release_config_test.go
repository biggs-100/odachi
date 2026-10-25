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
		"attestations: write",
		"sigstore/cosign-installer@",
		"anchore/sbom-action/download-syft@",
		"goreleaser/goreleaser-action@",
		"actions/attest@",
		"subject-checksums: ./dist/checksums.txt",
	} {
		if !strings.Contains(workflow, want) {
			t.Errorf("release.yml lost %q: the release would fail at the worst moment", want)
		}
	}

	assertActionsArePinned(t, workflow)
}

// assertActionsArePinned refuses a moving tag on any action of the job. Every
// step in a job with id-token: write can read the OIDC token, so a moved tag
// could ask Fulcio for this workflow's identity, sign its own checksums and
// rewrite the release: verify-blob would still pass. The whole file is checked,
// not a list, because the next action somebody adds inherits the identity too.
func assertActionsArePinned(t *testing.T, workflow string) {
	t.Helper()
	for _, line := range strings.Split(workflow, "\n") {
		at := strings.Index(line, "uses:")
		if at < 0 {
			continue
		}
		fields := strings.Fields(line[at+len("uses:"):])
		if len(fields) == 0 {
			t.Errorf("release.yml has a uses: line with no action: %q", line)
			continue
		}
		ref := fields[0]
		name, pin, ok := strings.Cut(ref, "@")
		if !ok || len(pin) != 40 {
			t.Errorf("release.yml uses %q, want a 40-character commit SHA: a tag can move", ref)
			continue
		}
		for _, r := range pin {
			if !strings.ContainsRune("0123456789abcdef", r) {
				t.Errorf("release.yml uses %q and %q is not a commit SHA", name, pin)
				break
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
