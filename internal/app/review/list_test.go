package review_test

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/app/review"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
)

// storeOtherTree is a verdict the store holds about bytes that are not the ones that
// stand: the predecessor a correction names, and the row a listing shows.
const storeOtherTree = "2222222222222222222222222222222222222222"

// TestTheStoreHasADoor: the store is keyed by tree, and a person needs a tree to
// correct or to read what a verdict judged. `--all` lists them newest first, with the
// five facts that make one recognizable, and marks the one this worktree answers.
func TestTheStoreHasADoor(t *testing.T) {
	p := newProject()
	// Two instants, because "newest first" is the promise that makes the listing
	// useful and a fixture where both share one cannot fail on it.
	older := verdictAt(t, storeOtherTree, "abc1234", now.Add(-time.Hour), []string{"reliability"},
		domainreview.Pass{Path: "a.go", Lens: "reliability", Blob: gitBlob})
	newer := verdictAt(t, currentTree, "abc1234", now, []string{"risk", "reliability"})
	p.files.files[storePath(storeOtherTree)] = []byte(older)
	p.files.files[storePath(currentTree)] = []byte(newer)

	result := p.run(review.Options{All: true})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	lines := result.Lines
	if len(lines) < 3 {
		t.Fatalf("the listing is short: %v", lines)
	}
	if !strings.Contains(lines[0], "2 verdict(s)") {
		t.Errorf("the header does not count the store: %q", lines[0])
	}
	joined := strings.Join(lines, "\n")
	if first, second := strings.Index(joined, currentTree), strings.Index(joined, storeOtherTree); first < 0 || second < 0 || first > second {
		t.Errorf("the listing is not newest first:\n%s", joined)
	}
	for _, want := range []string{currentTree, storeOtherTree, "risk, reliability", "reliability", "base abc1234"} {
		if !strings.Contains(joined, want) {
			t.Errorf("the listing does not carry %q:\n%s", want, joined)
		}
	}
	if !slices.ContainsFunc(lines, func(line string) bool {
		return strings.Contains(line, currentTree) && strings.Contains(line, "this tree")
	}) {
		t.Errorf("the listing does not mark the tree that stands:\n%s", joined)
	}
}

// TestAnEmptyStoreSaysSo: a clone with no verdict is not an error, and the verb says
// exactly that instead of printing a header over nothing.
func TestAnEmptyStoreSaysSo(t *testing.T) {
	p := newProject()

	result := p.run(review.Options{All: true})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, "no verdict") {
		t.Errorf("message = %q, want the empty store named", result.Message)
	}
}

// TestShowNamesTheDoor: the tree a correction needs a hash for is one verb away, and
// a `--show` that finds nothing says where the others are. A person reading the store
// directory by hand is not a door.
func TestShowNamesTheDoor(t *testing.T) {
	p := newProject()
	p.files.files[storePath(storeOtherTree)] = []byte(verdictFor(t, storeOtherTree, "abc1234", []string{"reliability"}))

	result := p.run(review.Options{Show: true})
	if result.Code == review.CodeOK {
		t.Fatalf("--show found a verdict that is not there: %s", result.Message)
	}

	lines := strings.Join(result.Lines, "\n")
	if !strings.Contains(lines, "odachi review --all") {
		t.Errorf("the refusal does not name the door:\n%s", lines)
	}
	if !strings.Contains(lines, "1 other verdict") {
		t.Errorf("the refusal does not count the others:\n%s", lines)
	}
}

// verdictAt seals one verdict at an instant, so a test can pin an order.
func verdictAt(t *testing.T, tree, base string, at time.Time, lenses []string, passes ...domainreview.Pass) string {
	t.Helper()
	sealed, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "before",
		Model:    "other/model",
		Lenses:   lenses,
		Passes:   passes,
	}, tree, base, at)
	if err != nil {
		t.Fatalf("sealing %s: %v", tree, err)
	}
	return string(domainreview.Render(sealed))
}
