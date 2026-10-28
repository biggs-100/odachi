package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// failingGit is a repository port that cannot answer, the way a corrupted clone
// or a git that refuses to run leaves it.
type failingGit struct{ err error }

func (f failingGit) CommonDir(string) (string, error) { return "", f.err }

// TestTheSelectionAdapterAnswersBothOnlyWhenThereIsNoRepository: a folder that
// never ran git init has no store, and "the project never chose" is the right
// answer there - the same answer as a missing file.
func TestTheSelectionAdapterAnswersBothOnlyWhenThereIsNoRepository(t *testing.T) {
	t.Parallel()

	got, err := (agentSelection{}).Written(t.TempDir())

	if err != nil {
		t.Fatalf("Written in a folder with no repository = error %v, want the empty answer", err)
	}
	if got.Agents != "" || got.ProjectAgents != "" {
		t.Errorf("Written = %+v, want the empty answer (both clients, the mirror in the tree)", got)
	}
}

// TestTheSelectionAdapterNamesAGitFailure: the answer "I could not look" is not
// the answer "there is nothing". A git that fails must reach the caller, or a
// verb falls back to both clients and rewrites the files of the one the person
// excluded.
func TestTheSelectionAdapterNamesAGitFailure(t *testing.T) {
	t.Parallel()

	failure := errors.New("fatal: detected dubious ownership in repository")
	_, err := (agentSelection{git: failingGit{err: failure}}).Written(t.TempDir())

	if !errors.Is(err, failure) {
		t.Fatalf("Written = error %v, want the git failure", err)
	}
}

// TestTheSelectionAdapterReadsTheStore: the happy path the two branches guard.
func TestTheSelectionAdapterReadsTheStore(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	initRepo(t, dir)
	storeDir := filepath.Join(dir, ".git", "odachi")
	if err := os.MkdirAll(storeDir, 0o755); err != nil {
		t.Fatalf("mkdir store: %v", err)
	}
	if err := os.WriteFile(filepath.Join(storeDir, "agents.toml"), []byte("agents = \"opencode\"\n"), 0o644); err != nil {
		t.Fatalf("seed selection: %v", err)
	}

	got, err := (agentSelection{}).Written(dir)
	if err != nil {
		t.Fatalf("Written: %v", err)
	}
	if got.Agents != "opencode" {
		t.Errorf("Written = %+v, want opencode", got)
	}
}
