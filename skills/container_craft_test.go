package skills_test

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
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
		"references/supply-chain.md":       {"Source: the Kubernetes images docs, 2026-11", "Source: Trivy's OS coverage page, 2026-11", "Source: the CNCF Distribution GC"},
		"references/hardening.md":          {"moby v27.3.1", "SP 800-190", "v1.9.0", "Source: the OWASP Docker Security Cheat Sheet", "Source: Docker rootless docs", "Source: the Kubernetes Secrets docs"},
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

// TestContainerCraftDeclaresItsTrigger: the trigger is the only thing that
// decides whether the skill loads, and the done-when asks for a test that pins it.
// A body whose situations nobody recognizes is a file nobody reads, however good
// the rules inside are (0333).
func TestContainerCraftDeclaresItsTrigger(t *testing.T) {
	body := string(mustSkill(t, "container-craft").Main())
	parts := strings.SplitN(body, "---", 3)
	if len(parts) < 3 {
		t.Fatalf("container-craft has no frontmatter to hold a trigger")
	}
	front := strings.Join(strings.Fields(parts[1]), " ")

	for _, want := range []string{"Trigger:", "a Dockerfile", "compose file", "docker run", "podman", "registry", "died at 3 a.m."} {
		if !strings.Contains(front, want) {
			t.Errorf("the trigger does not name %q, so a session in that situation never loads the skill:\n%s", want, front)
		}
	}
}

// TestEverySourceLineIsPinned: a rule about an outside system without a document
// and a date is already false, and the guard used to ask whether a page kept *a*
// citation -a page that lost one of eight stayed green, which is exactly the loss
// the property is about (0333). Every Source line a page carries is pinned in
// container_craft.sources, and a pin the page no longer carries fails too, so the
// baseline follows the pages down instead of drifting above them (0335).
func TestEverySourceLineIsPinned(t *testing.T) {
	pinned := readSources(t)
	year := regexp.MustCompile(`20\d\d`)
	pages := 0

	for _, file := range mustSkill(t, "container-craft").Files {
		if !strings.HasPrefix(file.Path, "references/") {
			continue
		}
		for _, line := range strings.Split(string(file.Content), "\n") {
			trimmed := strings.TrimSpace(line)
			if !strings.HasPrefix(trimmed, "Source:") {
				continue
			}
			if !year.MatchString(trimmed) {
				t.Errorf("%s has a Source line with no year: %q", file.Path, trimmed)
			}
			if !slices.Contains(pinned[file.Path], trimmed) {
				t.Errorf("%s carries a Source line the baseline does not pin: %q; add it to skills/container_craft.sources on purpose",
					file.Path, trimmed)
			}
		}
	}

	for _, page := range sortedPages(pinned) {
		pages++
		content := containerPage(t, page)
		for _, line := range pinned[page] {
			if !strings.Contains(content, line) {
				t.Errorf("%s lost its Source line: %q", page, line)
			}
		}
	}
	if pages == 0 {
		t.Fatal("the baseline pins no page: this test would pass without measuring anything")
	}
}

// readSources reads the pinned Source lines: one `page<TAB>line` per row, `#`
// comments ignored, the same shape internal/archtest/normative.baseline uses.
func readSources(t *testing.T) map[string][]string {
	t.Helper()

	data, err := os.ReadFile("container_craft.sources")
	if err != nil {
		t.Fatalf("read container_craft.sources: %v", err)
	}

	pinned := map[string][]string{}
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(strings.TrimSuffix(raw, "\r"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		page, source, ok := strings.Cut(line, "\t")
		if !ok {
			t.Fatalf("a pinned line is not page<TAB>line: %q", raw)
		}
		pinned[page] = append(pinned[page], strings.TrimSpace(source))
	}
	return pinned
}

// sortedPages names the pinned pages in a stable order, so a failure reads the
// same on every run.
func sortedPages(pinned map[string][]string) []string {
	pages := make([]string, 0, len(pinned))
	for page := range pinned {
		pages = append(pages, page)
	}
	sort.Strings(pages)
	return pages
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
