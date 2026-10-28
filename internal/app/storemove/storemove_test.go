package storemove_test

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/app/storemove"
)

const (
	dir   = "/proyecto"
	store = "/proyecto/.git/odachi"
)

// fakeFiles is the disk in memory: every path the migration touches, and the
// ledger a test reads to prove what it did or refused to do.
type fakeFiles struct {
	files map[string][]byte
	wrote []string
}

func newFakeFiles(seed map[string]string) *fakeFiles {
	f := &fakeFiles{files: map[string][]byte{}}
	for path, content := range seed {
		f.files[filepath.FromSlash(path)] = []byte(content)
	}
	return f
}

func (f *fakeFiles) Exists(path string) (bool, error) {
	if _, ok := f.files[path]; ok {
		return true, nil
	}
	// A directory is not a key: it exists when something lives under it.
	prefix := path + string(filepath.Separator)
	for key := range f.files {
		if strings.HasPrefix(key, prefix) {
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeFiles) Read(path string) ([]byte, error) {
	data, ok := f.files[path]
	if !ok {
		return nil, errors.New("missing file: " + path)
	}
	return data, nil
}

func (f *fakeFiles) Write(path string, data []byte) error {
	f.wrote = append(f.wrote, path)
	f.files[path] = append([]byte(nil), data...)
	return nil
}

func (f *fakeFiles) Remove(path string) error {
	if _, ok := f.files[path]; !ok {
		return fmt.Errorf("missing file: %s", path)
	}
	delete(f.files, path)
	return nil
}

func (f *fakeFiles) Rename(from, to string) error {
	data, ok := f.files[from]
	if !ok {
		return fmt.Errorf("missing file: %s", from)
	}
	if _, exists := f.files[to]; exists {
		return fmt.Errorf("already exists: %s", to)
	}
	f.files[to] = data
	delete(f.files, from)
	return nil
}

func (f *fakeFiles) List(listDir string) ([]string, error) {
	var names []string
	for path := range f.files {
		if filepath.Dir(path) == listDir {
			names = append(names, filepath.Base(path))
		}
	}
	slices.Sort(names)
	return names, nil
}

func (f *fakeFiles) Backup(path string) (string, error) {
	return "", errors.New("the migration never backs up: it moves or copies")
}

func (f *fakeFiles) ModTime(path string) (time.Time, error) { return time.Time{}, nil }

func at(path string) string { return filepath.FromSlash(path) }

func (f *fakeFiles) keys() []string {
	out := make([]string, 0, len(f.files))
	for key := range f.files {
		out = append(out, key)
	}
	slices.Sort(out)
	return out
}

// TestStoreMoveAdoptsTheVerdictsAndPrintsTheGitMoves is the whole migration in
// one run: the evidence lands in the clone's store byte for byte, and the rules
// are named for git to move because the tool never touches a tracked file.
func TestStoreMoveAdoptsTheVerdictsAndPrintsTheGitMoves(t *testing.T) {
	const verdict = `{"tree":"abc","findings":[]}`
	files := newFakeFiles(map[string]string{
		dir + "/.odachi/reviews/abc.json":   verdict,
		dir + "/.odachi/fence.toml":         "[[rule]]\nnot = \"pagos\"\n",
		dir + "/.odachi/roles/marketing.md": "---\nname: marketing\n---\n",
		dir + "/.odachi/agents.toml":        "agents = \"opencode\"\n",
		dir + "/odd/profile.toml":           "name = \"novato\"\n",
	})

	result := storemove.Run(opts(), files, nil)

	if result.Code != storemove.CodeOK {
		t.Fatalf("Code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if got := string(files.files[at(store+"/reviews/abc.json")]); got != verdict {
		t.Errorf("the verdict was not adopted byte for byte: %q", got)
	}
	if _, still := files.files[at(dir+"/.odachi/reviews/abc.json")]; still {
		t.Error("the old verdict is still there: it was a move, not a copy")
	}
	if _, present := files.files[at(dir+"/.odachi/agents.toml")]; !present {
		t.Errorf("the client choice is tracked: the tool must copy it and let git remove it; files: %v", files.keys())
	}
	if got := string(files.files[at(store+"/agents.toml")]); !strings.Contains(got, "opencode") {
		t.Errorf("the client choice was not copied into the store: %q", got)
	}

	joined := strings.Join(result.Lines, "\n")
	for _, want := range []string{
		"git mv .odachi/fence.toml odd/fence.toml",
		"git mv .odachi/roles odd/roles",
		"git rm --cached .odachi/agents.toml",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("the output does not carry %q:\n%s", want, joined)
		}
	}
	if strings.Contains(joined, "git mv .odachi/profile.toml") {
		t.Errorf("the profile is already in odd/ and the output still moves it:\n%s", joined)
	}
}

// TestStoreMoveRepeatsWithoutChange: a second run has nothing to adopt and says
// so, which is what makes the migration safe to run again.
func TestStoreMoveRepeatsWithoutChange(t *testing.T) {
	files := newFakeFiles(map[string]string{
		dir + "/.odachi/reviews/abc.json": `{"tree":"abc"}`,
	})

	if first := storemove.Run(opts(), files, nil); first.Code != storemove.CodeOK {
		t.Fatalf("first run: Code = %d (%s)", first.Code, first.Message)
	}
	second := storemove.Run(opts(), files, nil)

	if second.Code != storemove.CodeOK {
		t.Fatalf("second run: Code = %d (%s)", second.Code, second.Message)
	}
	if !strings.Contains(second.Message, "nothing to move") && !strings.Contains(second.Message, "already in the store") {
		t.Errorf("second run = %q, want a no-op", second.Message)
	}
}

// TestStoreMoveNeverOverwritesADifferentVerdict: two bytes for one tree is not a
// merge, it is a conflict the human has to see.
func TestStoreMoveNeverOverwritesADifferentVerdict(t *testing.T) {
	files := newFakeFiles(map[string]string{
		dir + "/.odachi/reviews/abc.json": `{"tree":"viejo"}`,
		store + "/reviews/abc.json":       `{"tree":"otro"}`,
	})

	result := storemove.Run(opts(), files, nil)

	if !strings.Contains(result.Message, "already holds other bytes") {
		t.Errorf("Message = %q, want it to name the conflict", result.Message)
	}
	if got := string(files.files[at(store+"/reviews/abc.json")]); !strings.Contains(got, "otro") {
		t.Errorf("the stored verdict was overwritten: %q", got)
	}
	if _, still := files.files[at(dir+"/.odachi/reviews/abc.json")]; !still {
		t.Error("the old verdict was removed after a refusal")
	}
}

// TestADryRunTouchesNothing: the human can read the plan before letting it happen.
func TestADryRunTouchesNothing(t *testing.T) {
	files := newFakeFiles(map[string]string{
		dir + "/.odachi/reviews/abc.json": `{"tree":"abc"}`,
	})
	before := len(files.files)

	dry := storemove.Options{Dir: dir, Store: store, DryRun: true}
	result := storemove.Run(dry, files, nil)

	if result.Code != storemove.CodeOK {
		t.Fatalf("Code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if len(files.files) != before {
		t.Errorf("the dry run touched the disk: %d files -> %d", before, len(files.files))
	}
	if len(files.wrote) != 0 {
		t.Errorf("the dry run wrote %v", files.wrote)
	}
	if !strings.Contains(strings.Join(result.Lines, "\n"), ".odachi/reviews/abc.json") {
		t.Errorf("the dry run does not name what it would move: %v", result.Lines)
	}
}

// opts is the ordinary invocation: the project and its store.
func opts() storemove.Options { return storemove.Options{Dir: dir, Store: store} }
