package change_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/change"
)

// A commit that names the risk it carries is the one declaration the gate can
// contrast with what the paths say; the name is the first word, so prose after it
// belongs to the human and not to the parser.
func TestARiskLineNamesTheRiskTheCommitCarries(t *testing.T) {
	risk, err := change.ParseRisk([]string{
		"move the session store\n\nThe old table is rewritten in place.\n\nRisk: data: a migration that rewrites the table\n",
	})
	if err != nil {
		t.Fatalf("ParseRisk: %v", err)
	}
	if !risk.Declared {
		t.Fatalf("the Risk: line was not read")
	}
	if !slices.Equal(risk.Named, []string{"data"}) {
		t.Fatalf("Named = %q, want [data]", risk.Named)
	}
}

// `none` answers the question and names nothing: it is the claim of harmlessness
// the note exists to contrast, so it is a declaration and never a name.
func TestNoneIsADeclarationWithNoName(t *testing.T) {
	risk, err := change.ParseRisk([]string{"a document\n\nRisk: none\n"})
	if err != nil {
		t.Fatalf("ParseRisk: %v", err)
	}
	if !risk.Declared {
		t.Fatalf("Risk: none was not read as a declaration")
	}
	if len(risk.Named) != 0 {
		t.Fatalf("Named = %q, want none", risk.Named)
	}
}

// A reasoned none keeps the human's explanation without turning it into a named
// risk. The separator is punctuation; the trimmed prose after it is the reason.
func TestRiskNoneCapturesWrittenReason(t *testing.T) {
	cases := map[string]string{
		"hyphen":  "Risk: none - documentation only\n",
		"colon":   "Risk: none: documentation only\n",
		"em dash": "Risk: none — documentation only\n",
		"spacing": "Risk: none —   documentation only  \n",
	}

	for name, message := range cases {
		t.Run(name, func(t *testing.T) {
			risk, err := change.ParseRisk([]string{message})
			if err != nil {
				t.Fatalf("ParseRisk: %v", err)
			}
			if !risk.Declared || len(risk.Named) != 0 {
				t.Fatalf("reasoned none declared %+v, want a declaration with no name", risk)
			}
			if risk.NoneReason != "documentation only" {
				t.Errorf("NoneReason = %q, want %q", risk.NoneReason, "documentation only")
			}
			if risk.NoneWithoutReason {
				t.Errorf("NoneWithoutReason = true, want false for a written reason")
			}
		})
	}
}

// Bare none still answers the question, but it must remain distinguishable from
// a none that carries a written explanation.
func TestRiskNoneWithoutReasonIsBare(t *testing.T) {
	risk, err := change.ParseRisk([]string{"a document\n\nRisk: none\n"})
	if err != nil {
		t.Fatalf("ParseRisk: %v", err)
	}
	if !risk.Declared || len(risk.Named) != 0 {
		t.Fatalf("bare none declared %+v, want a declaration with no name", risk)
	}
	if risk.NoneReason != "" {
		t.Errorf("NoneReason = %q, want empty", risk.NoneReason)
	}
	if !risk.NoneWithoutReason {
		t.Errorf("NoneWithoutReason = false, want true")
	}
}

// Across a range, the first written explanation wins and any bare none keeps
// the range marked as missing a reason.
func TestRiskNoneAcrossCommitsKeepsFirstReasonAndAnyBareDeclaration(t *testing.T) {
	risk, err := change.ParseRisk([]string{
		"first\n\nRisk: none\n",
		"second\n\nRisk: none - documentation only\n",
		"third\n\nRisk: none: generated reference\n",
	})
	if err != nil {
		t.Fatalf("ParseRisk: %v", err)
	}
	if risk.NoneReason != "documentation only" {
		t.Errorf("NoneReason = %q, want the first written reason", risk.NoneReason)
	}
	if !risk.NoneWithoutReason {
		t.Errorf("NoneWithoutReason = false, want true because one commit was bare")
	}
}

// Named-risk parsing still takes only the first word, strips its punctuation,
// deduplicates names and ignores the prose after them.
func TestRiskNamedParsingRemainsUnchanged(t *testing.T) {
	risk, err := change.ParseRisk([]string{
		"first\n\nRisk: data: a migration that rewrites the table\n",
		"second\n\nRisk: delivery - installer behavior\n",
		"third\n\nRisk: Data, the same migration\n",
	})
	if err != nil {
		t.Fatalf("ParseRisk: %v", err)
	}
	if !slices.Equal(risk.Named, []string{"data", "delivery"}) {
		t.Errorf("Named = %q, want [data delivery]", risk.Named)
	}
	if risk.NoneReason != "" || risk.NoneWithoutReason {
		t.Errorf("named risks populated none fields: %+v", risk)
	}
}

// Case does not decide: a human who writes `Data` named the same risk as `data`,
// and two commits that name it are one risk, not two.
func TestTheSameRiskTwiceIsOneRisk(t *testing.T) {
	risk, err := change.ParseRisk([]string{
		"first\n\nRisk: Data\n",
		"second\n\nRisk: data\n",
	})
	if err != nil {
		t.Fatalf("ParseRisk: %v", err)
	}
	if !slices.Equal(risk.Named, []string{"data"}) {
		t.Fatalf("Named = %q, want [data] once", risk.Named)
	}
}

// Two different names are two risks the range carries, in the order the commits
// first said them: a later reader must not have to guess which came first.
func TestTwoNamesAreKeptInOrder(t *testing.T) {
	risk, err := change.ParseRisk([]string{
		"first\n\nRisk: security\n",
		"second\n\nRisk: delivery\n",
	})
	if err != nil {
		t.Fatalf("ParseRisk: %v", err)
	}
	if !slices.Equal(risk.Named, []string{"security", "delivery"}) {
		t.Fatalf("Named = %q, want [security delivery]", risk.Named)
	}
}

// A message with no line is the past and the habit of many projects: it answers
// nothing, and the parser must not invent a declaration for it.
func TestNoRiskLineIsNotADeclaration(t *testing.T) {
	risk, err := change.ParseRisk([]string{"the release body comes from the changelog\n"})
	if err != nil {
		t.Fatalf("ParseRisk: %v", err)
	}
	if risk.Declared || len(risk.Named) != 0 {
		t.Fatalf("a message with no Risk: line declared something: %+v", risk)
	}
}

// An empty line is not `none`: it answered nothing, and reading it as harmless is
// the silence this parser exists to refuse.
func TestAnEmptyRiskLineIsRefused(t *testing.T) {
	_, err := change.ParseRisk([]string{"something\n\nRisk:   \n"})
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("error = %v, want the empty line named", err)
	}
}
