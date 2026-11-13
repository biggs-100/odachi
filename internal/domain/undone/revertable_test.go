package undone_test

import (
	"slices"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/undone"
)

// TestRevertableKeepsTheProductionGoFiles is the boundary of the undo: reverting
// a test removes the witness, and reverting a fixture turns a behavior into a
// broken test. Neither is the change this row judges.
func TestRevertableKeepsTheProductionGoFiles(t *testing.T) {
	paths := []string{
		"internal/app/check/check.go",
		"internal/app/check/check_test.go",
		"go.mod",
		"go.sum",
		"brief.md",
		"docs/decisions/0052-the-gate-runs-the-undo.md",
		"internal/probe/testdata/lenses/clean/cache.go",
		"internal/domain/undone/testdata/golden.txt",
		"internal\\adapters\\gitx\\copy.go",
		"",
	}

	got := undone.Revertable(paths)

	want := []string{
		"internal/app/check/check.go",
		"internal/adapters/gitx/copy.go",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("Revertable = %v, want %v", got, want)
	}
}
