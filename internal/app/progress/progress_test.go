package progress_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/progress"
)

// dir is the project root the verb sees. Every port is a double: progress is
// read-only, and a test that cannot write is the strongest way to say so.
const dir = "/proyecto"

// fakeFiles is the disk port in memory. It records writes so a test can prove
// there are none.
type fakeFiles struct {
	files  map[string][]byte
	writes []string
}

func newFakeFiles(seed map[string]string) *fakeFiles {
	f := &fakeFiles{files: map[string][]byte{}}
	for path, content := range seed {
		f.files[fullPath(path)] = []byte(content)
	}
	return f
}

func fullPath(rel string) string {
	return filepath.Join(dir, filepath.FromSlash(rel))
}

func (f *fakeFiles) Exists(path string) (bool, error) {
	_, ok := f.files[path]
	return ok, nil
}

func (f *fakeFiles) Read(path string) ([]byte, error) {
	data, ok := f.files[path]
	if !ok {
		return nil, errMissing(path)
	}
	return data, nil
}

func (f *fakeFiles) List(target string) ([]string, error) {
	prefix := strings.TrimSuffix(filepath.ToSlash(target), "/") + "/"
	var names []string
	for path := range f.files {
		slashed := filepath.ToSlash(path)
		if !strings.HasPrefix(slashed, prefix) {
			continue
		}
		rest := strings.TrimPrefix(slashed, prefix)
		if strings.Contains(rest, "/") {
			continue
		}
		names = append(names, rest)
	}
	sortStrings(names)
	return names, nil
}

func (f *fakeFiles) Write(path string, data []byte) error {
	f.writes = append(f.writes, path)
	f.files[path] = data
	return nil
}

// fakeGit answers the one fact progress needs from the repository.
type fakeGit struct {
	tree string
	err  error
}

func (g *fakeGit) TreeHash(root string) (string, error) {
	return g.tree, g.err
}

// missing is the error a read of a file that is not there returns.
type missing string

func (m missing) Error() string { return "missing file: " + string(m) }

func errMissing(path string) error { return missing(path) }

// sortStrings is a tiny insertion sort: the fixture never holds many names and
// the order has to be stable for the report to be comparable.
func sortStrings(names []string) {
	for i := 1; i < len(names); i++ {
		for j := i; j > 0 && names[j] < names[j-1]; j-- {
			names[j], names[j-1] = names[j-1], names[j]
		}
	}
}

// taskDoc is one feature with a done step - with its commit - and a step still
// owed.
const taskDoc = `# reserva

## Requirements (verbatim, from the user)
- R1 · "reservar turnos"

## Done when
- reservo un turno

## Steps
- [x] T1 · el turno se guarda   R1 · abc1234
- [ ] T2 · el turno se puede cancelar   R1

## History
L1 · "reservar turnos"
`

// project is one run's wiring.
type project struct {
	files *fakeFiles
	git   *fakeGit
}

func newProject() *project {
	return &project{
		files: newFakeFiles(map[string]string{"odd/tasks/0007-reserva.md": taskDoc}),
		git:   &fakeGit{tree: "4a1c9f7d2e3b5a6c8d9e0f1a2b3c4d5e6f7a8b9c"},
	}
}

func (p *project) run(opts progress.Options) progress.Result {
	opts.Dir = dir
	return progress.Run(opts, p.files, p.git)
}

// TestProgressNamesTheFeatureAndItsSteps: the report answers one question - where
// does the open feature stand - and it answers it with what the task doc already
// says: the name, how many steps are done, and the commit each one landed in.
func TestProgressNamesTheFeatureAndItsSteps(t *testing.T) {
	p := newProject()

	result := p.run(progress.Options{})

	lines := strings.Join(result.Lines, "\n")
	for _, want := range []string{
		"0007-reserva",
		"1 of 2",
		"abc1234",
		"el turno se guarda",
		"el turno se puede cancelar",
	} {
		if !strings.Contains(lines, want) {
			t.Errorf("the report does not say %q:\n%s", want, lines)
		}
	}
	if result.Code != 0 {
		t.Errorf("code = %d, want 0: a report is not a gate", result.Code)
	}
}
