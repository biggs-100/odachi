package taskdoc_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/biggz/odachi/internal/domain/taskdoc"
)

// example reads one of the canonical artifact examples.
func example(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("..", "..", "..", "testdata", "examples", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func TestParseCanonicalTaskDoc(t *testing.T) {
	got := taskdoc.Parse(example(t, "task.md"))

	if got.Title != "reserva-de-turno" {
		t.Errorf("Title = %q, want %q", got.Title, "reserva-de-turno")
	}

	// The fixture carries three requirements and three steps.
	if len(got.Requirements) != 3 {
		t.Fatalf("len(Requirements) = %d, want 3", len(got.Requirements))
	}
	if len(got.DoneWhen) != 3 {
		t.Errorf("len(DoneWhen) = %d, want 3", len(got.DoneWhen))
	}
	if len(got.Steps) != 3 {
		t.Fatalf("len(Steps) = %d, want 3", len(got.Steps))
	}
	if len(got.History) != 3 {
		t.Errorf("len(History) = %d, want 3", len(got.History))
	}

	wantFirstReq := taskdoc.Requirement{
		ID:   "R1",
		Text: "quiero que el paciente reserve un turno desde el telefono y que no se pisen dos turnos del mismo medico",
	}
	if got.Requirements[0] != wantFirstReq {
		t.Errorf("Requirements[0] = %+v, want %+v", got.Requirements[0], wantFirstReq)
	}
	if got.Requirements[2].ID != "R3" {
		t.Errorf("Requirements[2].ID = %q, want R3", got.Requirements[2].ID)
	}

	wantFirstStep := taskdoc.Step{
		ID:          "T1",
		Text:        "horario libre u ocupado en el dominio, sin base de datos todavía",
		Requirement: "R1",
		Commit:      "4a1c9f",
		Done:        true,
	}
	if got.Steps[0] != wantFirstStep {
		t.Errorf("Steps[0] = %+v, want %+v", got.Steps[0], wantFirstStep)
	}
	if got.Steps[1].Done {
		t.Errorf("Steps[1].Done = true, want false")
	}
	if got.Steps[1].Commit != "" {
		t.Errorf("Steps[1].Commit = %q, want empty", got.Steps[1].Commit)
	}

	next, ok := got.Next()
	if !ok {
		t.Fatalf("Next() reported no step, want T2")
	}
	if next.ID != "T2" {
		t.Errorf("Next().ID = %q, want T2", next.ID)
	}

	if done, total := got.Progress(); done != 1 || total != 3 {
		t.Errorf("Progress() = %d/%d, want 1/3", done, total)
	}
}

func TestNextIsFalseWhenEveryStepIsDone(t *testing.T) {
	md := `# demo

## Pasos
- [x] T1 · único   R1 · abc123
`
	got := taskdoc.Parse(md)
	if _, ok := got.Next(); ok {
		t.Fatalf("Next() reported a step, want none")
	}
	if done, total := got.Progress(); done != 1 || total != 1 {
		t.Errorf("Progress() = %d/%d, want 1/1", done, total)
	}
}

func TestRecordMarksStepDoneWithItsCommit(t *testing.T) {
	original := taskdoc.Parse(example(t, "task.md"))

	recorded, err := original.Record("T2", "abc123")
	if err != nil {
		t.Fatalf("Record(T2) error: %v", err)
	}

	if done, total := recorded.Progress(); done != 2 || total != 3 {
		t.Errorf("recorded Progress() = %d/%d, want 2/3", done, total)
	}
	if !recorded.Steps[1].Done || recorded.Steps[1].Commit != "abc123" {
		t.Errorf("Steps[1] = %+v, want done with commit abc123", recorded.Steps[1])
	}
	if next, ok := recorded.Next(); !ok || next.ID != "T3" {
		t.Errorf("recorded Next() = %+v, %v; want T3", next, ok)
	}

	// The receiver is a value: the original document must be untouched.
	if done, _ := original.Progress(); done != 1 {
		t.Errorf("original Progress() done = %d, want 1", done)
	}
	if original.Steps[1].Done || original.Steps[1].Commit != "" {
		t.Errorf("original Steps[1] mutated: %+v", original.Steps[1])
	}
}

func TestRecordUnknownIDIsAnErrorAndLeavesTheDocUnchanged(t *testing.T) {
	original := taskdoc.Parse(example(t, "task.md"))

	got, err := original.Record("T9", "abc")
	if err == nil {
		t.Fatalf("Record(T9) returned nil error, want an error")
	}
	if !strings.Contains(err.Error(), "T9") {
		t.Errorf("error %q does not name the unknown id", err)
	}
	if !reflect.DeepEqual(got, original) {
		t.Errorf("Record(T9) mutated the document on error")
	}
}

func TestRenderRoundTripsTheExampleDocuments(t *testing.T) {
	for _, name := range []string{"task.md", "brief.md"} {
		t.Run(name, func(t *testing.T) {
			parsed := taskdoc.Parse(example(t, name))
			again := taskdoc.Parse(parsed.Render())
			if !reflect.DeepEqual(again, parsed) {
				t.Errorf("Parse(Render(x)) != x\n--- got ---\n%+v\n--- want ---\n%+v\n--- rendered ---\n%s", again, parsed, parsed.Render())
			}
		})
	}
}

func TestRenderRoundTripsARecordedDocument(t *testing.T) {
	recorded, err := taskdoc.Parse(example(t, "task.md")).Record("T2", "abc123")
	if err != nil {
		t.Fatalf("Record(T2) error: %v", err)
	}
	if again := taskdoc.Parse(recorded.Render()); !reflect.DeepEqual(again, recorded) {
		t.Errorf("Parse(Render(recorded)) != recorded\n--- got ---\n%+v", again)
	}
}

func TestRenderIsStable(t *testing.T) {
	parsed := taskdoc.Parse(example(t, "task.md"))
	if first, second := parsed.Render(), taskdoc.Parse(parsed.Render()).Render(); first != second {
		t.Errorf("Render is not stable:\n--- first ---\n%s\n--- second ---\n%s", first, second)
	}
}

func TestParseMalformedInputNeverPanics(t *testing.T) {
	inputs := map[string]string{
		"empty":              "",
		"blank lines":        "\n\n\n",
		"bullet no section":  "- [x] T1 · huérfano   R1\n",
		"broken checkbox":    "## Pasos\n- [ T1 · roto\n",
		"step no separator":  "## Pasos\n- [ ] T1 sin separador\n",
		"requirement junk":   "## Requisitos (textual, del usuario)\n· \nR1\n",
		"history no id":      "## Historial\nsin identificador\n",
		"header only":        "#\n##\n###\n",
		"no newline at end":  "## Pasos\n- [ ] T1 · único   R1",
		"unknown section":    "## Otra cosa\n- [x] T1 · nada\n",
		"unterminated quote": "## Requisitos (textual, del usuario)\nR1 · \"sin cerrar\n",
	}
	for name, md := range inputs {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Parse panicked: %v", r)
				}
			}()
			// Whatever it returns must render and reparse without loss.
			parsed := taskdoc.Parse(md)
			if again := taskdoc.Parse(parsed.Render()); !reflect.DeepEqual(again, parsed) {
				t.Errorf("Parse(Render(x)) != x for %q\n--- got ---\n%+v\n--- rendered ---\n%s", md, again, parsed.Render())
			}
		})
	}
}

func TestParseIgnoresUnknownSections(t *testing.T) {
	md := `# demo

## Requisitos (textual, del usuario)
R1 · "algo"

## Notas
- esto no es un paso
- [x] T9 · tampoco   R1 · ffffff

## Pasos
- [ ] T1 · uno   R1
`
	got := taskdoc.Parse(md)

	if len(got.Steps) != 1 {
		t.Fatalf("len(Steps) = %d, want 1", len(got.Steps))
	}
	if got.Steps[0].ID != "T1" {
		t.Errorf("Steps[0].ID = %q, want T1", got.Steps[0].ID)
	}
	if len(got.Requirements) != 1 || got.Requirements[0].Text != "algo" {
		t.Errorf("Requirements = %+v, want one requirement with text algo", got.Requirements)
	}
}

func TestParseAcceptsCRLF(t *testing.T) {
	md := "# demo\r\n\r\n## Requisitos (textual, del usuario)\r\nR1 · \"algo\"\r\n\r\n## Pasos\r\n- [x] T1 · uno   R1 · abc123\r\n\r\n## Historial\r\nL1 · \"nota\"\r\n"
	got := taskdoc.Parse(md)

	if len(got.Steps) != 1 {
		t.Fatalf("len(Steps) = %d, want 1", len(got.Steps))
	}
	want := taskdoc.Step{ID: "T1", Text: "uno", Requirement: "R1", Commit: "abc123", Done: true}
	if got.Steps[0] != want {
		t.Errorf("Steps[0] = %+v, want %+v", got.Steps[0], want)
	}
	if len(got.History) != 1 || got.History[0] != `L1 · "nota"` {
		t.Errorf("History = %q, want [L1 · \"nota\"]", got.History)
	}
}
