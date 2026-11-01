package check_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/check"
)

// TestAReadyRunNamesTheFenceItObeyed: when the policy comes from a ref, the green
// has to say which ref. A local run against a stale base and a CI run against the
// base branch otherwise read as the same answer, and that is the green that lied.
// A run without a ref says nothing: the worktree's own fence needs no name.
func TestAReadyRunNamesTheFenceItObeyed(t *testing.T) {
	files := newFakeFiles(readyFiles())
	git := &fakeGit{
		tree:    currentTree,
		filesAt: map[string]string{"origin/main:odd/fence.toml": ""},
	}

	withRef := run(t, files, git, check.Options{FenceFrom: "origin/main", Base: "origin/main"})
	if withRef.Code != 0 {
		t.Fatalf("the trusted ref blocked a ready tree: %s", withRef.Message)
	}
	if joined := strings.Join(withRef.Lines, "\n"); !strings.Contains(joined, "fence: origin/main") {
		t.Errorf("the ready run does not name the fence it obeyed:\n%s", joined)
	}

	plain := run(t, files, git, check.Options{})
	if joined := strings.Join(plain.Lines, "\n"); strings.Contains(joined, "fence:") {
		t.Errorf("a run with no ref named a fence:\n%s", joined)
	}
}

// TestABlockedRunNamesTheFenceItObeyed: a block read from a ref is as much a
// claim about that ref as a green is. Without the name, a human cannot tell the
// base branch's policy from the branch's own copy, which is the whole complaint.
func TestABlockedRunNamesTheFenceItObeyed(t *testing.T) {
	files := newFakeFiles(readyFiles())
	git := &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/payments/stripe.go"},
		filesAt: map[string]string{
			"origin/main:odd/fence.toml": "[[rule]]\nnot = \"pagos\"\npaths = [\"**/payments/**\"]\n",
		},
	}

	result := run(t, files, git, check.Options{FenceFrom: "origin/main", Base: "origin/main"})

	if result.Code != 2 {
		t.Fatalf("the base fence did not block: %d (%s)", result.Code, result.Message)
	}
	if joined := strings.Join(result.Lines, "\n"); !strings.Contains(joined, "fence: origin/main") {
		t.Errorf("the blocked run does not name the fence it obeyed:\n%s", joined)
	}
}
