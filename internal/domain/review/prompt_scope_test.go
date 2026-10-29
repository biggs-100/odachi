package review_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/review"
)

// TestThePromptSaysWhenAPersonRaisedTheScope: the lenses section claims the
// evidence chose the lenses, and that sentence is false the moment somebody added
// one with --lens. A reviewer that is told the truth knows the addition is
// deliberate; one that is told the evidence asked for it reviews a question nobody
// asked.
func TestThePromptSaysWhenAPersonRaisedTheScope(t *testing.T) {
	in := review.PromptInput{
		Brief:  "# a-brief\n",
		Tree:   "4a1c9f7d",
		Base:   "HEAD",
		Paths:  []string{"a.go"},
		Diff:   "diff --git a/a.go b/a.go\n",
		Lenses: []review.Lens{{Name: review.Reliability, Body: "---\nname: reliability\n---\n\nbody\n"}},
		Scope: &review.Scope{
			Required: []string{review.Reliability},
			Added:    []string{review.Risk},
			Asked:    "risk",
			Because:  "the change touches the gate",
		},
	}

	text := review.Prompt(in)

	if !strings.Contains(text, "risk was added with --lens risk") {
		t.Errorf("the prompt does not say the scope was raised:\n%s", text)
	}
	if !strings.Contains(text, "asked for reliability") {
		t.Errorf("the prompt does not say what the evidence asked for:\n%s", text)
	}
	if !strings.Contains(text, "the change touches the gate") {
		t.Errorf("the prompt does not carry the caller's words:\n%s", text)
	}
	if strings.Contains(text, "chosen by the diff's evidence, not by anyone's taste") {
		t.Errorf("the prompt still claims nobody chose the lenses:\n%s", text)
	}
}

// TestThePromptKeepsItsWordsWhenNobodyRaisedTheScope: a run with no --lens prints
// byte for byte the prompt it printed before scopes existed, so a normal review is
// not re-measured against the size limit for a sentence that does not apply.
func TestThePromptKeepsItsWordsWhenNobodyRaisedTheScope(t *testing.T) {
	in := review.PromptInput{
		Brief:  "# a-brief\n",
		Tree:   "4a1c9f7d",
		Base:   "HEAD",
		Paths:  []string{"a.go"},
		Diff:   "diff --git a/a.go b/a.go\n",
		Lenses: []review.Lens{{Name: review.Reliability, Body: "---\nname: reliability\n---\n\nbody\n"}},
	}

	text := review.Prompt(in)

	if !strings.Contains(text, "chosen by the diff's evidence, not by anyone's taste") {
		t.Errorf("a run that raised nothing lost the sentence it always printed:\n%s", text)
	}
}
