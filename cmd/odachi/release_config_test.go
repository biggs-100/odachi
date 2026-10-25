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
		"goreleaser/goreleaser-action@",
		// The recovery path: the attestation is written after the release is
		// public, so a failure there needs a run that attests without publishing.
		"workflow_dispatch:",
		"gh release download",
	} {
		if !strings.Contains(workflow, want) {
			t.Errorf("release.yml lost %q: the release would fail at the worst moment", want)
		}
	}

	// The same strings live in both jobs, so a file-wide search lets the
	// recovery job's copies stand in for the publishing ones: delete the
	// attestation that gives a tag its provenance and the test would stay
	// green. Each job is checked on its own.
	publish, retry, ok := strings.Cut(workflow, "attest-retry:")
	if !ok {
		t.Fatalf("release.yml has no attest-retry job: a failed attestation would leave a release with no provenance")
	}
	for _, want := range []string{
		"attestations: write",
		"actions/attest@",
		"subject-checksums: ./dist/checksums.txt",
	} {
		if !strings.Contains(publish, want) {
			t.Errorf("the publishing job lost %q: a tag would publish without provenance", want)
		}
		if !strings.Contains(retry, want) {
			t.Errorf("the recovery job lost %q: a failed attestation could not be repaired", want)
		}
	}
	// A dispatch must not reach the job that publishes, and a tag push must not
	// reach the one that only attests.
	if !strings.Contains(publish, "if: github.event_name == 'push'") {
		t.Error("the publishing job is not push-only: a manual dispatch would publish again")
	}
	if !strings.Contains(retry, "if: github.event_name == 'workflow_dispatch'") {
		t.Error("the recovery job is not dispatch-only: a tag would run it beside the release")
	}

	// The published checksums are an input in the recovery job, so they have to
	// be verified before they are attested: attesting an unauthenticated file
	// would stamp provenance on bytes this workflow never built.
	verify := strings.Index(retry, "cosign verify-blob")
	attest := strings.Index(retry, "actions/attest@")
	if verify < 0 {
		t.Fatal("the recovery job does not verify the checksums bundle: it would attest whatever the release holds")
	}
	if !strings.Contains(retry, "checksums.txt.sigstore.json") {
		t.Error("the recovery job verifies nothing: the bundle it must check is not named")
	}
	if attest < 0 || verify > attest {
		t.Error("the recovery job attests before it verifies: the signature has to come first")
	}
	// Verifying without binding the identity is verifying nothing: any Fulcio
	// identity can produce a bundle, and the actor this path worries about can
	// run a workflow of their own. The identity must name this repository's
	// workflow at this tag, and the issuer must be GitHub's token service.
	for _, want := range []string{
		"--certificate-identity",
		// The whole binding in one string: this repository, this workflow, this
		// tag. A literal repo or a literal tag would otherwise pass a suffix.
		"https://github.com/${GH_REPO}/.github/workflows/release.yml@refs/tags/${TAG}",
		"--certificate-oidc-issuer",
		"https://token.actions.githubusercontent.com",
	} {
		if !strings.Contains(retry, want) {
			t.Errorf("the recovery job verifies with %q missing: it would accept a bundle from any identity", want)
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
