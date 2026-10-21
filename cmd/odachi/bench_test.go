package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// sandboxes lists the bench's temporary directories on this machine. It is how a
// test proves the instrument leaves the machine as it found it.
func sandboxes(t *testing.T) []string {
	t.Helper()
	entries, err := filepath.Glob(filepath.Join(os.TempDir(), "odachi-bench-*"))
	if err != nil {
		t.Fatalf("list the temp directory: %v", err)
	}
	return entries
}

// TestBenchRejectsAnUnknownJourney: a filter that matches nothing would report
// "todo bien" about a corpus that never ran, so it names the corpus instead.
func TestBenchRejectsAnUnknownJourney(t *testing.T) {
	before := sandboxes(t)

	code, _, stderr := runDir(t, t.TempDir(), []string{"bench", "--journeys", "inventado"}, "")
	if code != 1 {
		t.Fatalf("an unknown journey must exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "inventado") {
		t.Fatalf("the message must name the typo: %q", stderr)
	}
	if !strings.Contains(stderr, "repo-nuevo-start") {
		t.Fatalf("the message must name the corpus: %q", stderr)
	}
	if after := sandboxes(t); len(after) != len(before) {
		t.Fatalf("nothing may run when the filter is wrong: %d -> %d sandboxes", len(before), len(after))
	}
}

// TestBenchRemovesItsSandboxes: even a journey that cannot run leaves nothing
// behind. The assertion is on the real temp directory, not on a double.
func TestBenchRemovesItsSandboxes(t *testing.T) {
	before := sandboxes(t)

	missing := filepath.Join(t.TempDir(), "no-existe")
	code, _, stderr := runDir(t, t.TempDir(), []string{"bench", "--binary", missing, "--journeys", "repo-nuevo-start"}, "")
	if code != 1 {
		t.Fatalf("a binary that cannot run must fail the bench, got %d: %s", code, stderr)
	}
	if !strings.Contains(stderr, "no pude correr el paso 1") {
		t.Fatalf("the report must say the step could not run: %q", stderr)
	}

	after := sandboxes(t)
	if len(after) != len(before) {
		t.Fatalf("the bench left %d sandboxes behind: %v", len(after)-len(before), after)
	}
}

// TestBenchJSONSaysTheSameThing: --json carries the counts and the dead ends, so
// a machine reads the same report a human does.
func TestBenchJSONSaysTheSameThing(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "no-existe")
	code, stdout, stderr := runDir(t, t.TempDir(), []string{"bench", "--binary", missing, "--journeys", "repo-nuevo-start", "--json"}, "")
	if stderr != "" {
		t.Fatalf("--json writes nothing to stderr: %q", stderr)
	}
	if code != 1 {
		t.Fatalf("exit %d, want 1", code)
	}

	var payload struct {
		Code      int    `json:"code"`
		Message   string `json:"message"`
		Steps     int    `json:"steps"`
		Blocks    int    `json:"blocks"`
		InBand    int    `json:"in_band"`
		OutOfBand int    `json:"out_of_band"`
		Templated int    `json:"templated"`
		DeadEnd   int    `json:"dead_end"`
		Journeys  []struct {
			Name      string   `json:"name"`
			Steps     int      `json:"steps"`
			Blocks    int      `json:"blocks"`
			Surprises []string `json:"surprises"`
		} `json:"journeys"`
		DeadEnds []struct {
			Journey string `json:"journey"`
			Command string `json:"command"`
			Exit    int    `json:"exit"`
			Message string `json:"message"`
		} `json:"dead_ends"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("--json must be valid JSON: %v\n%s", err, stdout)
	}

	if payload.Code != 1 {
		t.Fatalf("the JSON must carry the exit code, got %d", payload.Code)
	}
	if len(payload.Journeys) != 1 || payload.Journeys[0].Name != "repo-nuevo-start" {
		t.Fatalf("the JSON must name the journey that ran: %+v", payload.Journeys)
	}
	if len(payload.Journeys[0].Surprises) == 0 {
		t.Fatalf("the JSON must say the step could not run: %+v", payload.Journeys[0])
	}
	if payload.DeadEnds == nil {
		t.Fatalf("dead_ends must be a list, empty or not: %s", stdout)
	}
}

// TestBenchUnknownFlag: the flag surface is small and a typo says so.
func TestBenchUnknownFlag(t *testing.T) {
	code, _, stderr := runDir(t, t.TempDir(), []string{"bench", "--inventado"}, "")
	if code != 1 || !strings.Contains(stderr, "flag inválido") {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
}

// TestBenchIsDispatched: the verb exists and does not fall through to the unknown
// command path.
func TestBenchIsDispatched(t *testing.T) {
	code, _, stderr := runDir(t, t.TempDir(), []string{"bench", "--journeys", "inventado"}, "")
	if strings.Contains(stderr, "comando desconocido") {
		t.Fatalf("bench must be dispatched: %q", stderr)
	}
	if code != 1 {
		t.Fatalf("exit %d, want 1", code)
	}
}
