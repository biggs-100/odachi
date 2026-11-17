package review_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/review"
)

// TestThePromptNamesTheFrozenCandidate: a reviewer allowed to read gets one
// exact directory - the tree the verdict names - and no other source. Without
// it the printed form must not claim the reviewer can read files.
func TestThePromptNamesTheFrozenCandidate(t *testing.T) {
	in := promptInput()
	in.FrozenTree = "C:/temp/odachi-candidate-abc"
	prompt := review.Prompt(in)

	for _, want := range []string{
		"## The frozen candidate",
		in.FrozenTree,
		"Read only this directory",
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("the prompt does not carry %q:\n%s", want, prompt)
		}
	}
}

// TestThePromptWithoutAFrozenTreeDoesNotInventOne: the bare form prepares no
// snapshot, so its prompt stays the question it always was.
func TestThePromptWithoutAFrozenTreeDoesNotInventOne(t *testing.T) {
	prompt := review.Prompt(promptInput())
	if strings.Contains(prompt, "## The frozen candidate") {
		t.Error("the prompt names a frozen tree that does not exist")
	}
}

// TestThePromptForbidsCommandsButAllowsFrozenReads: no shell and no edits, but
// the clause that banned all reads is gone when there is a frozen tree to read.
func TestThePromptForbidsCommandsButAllowsFrozenReads(t *testing.T) {
	in := promptInput()
	in.FrozenTree = "C:/temp/odachi-candidate-abc"
	prompt := review.Prompt(in)

	for _, want := range []string{"Do not run commands", "edit anything"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("the prompt lost %q:\n%s", want, prompt)
		}
	}
	if strings.Contains(prompt, "do not edit or read other files") {
		t.Error("the prompt still forbids reading the frozen candidate")
	}
}
