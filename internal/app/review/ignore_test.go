package review_test

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/review"
)

// TestAShowWithoutAStoreIsRefused: a store that was not resolved is not an empty
// one. Joining an empty root would put a verdict relative to the process
// directory, and the run says why instead.
func TestAShowWithoutAStoreIsRefused(t *testing.T) {
	p := newProject()

	result := p.runWithoutStore(review.Options{Show: true})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want a refusal (message: %s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, "clone's store") {
		t.Errorf("message = %q, want it to name the store", result.Message)
	}
	if writes := p.files.written(); len(writes) != 0 {
		t.Errorf("writes = %v, want none", writes)
	}
}

// TestTheVerdictIsWrittenOutsideTheWorktree: the whole point of the move. A
// review writes exactly the verdict, and the verdict is not a path inside the
// project: storing it cannot change the tree that names it, so no ignore rule has
// to keep it out.
func TestTheVerdictIsWrittenOutsideTheWorktree(t *testing.T) {
	p := newProject()
	p.files.files[fullPath("veredicto.json")] = []byte(`{"reviewer":"humano","inspection":{"status":"completed","paths":["a.go"]},"findings":[]}`)

	result := p.run(review.Options{Base: "HEAD", Attach: "veredicto.json"})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	writes := p.files.written()
	if want := relative(storePath(currentTree)); !slices.Equal(writes, []string{want}) {
		t.Fatalf("writes = %v, want exactly the verdict at %s", writes, want)
	}

	inProject := []string{}
	for _, written := range writes {
		slashed := filepath.ToSlash(written)
		if strings.HasPrefix(slashed, filepath.ToSlash(dir)+"/") && !strings.HasPrefix(slashed, relative(store)+"/") {
			inProject = append(inProject, slashed)
		}
	}
	if len(inProject) != 0 {
		t.Errorf("writes = %v, want nothing under the worktree", inProject)
	}
	if want := relative(storePath(currentTree)); !slices.Equal(result.Files, []string{want}) {
		t.Errorf("result.Files = %v, want the label of the file it wrote (%s)", result.Files, want)
	}
}

// TestASecondReviewOfTheSameTreeWritesNothingNew: the store is immutable, and
// that does not depend on an ignore rule either.
func TestASecondReviewOfTheSameTreeWritesNothingNew(t *testing.T) {
	p := newProject()
	p.files.files[fullPath("veredicto.json")] = []byte(`{"reviewer":"humano","inspection":{"status":"completed","paths":["a.go"]},"findings":[]}`)
	if first := p.run(review.Options{Base: "HEAD", Attach: "veredicto.json"}); first.Code != review.CodeOK {
		t.Fatalf("first review: code = %d (%s)", first.Code, first.Message)
	}

	second := p.run(review.Options{Base: "HEAD", Attach: "veredicto.json"})
	if second.Code == review.CodeOK {
		t.Fatalf("second review: code = 0, want a refusal")
	}
	if !strings.Contains(second.Message, "already has a verdict") {
		t.Errorf("message = %q, want it to say the tree already has one", second.Message)
	}
}
