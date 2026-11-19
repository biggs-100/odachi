package skills_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestContainerCraftTakesTheRulesThatCostANight: the door for an image, a run
// command or a dead container exists, and its rules are the ones whose absence is
// an incident rather than a preference. Pinned because a page that lost them would
// still read like advice (0333).
func TestContainerCraftTakesTheRulesThatCostANight(t *testing.T) {
	skill := mustSkill(t, "container-craft")
	body := string(skill.Main())
	flat := strings.Join(strings.Fields(body), " ")

	assertEveryReferenceIsRouted(t, skill, body)

	for _, want := range []string{
		"a digest is the bytes",
		"docker history",
		"COPY --from",
		"PID 1",
		"--memory-swap",
		"0.0.0.0",
		"down -v",
		"fourteen caps",
		"does not rotate",
		"promote the digest",
	} {
		if !strings.Contains(flat, want) {
			t.Errorf("container-craft does not say %q", want)
		}
	}
}

// TestContainerCraftCarriesItsDepth: every reference is reachable from a gate and
// listed, and each page still says the facts that make it worth loading -the ones a
// reader cannot guess and would otherwise get from the other engine's defaults
// (0333).
func TestContainerCraftCarriesItsDepth(t *testing.T) {
	want := map[string][]string{
		"references/images-and-build.md":   {"multi-stage", "COPY --from", ".dockerignore", "--mount=type=secret", "pipefail", "scratch", "musl", "ca-certificates"},
		"references/running-containers.md": {"on-failure", "HEALTHCHECK", "STOPSIGNAL", "--memory-swap", "0.0.0.0", "down -v", "json-file"},
		"references/docker-and-podman.md":  {"Quadlet", "podman-compose", "rootless", "socket"},
		"references/supply-chain.md":       {"digest", "cosign", "Notation", "SLSA", "CycloneDX", "SPDX", "admission"},
		"references/hardening.md":          {"CAP_NET_RAW", "--privileged", "no-new-privileges", "CIS", "NIST"},
		"references/pipelines.md":          {"promote", "TARGETARCH", "cache", "ImagePullBackOff"},
		"references/the-failures.md":       {"What catches it earlier", "json-file", "SIGKILL", "--memory-swap", "DOCKER-USER"},
	}
	for path, phrases := range want {
		content := containerPage(t, path)
		for _, phrase := range phrases {
			if !strings.Contains(strings.ToLower(content), strings.ToLower(phrase)) {
				t.Errorf("%s does not say %q", path, phrase)
			}
		}
	}
}

// TestContainerCraftCountsWhatItShips: the map says how many references the skill
// carries, and a count that disagrees with the folder is the line a maintainer
// trusts while it lies (0333).
func TestContainerCraftCountsWhatItShips(t *testing.T) {
	skill := mustSkill(t, "container-craft")

	references := 0
	for _, file := range skill.Files {
		if strings.HasPrefix(file.Path, "references/") {
			references++
		}
	}
	if references != 7 {
		t.Fatalf("container-craft carries %d references, and the map says seven", references)
	}

	data, err := os.ReadFile(filepath.Join("..", "docs", "impact-map.md"))
	if err != nil {
		t.Fatalf("read docs/impact-map.md: %v", err)
	}
	if !strings.Contains(string(data), "`container-craft` ships seven references") {
		t.Error("the impact map does not say container-craft ships seven references")
	}
}

// TestTheOtherDoorsKeepTheirOwnQuestions: `data` owns what a value is, `db-craft`
// owns how it is stored, and this door owns how it reaches a host. A trigger that
// claims another door leaves the agent with two answers for one situation (0333).
func TestTheOtherDoorsKeepTheirOwnQuestions(t *testing.T) {
	flat := strings.Join(strings.Fields(string(mustSkill(t, "container-craft").Main())), " ")

	for _, want := range []string{"`data`", "`db-craft`", "how it reaches a host"} {
		if !strings.Contains(flat, want) {
			t.Errorf("container-craft does not hand %q to its owner", want)
		}
	}
}

// TestEveryPageKeepsItsCitations: a rule about an outside system without a
// document and a date is already false, and a citation is exactly the property a
// list of nouns cannot notice losing -0333's first review found a page whose
// Podman defaults had lost theirs while every phrase test stayed green (0333).
func TestEveryPageKeepsItsCitations(t *testing.T) {
	cited := map[string][]string{
		"references/images-and-build.md":   {"Docker best practices, 2026-11", "Docker multi-stage docs"},
		"references/running-containers.md": {"Docker docs, 2026-11"},
		"references/docker-and-podman.md":  {"podman-run(1)", "podman-systemd.unit(5)", "2026-11"},
		"references/supply-chain.md":       {"trivy.dev", "docs.sigstore.dev", "CycloneDX 1.7"},
		"references/hardening.md":          {"moby v27.3.1", "SP 800-190", "v1.9.0"},
		"references/pipelines.md":          {"Docker multi-platform docs, 2026-11"},
		"references/the-failures.md":       {"Docker logging, 2026-11", "Distribution GC, 2026-11", "musl docs, 2026-11"},
	}
	for path, marks := range cited {
		content := containerPage(t, path)
		for _, mark := range marks {
			if !strings.Contains(content, mark) {
				t.Errorf("%s lost its citation %q: a rule about an outside system with no document and no date is already false", path, mark)
			}
		}
	}
}

// containerPage is one reference of container-craft, flattened to a single line so a
// phrase can be searched across the wrapping.
func containerPage(t *testing.T, path string) string {
	t.Helper()
	for _, file := range mustSkill(t, "container-craft").Files {
		if file.Path == path {
			return strings.Join(strings.Fields(string(file.Content)), " ")
		}
	}
	t.Fatalf("container-craft does not carry %s", path)
	return ""
}
