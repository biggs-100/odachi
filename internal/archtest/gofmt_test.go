package archtest_test

import (
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestEveryGoFileIsGofmtClean: nothing looked at formatting, so a file landed with CRLF in
// the blob and no final newline and gofmt -l . stayed red for fourteen releases. The rule is
// the one gofmt already has, checked in process: what format.Source would write is what the
// file holds. A checkout's line endings are the platform's, and this reads the worktree,
// which is what CI checks out.
func TestEveryGoFileIsGofmtClean(t *testing.T) {
	root := filepath.Join("..", "..")
	var unformatted []string
	checked := 0

	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := entry.Name()
		if entry.IsDir() {
			// The tool's own machinery and a clone's state are not source.
			if name == ".git" || name == "dist" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(name, ".go") {
			return nil
		}

		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		checked++
		if formatted, err := format.Source(source); err == nil && string(formatted) != string(source) {
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				rel = path
			}
			unformatted = append(unformatted, firstDifference(rel, string(source), string(formatted)))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking the tree: %v", err)
	}
	if checked == 0 {
		t.Fatal("no Go file was checked: the walk is looking at the wrong root")
	}
	if len(unformatted) > 0 {
		t.Fatalf("these files are not what gofmt writes:\n%s", strings.Join(unformatted, "\n"))
	}
}

// firstDifference names the file and the first line where the bytes differ, because a
// message that only says "not formatted" sends a person to run a command they already ran.
func firstDifference(rel, have, want string) string {
	haveLines := strings.Split(have, "\n")
	wantLines := strings.Split(want, "\n")
	for i := 0; i < len(haveLines) && i < len(wantLines); i++ {
		if haveLines[i] != wantLines[i] {
			return rel + ":" + itoa(i+1) + ": " + haveLines[i] + " -> " + wantLines[i]
		}
	}
	return rel + ": the file ends differently"
}

// itoa is small on purpose: one call site, no fmt import in a test that is read as a check.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
