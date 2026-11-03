package gitx_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/gitx"
)

// Messages answers the commit messages of a range, so a reader can find the
// trailer a change wrote about itself. The base commit is the boundary, not part
// of the answer.
func TestMessagesAnswersTheCommitsSinceTheBase(t *testing.T) {
	root, base := newRepo(t)
	write(t, filepath.Join(root, "notes.txt"), "segunda línea\n")
	git(t, root, "add", "-A")
	git(t, root, "commit", "-m", "second", "-m", "Odachi-Change: the second thing")

	messages, err := gitx.Messages(root, base)
	if err != nil {
		t.Fatalf("Messages: %v", err)
	}

	joined := strings.Join(messages, "\n")
	for _, want := range []string{"second", "Odachi-Change: the second thing"} {
		if !strings.Contains(joined, want) {
			t.Errorf("the range does not carry %q:\n%s", want, joined)
		}
	}
	if strings.Contains(joined, "base\n") {
		t.Errorf("the range carried the base commit:\n%s", joined)
	}
}
