package undo_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/app/lock"
	"github.com/biggs-100/odachi/internal/app/undo"
)

// fakeLock is the Lock port in memory: it answers with what the test decides and
// records what undo asked for.
type fakeLock struct {
	held bool
	line string
	err  error

	paths    []string
	owners   []string
	releases int
}

func (f *fakeLock) Acquire(path, owner string) (bool, string, error) {
	f.paths = append(f.paths, path)
	f.owners = append(f.owners, owner)
	if f.err != nil {
		return false, "", f.err
	}
	return f.held, f.line, nil
}

func (f *fakeLock) Release() error {
	f.releases++
	return nil
}

// TestUndoIsRefusedWhileAnotherRunHoldsTheLock: undo replaces an artifact and
// backs up what it replaces, so it writes more than one file and takes the lock.
// A refusal happens before any backup is even listed.
func TestUndoIsRefusedWhileAnotherRunHoldsTheLock(t *testing.T) {
	refusal := lock.Refusal(lock.Owner{Verb: "undo", PID: "5150", Instant: "2026-10-21T15:47:03Z"}, 12*time.Second)
	port := &fakeLock{held: false, line: refusal}
	files := newFakeFiles()
	files.put(briefRel, "brief nuevo", at(21, 23, 0))
	files.put(briefRel+".bak-"+digest("brief viejo"), "brief viejo", at(20, 9, 0))

	res := undo.Run(undo.Options{Dir: dir}, files, port)

	if res.Code != undo.CodeBlocked {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, undo.CodeBlocked, res.Message)
	}
	for _, want := range []string{"another odachi run is writing to this project", "undo, pid 5150", "12s ago", ".odachi/lock"} {
		if !strings.Contains(res.Message, want) {
			t.Errorf("Message = %q, want it to contain %q", res.Message, want)
		}
	}
	if len(files.touched) != 0 {
		t.Errorf("the blocked run touched the project: %v", files.touched)
	}
	if len(files.wrote) != 0 || len(files.backups) != 0 {
		t.Errorf("escrituras: %v, backups: %v; want none", files.wrote, files.backups)
	}
	if port.releases != 0 {
		t.Errorf("a refusal released the lock %d times; it was never ours", port.releases)
	}
	if len(port.owners) != 1 || port.owners[0] != "undo" {
		t.Errorf("Acquire owners = %v, want one call with %q", port.owners, "undo")
	}
}

// TestUndoAnnouncesAStaleLock: an abandoned lock is taken over and the line is
// what the human sees first.
func TestUndoAnnouncesAStaleLock(t *testing.T) {
	const notice = "there was an old lock from start, 31m12s ago; I took it"
	files := newFakeFiles()
	files.put(briefRel, "brief nuevo", at(21, 23, 0))
	files.put(briefRel+".bak-"+digest("brief viejo"), "brief viejo", at(20, 9, 0))
	port := &fakeLock{held: true, line: notice}

	res := undo.Run(undo.Options{Dir: dir, File: "brief"}, files, port)

	if res.Code != undo.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, undo.CodeOK, res.Message)
	}
	if !strings.Contains(res.Message, notice) {
		t.Errorf("Message = %q, want the announcement %q", res.Message, notice)
	}
	if len(res.Lines) == 0 || res.Lines[0] != notice {
		t.Errorf("Lines = %v, want the announcement first", res.Lines)
	}
	if port.releases != 1 {
		t.Errorf("port released %d times, want 1", port.releases)
	}
}

// TestUndoReleasesTheLockWhenThereIsNothingToUndo: the blocked path releases the
// lock too, or a project with no backups would stay locked for ten minutes.
func TestUndoReleasesTheLockWhenThereIsNothingToUndo(t *testing.T) {
	files := newFakeFiles()
	port := &fakeLock{held: true}

	res := undo.Run(undo.Options{Dir: dir}, files, port)

	if res.Code != undo.CodeBlocked {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, undo.CodeBlocked, res.Message)
	}
	if port.releases != 1 {
		t.Errorf("port released %d times, want exactly 1 after a failed run", port.releases)
	}
}

// TestUndoStopsWhenTheLockCannotBeTaken: no lock, no writes.
func TestUndoStopsWhenTheLockCannotBeTaken(t *testing.T) {
	files := newFakeFiles()
	files.put(briefRel, "brief nuevo", at(21, 23, 0))
	files.put(briefRel+".bak-"+digest("brief viejo"), "brief viejo", at(20, 9, 0))
	port := &fakeLock{err: errors.New("permission denied")}

	res := undo.Run(undo.Options{Dir: dir}, files, port)

	if res.Code != undo.CodeBlocked {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, undo.CodeBlocked, res.Message)
	}
	for _, want := range []string{".odachi/lock", "permission denied"} {
		if !strings.Contains(res.Message, want) {
			t.Errorf("Message = %q, want it to contain %q", res.Message, want)
		}
	}
	if len(files.touched) != 0 {
		t.Errorf("the run without a lock touched the project: %v", files.touched)
	}
}
