package budget_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/budget"
)

// TestIssuesIsSilentWhenThePinStillDescribesTheText: a pin that matches is a
// question already answered, and a test that speaks there would be noise. The
// ratchet only speaks about a disagreement.
func TestIssuesIsSilentWhenThePinStillDescribesTheText(t *testing.T) {
	t.Parallel()

	got := budget.Issues([]budget.Pin{{Artifact: "block:novato", Bytes: 5529}}, []budget.Pin{{Artifact: "block:novato", Bytes: 5529}})
	if len(got) != 0 {
		t.Errorf("Issues = %q, want silence", got)
	}
}

// TestIssuesNamesBothDirections: growing is a decision somebody has to take, and
// shrinking is one too -a pin that stays above the text stops describing it, and
// the next raise hides inside that gap (0295).
func TestIssuesNamesBothDirections(t *testing.T) {
	t.Parallel()

	grown := budget.Issues([]budget.Pin{{Artifact: "lens:risk", Bytes: 2300}}, []budget.Pin{{Artifact: "lens:risk", Bytes: 2221}})
	if len(grown) != 1 || !strings.Contains(grown[0], "grew by 79 bytes") || !strings.Contains(grown[0], "raise the pin on purpose") {
		t.Errorf("a grown surface = %q, want it named with the difference and the act", grown)
	}

	shrunk := budget.Issues([]budget.Pin{{Artifact: "index", Bytes: 3700}}, []budget.Pin{{Artifact: "index", Bytes: 3800}})
	if len(shrunk) != 1 || !strings.Contains(shrunk[0], "shrank by 100") || !strings.Contains(shrunk[0], "lower the pin") {
		t.Errorf("a shrunk surface = %q, want it named with the difference and the act", shrunk)
	}
}

// TestIssuesNamesASurfaceNobodyPinnedAndAPinNobodyMeasured: a measurement with no
// pin is a surface nobody watches, and a pin with no measurement is a number about
// a text that moved or disappeared. Deleting either side is how the list starts
// describing nothing (0295).
func TestIssuesNamesASurfaceNobodyPinnedAndAPinNobodyMeasured(t *testing.T) {
	t.Parallel()

	got := budget.Issues(
		[]budget.Pin{{Artifact: "block:novato", Bytes: 5529}},
		[]budget.Pin{{Artifact: "lens:readability", Bytes: 2138}},
	)
	if len(got) != 2 {
		t.Fatalf("Issues = %q, want one row per side", got)
	}
	joined := strings.Join(got, "\n")
	if !strings.Contains(joined, "block:novato is 5529 bytes and no pin names it") {
		t.Errorf("the unpinned surface is not named:\n%s", joined)
	}
	if !strings.Contains(joined, "the pin names lens:readability and nothing measured it") {
		t.Errorf("the orphan pin is not named:\n%s", joined)
	}
}

// TestReadRefusesALineThatIsNotTwoFields: a baseline that half-parses is a pin
// nobody can trust, and reading it as empty would make the ratchet pass.
func TestReadRefusesALineThatIsNotTwoFields(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "bytes.baseline")
	if err := os.WriteFile(path, []byte("block:novato\n"), 0o600); err != nil {
		t.Fatalf("writing the fixture: %v", err)
	}
	if _, err := budget.Read(path); err == nil {
		t.Fatal("Read accepted a row with no count")
	}
}

// TestReadIgnoresCommentsAndKeepsTheArtifacts: the file says why it exists, and a
// reader that choked on its own prose would push the explanation out of it.
func TestReadIgnoresCommentsAndKeepsTheArtifacts(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "bytes.baseline")
	text := "# a pin is exact, and both directions fail\n\nlens:risk\t2221\nblock:novato\t5529\n"
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatalf("writing the fixture: %v", err)
	}

	pins, err := budget.Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	want := []budget.Pin{{Artifact: "block:novato", Bytes: 5529}, {Artifact: "lens:risk", Bytes: 2221}}
	if !slices.Equal(pins, want) {
		t.Errorf("Read = %v, want %v", pins, want)
	}
}
