package check_test

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/app/check"
	"github.com/biggs-100/odachi/internal/domain/gate"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
)

// TestTheNewestVerdictIsTheOneTheStoreWroteLast: the gate reads the verdict the
// store wrote last, not the one whose answer declares the latest clock. A
// reviewer that writes a date where a time was expected must not send the gate
// back to a tree it already reviewed and then call the newer tree unreviewed.
func TestTheNewestVerdictIsTheOneTheStoreWroteLast(t *testing.T) {
	const base = "9a211be952e250bfc7adc0acdbe496c88143ff1f"
	const older = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	const newer = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, filepath.FromSlash("odd/tasks/0007-turnos.md"))] = []byte(reviewableDoc(base, true))
	files.files[reviewPath(older)] = []byte(verdictAtInstant(t, older, base, time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)))
	files.files[reviewPath(newer)] = []byte(verdictAtInstant(t, newer, base, time.Date(2026, 11, 7, 0, 0, 0, 0, time.UTC)))
	// The store wrote the correction last, and that is the order that matters.
	files.modTimes = map[string]time.Time{
		reviewPath(older): time.Date(2026, 11, 7, 10, 0, 0, 0, time.UTC),
		reviewPath(newer): time.Date(2026, 11, 7, 11, 0, 0, 0, time.UTC),
	}

	git := &fakeGit{
		tree: currentTree,
		betweenByFrom: map[string][]string{
			base:  {"internal/app/check/check.go"},
			older: {"internal/app/check/check.go"},
			newer: {"docs/next-session.md"},
		},
	}

	got := run(t, files, git, check.Options{ReviewMode: true})

	if got.Code != int(gate.Ready) {
		t.Fatalf("Code = %d, want ready (message %q)", got.Code, got.Message)
	}
	if joined := strings.Join(got.Lines, "\n"); strings.Contains(joined, "no verdict") {
		t.Fatalf("the gate read the older verdict and named the newer tree as unreviewed:\n%s", joined)
	}
}

// verdictAtInstant seals a verdict whose answer declares the given clock, so a
// test can make the declaration disagree with the store's write order.
func verdictAtInstant(t *testing.T, tree, base string, at time.Time) string {
	t.Helper()

	sealed, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "reviewer",
		Lenses:   []string{domainreview.Reliability},
	}, tree, base, at)
	if err != nil {
		t.Fatalf("sealing the seeded verdict: %v", err)
	}
	return string(domainreview.Render(sealed))
}
