package ask_test

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/ask"
	"github.com/biggs-100/odachi/internal/domain/taskdoc"
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

func (f *fakeFiles) Write(path string, data []byte) error {
	f.wrote = append(f.wrote, path)
	f.files[path] = data
	return nil
}

func (f *fakeFiles) content(t *testing.T, path string) string {
	t.Helper()
	data, ok := f.files[path]
	if !ok {
		t.Fatalf("%q does not exist; files: %v", path, f.paths())
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

// taskPath is where the first feature lives once odachi start wrote it.
var taskPath = filepath.Join("raiz", "odd", "tasks", "0001-reservar-turnos.md")

// seedDoc is a task document exactly as `odachi start` renders it, so the tests
// exercise the canonical bytes the tool itself produces.
func seedDoc(outcome string, history ...string) string {
	return taskdoc.TaskDoc{
		Title:        "reservar-turnos",
		Requirements: []taskdoc.Requirement{{ID: "R1", Text: outcome}},
		DoneWhen:     []string{"Reservo un turno y aparece en la agenda (R1)"},
		History:      history,
	}.Render()
}

const outcome = "quiero reservar turnos desde el teléfono"

// project seeds one started feature.
func project(history ...string) *fakeFiles {
	if len(history) == 0 {
		history = []string{`L1 · "` + outcome + `"`}
	}
	return newFakeFiles(map[string]string{taskPath: seedDoc(outcome, history...)})
}

// history returns the document's history lines as they now sit on disk.
func historyOf(t *testing.T, files *fakeFiles) []string {
	t.Helper()
	return taskdoc.Parse(files.content(t, taskPath)).History
}

func assertBlocked(t *testing.T, res ask.Result) string {
	t.Helper()
	if res.Code != ask.CodeBlocked {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, ask.CodeBlocked, res.Message)
	}
	if strings.TrimSpace(res.Message) == "" {
		t.Error("a blocked result must say why")
	}
	if strings.Contains(res.Message, "\n") {
		t.Errorf("the message must be one line: %q", res.Message)
	}
	return res.Message
}

// TestAskWithoutATaskDocumentNamesStart: no log to append to yet, and the
// message says how to create one.
func TestAskWithoutATaskDocumentNamesStart(t *testing.T) {
	files := newFakeFiles(nil)

	res := ask.Run(ask.Options{Dir: "raiz", Question: "¿Se pueden cancelar los turnos?"}, files)

	message := assertBlocked(t, res)
	if !strings.Contains(message, "odachi start") {
		t.Errorf("message %q does not name odachi start", message)
	}
	if len(files.wrote) != 0 {
		t.Errorf("a blocked ask wrote %v", files.wrote)
	}
}

// TestAskRecordsAQuestion is the shape the agent writes when it finds something
// only the human can decide.
func TestAskRecordsAQuestion(t *testing.T) {
	files := project()

	res := ask.Run(ask.Options{Dir: "raiz", Question: "  ¿Los turnos se pueden cancelar?  "}, files)

	if res.Code != ask.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, ask.CodeOK, res.Message)
	}

	want := `L2 · PREGUNTA P1 · "¿Los turnos se pueden cancelar?"`
	history := historyOf(t, files)
	if got := history[len(history)-1]; got != want {
		t.Fatalf("appended line = %q, want %q", got, want)
	}
	if history[0] != `L1 · "`+outcome+`"` {
		t.Errorf("the question was written over the first line: %v", history)
	}
	if !strings.Contains(res.Message, "P1") || !strings.Contains(res.Message, "--for P1") {
		t.Errorf("message %q does not tell the human how to answer P1", res.Message)
	}
	if res.Task != "odd/tasks/0001-reservar-turnos.md" {
		t.Errorf("Task = %q, want the task document relative to the project", res.Task)
	}
}

// TestAskNumbersQuestionsMonotonically: P numbers are per task document, so a
// second question is P2 even when the first one is still open.
func TestAskNumbersQuestionsMonotonically(t *testing.T) {
	files := project()

	for _, question := range []string{"¿Se pueden cancelar?", "¿Hace falta WhatsApp?"} {
		if res := ask.Run(ask.Options{Dir: "raiz", Question: question}, files); res.Code != ask.CodeOK {
			t.Fatalf("ask %q: code %d (message: %s)", question, res.Code, res.Message)
		}
	}

	history := historyOf(t, files)
	want := []string{
		`L1 · "` + outcome + `"`,
		`L2 · PREGUNTA P1 · "¿Se pueden cancelar?"`,
		`L3 · PREGUNTA P2 · "¿Hace falta WhatsApp?"`,
	}
	if !slices.Equal(history, want) {
		t.Fatalf("history =\n%v\nwant\n%v", history, want)
	}
}

// TestAskRecordsACraftDecision: the agent deciding alone, saying why. This is
// the tutor material, and it is never a question.
func TestAskRecordsACraftDecision(t *testing.T) {
	files := project()

	res := ask.Run(ask.Options{
		Dir:     "raiz",
		Craft:   "Guardé los turnos en SQLite",
		Because: "dos personas no pueden escribir el mismo archivo",
	}, files)

	if res.Code != ask.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, ask.CodeOK, res.Message)
	}

	want := `L2 · OFICIO · "Guardé los turnos en SQLite" (porque dos personas no pueden escribir el mismo archivo)`
	history := historyOf(t, files)
	if got := history[len(history)-1]; got != want {
		t.Fatalf("appended line = %q, want %q", got, want)
	}
	if len(taskdoc.Parse(files.content(t, taskPath)).OpenQuestions()) != 0 {
		t.Error("a craft decision was recorded as a question")
	}
}

// TestCraftWithoutBecauseIsBlocked: a decision with no reason teaches nothing,
// so the tool refuses it and names the flag.
func TestCraftWithoutBecauseIsBlocked(t *testing.T) {
	files := project()

	res := ask.Run(ask.Options{Dir: "raiz", Craft: "Guardé los turnos en SQLite"}, files)

	message := assertBlocked(t, res)
	if !strings.Contains(message, "--because") {
		t.Errorf("message %q does not name --because", message)
	}
	if len(files.wrote) != 0 {
		t.Errorf("a blocked craft decision wrote %v", files.wrote)
	}
}

// TestBecauseWithoutCraftIsBlocked: the reverse mistake, named the same way.
func TestBecauseWithoutCraftIsBlocked(t *testing.T) {
	files := project()

	message := assertBlocked(t, ask.Run(ask.Options{Dir: "raiz", Because: "porque sí"}, files))

	if !strings.Contains(message, "--craft") {
		t.Errorf("message %q does not name --craft", message)
	}
	if len(files.wrote) != 0 {
		t.Errorf("a blocked run wrote %v", files.wrote)
	}
}

// TestAnswerClosesTheOnlyOpenQuestion: with one open question, --for is
// unnecessary.
func TestAnswerClosesTheOnlyOpenQuestion(t *testing.T) {
	files := project(
		`L1 · "`+outcome+`"`,
		`L2 · PREGUNTA P1 · "¿Se pueden cancelar los turnos?"`,
	)

	res := ask.Run(ask.Options{Dir: "raiz", Answer: "No en la primera versión"}, files)

	if res.Code != ask.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, ask.CodeOK, res.Message)
	}

	want := `L3 · VOS P1 · "No en la primera versión"`
	history := historyOf(t, files)
	if got := history[len(history)-1]; got != want {
		t.Fatalf("appended line = %q, want %q", got, want)
	}
	doc := taskdoc.Parse(files.content(t, taskPath))
	if len(doc.OpenQuestions()) != 0 {
		t.Errorf("OpenQuestions = %+v, want none", doc.OpenQuestions())
	}
}

// TestTwoOpenQuestionsNeedFor: the tool never guesses which question the human
// just answered, and the message offers the choice.
func TestTwoOpenQuestionsNeedFor(t *testing.T) {
	files := project(
		`L1 · "`+outcome+`"`,
		`L2 · PREGUNTA P1 · "¿Se pueden cancelar los turnos?"`,
		`L3 · PREGUNTA P2 · "¿Hace falta WhatsApp?"`,
	)

	res := ask.Run(ask.Options{Dir: "raiz", Answer: "Sí"}, files)

	message := assertBlocked(t, res)
	if !strings.Contains(message, "--for P1") {
		t.Errorf("message %q does not name --for P1", message)
	}
	if !strings.Contains(message, "P2") {
		t.Errorf("message %q does not name the other open question P2", message)
	}
	if len(files.wrote) != 0 {
		t.Errorf("a blocked answer wrote %v", files.wrote)
	}
}

// TestAnswerForClosesTheRightOne: --for picks, and the other question stays
// open and openable later.
func TestAnswerForClosesTheRightOne(t *testing.T) {
	files := project(
		`L1 · "`+outcome+`"`,
		`L2 · PREGUNTA P1 · "¿Se pueden cancelar los turnos?"`,
		`L3 · PREGUNTA P2 · "¿Hace falta WhatsApp?"`,
	)

	res := ask.Run(ask.Options{Dir: "raiz", Answer: "Todavía no", For: "P2"}, files)

	if res.Code != ask.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, ask.CodeOK, res.Message)
	}

	want := `L4 · VOS P2 · "Todavía no"`
	history := historyOf(t, files)
	if got := history[len(history)-1]; got != want {
		t.Fatalf("appended line = %q, want %q", got, want)
	}

	open := taskdoc.Parse(files.content(t, taskPath)).OpenQuestions()
	if len(open) != 1 || open[0].ID != "P1" {
		t.Fatalf("OpenQuestions = %+v, want only P1", open)
	}
}

// TestAnswerForAnUnknownQuestionIsBlocked keeps --for honest.
func TestAnswerForAnUnknownQuestionIsBlocked(t *testing.T) {
	files := project(
		`L1 · "`+outcome+`"`,
		`L2 · PREGUNTA P1 · "¿Se pueden cancelar los turnos?"`,
	)

	message := assertBlocked(t, ask.Run(ask.Options{Dir: "raiz", Answer: "Sí", For: "P9"}, files))

	if !strings.Contains(message, "P9") {
		t.Errorf("message %q does not name P9", message)
	}
	if len(files.wrote) != 0 {
		t.Errorf("a blocked answer wrote %v", files.wrote)
	}
}

// TestAnswerWithNoOpenQuestionIsBlocked: nothing to answer, and the message
// names the command that opens one.
func TestAnswerWithNoOpenQuestionIsBlocked(t *testing.T) {
	files := project()

	message := assertBlocked(t, ask.Run(ask.Options{Dir: "raiz", Answer: "Sí"}, files))

	if !strings.Contains(message, "odachi ask") {
		t.Errorf("message %q does not name odachi ask", message)
	}
}

// TestForWithoutAnswerIsBlocked: --for alone names nothing to answer.
func TestForWithoutAnswerIsBlocked(t *testing.T) {
	files := project(
		`L1 · "`+outcome+`"`,
		`L2 · PREGUNTA P1 · "¿Se pueden cancelar los turnos?"`,
	)

	message := assertBlocked(t, ask.Run(ask.Options{Dir: "raiz", For: "P1"}, files))

	if !strings.Contains(message, "--answer") {
		t.Errorf("message %q does not name --answer", message)
	}
}

// TestBareAskNeedsAHuman: nothing to record is a missing human answer, not a
// broken command.
func TestBareAskNeedsAHuman(t *testing.T) {
	res := ask.Run(ask.Options{Dir: "raiz"}, project())

	if res.Code != ask.CodeNeedsAnswer {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, ask.CodeNeedsAnswer, res.Message)
	}
	if !strings.Contains(res.Message, `odachi ask "<question>"`) {
		t.Errorf("message %q does not show the shape of the command", res.Message)
	}
}

// TestMixedFormsAreBlocked: one call records one thing.
func TestMixedFormsAreBlocked(t *testing.T) {
	files := project()

	res := ask.Run(ask.Options{Dir: "raiz", Question: "¿Y esto?", Answer: "Sí", Craft: "x", Because: "y"}, files)

	if message := assertBlocked(t, res); !strings.Contains(message, "one form") {
		t.Errorf("message %q does not explain that one call records one thing", message)
	}
	if len(files.wrote) != 0 {
		t.Errorf("a blocked run wrote %v", files.wrote)
	}
}

// TestADocumentOutsideTheCanonicalShapeIsNotRewritten: the history is the
// human's memory, so the tool refuses to rewrite a document whose content it
// would drop.
func TestADocumentOutsideTheCanonicalShapeIsNotRewritten(t *testing.T) {
	handwritten := "# reservar-turnos\n\n## Requisitos (textual, del usuario)\nR1 · \"reservar turnos\"\n\n## Notas\nuna idea que el humano agregó a mano\n\n## Historial\nL1 · \"reservar turnos\"\n"
	files := newFakeFiles(map[string]string{taskPath: handwritten})

	res := ask.Run(ask.Options{Dir: "raiz", Question: "¿Se pueden cancelar?"}, files)

	message := assertBlocked(t, res)
	if !strings.Contains(message, "## History") {
		t.Errorf("message %q does not say where the notes belong", message)
	}
	if got := files.content(t, taskPath); got != handwritten {
		t.Errorf("the document was rewritten:\n%s", got)
	}
}
