package tasks_test

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/tasks"
)

// fakeFiles is the disk in memory: it lists directories and reads files, which
// is all this package needs.
type fakeFiles struct{ files map[string][]byte }

func newFakeFiles(seed map[string]string) *fakeFiles {
	f := &fakeFiles{files: map[string][]byte{}}
	for path, content := range seed {
		f.files[path] = []byte(content)
	}
	return f
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

func (f *fakeFiles) Read(path string) ([]byte, error) {
	data, ok := f.files[path]
	if !ok {
		return nil, fmt.Errorf("missing file: %s", path)
	}
	return data, nil
}

// doc renders a minimal canonical task document for the given outcome.
func doc(title, outcome string) string {
	return "# " + title + "\n\n## Requisitos (textual, del usuario)\nR1 · \"" + outcome + "\"\n\n## Historial\nL1 · \"" + outcome + "\"\n"
}

// TestCurrentIsTheHighestNumber: there is one task document in play, always the
// last feature the human opened.
// TestSkippedNamesTheDocumentsOdachiDoesNotRead: the folder may hold another
// shape - ODD names a feature without a number - and the tool says which files
// it cannot read instead of dropping them silently.
func TestSkippedNamesTheDocumentsOdachiDoesNotRead(t *testing.T) {
	files := newFakeFiles(map[string]string{
		filepath.Join("raiz", "odd", "tasks", "0007-turnos.md"):        "",
		filepath.Join("raiz", "odd", "tasks", "qonta-einvoicing.md"):   "",
		filepath.Join("raiz", "odd", "tasks", "qonta-pos.md.bak-2026"): "",
		filepath.Join("raiz", "odd", "tasks", "notas.txt"):             "",
	})

	skipped, err := tasks.Skipped(files, "raiz")
	if err != nil {
		t.Fatalf("Skipped: %v", err)
	}
	if !slices.Equal(skipped, []string{"qonta-einvoicing.md"}) {
		t.Errorf("Skipped = %v, want only the .md that is not numbered", skipped)
	}
	if line := tasks.SkippedLine(skipped); !strings.Contains(line, "qonta-einvoicing.md") || !strings.Contains(line, "1 document") {
		t.Errorf("SkippedLine = %q, want the count and the name", line)
	}
	if line := tasks.SkippedLine(nil); line != "" {
		t.Errorf("SkippedLine(nil) = %q, want silence", line)
	}
}

func TestCurrentIsTheHighestNumber(t *testing.T) {
	files := newFakeFiles(map[string]string{
		filepath.Join("raiz", "odd", "tasks", "0001-reservar-turnos.md"): doc("reservar-turnos", "reservar turnos"),
		filepath.Join("raiz", "odd", "tasks", "0002-facturar-obras.md"):  doc("facturar-obras", "facturar a las obras sociales"),
		filepath.Join("raiz", "odd", "tasks", "0007-cobrar-senas.md"):    doc("cobrar-senas", "cobrar señas"),
		// Housekeeping that must never be mistaken for a feature.
		filepath.Join("raiz", "odd", "tasks", "0008-viejo.md.bak-1a2b3c4d"): "respaldo",
		filepath.Join("raiz", "odd", "tasks", "notas.md"):                   "una nota suelta",
	})

	current, err := tasks.Current(files, "raiz")
	if err != nil {
		t.Fatalf("Current: %v", err)
	}

	if current.Number != 7 {
		t.Errorf("Number = %d, want 7", current.Number)
	}
	if current.Slug != "cobrar-senas" {
		t.Errorf("Slug = %q, want cobrar-senas", current.Slug)
	}
	if current.Rel != "odd/tasks/0007-cobrar-senas.md" {
		t.Errorf("Rel = %q, want odd/tasks/0007-cobrar-senas.md", current.Rel)
	}
	if current.Doc.Requirements[0].Text != "cobrar señas" {
		t.Errorf("the document was not parsed: %+v", current.Doc)
	}
	if !strings.HasSuffix(filepath.ToSlash(current.Path), "odd/tasks/0007-cobrar-senas.md") {
		t.Errorf("Path = %q, want the file on disk", current.Path)
	}
}

// TestCurrentWithoutADocumentNamesStart: a project with no feature yet is not a
// crash, and the message names the command that creates one.
func TestCurrentWithoutADocumentNamesStart(t *testing.T) {
	_, err := tasks.Current(newFakeFiles(nil), "raiz")
	if err == nil {
		t.Fatal("Current on an empty project = nil error, want one")
	}
	if !strings.Contains(err.Error(), "odachi start") {
		t.Errorf("error %q does not name odachi start", err)
	}
	if !strings.Contains(err.Error(), "odd/tasks") {
		t.Errorf("error %q does not name the directory it looked in", err)
	}
	if strings.Contains(err.Error(), "\n") {
		t.Errorf("error must be one line: %q", err)
	}
}

// TestAllIsSortedByNumber: summary reports every feature in the order the human
// opened them.
func TestAllIsSortedByNumber(t *testing.T) {
	files := newFakeFiles(map[string]string{
		filepath.Join("raiz", "odd", "tasks", "0002-facturar-obras.md"):  doc("facturar-obras", "facturar"),
		filepath.Join("raiz", "odd", "tasks", "0001-reservar-turnos.md"): doc("reservar-turnos", "reservar"),
		filepath.Join("raiz", "odd", "tasks", "0010-cobrar-senas.md"):    doc("cobrar-senas", "cobrar"),
	})

	all, err := tasks.All(files, "raiz")
	if err != nil {
		t.Fatalf("All: %v", err)
	}

	got := make([]int, 0, len(all))
	for _, d := range all {
		got = append(got, d.Number)
	}
	if want := []int{1, 2, 10}; !slices.Equal(got, want) {
		t.Fatalf("numbers = %v, want %v", got, want)
	}
}

// TestAllOnAnEmptyProjectIsEmpty: no feature is not a failure to report, it is
// a fact the caller narrates.
func TestAllOnAnEmptyProjectIsEmpty(t *testing.T) {
	all, err := tasks.All(newFakeFiles(nil), "raiz")
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("All = %+v, want nothing", all)
	}
}
