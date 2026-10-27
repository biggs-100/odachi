package role_test

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/role"
	domainrole "github.com/biggs-100/odachi/internal/domain/role"
)

// fakeFiles is the disk in memory. Writes are recorded so a test can assert
// that a blocked run touched nothing.
type fakeFiles struct {
	files map[string][]byte
	wrote []string
}

func newFakeFiles(files map[string]string) *fakeFiles {
	f := &fakeFiles{files: map[string][]byte{}}
	for path, content := range files {
		f.files[path] = []byte(content)
	}
	return f
}

func (f *fakeFiles) Exists(path string) (bool, error) {
	_, ok := f.files[path]
	return ok, nil
}

func (f *fakeFiles) Read(path string) ([]byte, error) {
	data, ok := f.files[path]
	if !ok {
		return nil, fmt.Errorf("missing file: %s", path)
	}
	return data, nil
}

func (f *fakeFiles) Write(path string, data []byte) error {
	f.wrote = append(f.wrote, path)
	f.files[path] = data
	return nil
}

func (f *fakeFiles) List(dir string) ([]string, error) {
	var names []string
	for path := range f.files {
		if filepath.Dir(path) == dir {
			names = append(names, filepath.Base(path))
		}
	}
	slices.Sort(names)
	return names, nil
}

// TestNewWritesASkeletonThatPassesItsOwnCheck: the file `role new` leaves is the
// first thing a human edits, so it parses and validates before they touch it.
func TestNewWritesASkeletonThatPassesItsOwnCheck(t *testing.T) {
	files := newFakeFiles(nil)

	res := role.New("raiz", "marketing", files)

	if res.Code != role.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, role.CodeOK, res.Message)
	}
	rel := role.Dir + "/marketing.md"
	path := filepath.Join("raiz", role.Dir, "marketing.md")
	parsed := domainrole.Parse(rel, string(files.files[path]))
	if problems := parsed.Problems("marketing"); len(problems) > 0 {
		t.Fatalf("the skeleton does not pass its own check: %+v", problems)
	}
	for _, section := range domainrole.Sections {
		if strings.TrimSpace(parsed.Bodies[section]) == "" {
			t.Errorf("the skeleton leaves %q empty", section)
		}
	}
	if !strings.Contains(res.Message, "odachi role marketing") {
		t.Errorf("message %q does not name the next command", res.Message)
	}
	if !slices.Equal(res.Files, []string{rel}) {
		t.Errorf("Files = %v, want [%s]", res.Files, rel)
	}
}

// TestNewRefusesABadSlug: the slug is the file, the agent and the skill, so it is
// validated before anything is written.
func TestNewRefusesABadSlug(t *testing.T) {
	files := newFakeFiles(nil)

	res := role.New("raiz", "Marketing Mala", files)

	if res.Code != role.CodeBlocked {
		t.Fatalf("Code = %d, want %d", res.Code, role.CodeBlocked)
	}
	if !strings.Contains(res.Message, domainrole.NamePattern) {
		t.Errorf("message %q does not name the shape a slug must have", res.Message)
	}
	if len(files.wrote) != 0 {
		t.Errorf("a blocked run wrote %v", files.wrote)
	}
}

// TestNewRefusesAnExistingRole: a source is the human's words; overwriting it is
// not a convenience.
func TestNewRefusesAnExistingRole(t *testing.T) {
	path := filepath.Join("raiz", role.Dir, "marketing.md")
	files := newFakeFiles(map[string]string{path: "the human's words"})

	res := role.New("raiz", "marketing", files)

	if res.Code != role.CodeBlocked {
		t.Fatalf("Code = %d, want %d", res.Code, role.CodeBlocked)
	}
	if !strings.Contains(res.Message, "already exists") {
		t.Errorf("message %q does not say the file is already there", res.Message)
	}
	if got := string(files.files[path]); got != "the human's words" {
		t.Errorf("the existing source was rewritten:\n%s", got)
	}
}
