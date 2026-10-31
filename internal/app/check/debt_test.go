package check_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/check"
)

// TestTheCloseNamesTheReviewDebtItLeaves: the rows of docs/review-debt.md are the
// list a person reads before choosing the next feature. The close names them, so the
// reading does not depend on somebody remembering the file exists; a row never
// blocks, and the ready tree stays ready.
func TestTheCloseNamesTheReviewDebtItLeaves(t *testing.T) {
	files := closedFiles()
	files["docs/review-debt.md"] = `# Review debt

## Open

- **internal/app/x.go:12** - the guard asks the wrong question, and a test would say
  so. Explained across two lines.
- **internal/app/y.go:3** - another row.

## Closed by feature 0001

- **internal/app/z.go:1** - paid, and it does not count.
`

	got := run(t, newFakeFiles(files), &fakeGit{tree: currentTree}, check.Options{})

	if got.Code != 0 {
		t.Fatalf("code = %d, want 0 (%s)", got.Code, got.Message)
	}
	lines := strings.Join(got.Lines, "\n")
	if !strings.Contains(lines, "review debt: 2 open row(s) in docs/review-debt.md") {
		t.Errorf("the close does not name the debt it leaves:\n%s", lines)
	}
	if !strings.Contains(lines, "internal/app/x.go:12") || !strings.Contains(lines, "internal/app/y.go:3") {
		t.Errorf("the note does not name the rows:\n%s", lines)
	}
	if strings.Contains(lines, "internal/app/z.go:1") {
		t.Errorf("a closed row was named as open:\n%s", lines)
	}
}

// TestAProjectWithoutDebtSaysNothing: a repository that never wrote the file, or
// whose Open section is empty, gets no line at all -a note about nothing is noise-.
func TestAProjectWithoutDebtSaysNothing(t *testing.T) {
	files := closedFiles()
	files["docs/review-debt.md"] = "# Review debt\n\n## Open\n\nNothing, after 0140.\n"

	got := run(t, newFakeFiles(files), &fakeGit{tree: currentTree}, check.Options{})

	if got.Code != 0 {
		t.Fatalf("code = %d, want 0 (%s)", got.Code, got.Message)
	}
	if strings.Contains(strings.Join(got.Lines, "\n"), "review debt:") {
		t.Errorf("a project with no open rows was told about debt:\n%s", strings.Join(got.Lines, "\n"))
	}
}
