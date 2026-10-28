package gitx_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/gitx"
)

// TestIgnoredAnswersForAPathThatDoesNotExist is the property the evidence
// depends on: the rule is asked before the file is written, so git has to answer
// for a name that is not there yet.
func TestIgnoredAnswersForAPathThatDoesNotExist(t *testing.T) {
	t.Parallel()
	root, _ := newRepo(t)
	write(t, filepath.Join(root, ".gitignore"), "/proof/\n")

	ignored, err := gitx.Ignored(root, "proof/0000-check.txt")
	if err != nil {
		t.Fatalf("Ignored(proof/0000-check.txt): %v", err)
	}
	if !ignored {
		t.Errorf("Ignored(proof/0000-check.txt) = false, want true under /proof/")
	}

	// The directory itself answers too: `check` asks about proof/ before it
	// knows which proof it is about to write.
	ignored, err = gitx.Ignored(root, "proof")
	if err != nil {
		t.Fatalf("Ignored(proof): %v", err)
	}
	if !ignored {
		t.Errorf("Ignored(proof) = false, want true under /proof/")
	}
}

// A path no rule covers is not ignored, and that is a false and not an error:
// git's exit 1 is an answer.
func TestIgnoredIsFalseWithoutARule(t *testing.T) {
	t.Parallel()
	root, _ := newRepo(t)
	write(t, filepath.Join(root, ".gitignore"), "/proof/\n")

	for _, path := range []string{"proof.txt", "notes.txt", "odd/tasks/0001-x.md"} {
		ignored, err := gitx.Ignored(root, path)
		if err != nil {
			t.Fatalf("Ignored(%s): %v", path, err)
		}
		if ignored {
			t.Errorf("Ignored(%s) = true, want false: no rule covers it", path)
		}
	}
}

// A file inside an ignored directory is ignored, which is what keeps a written
// verdict out of the tree it names.
func TestIgnoredSeesANestedIgnoredFile(t *testing.T) {
	t.Parallel()
	root, _ := newRepo(t)
	write(t, filepath.Join(root, ".gitignore"), "build/\n")
	if err := os.MkdirAll(filepath.Join(root, "build"), 0o755); err != nil {
		t.Fatalf("mkdir build: %v", err)
	}
	write(t, filepath.Join(root, "build", "salida.txt"), "binario\n")

	ignored, err := gitx.Ignored(root, "build/salida.txt")
	if err != nil {
		t.Fatalf("Ignored(build/salida.txt): %v", err)
	}
	if !ignored {
		t.Errorf("Ignored(build/salida.txt) = false, want true under build/")
	}
}

// Outside a repository git cannot answer, and a silent false there would be the
// worst answer: the caller would write evidence into a tree nothing describes.
func TestIgnoredOutsideARepositoryIsAnError(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	ignored, err := gitx.Ignored(root, "proof/0000-check.txt")

	if err == nil {
		t.Fatalf("Ignored outside a repository = (%v, nil), want an error", ignored)
	}
	if ignored {
		t.Errorf("Ignored outside a repository = true, want false with the error")
	}
	if !strings.Contains(err.Error(), "not a git repository") && !strings.Contains(err.Error(), "check-ignore") {
		t.Errorf("error %q does not carry git's own explanation", err)
	}
}
