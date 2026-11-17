package review_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/review"
)

// TestTheReviewerGetsTheFrozenTree: exec freezes the tree the verdict names,
// gives that directory to the reviewer in the prompt, and discards it when the
// run ends. A prompt naming no snapshot leaves the reviewer blind; one naming a
// different tree approves bytes it never judged.
func TestTheReviewerGetsTheFrozenTree(t *testing.T) {
	p := newProject()
	p.git.archive = "C:/temp/odachi-candidate-abc"
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

	result := p.run(review.Options{Exec: "revisor"})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if got := p.git.archived; len(got) != 1 || got[0] != currentTree {
		t.Errorf("archived = %v, want the verdict's tree %q", got, currentTree)
	}
	if !strings.Contains(p.shell.input, p.git.archive) {
		t.Errorf("the reviewer's prompt does not name the frozen directory:\n%s", p.shell.input)
	}
	if got := p.git.discarded; len(got) != 1 || got[0] != p.git.archive {
		t.Errorf("discarded = %v, want %q", got, p.git.archive)
	}
}

// TestAFrozenCandidateIsRequired: a repository that cannot export the tree must
// not start a reviewer that would then read the live worktree instead.
func TestAFrozenCandidateIsRequired(t *testing.T) {
	p := newProject()
	p.git.archiveErr = errors.New("git archive failed")

	result := p.run(review.Options{Exec: "revisor"})
	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 (%s)", result.Code, result.Message)
	}
	if len(p.shell.commands) != 0 {
		t.Error("the reviewer ran without a frozen candidate")
	}
	if len(p.git.discarded) != 0 {
		t.Error("a failed freeze discarded a directory it never made")
	}
}
