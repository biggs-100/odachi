package refusal_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/refusal"
)

// fixture is a package with the four shapes the walk separates: a refusal that
// names a command, one that names nothing, one built by a Sprintf, and one whose
// message is a value.
const fixture = `package fixture

import "fmt"

func blocked(string) {}

func run(message string) {
	_ = blocked("check the permissions and run odachi check again")
	_ = blocked("this one names nothing")
	_ = blocked(fmt.Sprintf("could not read %s", "x"))
	_ = blocked(message)
}
`

// writeFixture writes the package and returns its directory.
func writeFixture(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(fixture), 0o600); err != nil {
		t.Fatalf("writing the fixture: %v", err)
	}
	// A test file is not a refusal a human reads: the walk skips it.
	tests := strings.Replace(fixture, "package fixture", "package fixture_test", 1)
	if err := os.WriteFile(filepath.Join(dir, "fixture_test.go"), []byte(tests), 0o600); err != nil {
		t.Fatalf("writing the fixture's test: %v", err)
	}
	return dir
}

// TestWalkSeparatesTheFourShapes: the sentence is kept when the walk can read it and
// names no way out; a sentence that names an odachi command is not debt; a message
// built from a value is counted, because a sentence nobody can read is a sentence
// nobody can judge (0288).
func TestWalkSeparatesTheFourShapes(t *testing.T) {
	t.Parallel()

	found, err := refusal.Walk(writeFixture(t))
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	want := []string{"main.go\tcould not read %s", "main.go\tthis one names nothing"}
	if !slices.Equal(found.Sentences, want) {
		t.Errorf("sentences = %q, want %q", found.Sentences, want)
	}
	if found.Unreadable["main.go"] != 1 {
		t.Errorf("unreadable = %v, want one call in main.go", found.Unreadable)
	}
}

// TestWalkRefusesADirectoryItCannotRead: a walk that cannot see the package is not a
// package without refusals, and reporting zero would make the ratchet pass.
func TestWalkRefusesADirectoryItCannotRead(t *testing.T) {
	t.Parallel()

	if _, err := refusal.Walk(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("Walk answered a directory that does not exist")
	}
}

// TestGrewAndShrunkReadTheTwoDirections: a ratchet is only a ratchet if it refuses
// growth and reports the shrinking. The two readings are pure, so they are pinned
// here without touching the package they guard.
func TestGrewAndShrunkReadTheTwoDirections(t *testing.T) {
	t.Parallel()

	found := []string{"check.go\tkept", "check.go\tnew"}
	baseline := []string{"check.go\tgone", "check.go\tkept"}

	if got := refusal.Grew(found, baseline); !slices.Equal(got, []string{"check.go\tnew"}) {
		t.Errorf("Grew = %v, want only the row the baseline does not freeze", got)
	}
	if got := refusal.Shrunk(found, baseline); !slices.Equal(got, []string{"check.go\tgone"}) {
		t.Errorf("Shrunk = %v, want only the row that disappeared", got)
	}
	if got := refusal.Grew(found, found); len(got) != 0 {
		t.Errorf("Grew = %v, want nothing when the baseline is current", got)
	}
	if got := refusal.Shrunk(found, found); len(got) != 0 {
		t.Errorf("Shrunk = %v, want nothing when the baseline is current", got)
	}
}

// TestReadBaselineReadsTheTwoKinds: the file says what it freezes, and a line it
// cannot read stops the run instead of being skipped -a baseline half read is a
// number nobody can trust.
func TestReadBaselineReadsTheTwoKinds(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "refusals.baseline")
	body := "# why this file exists\n" +
		"sentence\tcheck.go\tthis one names nothing\n" +
		"count\tcheck.go\t5\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("writing the baseline: %v", err)
	}

	baseline, err := refusal.ReadBaseline(path)
	if err != nil {
		t.Fatalf("ReadBaseline: %v", err)
	}
	if !slices.Equal(baseline.Sentences, []string{"check.go\tthis one names nothing"}) {
		t.Errorf("sentences = %q", baseline.Sentences)
	}
	if baseline.Counts["check.go"] != 5 {
		t.Errorf("counts = %v, want five in check.go", baseline.Counts)
	}

	for name, bad := range map[string]string{
		"unknown kind": "typo\tcheck.go\tx\n",
		"two fields":   "sentence\tcheck.go\n",
		"no number":    "count\tcheck.go\tmany\n",
	} {
		broken := filepath.Join(dir, "broken-"+strings.ReplaceAll(name, " ", "-"))
		if err := os.WriteFile(broken, []byte(bad), 0o600); err != nil {
			t.Fatalf("%s: writing: %v", name, err)
		}
		if _, err := refusal.ReadBaseline(broken); err == nil {
			t.Errorf("%s: ReadBaseline accepted %q", name, bad)
		}
	}
}

// TestCountIssuesJudgesTheFourCases: a count the baseline never froze, one that
// grew, one that shrank, and a file that stopped building a refusal from a value.
// Each is a different act for the reader, and the wording lives here so the packages
// that consume the ratchet cannot drift.
func TestCountIssuesJudgesTheFourCases(t *testing.T) {
	t.Parallel()

	found := map[string]int{"new.go": 2, "grew.go": 4, "shrank.go": 1}
	frozen := map[string]int{"grew.go": 3, "shrank.go": 2, "gone.go": 5}

	got := refusal.CountIssues(found, frozen, "refusals.baseline")
	if len(got) != 4 {
		t.Fatalf("issues = %q, want one per case", got)
	}
	for _, want := range []string{
		"new.go has 2 refusal(s) the walk cannot read and no count in refusals.baseline",
		"grew.go has 4 refusal(s) the walk cannot read and refusals.baseline froze 3",
		"shrank.go now has 1 refusal(s) the walk cannot read; lower refusals.baseline",
		"gone.go now has 0 refusal(s) the walk cannot read; lower refusals.baseline",
	} {
		var seen bool
		for _, line := range got {
			if strings.HasPrefix(line, want) {
				seen = true
			}
		}
		if !seen {
			t.Errorf("issues do not carry %q:\n%s", want, strings.Join(got, "\n"))
		}
	}
	if lines := refusal.CountIssues(map[string]int{"a.go": 3}, map[string]int{"a.go": 3}, "refusals.baseline"); len(lines) != 0 {
		t.Errorf("issues = %q, want none when the count and the baseline agree", lines)
	}
}
