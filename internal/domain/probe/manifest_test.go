package probe_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/probe"
)

func valid() string {
	return `{
  "name": "lenses",
  "defects": [
    {"id": "a", "path": "server.go", "line": 3, "lens": "risk", "severity": "CRITICAL", "why": "the id is concatenated"}
  ]
}`
}

// TestParseDefaultsTheWindow: a manifest that names no window measures the same
// three lines around a defect the review floor calls context, so every probe
// compares like with like.
func TestParseDefaultsTheWindow(t *testing.T) {
	t.Parallel()
	manifest, err := probe.Parse([]byte(valid()))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if manifest.Window != probe.DefaultWindow {
		t.Errorf("Window = %d, want %d", manifest.Window, probe.DefaultWindow)
	}
	if manifest.Name != "lenses" || len(manifest.Defects) != 1 {
		t.Errorf("manifest = %+v, want one defect named lenses", manifest)
	}
}

// TestParseRefusesAManifestThatCannotBeScored: every refusal names the defect,
// because the person who wrote the manifest is the one who fixes it. A key that
// parses but cannot be scored produces a number nobody can trust.
func TestParseRefusesAManifestThatCannotBeScored(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		body   string
		reason string
	}{
		{"no defects", `{"defects": []}`, "names no defect"},
		{"no id", `{"defects": [{"path": "a.go", "line": 1, "lens": "risk", "severity": "WARNING", "why": "x"}]}`, "has no id"},
		{"repeated id", `{"defects": [
			{"id": "a", "path": "a.go", "line": 1, "lens": "risk", "severity": "WARNING", "why": "x"},
			{"id": "a", "path": "b.go", "line": 2, "lens": "risk", "severity": "WARNING", "why": "y"}]}`, "repeats an id"},
		{"no path", `{"defects": [{"id": "a", "line": 1, "lens": "risk", "severity": "WARNING", "why": "x"}]}`, "names no path"},
		{"no line", `{"defects": [{"id": "a", "path": "a.go", "line": 0, "lens": "risk", "severity": "WARNING", "why": "x"}]}`, "has line 0"},
		{"unknown lens", `{"defects": [{"id": "a", "path": "a.go", "line": 1, "lens": "tests", "severity": "WARNING", "why": "x"}]}`, "the lenses are risk, readability, reliability, resilience"},
		{"unknown severity", `{"defects": [{"id": "a", "path": "a.go", "line": 1, "lens": "risk", "severity": "MAJOR", "why": "x"}]}`, "the vocabulary is BLOCKER, CRITICAL, WARNING, SUGGESTION"},
		{"no why", `{"defects": [{"id": "a", "path": "a.go", "line": 1, "lens": "risk", "severity": "WARNING"}]}`, "does not say what it is"},
		{"negative window", `{"window": -1, "defects": [{"id": "a", "path": "a.go", "line": 1, "lens": "risk", "severity": "WARNING", "why": "x"}]}`, "window is -1"},
		{"misspelled key", `{"defect": []}`, "not readable"},
		{"not json", `defects: []`, "not readable"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			_, err := probe.Parse([]byte(c.body))
			if err == nil {
				t.Fatalf("Parse accepted a manifest that cannot be scored: %s", c.body)
			}
			if !strings.Contains(err.Error(), c.reason) {
				t.Errorf("error = %q, want it to name %q", err, c.reason)
			}
		})
	}
}
