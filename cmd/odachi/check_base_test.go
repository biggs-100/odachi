package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const claimedChangeLine = "bring back the wait state on the pay button"

// claimedChangeProject is a clean feature branch whose committed change names
// itself in a trailer. The fenced path is committed, so a candidate read from
// HEAD cannot see the defect this fixture is meant to expose.
func claimedChangeProject(t *testing.T, onDefault bool) (dir, base string) {
	t.Helper()

	dir = t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)
	closeFeature(t, dir)
	commitAll(t, dir, "base")
	gitIn(t, dir, "branch", "-M", "main")
	base = headOf(t, dir)

	origin := filepath.Join(t.TempDir(), "origin.git")
	if out, err := exec.Command("git", "init", "-q", "--bare", origin).CombinedOutput(); err != nil {
		t.Fatalf("git init --bare: %v: %s", err, out)
	}
	gitIn(t, dir, "remote", "add", "origin", origin)
	gitIn(t, dir, "push", "-q", "origin", "main")
	gitIn(t, dir, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")
	if !onDefault {
		gitIn(t, dir, "checkout", "-q", "-b", "feature")
	}

	writeFile(t, dir, "internal/payments/stripe.go", "package payments\n")
	commitAll(t, dir, "the wait state\n\nOdachi-Change: "+claimedChangeLine+"\n\nRisk: none - the fixture declares no named risk\n")
	return dir, base
}

// TestCheckResolvesChangeBaseBeforeCandidateSnapshot is the clean-branch bug:
// the trailer names the feature and its merge-base, and that base has to be
// chosen before the fence and review read the candidate.
func TestCheckResolvesChangeBaseBeforeCandidateSnapshot(t *testing.T) {
	t.Parallel()

	t.Run("check sees the committed fenced path", func(t *testing.T) {
		dir, _ := claimedChangeProject(t, false)
		if status := gitIn(t, dir, "status", "--porcelain"); status != "" {
			t.Fatalf("the fixture is not a clean feature branch: %q", status)
		}

		code, _, errOut := runDir(t, dir, []string{"check"}, "")
		if code != 2 {
			t.Fatalf("check = %d, want the fence decision 2 (stderr: %q)", code, errOut)
		}
		for _, want := range []string{"out of scope", "pagos", "internal/payments/stripe.go"} {
			if !strings.Contains(errOut, want) {
				t.Errorf("the fence refusal does not name %q: %q", want, errOut)
			}
		}
	})

	t.Run("bare review shares the trailer base", func(t *testing.T) {
		dir, base := claimedChangeProject(t, false)

		code, out, errOut := runDir(t, dir, []string{"review"}, "")
		if code != 0 {
			t.Fatalf("review = %d, want 0 (stdout: %q, stderr: %q)", code, out, errOut)
		}
		for _, want := range []string{"base " + base, "internal/payments/stripe.go", claimedChangeLine} {
			if !strings.Contains(out, want) {
				t.Errorf("the review prompt does not name %q:\n%s", want, out)
			}
		}
	})
}

// TestReviewRefusesTrailerOnDefaultBranchBeforeSnapshot pins the review half of
// check's branch boundary: a trailer on main is not a document-less feature.
func TestReviewRefusesTrailerOnDefaultBranchBeforeSnapshot(t *testing.T) {
	t.Parallel()
	dir, _ := claimedChangeProject(t, true)

	code, _, errOut := runDir(t, dir, []string{"review"}, "")
	if code != 1 {
		t.Fatalf("review = %d, want 1 (stderr: %q)", code, errOut)
	}
	for _, want := range []string{"branch", "main"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("the default-branch refusal does not name %q: %q", want, errOut)
		}
	}
	if strings.Contains(errOut, "nothing to review") {
		t.Errorf("the raw HEAD candidate hid the trailer's branch refusal: %q", errOut)
	}
}

// TestUnfetchedOriginBlocksCheckAndReview keeps an existing remote distinct
// from no remote: an unresolved default must name both ways to disambiguate it.
func TestUnfetchedOriginBlocksCheckAndReview(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)
	closeFeature(t, dir)
	commitAll(t, dir, "base")
	gitIn(t, dir, "branch", "-M", "main")
	origin := filepath.Join(t.TempDir(), "origin.git")
	if out, err := exec.Command("git", "init", "-q", "--bare", origin).CombinedOutput(); err != nil {
		t.Fatalf("git init --bare: %v: %s", err, out)
	}
	gitIn(t, dir, "remote", "add", "origin", origin)

	for _, tc := range []struct {
		name string
		args []string
	}{
		{name: "check", args: []string{"check"}},
		{name: "review", args: []string{"review"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			code, _, errOut := runDir(t, dir, tc.args, "")
			if code != 1 {
				t.Fatalf("%s = %d, want 1 (stderr: %q)", tc.name, code, errOut)
			}
			for _, want := range []string{"git fetch origin", "--base"} {
				if !strings.Contains(errOut, want) {
					t.Errorf("the unresolved-remote refusal does not name %q: %q", want, errOut)
				}
			}
		})
	}

	// An explicit base is the other valid answer; it must not probe the remote.
	writeFile(t, dir, "internal/payments/stripe.go", "package payments\n")
	if code, _, errOut := runDir(t, dir, []string{"check", "--base", "HEAD"}, ""); code != 2 || strings.Contains(errOut, "git fetch origin") {
		t.Fatalf("check --base HEAD = %d, want the explicit fence decision without a remote probe (stderr: %q)", code, errOut)
	}
	if code, out, errOut := runDir(t, dir, []string{"review", "--base", "HEAD"}, ""); code != 0 || !strings.Contains(out, "internal/payments/stripe.go") {
		t.Fatalf("review --base HEAD = %d, want the explicit candidate (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if code, _, errOut := runDir(t, dir, []string{"check", "--base", ""}, ""); code != 2 || strings.Contains(errOut, "git fetch origin") {
		t.Fatalf("check --base \"\" = %d, want the explicit empty-tree fence decision without a remote probe (stderr: %q)", code, errOut)
	}

	// An explicit but unreadable ref still fails closed, and its refusal names
	// the flag that chooses a different range instead of probing origin.
	for _, tc := range []struct {
		name string
		args []string
	}{
		{name: "check", args: []string{"check", "--base", "missing"}},
		{name: "review", args: []string{"review", "--base", "missing"}},
	} {
		t.Run(tc.name+" names the explicit base refusal", func(t *testing.T) {
			code, _, errOut := runDir(t, dir, tc.args, "")
			if code != 1 {
				t.Fatalf("%s = %d, want 1 (stderr: %q)", tc.name, code, errOut)
			}
			for _, want := range []string{"missing", "--base"} {
				if !strings.Contains(errOut, want) {
					t.Errorf("the explicit-base refusal does not name %q: %q", want, errOut)
				}
			}
		})
	}
}

// TestExplicitHEADKeepsTheUnbornRepositoryFallback: HEAD names no commit range,
// so it must not turn the existing empty-tree fallback into a git-log failure.
func TestExplicitHEADKeepsTheUnbornRepositoryFallback(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)
	closeFeature(t, dir)

	if code, _, errOut := runDir(t, dir, []string{"check", "--base", "HEAD"}, ""); code != 1 || !strings.Contains(errOut, "the proof for this change is missing") {
		t.Fatalf("check --base HEAD without a commit = %d, want the existing proof refusal (stderr: %q)", code, errOut)
	}
	if code, out, errOut := runDir(t, dir, []string{"review", "--base", "HEAD"}, ""); code != 0 || !strings.Contains(out, "reservar turnos") {
		t.Fatalf("review --base HEAD without a commit = %d, want the empty-tree candidate (stdout: %q, stderr: %q)", code, out, errOut)
	}
}

// TestNoOriginKeepsTheTrailerLaneClosed: no remote is not an unresolved remote.
// The default stays HEAD until the human names a base explicitly.
func TestNoOriginKeepsTheTrailerLaneClosed(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)
	closeFeature(t, dir)
	commitAll(t, dir, "base")
	gitIn(t, dir, "branch", "-M", "main")
	gitIn(t, dir, "checkout", "-q", "-b", "feature")
	writeFile(t, dir, "internal/payments/stripe.go", "package payments\n")
	commitAll(t, dir, "the wait state\n\nOdachi-Change: "+claimedChangeLine+"\n")

	if code, _, errOut := runDir(t, dir, []string{"check"}, ""); code != 1 || !strings.Contains(errOut, "the proof for this change is missing") || strings.Contains(errOut, "git fetch origin") {
		t.Fatalf("check without origin = %d, want the existing closed lane (stderr: %q)", code, errOut)
	}
	if code, _, errOut := runDir(t, dir, []string{"review"}, ""); code != 1 || !strings.Contains(errOut, "nothing to review") {
		t.Fatalf("review without origin = %d, want the existing HEAD candidate (stderr: %q)", code, errOut)
	}

	if code, _, errOut := runDir(t, dir, []string{"check", "--base", "main"}, ""); code != 2 {
		t.Fatalf("check --base main = %d, want the explicit fence decision (stderr: %q)", code, errOut)
	}
	if code, out, errOut := runDir(t, dir, []string{"review", "--base", "main"}, ""); code != 0 || !strings.Contains(out, claimedChangeLine) {
		t.Fatalf("review --base main = %d, want the explicit trailer candidate (stdout: %q, stderr: %q)", code, out, errOut)
	}
}
