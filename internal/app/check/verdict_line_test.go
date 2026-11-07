package check_test

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/app/check"
	"github.com/biggs-100/odachi/internal/domain/gate"
)

// TestAReadyCloseNamesTheVerdictsTree: the proof names one tree and the verdict
// judged another, because the trees after the close only move documents. At the
// close both are printed, so a reader can tell which bytes were reviewed and
// which were published.
func TestAReadyCloseNamesTheVerdictsTree(t *testing.T) {
	const base = "9a211be952e250bfc7adc0acdbe496c88143ff1f"
	const judged = "cccccccccccccccccccccccccccccccccccccccc"

	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, filepath.FromSlash("odd/tasks/0007-turnos.md"))] = []byte(reviewableDoc(base, true))
	files.files[reviewPath(judged)] = []byte(verdictAtInstant(t, judged, base, time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)))
	files.modTimes = map[string]time.Time{
		reviewPath(judged): time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC),
	}

	git := &fakeGit{
		tree: currentTree,
		betweenByFrom: map[string][]string{
			base:   {"internal/app/check/check.go"},
			judged: {},
		},
	}

	got := run(t, files, git, check.Options{})

	if got.Code != int(gate.Ready) {
		t.Fatalf("Code = %d, want ready (message %q)", got.Code, got.Message)
	}
	if joined := strings.Join(got.Lines, "\n"); !strings.Contains(joined, "verdict: "+judged) {
		t.Fatalf("the ready output does not name the verdict's tree:\n%s", joined)
	}
}

// TestACloseWithoutAVerdictSaysNothing: a silence is not a claim. The review note
// is what says a verdict is missing; this line never invents one.
func TestACloseWithoutAVerdictSaysNothing(t *testing.T) {
	const base = "9a211be952e250bfc7adc0acdbe496c88143ff1f"

	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, filepath.FromSlash("odd/tasks/0007-turnos.md"))] = []byte(reviewableDoc(base, true))

	got := run(t, files, &fakeGit{tree: currentTree}, check.Options{})

	if got.Code != int(gate.Ready) {
		t.Fatalf("Code = %d, want ready (message %q)", got.Code, got.Message)
	}
	if joined := strings.Join(got.Lines, "\n"); strings.Contains(joined, "verdict:") {
		t.Fatalf("the ready output names a verdict that does not exist:\n%s", joined)
	}
}
