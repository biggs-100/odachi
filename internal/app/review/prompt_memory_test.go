package review_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/review"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
)

// TestThePromptCarriesOnlyTheCleanRowsItCanUse: a clean row says "this file was read and
// nothing was found". It matters when the file is in the diff, and it is noise when it is
// not - 99.7 % of the carried passes name a path the candidate never touched. The prompt
// carries the ones it can use; the store keeps them all, so the verdict still inherits
// every row.
func TestThePromptCarriesOnlyTheCleanRowsItCanUse(t *testing.T) {
	p := newLensProject()
	p.git.blobs = map[string]string{"a.go": gitBlob, "internal/other.go": gitBlob}

	stored, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "primero",
		Lenses:   []string{domainreview.Reliability},
		Passes: []domainreview.Pass{
			{Path: "a.go", Blob: gitBlob, Lens: domainreview.Reliability},
			{Path: "internal/other.go", Blob: gitBlob, Lens: domainreview.Reliability},
		},
	}, otherTree, "HEAD", now)
	if err != nil {
		t.Fatalf("seeding the verdict: %v", err)
	}
	p.files.files[storePath(otherTree)] = domainreview.Render(stored)

	result := p.run(review.Options{})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	if !strings.Contains(result.Prompt, "- a.go (reliability) audited clean") {
		t.Errorf("the prompt lost the clean row of the path that changed:\n%s", result.Prompt)
	}
	if strings.Contains(result.Prompt, "internal/other.go") {
		t.Errorf("the prompt carries a clean row about a path the candidate never touched")
	}

	// The store is untouched: the row is still there for the verdict to inherit.
	data, err := domainreview.Parse(p.files.files[storePath(otherTree)])
	if err != nil {
		t.Fatalf("the stored verdict does not parse: %v", err)
	}
	if len(data.Passes) != 2 {
		t.Errorf("the store lost a row: %d passes, want 2", len(data.Passes))
	}
}
