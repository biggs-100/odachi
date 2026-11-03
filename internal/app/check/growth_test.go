package check_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/check"
	"github.com/biggs-100/odachi/internal/app/clean"
)

// fakeGrowth is the store's accounting as the verb that owns the policy answers
// it: the numbers a close reads, and whether they could be read at all.
type fakeGrowth struct {
	growth clean.Growth
	err    error
	asked  []string
}

func (g *fakeGrowth) Measure(store string) (clean.Growth, error) {
	g.asked = append(g.asked, store)
	return g.growth, g.err
}

// withBase is the fixture document with a base and a state: the close knows which
// feature a tree belongs to only when the document names where it started.
func withBase(done bool) map[string]string {
	files := readyFiles()
	files["odd/tasks/0007-turnos.md"] = reviewableDoc(otherTree, done)
	return files
}

// TestTheCloseNamesTheStoreGrowth: a feature's last tree is the moment a person
// decides what to do next, and the store is the one thing that grows without
// anybody noticing. The close names it where the person is already reading, and
// the switch stays `odachi clean --apply`.
func TestTheCloseNamesTheStoreGrowth(t *testing.T) {
	growth := &fakeGrowth{growth: clean.Growth{
		Files:       1352,
		Bytes:       36595200,
		DoomedFiles: 581,
		DoomedBytes: 35651584,
	}}

	got := run(t, newFakeFiles(withBase(true)), &fakeGit{tree: currentTree}, check.Options{Growth: growth})

	if got.Code != 0 {
		t.Fatalf("code = %d, want 0 (%s)", got.Code, got.Message)
	}
	want := "store: 34.9 MB in 1352 files; odachi clean --apply would free 34.0 MB in 581 files"
	if !strings.Contains(strings.Join(got.Lines, "\n"), want) {
		t.Errorf("the close does not name the store's growth:\n%s", strings.Join(got.Lines, "\n"))
	}
	if len(growth.asked) != 1 || growth.asked[0] != store {
		t.Errorf("the close measured %v, want the clone's store %q", growth.asked, store)
	}
}

// TestAnOpenTreeDoesNotMeasureTheStore: the step's trees are not a close, and a
// measurement on every commit would be a cost nobody asked for. Only the tree
// that ends the feature asks.
func TestAnOpenTreeDoesNotMeasureTheStore(t *testing.T) {
	growth := &fakeGrowth{growth: clean.Growth{Files: 1, Bytes: 10}}

	got := run(t, newFakeFiles(withBase(false)), &fakeGit{tree: currentTree}, check.Options{Growth: growth})

	if got.Code != 0 {
		t.Fatalf("code = %d, want 0 (%s)", got.Code, got.Message)
	}
	if len(growth.asked) != 0 {
		t.Errorf("an open tree measured the store %d time(s), want none", len(growth.asked))
	}
	if strings.Contains(strings.Join(got.Lines, "\n"), "store:") {
		t.Errorf("an open tree was told about the store:\n%s", strings.Join(got.Lines, "\n"))
	}
}

// TestACloseSaysWhenTheStoreCannotBeMeasured: a read that fails is not a store of
// zero bytes. The close says so, keeps its answer, and keeps its exit code: the
// growth is a note, never a gate.
func TestACloseSaysWhenTheStoreCannotBeMeasured(t *testing.T) {
	growth := &fakeGrowth{err: errors.New("scratch is deeper than 6 levels")}

	got := run(t, newFakeFiles(withBase(true)), &fakeGit{tree: currentTree}, check.Options{Growth: growth})

	if got.Code != 0 {
		t.Fatalf("code = %d, want 0 (%s)", got.Code, got.Message)
	}
	if !strings.Contains(strings.Join(got.Lines, "\n"), "store: could not be measured: scratch is deeper than 6 levels") {
		t.Errorf("the close hides a store it could not measure:\n%s", strings.Join(got.Lines, "\n"))
	}
}
