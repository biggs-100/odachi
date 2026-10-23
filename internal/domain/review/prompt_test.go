package review_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/review"
)

// promptInput is a candidate with everything the review needs to be possible.
func promptInput() review.PromptInput {
	return review.PromptInput{
		Brief: "# turnos\n\n## Qué tiene que existir\n- reservar turnos\n",
		Tree:  tree,
		Base:  "HEAD",
		Paths: []string{"internal/app/check/check.go", "internal/domain/review/review.go"},
		Diff:  "diff --git a/a.go b/a.go\n--- a/a.go\n+++ b/a.go\n@@ -1 +1 @@\n-viejo\n+nuevo\n",
	}
}

// TestPromptCarriesTheBrief: a review without the vision it is judged against
// is a style opinion. The brief goes in verbatim.
func TestPromptCarriesTheBrief(t *testing.T) {
	prompt := review.Prompt(promptInput())

	if !strings.Contains(prompt, "## Qué tiene que existir") {
		t.Error("the prompt does not carry the brief")
	}
	if !strings.Contains(prompt, "- reservar turnos") {
		t.Error("the prompt does not carry the brief's list")
	}
}

// TestPromptCarriesTheCandidate: which tree, against which base, which paths and
// the diff itself.
func TestPromptCarriesTheCandidate(t *testing.T) {
	prompt := review.Prompt(promptInput())

	for _, want := range []string{
		tree,
		"HEAD",
		"internal/app/check/check.go",
		"internal/domain/review/review.go",
		"-viejo",
		"+nuevo",
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("the prompt does not contain %q", want)
		}
	}
}

// TestPromptCarriesTheFiveRules is the unit's reason to exist: an LLM review
// that reports style is noise, and noise is what makes people ignore reviews.
// Each rule is pinned by the clause that carries its meaning.
func TestPromptCarriesTheFiveRules(t *testing.T) {
	prompt := review.Prompt(promptInput())

	rules := []struct {
		name string
		mark string
	}{
		{"1 · precision: when in doubt, stay quiet", "When in doubt, say nothing"},
		{"1 · style banned", "Style and preference are banned"},
		{"2 · severity vocabulary", "BLOCKER, CRITICAL, WARNING, SUGGESTION"},
		{"3 · canonical empty result", "No findings."},
		{"4 · causality", "introduced it, activated it or worsened it"},
		{"4 · what was already there is follow-up", `"causal": false`},
		{"5 · one pass only", "One pass per lens"},
	}

	for _, rule := range rules {
		if !strings.Contains(prompt, rule.mark) {
			t.Errorf("the prompt is missing rule %s (%q)", rule.name, rule.mark)
		}
	}
}

// TestPromptCarriesTheSchema: the answer has a shape, and the prompt says it.
// The keys are the ones the domain parses, or the answer cannot be stored.
func TestPromptCarriesTheSchema(t *testing.T) {
	prompt := review.Prompt(promptInput())

	for _, key := range []string{`"tree"`, `"base"`, `"reviewer"`, `"at"`, `"findings"`, `"verdict"`, `"path"`, `"line"`, `"severity"`, `"causal"`, `"why"`} {
		if !strings.Contains(prompt, key) {
			t.Errorf("the schema in the prompt does not carry %s", key)
		}
	}
	for _, value := range []string{review.Approved, review.ChangesRequested} {
		if !strings.Contains(prompt, value) {
			t.Errorf("the prompt does not name the verdict %q", value)
		}
	}
}

// TestPromptAsksForOnlyTheJSON: the answer is machine-read, so prose around it
// is a parse error waiting to happen.
func TestPromptAsksForOnlyTheJSON(t *testing.T) {
	prompt := review.Prompt(promptInput())
	if !strings.Contains(prompt, "without text around it") {
		t.Error("the prompt does not ask for the JSON alone")
	}
}

// TestPromptSaysTheVerdictIsDerived: the reviewer supplies findings, not the
// verdict. Storing an administered verdict is exactly what this unit removed.
func TestPromptSaysTheVerdictIsDerived(t *testing.T) {
	prompt := review.Prompt(promptInput())
	if !strings.Contains(prompt, "is derived") {
		t.Error("the prompt does not say the verdict is derived from the findings")
	}
}
