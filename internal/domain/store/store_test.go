package store_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/store"
)

// fakeGit answers the one question the store asks.
type fakeGit struct {
	common string
	err    error
	asked  string
}

func (f *fakeGit) CommonDir(dir string) (string, error) {
	f.asked = dir
	return f.common, f.err
}

// TestRootHangsFromTheGitDirectory: the store is the git directory plus one
// folder, and the caller's directory is what git is asked about.
func TestRootHangsFromTheGitDirectory(t *testing.T) {
	common := t.TempDir()
	git := &fakeGit{common: common}

	root, err := store.Root("/un/proyecto", git)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	if want := filepath.Join(common, store.DirName); root != want {
		t.Errorf("Root = %q, want %q", root, want)
	}
	if git.asked != "/un/proyecto" {
		t.Errorf("CommonDir asked about %q, want the caller's directory", git.asked)
	}
}

// TestRootCreatesNothing: resolving the store is a read. A verb that only asks
// where the store is must not leave a folder behind, or every check would write.
func TestRootCreatesNothing(t *testing.T) {
	common := t.TempDir()

	root, err := store.Root("/un/proyecto", &fakeGit{common: common})
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	if _, statErr := os.Stat(root); !errors.Is(statErr, os.ErrNotExist) {
		t.Errorf("Root created %s; resolving the store must write nothing", root)
	}
}

// TestRootRefusesWithoutAGitDirectory: no git means no store, and the refusal
// names the reason. Guessing a location would put the evidence somewhere the
// next run cannot find.
func TestRootRefusesWithoutAGitDirectory(t *testing.T) {
	_, err := store.Root("/un/proyecto", &fakeGit{err: errors.New("fatal: not a git repository")})
	if err == nil {
		t.Fatal("Root without a git directory = nil error, want a refusal")
	}
}

// TestRootKeepsTheSentinel: a caller that has to tell "there is no store here"
// from "git could not answer" gets the distinction through the wrap, because the
// error is wrapped with %w and not flattened into a sentence.
func TestRootKeepsTheSentinel(t *testing.T) {
	sentinel := errors.New("not a git repository")
	_, err := store.Root("/un/proyecto", &fakeGit{err: sentinel})

	if !errors.Is(err, sentinel) {
		t.Fatalf("Root error = %v, want it to wrap the adapter's sentinel", err)
	}
}
