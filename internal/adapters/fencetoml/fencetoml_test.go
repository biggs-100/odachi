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
		loose []string
	}{
		{name: "no rules and nothing loose"},
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
			name:  "a template nobody gave paths",
			rules: []fence.Rule{{Not: "no hay migraciones", Paths: []string{"internal/db/migrations/**"}}},
			loose: []string{"no se toca el esquema: la migración va en otra feature"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := fencetoml.Decode([]byte(fencetoml.Render(c.rules, c.loose)))
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
	got := fencetoml.Render(nil, nil)
	if !strings.Contains(got, "Empty on purpose") {
		t.Errorf("the empty fence does not explain itself:\n%s", got)
	}
	if _, err := fencetoml.Decode([]byte(got)); err != nil {
		t.Errorf("the teaching text does not decode: %v", err)
	}
}
