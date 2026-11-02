package plan_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/plan"
)

// dirtyGit is a worktree that is not the commit: HEAD carries one tree and the
// worktree another, which is the state the task doc itself creates.
type dirtyGit struct{}

func (dirtyGit) HeadTree(string) (string, error)       { return "aaaaaaaa", nil }
func (dirtyGit) WorktreeTree(string) (string, error)   { return "bbbbbbbb", nil }
func (dirtyGit) TreeOf(string, string) (string, error) { return "", errors.New("no such commit") }

// TestPlanNamesTheFlagWhenItCannotRecordTheTree: the tool already knew it was
// not recording anything; naming the flag that would is the difference between
// a note and a dead end.
func TestPlanNamesTheFlagWhenItCannotRecordTheTree(t *testing.T) {
	files := project()

	res := plan.Run(plan.Options{Dir: "raiz", Done: "T2", Git: dirtyGit{}}, files)

	if res.Code != plan.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, plan.CodeOK, res.Message)
	}
	if !strings.Contains(res.Message, "not the commit") {
		t.Errorf("message %q does not say why the tree was not recorded", res.Message)
	}
	if !strings.Contains(res.Message, "--commit") {
		t.Errorf("message %q does not name --commit: the human is left with a mark nobody can write", res.Message)
	}
	step := lastStep(t, files)
	if !step.Done {
		t.Error("Done = false, want true: the step still closes")
	}
	if step.Commit != "" {
		t.Errorf("Commit = %q, want empty: nothing is guessed", step.Commit)
	}
}
