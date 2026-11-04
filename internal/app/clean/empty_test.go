package clean_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/app/clean"
)

// TestApplyEmptiesAClassByTheTree: a class with no lifetime is a folder, not a
// list of files, so the subtree the walk never reached goes with it. The classes
// with a lifetime keep their file-by-file decision: a record is evidence and one
// file is kept or doomed on its own.
func TestApplyEmptiesAClassByTheTree(t *testing.T) {
	record := ".git/odachi/reviews/" + strings.Repeat("a", 40) + ".json"
	f := newFake(map[string]string{
		".git/odachi/scratch/plan.txt":                             "plan",
		".git/odachi/scratch/a/b/c/d/e/f/deep.txt":                 "deep",
		".git/odachi/answers/" + strings.Repeat("b", 40) + ".json": "{}",
		record: "{}",
	})
	f.put(record, "{}", time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC))

	result := apply(f, clean.Options{Apply: true})

	if result.Code != clean.CodeOK {
		t.Fatalf("code = %d, want 0: %s", result.Code, result.Message)
	}
	if len(f.removedTrees) != 2 {
		t.Fatalf("emptied %v, want the scratch and the answers folders", f.removedTrees)
	}
	for _, root := range f.removedTrees {
		if strings.Contains(root, "reviews") {
			t.Errorf("the reviews folder was emptied as a tree: %s", root)
		}
	}
	for _, gone := range []string{".git/odachi/scratch/plan.txt", ".git/odachi/scratch/a/b/c/d/e/f/deep.txt"} {
		if f.has(gone) {
			t.Errorf("%s survived the sweep", gone)
		}
	}
	for path := range f.files {
		if strings.Contains(path, "answers") {
			t.Errorf("the answer scaffolding survived the sweep: %s", path)
		}
	}
	if !f.has(record) {
		t.Errorf("the kept review record was removed")
	}
	if !strings.Contains(result.Message, "emptied 2 folders") {
		t.Errorf("message = %q, want it to name the two emptied folders", result.Message)
	}
	if !strings.Contains(result.Message, "1 folder beyond the shape went with them") {
		t.Errorf("message = %q, want the folder beyond the walk named", result.Message)
	}
}

// TestApplyRefusesWhenAFolderCannotBeEmptied: a folder that cannot be removed is
// a refusal, not a half-sweep, and it names the path.
func TestApplyRefusesWhenAFolderCannotBeEmptied(t *testing.T) {
	f := newFake(map[string]string{
		".git/odachi/scratch/plan.txt":                             "plan",
		".git/odachi/reviews/" + strings.Repeat("a", 40) + ".json": "{}",
	})
	f.refuseTree(".git/odachi/scratch", errors.New("file in use"))

	result := apply(f, clean.Options{Apply: true})

	if result.Code != clean.CodeBlocked {
		t.Fatalf("code = %d, want blocked: %s", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, "scratch") || !strings.Contains(result.Message, "file in use") {
		t.Errorf("message = %q, want the path and the cause", result.Message)
	}
}
