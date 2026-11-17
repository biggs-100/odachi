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

// TestTheFrozenCandidateCostsAPathNotATree: the prompt grows by one directory
// name, not by the bytes of the files the reviewer will read. A prompt that
// carried the tree would grow with every candidate and press the same wall.
func TestTheFrozenCandidateCostsAPathNotATree(t *testing.T) {
	bare := review.Prompt(promptInput())
	in := promptInput()
	in.FrozenTree = "C:/temp/odachi-candidate-with-a-deliberately-long-name-0123456789"
	frozen := review.Prompt(in)

	growth := len(frozen) - len(bare)
	if growth > 1024 {
		t.Errorf("the frozen candidate added %d bytes; it must name a path, not carry the tree", growth)
	}
	if strings.Contains(frozen, "package main") || strings.Contains(frozen, "func ") {
		t.Error("the prompt carries file contents instead of the directory to read")
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
