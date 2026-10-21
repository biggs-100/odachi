package taskdoc

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// The record kinds of Contract IV. The `## Historial` is the only log the
// project has: three line shapes the agent writes, plus the human's own notes.
const (
	// KindQuestion is `L7 · PREGUNTA P1 · "¿Los turnos se pueden cancelar?"`.
	KindQuestion = "question"
	// KindAnswer is `L8 · VOS P1 · "No en la primera versión"`.
	KindAnswer = "answer"
	// KindCraft is `L9 · OFICIO · "decisión" (porque razón)`: the agent
	// deciding alone and saying why. It is never a question.
	KindCraft = "craft"
	// KindNote is anything else, including a line the human typed by hand.
	KindNote = "note"
)

// Event is one parsed history line. Line is the L number, 0 when the line has
// none. ID is the P number for questions and answers. Because carries the
// reason of a craft decision.
type Event struct {
	Line              int
	Kind              string
	ID, Text, Because string
}

// Question is one open vision question: the agent asked it, the human has not
// answered it yet.
type Question struct{ ID, Text string }

// The history grammar, in one place: a line number, the three record shapes,
// and the optional reason a craft decision carries.
const (
	recordQuestion = "PREGUNTA "
	recordAnswer   = "VOS "
	recordCraft    = "OFICIO"
	reasonOpen     = "(porque "
)

var (
	// lineNumberPattern is the "L7" that opens every history line.
	lineNumberPattern = regexp.MustCompile(`^L\d+$`)

	// questionIDPattern is the "P1" of a question or an answer.
	questionIDPattern = regexp.MustCompile(`^P\d+$`)
)

// Events returns the history as records, in file order. A malformed line is a
// note rather than an error: this file is written by an agent and read by
// humans, and nothing a human typed may break the parse.
func (d TaskDoc) Events() []Event {
	events := make([]Event, 0, len(d.History))
	for _, line := range d.History {
		events = append(events, parseEvent(line))
	}
	return events
}

// AppendHistory adds one line after the last one. It never rewrites, reorders
// or drops what is already there: the log is the memory of the work. A blank
// line is ignored instead of growing the log.
func (d TaskDoc) AppendHistory(line string) TaskDoc {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return d
	}

	history := make([]string, len(d.History), len(d.History)+1)
	copy(history, d.History)
	d.History = append(history, trimmed)
	return d
}

// NextLine is the first unused L number.
func (d TaskDoc) NextLine() int {
	highest := 0
	for _, event := range d.Events() {
		if event.Line > highest {
			highest = event.Line
		}
	}
	return highest + 1
}

// NextQuestion is the first unused P number of this task document.
func (d TaskDoc) NextQuestion() int {
	highest := 0
	for _, event := range d.Events() {
		if event.Kind != KindQuestion && event.Kind != KindAnswer {
			continue
		}
		number, ok := questionNumber(event.ID)
		if ok && number > highest {
			highest = number
		}
	}
	return highest + 1
}

// OpenQuestions returns the questions no VOS line answers, in the order they
// were asked.
func (d TaskDoc) OpenQuestions() []Question {
	events := d.Events()

	answered := map[string]bool{}
	for _, event := range events {
		if event.Kind == KindAnswer {
			answered[event.ID] = true
		}
	}

	open := make([]Question, 0, len(events))
	for _, event := range events {
		if event.Kind != KindQuestion || answered[event.ID] {
			continue
		}
		open = append(open, Question{ID: event.ID, Text: event.Text})
	}
	return open
}

// CloseQuestion appends the human's answer to an open question. An id that was
// never asked, or one that is already answered, is an error: the log only
// grows, so answering twice would leave two answers for one question.
func (d TaskDoc) CloseQuestion(id, answer string) (TaskDoc, error) {
	normalized := strings.ToUpper(strings.TrimSpace(id))

	if _, ok := questionNumber(normalized); !ok {
		return d, fmt.Errorf("no entendí %q como número de pregunta; usá --for P1 (o el que corresponda)", id)
	}

	asked := false
	for _, event := range d.Events() {
		if event.ID != normalized || (event.Kind != KindQuestion && event.Kind != KindAnswer) {
			continue
		}
		if event.Kind == KindAnswer {
			return d, fmt.Errorf("la pregunta %s ya está contestada; para abrir otra corré odachi ask \"<pregunta>\"", normalized)
		}
		asked = true
	}
	if !asked {
		return d, fmt.Errorf("este task doc no tiene la pregunta %s; abrila con odachi ask \"<pregunta>\"", normalized)
	}

	line := fmt.Sprintf("L%d · VOS %s · %s", d.NextLine(), normalized, quoted(answer))
	return d.AppendHistory(line), nil
}

// parseEvent reads one history line. Everything it cannot attribute becomes a
// note that keeps the line verbatim.
func parseEvent(line string) Event {
	raw := strings.TrimSpace(line)

	id, payload, ok := cut(raw)
	if !ok || !lineNumberPattern.MatchString(id) {
		return Event{Kind: KindNote, Text: raw}
	}
	number, err := strconv.Atoi(strings.TrimPrefix(id, "L"))
	if err != nil {
		return Event{Kind: KindNote, Text: raw}
	}

	event := Event{Line: number, Kind: KindNote}
	body := strings.TrimSpace(payload)

	switch {
	case strings.HasPrefix(body, recordQuestion):
		if question, ok := parseQuestion(KindQuestion, body, recordQuestion); ok {
			question.Line = event.Line
			return question
		}
	case strings.HasPrefix(body, recordAnswer):
		if answer, ok := parseQuestion(KindAnswer, body, recordAnswer); ok {
			answer.Line = event.Line
			return answer
		}
	case body == recordCraft, strings.HasPrefix(body, recordCraft+" "):
		event.Kind = KindCraft
		event.Text, event.Because = parseCraft(strings.TrimSpace(strings.TrimPrefix(body, recordCraft)))
		return event
	}

	event.Text = body
	return event
}

// parseQuestion reads "PREGUNTA P1 · "text"" and "VOS P1 · "text"".
func parseQuestion(kind, body, record string) (Event, bool) {
	id, text, ok := cut(strings.TrimSpace(strings.TrimPrefix(body, record)))
	if !ok || !questionIDPattern.MatchString(id) {
		return Event{}, false
	}
	return Event{Kind: kind, ID: id, Text: unquote(text)}, true
}

// parseCraft reads `"decisión" (porque razón)`. The reason is optional: a craft
// line without one is still a craft decision, just a worse one.
func parseCraft(body string) (text, because string) {
	body = strings.TrimSpace(strings.TrimPrefix(body, "·"))
	const closing = ")"

	index := strings.LastIndex(body, reasonOpen)
	if index < 0 || !strings.HasSuffix(body, closing) {
		return unquote(body), ""
	}

	text = strings.TrimSpace(body[:index])
	because = strings.TrimSpace(strings.TrimSuffix(body[index+len(reasonOpen):], closing))
	return unquote(text), because
}

// unquote drops the quotes the canonical shape puts around the human's words.
func unquote(text string) string {
	return strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(text), `"`), `"`))
}

// quoted renders one line inside the canonical quotes, on a single line: the
// history is read one line at a time, so a multi-line answer would break it.
func quoted(text string) string {
	return `"` + strings.Join(strings.Fields(text), " ") + `"`
}

// questionNumber reads the number of a P id.
func questionNumber(id string) (int, bool) {
	if !questionIDPattern.MatchString(id) {
		return 0, false
	}
	number, err := strconv.Atoi(strings.TrimPrefix(id, "P"))
	if err != nil {
		return 0, false
	}
	return number, true
}
