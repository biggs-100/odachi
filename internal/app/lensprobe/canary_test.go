package lensprobe_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/lensprobe"
)

// TestTheCanaryTellsAccessFromIsolation: a reviewer that quotes the line inside
// the frozen tree and refuses the one outside is using the candidate; one that
// cannot quote the inside line measured nothing, and one that quotes the outside
// line read more than it was given. Both legs matter.
func TestTheCanaryTellsAccessFromIsolation(t *testing.T) {
	c := lensprobe.Canary{Inside: "ODACHI-CANARY-INSIDE", Outside: "ODACHI-CANARY-OUTSIDE"}

	cases := []struct {
		name   string
		answer string
		want   string
	}{
		{"attached and scoped", "inside: ODACHI-CANARY-INSIDE; outside: I cannot read that path", "canary: attached and scoped"},
		{"not attached", "outside: I cannot read that path", "canary: not attached"},
		{"too wide", "inside: ODACHI-CANARY-INSIDE; outside: ODACHI-CANARY-OUTSIDE", "canary: too wide"},
		{"no answer", "", "canary: no answer"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := c.Read(tc.answer); got != tc.want {
				t.Errorf("Read(%q) = %q, want %q", tc.answer, got, tc.want)
			}
		})
	}
}

// TestTheCanaryPromptNamesBothLegs: the question has to ask for the inside line
// and the outside one, or a yes cannot be told from a refusal.
func TestTheCanaryPromptNamesBothLegs(t *testing.T) {
	c := lensprobe.Canary{Inside: "INSIDE-MARK", Outside: "OUTSIDE-MARK"}
	prompt := c.Prompt()
	for _, want := range []string{c.Inside, c.Outside, "verbatim"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("the canary prompt does not name %q:\n%s", want, prompt)
		}
	}
}
