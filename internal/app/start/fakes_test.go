package start_test

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/start"
)

// fakeFiles is the Files port in memory. Every call is recorded in a shared
// event log so a test can assert ordering (backup before write, answers shown
// before writing) and not only final content.
type fakeFiles struct {
	files  map[string][]byte
	events *[]string
}

func newFakeFiles(seed map[string]string, events *[]string) *fakeFiles {
	f := &fakeFiles{files: map[string][]byte{}, events: events}
	for path, content := range seed {
		f.files[f.key(path)] = []byte(content)
	}
	return f
}

// key normalizes a path the way the port documents relative paths: with forward
// slashes on every OS. The real adapter accepts native separators; the double
// models the contract, so a test with a slash path passes on Windows too.
func (f *fakeFiles) key(path string) string { return filepath.ToSlash(path) }

func (f *fakeFiles) record(event string) {
	if f.events != nil {
		*f.events = append(*f.events, event)
	}
}

func (f *fakeFiles) Exists(path string) (bool, error) {
	_, ok := f.files[f.key(path)]
	f.record("exists " + path)
	return ok, nil
}

func (f *fakeFiles) Read(path string) ([]byte, error) {
	data, ok := f.files[f.key(path)]
	if !ok {
		return nil, fmt.Errorf("archivo inexistente: %s", path)
	}
	f.record("read " + path)
	return data, nil
}

func (f *fakeFiles) List(dir string) ([]string, error) {
	var names []string
	for path := range f.files {
		if filepath.ToSlash(filepath.Dir(path)) == f.key(dir) {
			names = append(names, filepath.Base(path))
		}
	}
	slices.Sort(names)
	f.record("list " + dir)
	return names, nil
}

func (f *fakeFiles) Backup(path string) (string, error) {
	if _, ok := f.files[f.key(path)]; !ok {
		return "", fmt.Errorf("no se puede respaldar %s: no existe", path)
	}
	f.record("backup " + path)
	return path + ".bak-prueba", nil
}

func (f *fakeFiles) Write(path string, data []byte) error {
	f.record("write " + path)
	f.files[f.key(path)] = data
	return nil
}

func (f *fakeFiles) content(t *testing.T, path string) string {
	t.Helper()
	data, ok := f.files[f.key(path)]
	if !ok {
		t.Fatalf("el archivo %q no fue escrito; archivos: %v", path, f.paths())
	}
	return string(data)
}

func (f *fakeFiles) paths() []string {
	out := make([]string, 0, len(f.files))
	for path := range f.files {
		out = append(out, path)
	}
	slices.Sort(out)
	return out
}

// scriptedInterview is the Interview test double: it answers from a queue, in
// order, and records every question it was asked and every line it was shown.
type scriptedInterview struct {
	answers   []string
	questions []start.Question
	said      []string
	events    *[]string
}

func newScriptedInterview(events *[]string, answers ...string) *scriptedInterview {
	return &scriptedInterview{answers: answers, events: events}
}

func (s *scriptedInterview) Ask(q start.Question) (string, error) {
	s.questions = append(s.questions, q)
	if s.events != nil {
		*s.events = append(*s.events, "ask")
	}
	if len(s.answers) == 0 {
		return "", fmt.Errorf("entrevista guionada sin respuestas: pregunta %q", q.Text)
	}
	answer := s.answers[0]
	s.answers = s.answers[1:]
	return answer, nil
}

func (s *scriptedInterview) Say(line string) error {
	s.said = append(s.said, line)
	if s.events != nil {
		*s.events = append(*s.events, "say "+line)
	}
	return nil
}

func firstIndex(events []string, prefix string) int {
	for i, event := range events {
		if strings.HasPrefix(event, prefix) {
			return i
		}
	}
	return -1
}

func assertOneLine(t *testing.T, message string) {
	t.Helper()
	if strings.TrimSpace(message) == "" {
		t.Error("mensaje vacío: todo mensaje tiene que decir algo")
	}
	if strings.Contains(message, "\n") {
		t.Errorf("el mensaje debe ser una sola línea, got %q", message)
	}
}
