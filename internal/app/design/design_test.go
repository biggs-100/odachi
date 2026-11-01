package design_test

import (
	"path/filepath"
	"strings"
	"testing"

	appdesign "github.com/biggs-100/odachi/internal/app/design"
)

// fakeFiles is the disk port as data: enough to read the folders and rewrite one
// file, with the writes and backups a test can inspect.
type fakeFiles struct {
	files   map[string]string
	dirs    map[string]bool
	writes  []string
	backups []string
}

func newFakeFiles() *fakeFiles {
	return &fakeFiles{files: map[string]string{}, dirs: map[string]bool{}}
}

func (f *fakeFiles) put(rel, content string) *fakeFiles {
	path := filepath.Join("raiz", filepath.FromSlash(rel))
	f.files[path] = content
	for dir := filepath.Dir(path); len(dir) > len("raiz"); dir = filepath.Dir(dir) {
		f.dirs[dir] = true
	}
	return f
}

func (f *fakeFiles) List(dir string) ([]string, error) {
	names := []string{}
	for path := range f.files {
		if filepath.Dir(path) == dir {
			names = append(names, filepath.Base(path))
		}
	}
	for path := range f.dirs {
		if filepath.Dir(path) == dir {
			names = append(names, filepath.Base(path))
		}
	}
	return names, nil
}

func (f *fakeFiles) Exists(path string) (bool, error) {
	if _, ok := f.files[path]; ok {
		return true, nil
	}
	return f.dirs[path], nil
}

func (f *fakeFiles) Read(path string) ([]byte, error) { return []byte(f.files[path]), nil }

func (f *fakeFiles) Write(path string, data []byte) error {
	f.files[path] = string(data)
	f.writes = append(f.writes, path)
	return nil
}

func (f *fakeFiles) Backup(path string) (string, error) {
	f.backups = append(f.backups, path)
	return path + ".bak", nil
}

const manifest = `{"title":"Pricing page","brief":"b","artboards":[{"file":"01-a.html","title":"A","tradeoff":"dense","width":1440,"height":900}]}`

const comments = `{"version":1,"comments":[{"id":"c1","file":"01-a.html","selector":"h1","excerpt":"x","comment":"shorter","createdAt":"t","status":"open","sentAt":null,"resolvedAt":null}]}`

func TestListReportsTheDesignsAndTheirOpenComments(t *testing.T) {
	files := newFakeFiles().
		put("design/pricing/design.json", manifest).
		put("design/pricing/comments.json", comments)

	got := appdesign.Run(appdesign.Options{Dir: "raiz", Action: appdesign.ActionList}, files)

	if got.Code != appdesign.CodeOK {
		t.Fatalf("Code = %d: %s", got.Code, got.Message)
	}
	if len(got.Designs) != 1 || got.Designs[0].Slug != "pricing" {
		t.Fatalf("designs = %+v", got.Designs)
	}
	joined := strings.Join(got.Lines, "\n")
	for _, want := range []string{"pricing", "Pricing page", "01-a.html", "1 open comment"} {
		if !strings.Contains(joined, want) {
			t.Errorf("the report does not say %q:\n%s", want, joined)
		}
	}
}

// A design folder without a manifest is not a design: it is a folder the list
// skips, and the command does not invent one.
func TestAFolderWithoutAManifestIsNotADesign(t *testing.T) {
	files := newFakeFiles().put("design/wip/01-a.html", "<html></html>")

	got := appdesign.Run(appdesign.Options{Dir: "raiz", Action: appdesign.ActionList}, files)

	if len(got.Designs) != 0 {
		t.Fatalf("a folder with no manifest was reported: %+v", got.Designs)
	}
	if got.Code != appdesign.CodeOK {
		t.Fatalf("Code = %d", got.Code)
	}
}

// Resolving writes the file, keeps the history, and names the ids nobody carries.
func TestResolveWritesTheFileAndNamesTheIdsItDidNotFind(t *testing.T) {
	files := newFakeFiles().
		put("design/pricing/design.json", manifest).
		put("design/pricing/comments.json", comments)

	got := appdesign.Run(appdesign.Options{
		Dir:     "raiz",
		Action:  appdesign.ActionComments,
		Slug:    "pricing",
		Resolve: []string{"c1", "typo"},
		At:      "2026-11-01T00:00:00.000Z",
	}, files)

	if got.Code != appdesign.CodeOK {
		t.Fatalf("Code = %d: %s", got.Code, got.Message)
	}
	written := files.files[filepath.Join("raiz", "design", "pricing", "comments.json")]
	if !strings.Contains(written, `"status": "resolved"`) || !strings.Contains(written, "2026-11-01T00:00:00.000Z") {
		t.Fatalf("the file does not carry the resolution:\n%s", written)
	}
	if !strings.Contains(strings.Join(got.Lines, "\n"), "typo") {
		t.Fatalf("the report does not name the id nobody carries: %v", got.Lines)
	}
	if len(files.backups) != 1 {
		t.Errorf("the file was written without a backup: %v", files.backups)
	}
}

// An agent that names a design that does not exist is told so, by name.
func TestCommentsOnADesignThatDoesNotExistIsBlocked(t *testing.T) {
	files := newFakeFiles()

	got := appdesign.Run(appdesign.Options{Dir: "raiz", Action: appdesign.ActionComments, Slug: "nope"}, files)

	if got.Code != appdesign.CodeBlocked {
		t.Fatalf("Code = %d, want blocked", got.Code)
	}
	if !strings.Contains(got.Message, "nope") {
		t.Errorf("the refusal does not name the design: %q", got.Message)
	}
}

// A resolve that names no comment writes nothing: an agent that mistypes an id reads
// a refusal, not exit 0 over a rewritten file.
func TestResolveWithNoMatchingIDWritesNothing(t *testing.T) {
	files := newFakeFiles().
		put("design/pricing/design.json", manifest).
		put("design/pricing/comments.json", comments)

	got := appdesign.Run(appdesign.Options{
		Dir:     "raiz",
		Action:  appdesign.ActionComments,
		Slug:    "pricing",
		Resolve: []string{"nope"},
		At:      "2026-11-01T00:00:00.000Z",
	}, files)

	if got.Code != appdesign.CodeBlocked {
		t.Fatalf("Code = %d, want blocked", got.Code)
	}
	if len(files.writes) != 0 || len(files.backups) != 0 {
		t.Fatalf("a refusal wrote %v and backed up %v", files.writes, files.backups)
	}
	if !strings.Contains(got.Message, "nope") {
		t.Errorf("the refusal does not name the id: %q", got.Message)
	}
}

// racingFiles hands a different comments file on the read before the write: the
// canvas appending a comment between the command's read and its write.
type racingFiles struct {
	*fakeFiles
	reads int
	next  string
}

func (f *racingFiles) Read(path string) ([]byte, error) {
	if strings.HasSuffix(filepath.ToSlash(path), "comments.json") {
		f.reads++
		if f.reads > 1 {
			return []byte(f.next), nil
		}
	}
	return f.fakeFiles.Read(path)
}

// A comment another writer added between the read and the write is not lost: the
// command notices the bytes moved and redoes the work on the fresh file.
func TestResolveDoesNotSwallowACommentWrittenInBetween(t *testing.T) {
	base := newFakeFiles().
		put("design/pricing/design.json", manifest).
		put("design/pricing/comments.json", comments)
	files := &racingFiles{fakeFiles: base, next: `{"version":1,"comments":[
		{"id":"c1","file":"01-a.html","selector":"h1","excerpt":"x","comment":"shorter","createdAt":"t","status":"open","sentAt":null,"resolvedAt":null},
		{"id":"c2","file":"01-a.html","selector":"h2","excerpt":"y","comment":"from the canvas","createdAt":"t","status":"open","sentAt":null,"resolvedAt":null}
	]}`}

	got := appdesign.Run(appdesign.Options{
		Dir:     "raiz",
		Action:  appdesign.ActionComments,
		Slug:    "pricing",
		Resolve: []string{"c1"},
		At:      "2026-11-01T00:00:00.000Z",
	}, files)

	if got.Code != appdesign.CodeOK {
		t.Fatalf("Code = %d: %s", got.Code, got.Message)
	}
	written := files.files[filepath.Join("raiz", "design", "pricing", "comments.json")]
	if !strings.Contains(written, `"id": "c2"`) {
		t.Fatalf("the comment the canvas wrote was swallowed:\n%s", written)
	}
	if !strings.Contains(written, `"status": "resolved"`) {
		t.Fatalf("the resolve did not land on the fresh file:\n%s", written)
	}
}
