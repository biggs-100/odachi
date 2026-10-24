package check_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/check"
)

// TestTheFenceCanComeFromATrustedRef: a CI gate reads the base branch's fence,
// so the change that asks to pass cannot loosen the rule it must obey.
func TestTheFenceCanComeFromATrustedRef(t *testing.T) {
	files := newFakeFiles(readyFiles())
	git := &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/payments/stripe.go"},
		filesAt: map[string]string{
			"origin/main:.odachi/fence.toml": "[[rule]]\nnot = \"pagos\"\npaths = [\"**/payments/**\"]\n",
		},
	}

	worktree := run(t, files, git, check.Options{})
	if worktree.Code != 0 {
		t.Fatalf("without --fence-from: code = %d (%s)", worktree.Code, worktree.Message)
	}

	trusted := run(t, files, git, check.Options{FenceFrom: "origin/main"})
	if trusted.Code == 0 {
		t.Fatal("the trusted fence did not block the change")
	}
	if !strings.Contains(trusted.Message, "pagos") {
		t.Errorf("message = %q, want the rule named", trusted.Message)
	}
}

// TestAnUnreadableFenceRefIsABlock: a typo in the ref must not read as "there
// is no fence". Turning the gate off by mistake is the failure this refuses.
func TestAnUnreadableFenceRefIsABlock(t *testing.T) {
	files := newFakeFiles(readyFiles())
	git := &fakeGit{tree: currentTree, fileAtErr: errors.New("fatal: not a valid object name")}

	result := run(t, files, git, check.Options{FenceFrom: "origin/mian"})

	if result.Code == 0 {
		t.Fatal("a bad ref read as a project without a fence")
	}
	if !strings.Contains(result.Message, "--fence-from") {
		t.Errorf("message = %q, want the flag named", result.Message)
	}
}
