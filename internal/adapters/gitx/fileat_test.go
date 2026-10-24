package gitx_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/gitx"
)

// TestFileAtReadsARevisionAndItsAbsence: the gate reads the base branch's fence
// from a ref. A missing file at a valid rev is "no fence there"; a rev that does
// not exist is an error, because a typo must not turn the gate off.
func TestFileAtReadsARevisionAndItsAbsence(t *testing.T) {
	t.Parallel()
	root, _ := newRepo(t)

	data, found, err := gitx.FileAt(root, "HEAD", "notes.txt")
	if err != nil {
		t.Fatalf("FileAt(HEAD, notes.txt): %v", err)
	}
	if !found || !strings.Contains(string(data), "primera línea") {
		t.Errorf("FileAt = %q, found %v; want notes.txt at HEAD", data, found)
	}

	if _, found, err := gitx.FileAt(root, "HEAD", "no-such-file.txt"); err != nil || found {
		t.Errorf("missing file: found %v, err %v; want false/nil", found, err)
	}

	if _, _, err := gitx.FileAt(root, "no-such-rev", "notes.txt"); err == nil {
		t.Error("a ref that does not exist did not fail")
	}
	if _, _, err := gitx.FileAt(root, "", "notes.txt"); err == nil {
		t.Error("an empty ref did not fail")
	}
}
