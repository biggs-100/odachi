package fencetoml_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/fencetoml"
	"github.com/biggs-100/odachi/internal/domain/fence"
)

// TestRoundTripKeepsEveryRule is the property the merge will rest on: the file a
// run writes is the file the gate reads, rule for rule.
func TestRoundTripKeepsEveryRule(t *testing.T) {
	cases := []struct {
		name  string
		rules []fence.Rule
	}{
		{name: "no rules at all"},
		{
			name:  "one rule",
			rules: []fence.Rule{{Not: "no se toca el fixture", Paths: []string{"internal/probe/testdata/**"}}},
		},
		{
			name: "paths and dependencies",
			rules: []fence.Rule{{
				Not:   "pagos",
				Paths: []string{"payment*", "**/payments/**"},
				Deps:  []string{"stripe", "mercadopago"},
			}},
		},
		{
			name:  "a rule a human wrote",
			rules: []fence.Rule{{Not: "no hay migraciones", Paths: []string{"internal/db/migrations/**"}}},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := fencetoml.Decode([]byte(fencetoml.Render(c.rules)))
			if err != nil {
				t.Fatalf("Decode(Render(rules)) error: %v", err)
			}
			if len(got) != len(c.rules) {
				t.Fatalf("the round trip returned %d rules, want %d: %+v", len(got), len(c.rules), got)
			}
			for i, want := range c.rules {
				if got[i].Not != want.Not {
					t.Errorf("rule %d reads %q, want %q", i, got[i].Not, want.Not)
				}
				if !slices.Equal(got[i].Paths, want.Paths) {
					t.Errorf("rule %d reads paths %v, want %v", i, got[i].Paths, want.Paths)
				}
				if !slices.Equal(got[i].Deps, want.Deps) {
					t.Errorf("rule %d reads deps %v, want %v", i, got[i].Deps, want.Deps)
				}
			}
		})
	}
}

// TestDecodeRejectsAFileItCannotRead: the gate must never run with an empty fence
// because the file was broken and nobody said so.
func TestDecodeRejectsAFileItCannotRead(t *testing.T) {
	for _, bad := range []string{"[[rule]]\nnot =\n", "[[rule]]\nnot = \"x\"\npaths = 3\n"} {
		if _, err := fencetoml.Decode([]byte(bad)); err == nil {
			t.Errorf("Decode accepted a file that is not a fence: %q", bad)
		}
	}
}

// TestRenderTeachesTheFenceWithNothingToWrite: an empty file teaches nothing, so
// the run that chose no non-goals writes what the fence is for.
func TestRenderTeachesTheFenceWithNothingToWrite(t *testing.T) {
	got := fencetoml.Render(nil)
	if !strings.Contains(got, "Empty on purpose") {
		t.Errorf("the empty fence does not explain itself:\n%s", got)
	}
	if _, err := fencetoml.Decode([]byte(got)); err != nil {
		t.Errorf("the teaching text does not decode: %v", err)
	}
}

// TestRenderWritesRulesAndNoComment: the fence holds the rules somebody gave paths
// for and nothing else. A phrase that names no path is not the renderer's business:
// it lives in the brief, and writing it here as a template made the file ask the
// same question on every run -ninety-one times in this repository, before the run
// said anything useful- (0299).
func TestRenderWritesRulesAndNoComment(t *testing.T) {
	got := fencetoml.Render([]fence.Rule{
		{Not: "no se toca el fixture", Paths: []string{"internal/probe/testdata/**"}},
		{Not: "no hay migraciones", Paths: []string{"internal/db/migrations/**"}},
	})

	for _, line := range strings.Split(got, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			t.Errorf("the fence carries a comment beside its rules: %q\n%s", line, got)
		}
	}
	if strings.Count(got, "[[rule]]") != 2 {
		t.Errorf("the fence does not carry its two rules:\n%s", got)
	}
}
