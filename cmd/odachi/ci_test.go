package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gitIn runs one git command in dir and returns its trimmed stdout. Tests use it
// only to build and inspect the temporary repository.
func gitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

// projectWithOrigin builds the state a pushed feature is in: a project at its
// base commit, an `origin` that carries that commit as origin/main, and a
// feature branch on top of it. The fence at the base forbids payments; the
// branch empties it and adds a file where it said no.
//
// That is the defect --ci exists for: the branch loosened its own copy of the
// policy, so a gate that reads the branch's fence says ready, and the gate CI
// runs -built from the base, reading the base fence- refuses the same tree.
func projectWithOrigin(t *testing.T, dir string) {
	t.Helper()

	initRepo(t, dir)
	gitIn(t, dir, "config", "user.name", "odachi test")
	gitIn(t, dir, "config", "user.email", "test@odachi.local")

	seedProject(t, dir)
	gitIn(t, dir, "add", "-A")
	gitIn(t, dir, "commit", "-q", "-m", "base")
	gitIn(t, dir, "branch", "-M", "main")

	origin := filepath.Join(t.TempDir(), "origin.git")
	if out, err := exec.Command("git", "init", "-q", "--bare", origin).CombinedOutput(); err != nil {
		t.Fatalf("git init --bare: %v: %s", err, out)
	}
	gitIn(t, dir, "remote", "add", "origin", origin)
	gitIn(t, dir, "push", "-q", "origin", "main")
	// A clone writes this symref itself; the test writes it because it built the
	// remote by hand. It is what RemoteBase reads first.
	gitIn(t, dir, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")

	gitIn(t, dir, "checkout", "-q", "-b", "feature")
	writeFile(t, dir, "odd/fence.toml", "# this branch emptied the policy it must obey\n")
	writeFile(t, dir, "internal/payments/stripe.go", "package payments\n")
	gitIn(t, dir, "add", "-A")
	gitIn(t, dir, "commit", "-q", "-m", "the feature, with its own fence")
}

// TestCINamesTheRemoteBaseInsteadOfASecondOne: --ci decides where the policy and
// the base come from, so a --base or --fence-from beside it says two things at
// once. The run refuses and names both flags rather than picking one.
func TestCINamesTheRemoteBaseInsteadOfASecondOne(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)

	for _, second := range []string{"--base", "--fence-from"} {
		code, _, errOut := runDir(t, dir, []string{"check", "--ci", second, "origin/main"}, "")

		if code != 1 {
			t.Fatalf("check --ci %s: code = %d, want 1 (stderr: %q)", second, code, errOut)
		}
		if !strings.Contains(errOut, "--ci") || !strings.Contains(errOut, second) {
			t.Errorf("stderr %q does not name both --ci and %s", errOut, second)
		}
	}
}

// TestCIWithoutARemoteNamesTheFetch: a clone with no origin cannot answer what CI
// would judge, and the run says the command that would create the ref instead of
// falling back to the branch's own fence.
func TestCIWithoutARemoteNamesTheFetch(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)

	code, _, errOut := runDir(t, dir, []string{"check", "--ci"}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1 (stderr: %q)", code, errOut)
	}
	if !strings.Contains(errOut, "git fetch") {
		t.Errorf("stderr %q does not name the fetch that would create the base", errOut)
	}
}

// TestCIReadsTheFenceFromTheRemoteBase is the defect of the day: the branch's own
// fence says the change is fine, and the policy CI reads from the base says no.
// --fence-from HEAD is the local answer that passed while CI refused, so the test
// pins both: ready from the branch's fence, blocked from the base's.
func TestCIReadsTheFenceFromTheRemoteBase(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	projectWithOrigin(t, dir)

	code, _, errOut := runDir(t, dir, []string{"check", "--fence-from", "HEAD"}, "")
	if code != 0 {
		t.Fatalf("check --fence-from HEAD: code = %d, want 0 (stderr: %q)", code, errOut)
	}

	code, _, errOut = runDir(t, dir, []string{"check", "--ci"}, "")
	if code != 2 {
		t.Fatalf("check --ci: code = %d, want 2 (stderr: %q)", code, errOut)
	}
	if !strings.Contains(errOut, "out of scope") || !strings.Contains(errOut, "pagos") {
		t.Errorf("stderr %q does not say what the base fence forbids", errOut)
	}
}
