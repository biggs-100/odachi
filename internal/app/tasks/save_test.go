package tasks_test

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/tasks"
	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// writeFake is the disk in memory with a ledger: a test can assert what was
// written, what was backed up and in which order.
type writeFake struct {
	files   map[string][]byte
	wrote   []string
	backups []string
	order   []string
}

func newWriteFake(seed map[string]string) *writeFake {
	f := &writeFake{files: map[string][]byte{}}
	for path, content := range seed {
		f.files[path] = []byte(content)
	}
	return f
}

func (f *writeFake) List(dir string) ([]string, error) {
	var names []string
	for path := range f.files {
		if filepath.Dir(path) == dir {
			names = append(names, filepath.Base(path))
		}
	}
	slices.Sort(names)
	return names, nil
}

func (f *writeFake) Read(path string) ([]byte, error) {
	data, ok := f.files[path]
	if !ok {
		return nil, fmt.Errorf("archivo inexistente: %s", path)
	}
	return data, nil
}

func (f *writeFake) Write(path string, data []byte) error {
	f.wrote = append(f.wrote, path)
	f.order = append(f.order, "write "+path)
	f.files[path] = data
	return nil
}

func (f *writeFake) Backup(path string) (string, error) {
	f.backups = append(f.backups, path)
	f.order = append(f.order, "backup "+path)
	return path + ".bak-1a2b3c4d", nil
}

// saveDoc is the canonical document a first feature has after `odachi start`.
const saveRaw = "# reserva-de-turno\n\n" +
	"## Requisitos (textual, del usuario)\nR1 · \"reservar turnos\"\n\n" +
	"## Listo cuando\n- Reservo un turno (R1)\n\n" +
	"## Pasos\n- [x] T1 · horario libre   R1 · 4a1c9f\n\n" +
	"## Historial\nL1 · \"reservar turnos\"\n"

func saveDocument(t *testing.T) tasks.Document {
	t.Helper()
	path := filepath.Join("raiz", "odd", "tasks", "0001-reserva-de-turno.md")
	return tasks.Document{
		Path:   path,
		Rel:    "odd/tasks/0001-reserva-de-turno.md",
		Number: 1,
		Slug:   "reserva-de-turno",
		Raw:    saveRaw,
		Doc:    taskdoc.Parse(saveRaw),
	}
}

// TestSaveWritesTheCanonicalRender: the bytes on disk are exactly what Render
// produces, and an ordinary rewrite needs no backup.
func TestSaveWritesTheCanonicalRender(t *testing.T) {
	files := newWriteFake(map[string]string{filepath.Join("raiz", "odd", "tasks", "0001-reserva-de-turno.md"): saveRaw})
	current := saveDocument(t)
	grown := current.Doc.AppendHistory(`L2 · "una nota"`)

	if err := tasks.Save(files, current, grown); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if got := string(files.files[current.Path]); got != grown.Render() {
		t.Errorf("written bytes =\n%s\nwant\n%s", got, grown.Render())
	}
	if len(files.backups) != 0 {
		t.Errorf("Save made backups %v; only SaveWithBackup does that", files.backups)
	}
}

// TestSaveRefusesToDropALine: the canonical renderer knows only the canonical
// sections, so a document with an extra one is not rewritten.
func TestSaveRefusesToDropALine(t *testing.T) {
	handwritten := saveRaw + "\n## Notas\nuna idea que el humano agregó a mano\n"
	path := filepath.Join("raiz", "odd", "tasks", "0001-reserva-de-turno.md")
	files := newWriteFake(map[string]string{path: handwritten})
	current := tasks.Document{Path: path, Rel: "odd/tasks/0001-reserva-de-turno.md", Raw: handwritten, Doc: taskdoc.Parse(handwritten)}

	err := tasks.Save(files, current, current.Doc.AppendHistory(`L2 · "una nota"`))

	if err == nil {
		t.Fatal("Save on a document with an extra section = nil error, want one")
	}
	if !strings.Contains(err.Error(), "## Historial") {
		t.Errorf("error %q does not say where the notes belong", err)
	}
	if !strings.Contains(err.Error(), "## Notas") {
		t.Errorf("error %q does not name the section that would be dropped", err)
	}
	if len(files.wrote) != 0 {
		t.Errorf("a refused Save wrote %v", files.wrote)
	}
	if got := string(files.files[path]); got != handwritten {
		t.Errorf("the document was rewritten:\n%s", got)
	}
}

// TestSaveWithBackupCopiesAsideBeforeWriting: the human's file is never the
// price of a rewrite.
func TestSaveWithBackupCopiesAsideBeforeWriting(t *testing.T) {
	path := filepath.Join("raiz", "odd", "tasks", "0001-reserva-de-turno.md")
	files := newWriteFake(map[string]string{path: saveRaw})
	current := saveDocument(t)
	grown := current.Doc.AppendHistory(`L2 · "una nota"`)

	if err := tasks.SaveWithBackup(files, current, grown); err != nil {
		t.Fatalf("SaveWithBackup: %v", err)
	}

	if want := []string{"backup " + path, "write " + path}; !slices.Equal(files.order, want) {
		t.Fatalf("order = %v, want %v", files.order, want)
	}
	if got := string(files.files[path]); got != grown.Render() {
		t.Errorf("written bytes =\n%s\nwant\n%s", got, grown.Render())
	}
}

// TestSaveWithBackupGuardsBeforeBackingUp: a refused write leaves no trace, not
// even a backup, because nothing was about to be replaced.
func TestSaveWithBackupGuardsBeforeBackingUp(t *testing.T) {
	handwritten := saveRaw + "\n## Notas\nuna idea que el humano agregó a mano\n"
	path := filepath.Join("raiz", "odd", "tasks", "0001-reserva-de-turno.md")
	files := newWriteFake(map[string]string{path: handwritten})
	current := tasks.Document{Path: path, Rel: "odd/tasks/0001-reserva-de-turno.md", Raw: handwritten, Doc: taskdoc.Parse(handwritten)}

	err := tasks.SaveWithBackup(files, current, current.Doc.AppendHistory(`L2 · "una nota"`))

	if err == nil {
		t.Fatal("SaveWithBackup on a document with an extra section = nil error, want one")
	}
	if len(files.backups) != 0 {
		t.Errorf("a refused SaveWithBackup made backups %v", files.backups)
	}
	if len(files.wrote) != 0 {
		t.Errorf("a refused SaveWithBackup wrote %v", files.wrote)
	}
}

// TestSaveNamesTheFileWhenTheWriteFails: the human needs to know which file the
// disk refused, not just that something went wrong.
func TestSaveNamesTheFileWhenTheWriteFails(t *testing.T) {
	path := filepath.Join("raiz", "odd", "tasks", "0001-reserva-de-turno.md")
	files := &failingWrite{writeFake: newWriteFake(map[string]string{path: saveRaw})}
	current := saveDocument(t)

	err := tasks.Save(files, current, current.Doc.AppendHistory(`L2 · "una nota"`))

	if err == nil {
		t.Fatal("Save with a failing writer = nil error, want one")
	}
	if !strings.Contains(err.Error(), "no pude escribir") || !strings.Contains(err.Error(), current.Rel) {
		t.Errorf("error %q does not name the file it could not write", err)
	}
}

// failingWrite is a disk whose write always fails, to keep the message honest.
type failingWrite struct{ *writeFake }

func (f *failingWrite) Write(string, []byte) error {
	return fmt.Errorf("disco lleno")
}

// TestSaveRefusesADuplicateTitle: Parse keeps the last title, so a second one
// would silently replace the first. The guard treats that as a loss too.
func TestSaveRefusesADuplicateTitle(t *testing.T) {
	handwritten := saveRaw + "\n# otro-titulo\n"
	path := filepath.Join("raiz", "odd", "tasks", "0001-reserva-de-turno.md")
	files := newWriteFake(map[string]string{path: handwritten})
	current := tasks.Document{Path: path, Rel: "odd/tasks/0001-reserva-de-turno.md", Raw: handwritten, Doc: taskdoc.Parse(handwritten)}

	err := tasks.Save(files, current, current.Doc.AppendHistory(`L2 · "una nota"`))

	if err == nil {
		t.Fatal("Save on a document with two titles = nil error, want one")
	}
	if !strings.Contains(err.Error(), "# reserva-de-turno") {
		t.Errorf("error %q does not name the title that would be lost", err)
	}
	if !strings.Contains(err.Error(), "## Historial") {
		t.Errorf("error %q does not say where the notes belong", err)
	}
	if len(files.wrote) != 0 {
		t.Errorf("a refused Save wrote %v", files.wrote)
	}
}
