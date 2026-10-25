package review_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/review"
)

// lensBody is what the app hands the prompt for one lens: the file as it lives
// in internal/assets/review, frontmatter included. The domain strips it, so the
// prompt never reads like a file dump.
const lensBody = `---
name: risk
description: Trigger: revisar seguridad, permisos y datos sensibles.
---

Heredás las cinco reglas y el piso de severidad: WARNING y SUGGESTION nunca bloquean.

## Qué mirar
- una consulta armada concatenando entrada del request

## Qué NO reportar
- preferencias de estilo
`

// TestPromptWithoutLensesIsTheSamePromptAsBefore: the bare verb keeps one
// prompt. A project whose change asks for no lens gets exactly the review it got
// before lenses existed.
func TestPromptWithoutLensesIsTheSamePromptAsBefore(t *testing.T) {
	prompt := review.Prompt(promptInput())

	if strings.Contains(prompt, "## The lenses") {
		t.Error("the prompt of a candidate with no lens announces lenses")
	}
	if !strings.Contains(prompt, "## The six rules") {
		t.Error("the prompt lost the six rules")
	}
	if !strings.Contains(prompt, "## What to return") {
		t.Error("the prompt lost the answer shape")
	}
}

// TestPromptCarriesTheLensAndItsRules: a lens is the specialist's whole job, so
// its rules and its "what not to report" go in verbatim, without the frontmatter.
func TestPromptCarriesTheLensAndItsRules(t *testing.T) {
	in := promptInput()
	in.Lenses = []review.Lens{{Name: review.Risk, Body: lensBody}}

	prompt := review.Prompt(in)

	for _, want := range []string{
		"## The lenses of this tree",
		review.Risk,
		"## Qué mirar",
		"una consulta armada concatenando entrada del request",
		"## Qué NO reportar",
		"preferencias de estilo",
		"WARNING y SUGGESTION nunca bloquean",
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("the prompt does not contain %q\n%s", want, prompt)
		}
	}
	if strings.Contains(prompt, "name: risk") || strings.Contains(prompt, "description: Trigger:") {
		t.Errorf("the prompt carries the lens frontmatter, which is metadata for the registry, not for the model:\n%s", prompt)
	}
}

// TestPromptNamesTheLensesInTheHeader: a prompt read a month later says which
// specialists produced it, in the header, without reading the body.
func TestPromptNamesTheLensesInTheHeader(t *testing.T) {
	in := promptInput()
	in.Lenses = []review.Lens{
		{Name: review.Readability, Body: lensBody},
		{Name: review.Reliability, Body: lensBody},
	}

	header := strings.SplitN(review.Prompt(in), "\n", 2)[0]
	for _, want := range []string{review.Readability, review.Reliability} {
		if !strings.Contains(header, want) {
			t.Errorf("header %q does not name the lens %q", header, want)
		}
	}
}

// TestPromptKeepsTheSixRulesOnce: the lenses inherit the six rules from one
// place. Printing them per lens would let them drift.
func TestPromptKeepsTheSixRulesOnce(t *testing.T) {
	in := promptInput()
	in.Lenses = []review.Lens{
		{Name: review.Risk, Body: lensBody},
		{Name: review.Readability, Body: lensBody},
	}

	prompt := review.Prompt(in)
	if count := strings.Count(prompt, "## The six rules"); count != 1 {
		t.Errorf("the six rules appear %d times, want once", count)
	}
}

// TestPromptForbidsTheReviewerFromActing: the reviewer reads and answers. A
// command that also runs commands, edits files or reads other files puts its
// work in the output, and the output is the verdict.
func TestPromptForbidsTheReviewerFromActing(t *testing.T) {
	prompt := review.Prompt(promptInput())

	for _, want := range []string{
		"not an operator",
		"Do not run commands",
		"do not edit or read other files",
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("the prompt does not carry %q:\n%s", want, prompt)
		}
	}
}

// TestPromptTellsTheReviewerToAnswerPerLens: with more than one lens the answer
// has to be one pass per lens, not a single blurred sweep.
func TestPromptTellsTheReviewerToAnswerPerLens(t *testing.T) {
	in := promptInput()
	in.Lenses = []review.Lens{{Name: review.Risk, Body: lensBody}, {Name: review.Resilience, Body: lensBody}}

	prompt := review.Prompt(in)
	if !strings.Contains(prompt, "one pass per lens") {
		t.Errorf("the prompt does not ask for one pass per lens:\n%s", prompt)
	}
}

// TestVerdictRecordsTheLensesItCovered: the gate asks for a verdict per required
// lens, so the verdict has to say which lenses it answered.
func TestVerdictRecordsTheLensesItCovered(t *testing.T) {
	sealed, err := review.Seal(review.Verdict{
		Reviewer: "claude",
		Lenses:   []string{review.Reliability},
	}, tree, "HEAD", now)
	if err != nil {
		t.Fatalf("Seal = %v", err)
	}

	rendered := string(review.Render(sealed))
	if !strings.Contains(rendered, `"lenses"`) || !strings.Contains(rendered, review.Reliability) {
		t.Fatalf("the stored verdict does not record its lenses:\n%s", rendered)
	}

	back, err := review.Parse([]byte(rendered))
	if err != nil {
		t.Fatalf("Parse = %v", err)
	}
	if len(back.Lenses) != 1 || back.Lenses[0] != review.Reliability {
		t.Fatalf("round trip lenses = %v", back.Lenses)
	}
}

// TestVerdictRefusesAnInventedLens: the closed vocabulary holds for lenses too.
// A typo in an attached verdict must not look like coverage.
func TestVerdictRefusesAnInventedLens(t *testing.T) {
	verdict := review.Verdict{Reviewer: "humano", Lenses: []string{"seguridad"}}
	if err := verdict.Validate(); err == nil {
		t.Fatal("Validate accepted a lens that does not exist")
	}
}

// TestCoversReportsWhichLensesAreMissing: the gate and --show ask the same
// question, so it is answered in one place.
func TestCoversReportsWhichLensesAreMissing(t *testing.T) {
	verdict := review.Verdict{Lenses: []string{review.Reliability}}

	cases := map[string]struct {
		required []string
		want     []string
	}{
		"covered":             {[]string{review.Reliability}, nil},
		"one missing":         {[]string{review.Reliability, review.Risk}, []string{review.Risk}},
		"both missing":        {[]string{review.Risk, review.Resilience}, []string{review.Risk, review.Resilience}},
		"nothing is required": {nil, nil},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := verdict.Missing(tc.required)
			if strings.Join(got, ",") != strings.Join(tc.want, ",") {
				t.Errorf("Missing(%v) = %v, want %v", tc.required, got, tc.want)
			}
		})
	}
}

// TestSniffTellsTheTwoStoredDocumentsApart: a verdict and a refutation are both
// JSON a human can write, and --attach stores either. The key is what says which.
func TestSniffTellsTheTwoStoredDocumentsApart(t *testing.T) {
	verdictJSON := `{"reviewer":"claude","findings":[]}`
	refutationJSON := `{"refutations":[{"id":"F1","verdict":"stands","why":"no alcanza"}]}`

	if got := review.Sniff([]byte(verdictJSON)); got != review.KindVerdict {
		t.Errorf("Sniff(verdict) = %q, want %q", got, review.KindVerdict)
	}
	if got := review.Sniff([]byte(refutationJSON)); got != review.KindRefutation {
		t.Errorf("Sniff(refutation) = %q, want %q", got, review.KindRefutation)
	}
	for _, text := range []string{"", "No findings.", "basura", `{"tree":"` + tree + `"}`} {
		if got := review.Sniff([]byte(text)); got != review.KindUnknown {
			t.Errorf("Sniff(%q) = %q, want %q", text, got, review.KindUnknown)
		}
	}
}
