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
	// KindChange is `L10 · CAMBIO R1 · "el turno dura 45 minutos"`: the human
	// changed a requirement, and their words are kept as they gave them.
	KindChange = "change"
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
	// Options is the closed answer domain a PREGUNTA declares. It is empty on
	// every other record and on a question that declares no options: the suffix
	// is optional so a line written before options existed still parses (0345).
	Options []string
}

// Question is one open vision question: the agent asked it, the human has not
// answered it yet. Options is empty when the question accepts any answer.
type Question struct {
	ID, Text string
	Options  []string
}

// The history grammar, in one place: a line number, the three record shapes,
// and the optional reason a craft decision carries.
//
// A PREGUNTA may declare a closed answer domain as an optional suffix after the
// quoted question text:
//
//	L7 · PREGUNTA P1 · "¿Los turnos se pueden cancelar?" · ["sí" | "no"]
//
// The suffix is a ` · `, then a bracketed, pipe-separated list of quoted
// options. It is invisible when the question declares none, so a PREGUNTA
// written before options existed keeps its exact bytes and Parse(Render(d))
// stays the identity (0345).
const (
	recordQuestion = "PREGUNTA "
	recordAnswer   = "VOS "
	recordCraft    = "OFICIO"
	recordChange   = "CAMBIO "
	reasonOpen     = "(porque "
)

var (
	// lineNumberPattern is the "L7" that opens every history line.
	lineNumberPattern = regexp.MustCompile(`^L\d+$`)

	// questionIDPattern is the "P1" of a question or an answer.
	questionIDPattern = regexp.MustCompile(`^P\d+$`)

	// requirementIDPattern is the "R1" a change line has to name: a change is
	// about a requirement, and anything else is prose the human wrote. A human
	// typing `CAMBIO r1` by hand means R1, so the case does not matter.
	requirementIDPattern = regexp.MustCompile(`(?i)^r\d+$`)

	// questionOptionsPattern is the optional PREGUNTA suffix: a ` · ` then a
	// bracketed list of quoted options at the very end of the line. Anchoring on
	// the trailing bracket keeps a question whose text merely contains `·` or
	// `[` parsing as text (0345).
	questionOptionsPattern = regexp.MustCompile(`\s*·\s*\[(.*)\]$`)
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
		open = append(open, Question{ID: event.ID, Text: event.Text, Options: event.Options})
	}
	return open
}

// CloseQuestion appends the human's answer to an open question. An id that was
// never asked, or one that is already answered, is an error: the log only
// grows, so answering twice would leave two answers for one question.
func (d TaskDoc) CloseQuestion(id, answer string) (TaskDoc, error) {
	normalized := strings.ToUpper(strings.TrimSpace(id))

	if _, ok := questionNumber(normalized); !ok {
		return d, fmt.Errorf("could not read %q as a question number; use --for P1 (or the matching one)", id)
	}

	asked := false
	var options []string
	for _, event := range d.Events() {
		if event.ID != normalized || (event.Kind != KindQuestion && event.Kind != KindAnswer) {
			continue
		}
		if event.Kind == KindAnswer {
			return d, fmt.Errorf("question %s is already answered; to open another one run odachi ask \"<question>\"", normalized)
		}
		asked = true
		options = event.Options
	}
	if !asked {
		return d, fmt.Errorf("this task doc has no question %s; open it with odachi ask \"<question>\"", normalized)
	}
	// A question that declares no options keeps the old verbatim behavior. One
	// that does accepts only an option as a whole answer; a longer word merely
	// containing an option ("así" for "sí") is refused, with no mutation (0345).
	if len(options) > 0 && !answerAllowed(answer, options) {
		return d, fmt.Errorf("question %s only accepts one of these answers: %s", normalized, strings.Join(options, ", "))
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
	case strings.HasPrefix(body, recordChange):
		if change, ok := parseChange(body); ok {
			change.Line = event.Line
			return change
		}
	}

	event.Text = body
	return event
}

// parseQuestion reads `PREGUNTA P1 · "text"` and `VOS P1 · "text"`. A question
// may carry the optional options suffix documented above; an answer never does.
func parseQuestion(kind, body, record string) (Event, bool) {
	id, rest, ok := cut(strings.TrimSpace(strings.TrimPrefix(body, record)))
	if !ok || !questionIDPattern.MatchString(id) {
		return Event{}, false
	}

	event := Event{Kind: kind, ID: id}
	if kind == KindQuestion {
		if loc := questionOptionsPattern.FindStringSubmatchIndex(rest); loc != nil {
			event.Options = parseOptions(rest[loc[2]:loc[3]])
			rest = strings.TrimSpace(rest[:loc[0]])
		}
	}
	event.Text = unquote(rest)
	return event, true
}

// parseOptions reads the bracketed, pipe-separated options of a PREGUNTA suffix
// into their declared order. A blank item is dropped rather than becoming an
// option nobody can answer with.
func parseOptions(raw string) []string {
	var options []string
	for _, part := range strings.Split(raw, "|") {
		if option := unquote(part); option != "" {
			options = append(options, option)
		}
	}
	return options
}

// RenderQuestionOptions writes the optional options suffix of a PREGUNTA line.
// It returns the empty string when the question declares none, so a question
// written without options renders byte for byte as it always did (0345).
func RenderQuestionOptions(options []string) string {
	if len(options) == 0 {
		return ""
	}
	quotedOptions := make([]string, 0, len(options))
	for _, option := range options {
		quotedOptions = append(quotedOptions, quoted(option))
	}
	return sep + "[" + strings.Join(quotedOptions, " | ") + "]"
}

// answerAllowed reports whether the answer names one of the declared options as
// a whole answer. A longer word that only contains an option ("así" for "sí")
// is not a match, so a substring test would accept what must be refused (0345).
func answerAllowed(answer string, options []string) bool {
	normalized := normalizeOption(answer)
	for _, option := range options {
		if normalizeOption(option) == normalized {
			return true
		}
	}
	return false
}

// normalizeOption lowercases and collapses whitespace so " Sí " and "sí" agree.
func normalizeOption(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(value), " "))
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

// parseChange reads `CAMBIO R1 · "the words the human gave"`. The requirement id
// is what makes the reopen derivable; a line that names anything else, or
// nothing, is a note.
func parseChange(body string) (Event, bool) {
	id, text, ok := cut(strings.TrimSpace(strings.TrimPrefix(body, recordChange)))
	if !ok || !requirementIDPattern.MatchString(id) {
		return Event{}, false
	}
	return Event{Kind: KindChange, ID: strings.ToUpper(id), Text: unquote(text)}, true
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
