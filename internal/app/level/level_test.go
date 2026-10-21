package level_test

import (
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/biggz/odachi/internal/app/level"
	"github.com/biggz/odachi/internal/domain/profile"
	"github.com/biggz/odachi/internal/domain/taskdoc"
)

// fakeFiles is the disk in memory with an ordered event log, so a test can
// assert that the backup happens before the rewrite and not only after.
type fakeFiles struct {
	files  map[string][]byte
	events []string
}

func newFakeFiles(seed map[string]string) *fakeFiles {
	f := &fakeFiles{files: map[string][]byte{}}
	for path, content := range seed {
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
		return nil, fmt.Errorf("archivo inexistente: %s", path)
	}
	return data, nil
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

func (f *fakeFiles) Backup(path string) (string, error) {
	data, ok := f.files[path]
	if !ok {
		return "", fmt.Errorf("no se puede respaldar %s: no existe", path)
	}
	f.events = append(f.events, "backup "+path)
	backup := path + ".bak-1a2b3c4d"
	f.files[backup] = data
	return backup, nil
}

func (f *fakeFiles) Write(path string, data []byte) error {
	f.events = append(f.events, "write "+path)
	f.files[path] = data
	return nil
}

func (f *fakeFiles) content(t *testing.T, path string) string {
	t.Helper()
	data, ok := f.files[path]
	if !ok {
		t.Fatalf("%q no existe; archivos: %v", path, keys(f.files))
	}
	return string(data)
}

func keys(files map[string][]byte) []string {
	out := make([]string, 0, len(files))
	for path := range files {
		out = append(out, path)
	}
	slices.Sort(out)
	return out
}

// profilePath is where the resolved profile lives.
var profilePath = filepath.Join("raiz", ".odachi", "profile.toml")

// taskPath is odd/tasks/<NNNN>-<slug>.md inside the fake project.
func taskPath(number int, slug string) string {
	return filepath.Join("raiz", "odd", "tasks", fmt.Sprintf("%04d-%s.md", number, slug))
}

// feature renders a canonical task document with the given steps and history.
func feature(outcome string, doneSteps, totalSteps int, history ...string) string {
	doc := taskdoc.TaskDoc{
		Title:        "feature",
		Requirements: []taskdoc.Requirement{{ID: "R1", Text: outcome}},
		History:      history,
	}
	for i := 1; i <= totalSteps; i++ {
		doc.Steps = append(doc.Steps, taskdoc.Step{
			ID:     "T" + strconv.Itoa(i),
			Text:   "paso " + strconv.Itoa(i),
			Done:   i <= doneSteps,
			Commit: "4a1c9f",
		})
	}
	return doc.Render()
}

// builtin resolves one shipped profile.
func builtin(t *testing.T, name string) profile.Profile {
	t.Helper()
	p, err := profile.Resolve(profile.Builtin(), name)
	if err != nil {
		t.Fatalf("Resolve(%q): %v", name, err)
	}
	return p
}

// TestLevelWithoutEvidenceShowsTheFactsAndProposesNothing: a beginner with one
// feature started is not told to graduate.
func TestLevelWithoutEvidenceShowsTheFactsAndProposesNothing(t *testing.T) {
	files := newFakeFiles(map[string]string{
		taskPath(1, "reservar-turnos"): feature("reservar turnos", 1, 2,
			`L1 · "reservar turnos"`,
			`L2 · PREGUNTA P1 · "¿Se pueden cancelar?"`,
			`L3 · VOS P1 · "No todavía"`,
			`L4 · OFICIO · "usé SQLite" (porque no hay dos escritores)`,
		),
	})

	res := level.Run(level.Options{Dir: "raiz", Current: builtin(t, "novato")}, files)

	if res.Code != level.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, level.CodeOK, res.Message)
	}
	if len(files.events) != 0 {
		t.Errorf("level wrote without --to: %v", files.events)
	}

	block := strings.Join(res.Lines, "\n")
	for _, want := range []string{"perfil: novato", "features con pasos: 1", "preguntas contestadas: 1", "decisiones de oficio: 1", "preguntas abiertas: 0"} {
		if !strings.Contains(block, want) {
			t.Errorf("block does not report %q:\n%s", want, block)
		}
	}
	if strings.Contains(block, "--to") {
		t.Errorf("level proposed a graduation with no evidence behind it:\n%s", block)
	}
}

// TestLevelProposesSeniorWhenEarned: three features with a done step and five
// answered vision questions earn the proposal, and the tool says how to accept
// it without accepting it.
func TestLevelProposesSeniorWhenEarned(t *testing.T) {
	history := []string{`L1 · "feature"`}
	for i := 1; i <= 5; i++ {
		id := "P" + strconv.Itoa(i)
		history = append(history,
			fmt.Sprintf(`L%d · PREGUNTA %s · "¿pregunta %d?"`, 2*i, id, i),
			fmt.Sprintf(`L%d · VOS %s · "respuesta %d"`, 2*i+1, id, i),
		)
	}

	files := newFakeFiles(map[string]string{
		taskPath(1, "reservar-turnos"): feature("reservar turnos", 1, 2, history...),
		taskPath(2, "facturar-obras"):  feature("facturar obras", 1, 1, `L1 · "facturar obras"`),
		taskPath(3, "cobrar-senas"):    feature("cobrar senas", 1, 3, `L1 · "cobrar senas"`),
		// A feature with no done step is not counted: work in progress is not
		// evidence of a finished feature.
		taskPath(4, "agenda-del-dia"): feature("agenda del dia", 0, 2, `L1 · "agenda"`),
	})

	res := level.Run(level.Options{Dir: "raiz", Current: builtin(t, "novato")}, files)

	if res.Code != level.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, level.CodeOK, res.Message)
	}

	block := strings.Join(res.Lines, "\n")
	if !strings.Contains(block, "features con pasos: 3") {
		t.Errorf("block does not count three features with a done step:\n%s", block)
	}
	if !strings.Contains(block, "preguntas contestadas: 5") {
		t.Errorf("block does not count five answered questions:\n%s", block)
	}
	if !strings.Contains(block, "odachi level --to senior") {
		t.Errorf("block does not say how to accept the proposal:\n%s", block)
	}
	if len(files.events) != 0 {
		t.Errorf("level wrote the profile on its own: %v", files.events)
	}
}

// TestLevelForASeniorProposesNothing: the graduation is offered once, to a
// novato. A senior is not told that something is still missing.
func TestLevelForASeniorProposesNothing(t *testing.T) {
	files := newFakeFiles(map[string]string{
		taskPath(1, "reservar-turnos"): feature("reservar turnos", 1, 1, `L1 · "reservar turnos"`),
	})

	res := level.Run(level.Options{Dir: "raiz", Current: builtin(t, "senior")}, files)

	if res.Code != level.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, level.CodeOK, res.Message)
	}
	block := strings.Join(res.Lines, "\n")
	if strings.Contains(block, "--to") {
		t.Errorf("level proposed a profile to someone who is not a novato:\n%s", block)
	}
	if strings.Contains(block, "todavía no está ganado") {
		t.Errorf("level told a senior that senior is not earned yet:\n%s", block)
	}
}

// TestLevelToBacksUpBeforeWriting is the whole contract of `level --to`: the
// human's file is copied aside, then replaced, and the new file keeps all eight
// fields.
func TestLevelToBacksUpBeforeWriting(t *testing.T) {
	before := "name         = \"novato\"\nextends      = \"\"\nbrief_source = \"interview\"\nquestions    = \"one-by-one\"\nassumptions  = \"never\"\noutput       = \"line\"\nproof        = \"command\"\nexplain      = \"why\"\n"
	files := newFakeFiles(map[string]string{profilePath: before})

	res := level.Run(level.Options{Dir: "raiz", Current: builtin(t, "novato"), To: "senior"}, files)

	if res.Code != level.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, level.CodeOK, res.Message)
	}

	if len(files.events) != 2 {
		t.Fatalf("events = %v, want exactly a backup and a write", files.events)
	}
	if !strings.HasPrefix(files.events[0], "backup ") {
		t.Errorf("first event = %q, want the backup: the human's file is never the price of the change", files.events[0])
	}
	if files.content(t, profilePath+".bak-1a2b3c4d") != before {
		t.Errorf("the backup does not hold the previous profile")
	}

	after := files.content(t, profilePath)
	lines := strings.Split(strings.TrimSuffix(after, "\n"), "\n")
	if len(lines) != 8 {
		t.Fatalf("the new profile has %d lines, want all eight fields:\n%s", len(lines), after)
	}

	got := map[string]string{}
	for _, line := range lines {
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			t.Fatalf("line %q is not a key = value line", line)
		}
		got[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), `"`)
	}
	want := map[string]string{
		"name":         "senior",
		"extends":      "novato",
		"brief_source": "repo",
		"questions":    "batched",
		"assumptions":  "declare",
		"output":       "detail",
		"proof":        "suite-or-ci",
		"explain":      "none",
	}
	for key, value := range want {
		if got[key] != value {
			t.Errorf("profile.%s = %q, want %q", key, got[key], value)
		}
	}

	block := strings.Join(res.Lines, "\n")
	for _, want := range []string{"questions", "one-by-one", "batched", "proof", "suite-or-ci"} {
		if !strings.Contains(block, want) {
			t.Errorf("the change block does not report %q:\n%s", want, block)
		}
	}
	if res.Message == "" || strings.Contains(res.Message, "\n") {
		t.Errorf("message %q must be one line", res.Message)
	}
}

// TestLevelToWithoutAProfileFileWritesOne: a project that never wrote
// .odachi/profile.toml gets it now, with nothing to back up.
func TestLevelToWithoutAProfileFileWritesOne(t *testing.T) {
	files := newFakeFiles(nil)

	res := level.Run(level.Options{Dir: "raiz", Current: builtin(t, "novato"), To: "senior"}, files)

	if res.Code != level.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, level.CodeOK, res.Message)
	}
	if len(files.events) != 1 || !strings.HasPrefix(files.events[0], "write ") {
		t.Fatalf("events = %v, want one write and no backup", files.events)
	}
	if !strings.Contains(files.content(t, profilePath), `name         = "senior"`) {
		t.Errorf("profile was not written:\n%s", files.content(t, profilePath))
	}
}

// TestLevelToUnknownProfileNamesTheAvailableOnes: a typo is blocked, and the
// message offers what exists instead of a dead end.
func TestLevelToUnknownProfileNamesTheAvailableOnes(t *testing.T) {
	files := newFakeFiles(nil)

	res := level.Run(level.Options{Dir: "raiz", Current: builtin(t, "novato"), To: "sinior"}, files)

	if res.Code != level.CodeBlocked {
		t.Fatalf("Code = %d, want %d", res.Code, level.CodeBlocked)
	}
	if strings.Contains(res.Message, "\n") {
		t.Errorf("message must be one line: %q", res.Message)
	}
	for _, want := range []string{"sinior", "novato", "senior", "ci"} {
		if !strings.Contains(res.Message, want) {
			t.Errorf("message %q does not mention %q", res.Message, want)
		}
	}
	if len(files.events) != 0 {
		t.Errorf("a blocked level wrote %v", files.events)
	}
}

// TestLevelToTheSameProfileWritesNothing: applying what is already in force is
// not an event.
func TestLevelToTheSameProfileWritesNothing(t *testing.T) {
	files := newFakeFiles(map[string]string{profilePath: "name = \"senior\"\n"})

	res := level.Run(level.Options{Dir: "raiz", Current: builtin(t, "senior"), To: "senior"}, files)

	if res.Code != level.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, level.CodeOK, res.Message)
	}
	if len(files.events) != 0 {
		t.Errorf("level rewrote an unchanged profile: %v", files.events)
	}
	if !strings.Contains(res.Message, "senior") {
		t.Errorf("message %q does not say the profile was already senior", res.Message)
	}
}
