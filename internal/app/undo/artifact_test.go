package undo_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/undo"
)

// otherTaskRel is a second feature's document: several of them with backups is
// the state `--file task` could not get out of.
const otherTaskRel = "odd/tasks/0002-otra-cosa.md"

// twoTasks seeds two task documents, each with one version in the store.
func twoTasks() *fakeFiles {
	return newFakeFiles().
		put(taskRel, "la primera, como está", at(22, 0, 0)).
		putStore(taskRel, "la primera, como estaba", at(20, 0, 0)).
		put(otherTaskRel, "la segunda, como está", at(22, 0, 0)).
		putStore(otherTaskRel, "la segunda, como estaba", at(21, 0, 0))
}

// TestUndoAsksWhichTaskInsteadOfNamingFile: with several documents of the same
// kind, --file is the flag the run already has, so naming it is the dead end
// 0164 exists to remove.
func TestUndoAsksWhichTaskInsteadOfNamingFile(t *testing.T) {
	files := twoTasks()

	res := undo.Run(undo.Options{Dir: dir, Store: store, File: "task"}, files, nil)

	if res.Code != undo.CodeBlocked {
		t.Fatalf("Code = %d, want %d: it guessed which document (message: %s)", res.Code, undo.CodeBlocked, res.Message)
	}
	if !strings.Contains(res.Message, "--artifact") {
		t.Errorf("message %q does not name --artifact, the flag that answers it", res.Message)
	}
	for _, want := range []string{taskRel, otherTaskRel} {
		if !strings.Contains(res.Message, want) {
			t.Errorf("message %q does not name %s", res.Message, want)
		}
	}
	if strings.Contains(res.Message, "with --file") {
		t.Errorf("message %q sends the reader back to the flag they already used", res.Message)
	}
	if len(files.wrote) != 0 {
		t.Errorf("a refused undo wrote %v", files.wrote)
	}
}

// TestUndoRestoresTheArtifactItIsTold: naming the document is the way out of a
// project with several of them, and the restore backs up what it replaces.
func TestUndoRestoresTheArtifactItIsTold(t *testing.T) {
	files := twoTasks()

	res := undo.Run(undo.Options{Dir: dir, Store: store, File: "task", Artifact: otherTaskRel}, files, nil)

	if res.Code != undo.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, undo.CodeOK, res.Message)
	}
	if !strings.Contains(res.Message, otherTaskRel) {
		t.Errorf("message %q does not name the document it restored", res.Message)
	}
	if got := string(files.files[filepath.Join(dir, filepath.FromSlash(otherTaskRel))]); got != "la segunda, como estaba" {
		t.Errorf("the second document holds %q, want the version it was told", got)
	}
	if got := string(files.files[filepath.Join(dir, filepath.FromSlash(taskRel))]); got != "la primera, como está" {
		t.Errorf("the first document holds %q: it was not the one named", got)
	}
	if len(files.backups) != 1 {
		t.Errorf("backups = %v, want the one it replaced", files.backups)
	}
}

// TestUndoAcceptsABareName: a human types the file name they see in `--list`,
// not always the whole path, and both mean the same document.
func TestUndoAcceptsABareName(t *testing.T) {
	files := twoTasks()

	res := undo.Run(undo.Options{Dir: dir, Store: store, Artifact: "0002-otra-cosa.md"}, files, nil)

	if res.Code != undo.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, undo.CodeOK, res.Message)
	}
	if got := string(files.files[filepath.Join(dir, filepath.FromSlash(otherTaskRel))]); got != "la segunda, como estaba" {
		t.Errorf("the second document holds %q, want the version it was told", got)
	}
}

// TestUndoRefusesAnArtifactItKeepsNoVersionOf: naming something that has no
// backup is refused with the ones that do, instead of an empty answer.
func TestUndoRefusesAnArtifactItKeepsNoVersionOf(t *testing.T) {
	files := twoTasks()

	res := undo.Run(undo.Options{Dir: dir, Store: store, Artifact: "odd/tasks/9999-no-existe.md"}, files, nil)

	if res.Code != undo.CodeBlocked {
		t.Fatalf("Code = %d, want %d: it restored something nobody named", res.Code, undo.CodeBlocked)
	}
	for _, want := range []string{taskRel, otherTaskRel} {
		if !strings.Contains(res.Message, want) {
			t.Errorf("message %q does not name %s", res.Message, want)
		}
	}
	if len(files.wrote) != 0 {
		t.Errorf("a refused undo wrote %v", files.wrote)
	}
}

// TestUndoRefusesAnArtifactOutsideTheKindItWasTold: --file and --artifact are
// two ways of naming one document, and a contradiction is not a request.
func TestUndoRefusesAnArtifactOutsideTheKindItWasTold(t *testing.T) {
	files := twoTasks()

	res := undo.Run(undo.Options{Dir: dir, Store: store, File: "brief", Artifact: otherTaskRel}, files, nil)

	if res.Code != undo.CodeBlocked {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, undo.CodeBlocked, res.Message)
	}
	if !strings.Contains(res.Message, "brief") {
		t.Errorf("message %q does not name the artifact the run asked for", res.Message)
	}
}
