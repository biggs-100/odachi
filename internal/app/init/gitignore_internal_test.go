package init

import (
	"testing"
)

// block is the canonical block, spelled once so every expectation below is the
// same lines a human will read in their .gitignore.
const block = "# odachi:begin\n/proof/\n*.bak-*\n# odachi:end\n"

// TestIgnoreBlockAloneInAMissingFile: a project with no .gitignore gets the
// block and nothing else.
func TestIgnoreBlockAloneInAMissingFile(t *testing.T) {
	updated, changed := withIgnoreBlock(nil)

	if !changed {
		t.Errorf("changed = false, want true for a project with no .gitignore")
	}
	if string(updated) != block {
		t.Errorf("updated = %q, want %q", updated, block)
	}
}

// TestIgnoreBlockAppendedAfterTheExistingContent: the human's rules stay byte
// for byte, and exactly one blank line separates them from the block.
func TestIgnoreBlockAppendedAfterTheExistingContent(t *testing.T) {
	updated, changed := withIgnoreBlock([]byte("node_modules/\n.env\n"))

	if !changed {
		t.Errorf("changed = false, want true when the block is missing")
	}
	want := "node_modules/\n.env\n\n" + block
	if string(updated) != want {
		t.Errorf("updated = %q, want %q", updated, want)
	}
}

// TestIgnoreBlockAppendedAfterALineWithoutANewline: a last line with no closing
// newline keeps every byte and is not glued to the marker.
func TestIgnoreBlockAppendedAfterALineWithoutANewline(t *testing.T) {
	updated, changed := withIgnoreBlock([]byte("node_modules/"))

	if !changed {
		t.Errorf("changed = false, want true when the block is missing")
	}
	want := "node_modules/\n\n" + block
	if string(updated) != want {
		t.Errorf("updated = %q, want %q", updated, want)
	}
}

// TestIgnoreBlockIsReplacedInPlace: an upgrade refreshes the block where it is,
// so whatever the human wrote above and below survives.
func TestIgnoreBlockIsReplacedInPlace(t *testing.T) {
	current := "node_modules/\n\n# odachi:begin\n/proof/\n# odachi:end\n\ndist/\n"

	updated, changed := withIgnoreBlock([]byte(current))

	if !changed {
		t.Errorf("changed = false, want true when the block is out of date")
	}
	want := "node_modules/\n\n" + block + "\ndist/\n"
	if string(updated) != want {
		t.Errorf("updated = %q, want %q", updated, want)
	}
}

// TestIgnoreBlockIsIdempotent: the second run is a no-op, which is what makes a
// rerun of `odachi init` write nothing at all.
func TestIgnoreBlockIsIdempotent(t *testing.T) {
	once, changed := withIgnoreBlock([]byte("dist/\n"))
	if !changed {
		t.Fatalf("the first run changed nothing")
	}

	twice, changed := withIgnoreBlock(once)

	if changed {
		t.Errorf("the second run changed the file: %q", twice)
	}
	if string(twice) != string(once) {
		t.Errorf("second run = %q, want %q", twice, once)
	}
}
