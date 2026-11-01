package gitx_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/gitx"
)

// newClone builds a real repository with an `origin` remote: a seed repo with the
// named branch, cloned into a folder. The clone is what CI's base ref looks like
// from the machine that pushes, and cloning a local path needs no network.
func newClone(t *testing.T, branch string) (root string) {
	t.Helper()

	seed, _ := newRepo(t)
	git(t, seed, "branch", "-M", branch)

	parent := t.TempDir()
	root = filepath.Join(parent, "clone")
	git(t, parent, "clone", "--quiet", seed, root)

	return root
}

// TestRemoteBaseIsTheRemoteDefaultBranch: CI builds from the base branch and
// reads the fence there, so a local run that asks for the same answer must name
// the remote's own HEAD, not a branch spelled in this code. A remote on `trunk`
// is the proof: nothing here assumes `main`.
func TestRemoteBaseIsTheRemoteDefaultBranch(t *testing.T) {
	t.Parallel()
	root := newClone(t, "trunk")

	got, err := gitx.RemoteBase(root)
	if err != nil {
		t.Fatalf("RemoteBase: %v", err)
	}
	if got != "origin/trunk" {
		t.Errorf("RemoteBase = %q, want origin/trunk", got)
	}
}

// TestRemoteBaseFallsBackToMainWithoutSymref: a clone that lost origin/HEAD, or
// one made by an older git, still has the branch itself. The fallback is a
// spelling of the usual base, not a guess about the project.
func TestRemoteBaseFallsBackToMainWithoutSymref(t *testing.T) {
	t.Parallel()
	root := newClone(t, "main")
	git(t, root, "symbolic-ref", "--delete", "refs/remotes/origin/HEAD")

	got, err := gitx.RemoteBase(root)
	if err != nil {
		t.Fatalf("RemoteBase: %v", err)
	}
	if got != "origin/main" {
		t.Errorf("RemoteBase = %q, want origin/main", got)
	}
}

// TestRemoteBaseNamesTheFetchThatWouldCreateIt: a clone with no origin cannot
// answer, and the answer must be the command that fixes it. Guessing a base
// here would be the defect: a local green against a base CI never uses.
func TestRemoteBaseNamesTheFetchThatWouldCreateIt(t *testing.T) {
	t.Parallel()
	root, _ := newRepo(t)

	_, err := gitx.RemoteBase(root)
	if err == nil {
		t.Fatal("a repository with no origin answered a base")
	}
	for _, want := range []string{"origin", "git fetch"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %q, want %q named", err, want)
		}
	}
}

// TestRemoteBaseRefusesARemoteWithNoKnownBase: origin exists and carries no
// branch this code can name and no symref to read. The error says what was
// looked for instead of picking one.
func TestRemoteBaseRefusesARemoteWithNoKnownBase(t *testing.T) {
	t.Parallel()
	root := newClone(t, "trunk")
	git(t, root, "symbolic-ref", "--delete", "refs/remotes/origin/HEAD")

	_, err := gitx.RemoteBase(root)
	if err == nil {
		t.Fatal("a remote with no main and no symref answered a base")
	}
	for _, want := range []string{"origin/main", "origin/master"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %q, want %q named", err, want)
		}
	}
}
