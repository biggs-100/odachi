package gitx_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/gitx"
)

// TestOnDefaultBranchNamesTheBranchTheBaseNames: the answer does not spell
// either name. A clone whose default branch is `trunk` answers `trunk`, and
// `origin/trunk` names it.
func TestOnDefaultBranchNamesTheBranchTheBaseNames(t *testing.T) {
	t.Parallel()
	root := newClone(t, "trunk")

	name, on, err := gitx.OnDefaultBranch(root, "origin/trunk")
	if err != nil {
		t.Fatalf("OnDefaultBranch: %v", err)
	}
	if !on || name != "trunk" {
		t.Errorf("OnDefaultBranch = (%q, %v), want (trunk, true)", name, on)
	}
}

// TestOnDefaultBranchSeesADetachedHeadOnTheDefaultBranch: a commit made straight
// to trunk and then checked out detached is still a commit on trunk. Without a
// branch name, the refs answer: the local trunk contains HEAD.
func TestOnDefaultBranchSeesADetachedHeadOnTheDefaultBranch(t *testing.T) {
	t.Parallel()
	root := newClone(t, "trunk")
	write(t, root+"/notes.txt", "a change on the default branch\n")
	git(t, root, "commit", "-am", "straight to trunk")
	git(t, root, "checkout", "--detach", "HEAD")

	name, on, err := gitx.OnDefaultBranch(root, "origin/trunk")
	if err != nil {
		t.Fatalf("OnDefaultBranch: %v", err)
	}
	if !on || name != "trunk" {
		t.Errorf("OnDefaultBranch = (%q, %v), want (trunk, true)", name, on)
	}
}

// TestOnDefaultBranchLetsADetachedFeatureTipPass: the feature branch's tip is not
// one of trunk's commits, so the trailer lane stays open. Detaching is not the
// question; which commits are is.
func TestOnDefaultBranchLetsADetachedFeatureTipPass(t *testing.T) {
	t.Parallel()
	root := newClone(t, "trunk")
	git(t, root, "checkout", "-b", "feat/wait")
	write(t, root+"/notes.txt", "the wait state\n")
	git(t, root, "commit", "-am", "the wait state")
	git(t, root, "checkout", "--detach", "HEAD")

	name, on, err := gitx.OnDefaultBranch(root, "origin/trunk")
	if err != nil {
		t.Fatalf("OnDefaultBranch: %v", err)
	}
	if on {
		t.Errorf("OnDefaultBranch = (%q, %v), want no: a feature tip is not the default branch", name, on)
	}
}

// TestOnDefaultBranchLetsAFeatureBranchThatInheritsTheUpstreamPass: `git
// checkout -b feat origin/main` records origin/main as feat's upstream, so the
// upstream is not what makes a branch the default one; the name is. Both answers
// are checked: attached, and detached at the feature's own tip.
func TestOnDefaultBranchLetsAFeatureBranchThatInheritsTheUpstreamPass(t *testing.T) {
	t.Parallel()
	root := newClone(t, "trunk")
	git(t, root, "checkout", "-b", "feat/wait", "origin/trunk")
	write(t, root+"/notes.txt", "the wait state\n")
	git(t, root, "commit", "-am", "the wait state")

	name, on, err := gitx.OnDefaultBranch(root, "origin/trunk")
	if err != nil {
		t.Fatalf("OnDefaultBranch: %v", err)
	}
	if on {
		t.Errorf("OnDefaultBranch = (%q, %v), want no: a feature branch inherits the base as upstream", name, on)
	}

	git(t, root, "checkout", "--detach", "HEAD")

	name, on, err = gitx.OnDefaultBranch(root, "origin/trunk")
	if err != nil {
		t.Fatalf("OnDefaultBranch detached: %v", err)
	}
	if on {
		t.Errorf("OnDefaultBranch = (%q, %v), want no: a detached feature tip is not the default branch", name, on)
	}
}

// TestOnDefaultBranchAnswersNoWithoutALocalDefault: CI checks out a pull request
// detached and keeps no local trunk. The commits are not on the default branch
// there, so the lane reads the trailer; refusing every detached checkout would
// make the lane unverifiable on a pull request.
func TestOnDefaultBranchAnswersNoWithoutALocalDefault(t *testing.T) {
	t.Parallel()
	root := newClone(t, "trunk")
	git(t, root, "checkout", "--detach", "origin/trunk")
	git(t, root, "branch", "-D", "trunk")

	name, on, err := gitx.OnDefaultBranch(root, "origin/trunk")
	if err != nil {
		t.Fatalf("OnDefaultBranch: %v", err)
	}
	if on {
		t.Errorf("OnDefaultBranch = (%q, %v), want no: there is no local default branch", name, on)
	}
}

// TestOnDefaultBranchNamesTheReasonWhenGitCannotAnswer: a folder that is not a
// repository cannot answer, and the error is the answer - never a silent "no"
// that would let a trailer through.
func TestOnDefaultBranchNamesTheReasonWhenGitCannotAnswer(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	_, _, err := gitx.OnDefaultBranch(root, "origin/trunk")
	if err == nil {
		t.Fatal("a folder that is not a repository answered")
	}
	if !strings.Contains(err.Error(), "git") {
		t.Errorf("error = %q, want the command that failed", err)
	}
}
