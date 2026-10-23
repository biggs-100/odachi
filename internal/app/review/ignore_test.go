package review_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/biggs-100/odachi/internal/app/review"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
)

// TestPromptWritesNothingWhenTheRepositoryIgnoresTheStore: a project that ran
// `odachi init` already has the store ignored by its own .gitignore, so the verb
// writes nothing. A second ignore file would be a new path in the worktree, which
// is exactly the tree shift this rule exists to remove.
func TestPromptWritesNothingWhenTheRepositoryIgnoresTheStore(t *testing.T) {
	p := newProject()
	p.git.ignored = true

	result := p.run(review.Options{})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if !slices.Equal(p.git.ignoredAs, []string{domainreview.Dir}) {
		t.Errorf("the git port was asked about %v, want the store directory", p.git.ignoredAs)
	}
	if writes := p.files.written(); len(writes) != 0 {
		t.Fatalf("writes = %v, want none when the repository already ignores the store", writes)
	}
	if len(result.Files) != 0 {
		t.Errorf("result.Files = %v, want none: this run wrote nothing", result.Files)
	}
}

// TestStoreFallsBackToItsOwnIgnoreFileWhenTheProbeFails: a probe that cannot
// answer is not permission to write a verdict into the tree it names. The
// fallback is the per-directory ignore file, written before the hash, as before.
func TestStoreFallsBackToItsOwnIgnoreFileWhenTheProbeFails(t *testing.T) {
	p := newProject()
	p.git.ignoredErr = errors.New("git is not installed")
	p.files.files[fullPath("veredicto.json")] = []byte(`{"reviewer":"humano","findings":[]}`)

	result := p.run(review.Options{Base: "HEAD", Attach: "veredicto.json"})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	want := domainreview.Dir + "/" + domainreview.IgnoreName
	written := p.files.written()
	if !slices.Contains(written, want) {
		t.Fatalf("writes = %v, want the store's %s after a failed probe", written, want)
	}
	// The order is still the trick: the ignore file comes before the verdict, or
	// the verdict would move the tree that names it.
	if slices.Index(written, want) > slices.Index(written, relative(storePath(currentTree))) {
		t.Errorf("writes = %v, want the ignore file before the verdict", written)
	}
}
