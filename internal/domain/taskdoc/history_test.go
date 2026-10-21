package taskdoc_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// withHistory builds a minimal task document whose `## Historial` carries the
// given lines, in order.
func withHistory(lines ...string) taskdoc.TaskDoc {
	return taskdoc.Parse("# reserva-de-turno\n\n## Historial\n" + strings.Join(lines, "\n") + "\n")
}

// TestEventsParseTheThreeRecordShapes pins Contract IV: the history is an
// append-only log with exactly three deterministic line shapes, and everything
// else is a note the human wrote.
func TestEventsParseTheThreeRecordShapes(t *testing.T) {
	doc := withHistory(
		`L7 · PREGUNTA P1 · "¿Los turnos se pueden cancelar?"`,
		`L8 · VOS P1 · "No en la primera versión"`,
		`L9 · OFICIO · "Guardé los turnos en SQLite" (porque dos personas no pueden escribir el mismo archivo)`,
		`L10 · se decide que un turno dura 30 minutos`,
	)

	want := []taskdoc.Event{
		{Line: 7, Kind: taskdoc.KindQuestion, ID: "P1", Text: "¿Los turnos se pueden cancelar?"},
		{Line: 8, Kind: taskdoc.KindAnswer, ID: "P1", Text: "No en la primera versión"},
		{Line: 9, Kind: taskdoc.KindCraft, Text: "Guardé los turnos en SQLite", Because: "dos personas no pueden escribir el mismo archivo"},
		{Line: 10, Kind: taskdoc.KindNote, Text: "se decide que un turno dura 30 minutos"},
	}

	if got := doc.Events(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Events() =\n%+v\nwant\n%+v", got, want)
	}
}

// TestEventsTolerateMalformedLines: the agent writes this file, so a line the
// human typed by hand must never break the parse. It becomes a note.
func TestEventsTolerateMalformedLines(t *testing.T) {
	doc := withHistory(
		`L1 · "quiero reservar turnos"`,
		`una línea que el humano agregó a mano`,
		`L3 · PREGUNTA P1 sin el separador`,
	)

	events := doc.Events()
	if len(events) != 3 {
		t.Fatalf("Events() returned %d events, want 3: %+v", len(events), events)
	}
	for i, event := range events {
		if event.Kind != taskdoc.KindNote {
			t.Errorf("event %d Kind = %q, want %q", i, event.Kind, taskdoc.KindNote)
		}
	}
	if events[0].Line != 1 {
		t.Errorf("event 0 Line = %d, want 1", events[0].Line)
	}
	if events[1].Line != 0 {
		t.Errorf("event 1 Line = %d, want 0 for a line with no L number", events[1].Line)
	}
	if events[1].Text != "una línea que el humano agregó a mano" {
		t.Errorf("event 1 Text = %q, want the raw line", events[1].Text)
	}
	if events[2].Text != "PREGUNTA P1 sin el separador" {
		t.Errorf("event 2 Text = %q, want the raw payload", events[2].Text)
	}
}

// TestAppendHistoryIsAppendOnly: the log is the memory of the work, so a new
// line never rewrites, reorders or drops an old one.
func TestAppendHistoryIsAppendOnly(t *testing.T) {
	doc := withHistory(`L1 · "quiero reservar turnos"`, `L2 · se decide que un turno dura 30 minutos`)
	before := append([]string(nil), doc.History...)

	grown := doc.AppendHistory(`L3 · PREGUNTA P1 · "¿Se puede cancelar?"`)

	if len(grown.History) != len(before)+1 {
		t.Fatalf("len(History) = %d, want %d", len(grown.History), len(before)+1)
	}
	if !reflect.DeepEqual(grown.History[:len(before)], before) {
		t.Errorf("AppendHistory rewrote existing lines:\n got %v\nwant %v", grown.History[:len(before)], before)
	}
	if grown.History[len(before)] != `L3 · PREGUNTA P1 · "¿Se puede cancelar?"` {
		t.Errorf("appended line = %q", grown.History[len(before)])
	}
	if !reflect.DeepEqual(doc.History, before) {
		t.Errorf("AppendHistory mutated the receiver: %v", doc.History)
	}

	// A line arriving with stray whitespace is still stored exactly once, so a
	// rerun cannot grow the log with a blank line.
	if padded := doc.AppendHistory("   "); len(padded.History) != len(before) {
		t.Errorf("a blank line grew the log to %d lines; want %d", len(padded.History), len(before))
	}
}

// TestNextLineIsMonotonic: L numbers never repeat and never go back.
func TestNextLineIsMonotonic(t *testing.T) {
	cases := []struct {
		name string
		doc  taskdoc.TaskDoc
		want int
	}{
		{"empty", taskdoc.TaskDoc{}, 1},
		{"one line", withHistory(`L1 · "hola"`), 2},
		{"gap", withHistory(`L1 · "hola"`, `L9 · "chau"`), 10},
		{"trailing note", withHistory(`L1 · "hola"`, `sin número`), 2},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.doc.NextLine(); got != c.want {
				t.Errorf("NextLine() = %d, want %d", got, c.want)
			}
		})
	}
}

// TestNextQuestionIsMonotonic: P numbers are per task document and only count
// questions the agent asked, answered or not.
func TestNextQuestionIsMonotonic(t *testing.T) {
	cases := []struct {
		name string
		doc  taskdoc.TaskDoc
		want int
	}{
		{"empty", taskdoc.TaskDoc{}, 1},
		{
			name: "one open",
			doc:  withHistory(`L1 · PREGUNTA P1 · "¿a?"`),
			want: 2,
		},
		{
			name: "answered still counts",
			doc:  withHistory(`L1 · PREGUNTA P1 · "¿a?"`, `L2 · VOS P1 · "sí"`),
			want: 2,
		},
		{
			name: "craft is not a question",
			doc:  withHistory(`L1 · OFICIO · "usé SQLite" (porque no hay dos escritores)`),
			want: 1,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.doc.NextQuestion(); got != c.want {
				t.Errorf("NextQuestion() = %d, want %d", got, c.want)
			}
		})
	}
}

// TestOpenQuestions: a question is open while no VOS line answers it.
func TestOpenQuestions(t *testing.T) {
	doc := withHistory(
		`L1 · PREGUNTA P1 · "¿Se puede cancelar?"`,
		`L2 · PREGUNTA P2 · "¿Hace falta WhatsApp?"`,
		`L3 · VOS P1 · "No en la primera versión"`,
		`L4 · OFICIO · "Guardé los turnos en SQLite" (porque no hay dos escritores)`,
	)

	want := []taskdoc.Question{{ID: "P2", Text: "¿Hace falta WhatsApp?"}}
	if got := doc.OpenQuestions(); !reflect.DeepEqual(got, want) {
		t.Fatalf("OpenQuestions() = %+v, want %+v", got, want)
	}

	if got := withHistory(`L1 · PREGUNTA P1 · "¿a?"`, `L2 · VOS P1 · "b"`).OpenQuestions(); len(got) != 0 {
		t.Errorf("OpenQuestions() = %+v, want none once every question is answered", got)
	}
}

// TestCloseQuestionAppendsTheAnswer: closing never edits the question, it adds
// the VOS line right after the log.
func TestCloseQuestionAppendsTheAnswer(t *testing.T) {
	doc := withHistory(
		`L1 · PREGUNTA P1 · "¿Se puede cancelar?"`,
		`L2 · PREGUNTA P2 · "¿Hace falta WhatsApp?"`,
		`L3 · VOS P1 · "No en la primera versión"`,
	)

	closed, err := doc.CloseQuestion("P2", "Todavía no")
	if err != nil {
		t.Fatalf("CloseQuestion: %v", err)
	}

	want := `L4 · VOS P2 · "Todavía no"`
	if got := closed.History[len(closed.History)-1]; got != want {
		t.Errorf("appended line = %q, want %q", got, want)
	}
	if len(closed.History) != len(doc.History)+1 {
		t.Fatalf("len(History) = %d, want %d", len(closed.History), len(doc.History)+1)
	}
	if !reflect.DeepEqual(closed.History[:len(doc.History)], doc.History) {
		t.Errorf("CloseQuestion rewrote the log:\n got %v\nwant %v", closed.History, doc.History)
	}
	if len(closed.OpenQuestions()) != 0 {
		t.Errorf("OpenQuestions() = %+v after closing, want none", closed.OpenQuestions())
	}
	if got := closed.NextLine(); got != 5 {
		t.Errorf("NextLine() = %d after closing, want 5", got)
	}
}

// TestCloseQuestionRejectsTheImpossible: an id that was never asked, or one
// already answered, is an error the CLI turns into one Spanish line.
func TestCloseQuestionRejectsTheImpossible(t *testing.T) {
	doc := withHistory(
		`L1 · PREGUNTA P1 · "¿Se puede cancelar?"`,
		`L2 · VOS P1 · "No en la primera versión"`,
	)

	cases := []struct {
		name string
		id   string
		want string
	}{
		{"unknown", "P9", "P9"},
		{"already answered", "P1", "P1"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := doc.CloseQuestion(c.id, "lo que sea")
			if err == nil {
				t.Fatalf("CloseQuestion(%q) = nil error, want an error", c.id)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("error %q does not name %q", err, c.want)
			}
			if strings.Contains(err.Error(), "\n") {
				t.Errorf("error must be one line: %q", err)
			}
		})
	}
}

// TestHistoryRoundTripsByteExactly: growing the log must not disturb what the
// parser already reads. An existing document with history lines is rewritten
// byte for byte.
func TestHistoryRoundTripsByteExactly(t *testing.T) {
	source := `# reserva-de-turno

## Requisitos (textual, del usuario)
R1 · "quiero reservar turnos"

## Listo cuando
- Reservo y aparece en la agenda (R1)

## Pasos
- [x] T1 · agenda del día   R1 · 4a1c9f

## Historial
L1 · "quiero reservar turnos"
L2 · PREGUNTA P1 · "¿Los turnos se pueden cancelar?"
L3 · VOS P1 · "No en la primera versión"
L4 · OFICIO · "Guardé los turnos en SQLite" (porque dos personas no pueden escribir el mismo archivo)
`

	doc := taskdoc.Parse(source)
	if got := doc.Render(); got != source {
		t.Fatalf("Render() changed an existing document:\n--- got ---\n%s\n--- want ---\n%s", got, source)
	}

	grown := doc.AppendHistory(`L5 · PREGUNTA P2 · "¿Hace falta WhatsApp?"`)
	reparsed := taskdoc.Parse(grown.Render())

	if !reflect.DeepEqual(reparsed.History, grown.History) {
		t.Errorf("history after a round trip =\n%v\nwant\n%v", reparsed.History, grown.History)
	}
	if !reflect.DeepEqual(reparsed, grown) {
		t.Errorf("Parse(Render(d)) != d for a grown document")
	}
	if got := reparsed.Render(); got != grown.Render() {
		t.Errorf("Render is not stable after a round trip:\n%s", got)
	}
}
