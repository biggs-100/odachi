package start

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/biggs-100/odachi/internal/domain/brief"
)

// field identifies one of the seven answers, in the order the interview asks
// them.
type field int

const (
	fieldOutcome field = iota
	fieldAudience
	fieldDone
	fieldExclude
	fieldSensitive
	fieldDeadline
	fieldLearn
)

// questionSpec is one question and the answer field it fills.
type questionSpec struct {
	field   field
	text    string
	why     string
	options []string
}

// The seven questions, verbatim from Contract III. The why is one line and is
// shown with the question: it is the tutoring, not a lecture.
var specs = []questionSpec{
	{
		field: fieldOutcome,
		text:  "What do you want to exist when we finish? (one sentence, in your own words)",
		why:   "without this there is nothing to build; and in your own words it is what you read later to know we are on track",
	},
	{
		field:   fieldAudience,
		text:    "Who uses it, and from where? (1 phone · 2 computer · 3 both)",
		why:     "it changes everything that comes after, and it is the most expensive decision to change late",
		options: []string{"phone", "cpu", "both"},
	},
	{
		field: fieldDone,
		text:  "How will you know it worked? (I propose three; you can write your own)",
		why:   `this is called an acceptance criterion: it separates "it works" from "I think it works"`,
	},
	{
		field:   fieldExclude,
		text:    "What do you NOT want in this first version? (choose as many as you like)",
		why:     "it is what prevents 40 half-done things; a senior says no, and here the fence enforces it",
		options: presetKeys(),
	},
	{
		field:   fieldSensitive,
		text:    "Is there sensitive data? (1 none · 2 names · 3 health · 4 money)",
		why:     "it defines what cannot leave the project from day one",
		options: []string{"none", "names", "health", "money"},
	},
	{
		field:   fieldDeadline,
		text:    "How much time do you have for the first usable version? (1 today · 2 this week · 3 no rush)",
		why:     "it defines how much we cut; a project that never finishes teaches nothing",
		options: []string{"today", "week", "relaxed"},
	},
	{
		field: fieldLearn,
		text:  "What would you like to learn along the way? (choose one or two)",
		why:   "the tutor needs to know what to explain more; this is what takes you past novice",
	},
}

// questionSpecs returns the seven questions in order.
func questionSpecs() []questionSpec {
	return specs
}

// retryOutcome rejects an empty question 1: the whole project lives in that
// sentence, so it is never silently accepted.
const retryOutcome = "answer 1 cannot be empty: it is the sentence you will read later; write it again"

// answers is every answer gathered so far. An empty value means "not answered
// yet"; only the five brief-shaping answers are required.
type answers struct {
	outcome   string
	audience  string // phone | cpu | both
	done      []string
	exclude   []string
	sensitive string // none | names | health | money
	deadline  string // today | week | relaxed
	learn     []string
}

// has reports whether the field already has an answer. Exclude and learn are
// optional, so they never block a run.
func (a answers) has(f field) bool {
	switch f {
	case fieldOutcome:
		return a.outcome != ""
	case fieldAudience:
		return a.audience != ""
	case fieldDone:
		return len(a.done) > 0
	case fieldSensitive:
		return a.sensitive != ""
	case fieldDeadline:
		return a.deadline != ""
	default:
		return false
	}
}

// missing lists the required answers that are still empty, in question order.
func (a answers) missing() []field {
	var out []field
	for _, f := range []field{fieldOutcome, fieldAudience, fieldDone, fieldSensitive, fieldDeadline} {
		if !a.has(f) {
			out = append(out, f)
		}
	}
	return out
}

// readableFlags turns missing fields into the flags that would answer them.
func readableFlags(fields []field) []string {
	names := map[field]string{
		fieldOutcome:   "--outcome",
		fieldAudience:  "--audience",
		fieldDone:      "--done",
		fieldExclude:   "--exclude",
		fieldSensitive: "--sensitive",
		fieldDeadline:  "--deadline",
		fieldLearn:     "--learn",
	}
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		out = append(out, names[f])
	}
	return out
}

// decisions renders the answers that become durable decisions in brief.md. The
// human reads these lines later, so they say the decision, not the question.
func (a answers) decisions() []string {
	out := []string{
		audienceDecision(a.audience),
		sensitiveDecision(a.sensitive),
		deadlineDecision(a.deadline),
	}
	for _, item := range a.learn {
		out = append(out, "Wants to learn: "+strings.TrimSuffix(item, ".")+".")
	}
	return out
}

func audienceDecision(audience string) string {
	switch audience {
	case "phone":
		return "It is used from the phone."
	case "cpu":
		return "It is used from the computer."
	default:
		return "It is used from the phone and the computer."
	}
}

func sensitiveDecision(sensitive string) string {
	if sensitive == "none" {
		return "There is no sensitive data in the first version."
	}
	return "Sensitive data: " + sensitiveLabel(sensitive) + "; it does not leave the project from day one."
}

func deadlineDecision(deadline string) string {
	return "First usable version: " + deadlineLabel(deadline) + "."
}

func audienceLabel(audience string) string {
	switch audience {
	case "phone":
		return "the phone"
	case "cpu":
		return "the computer"
	default:
		return "the phone and the computer"
	}
}

func sensitiveLabel(sensitive string) string {
	switch sensitive {
	case "names":
		return "names"
	case "health":
		return "health"
	case "money":
		return "money"
	default:
		return "none"
	}
}

func deadlineLabel(deadline string) string {
	switch deadline {
	case "week":
		return "this week"
	case "relaxed":
		return "no rush"
	default:
		return "today"
	}
}

// answersFromFlags validates what the command line supplied. The write path is
// lenient about wording but never about silently accepting nonsense.
func answersFromFlags(opts Options) (answers, error) {
	var a answers

	if opts.Outcome != "" {
		outcome := strings.TrimSpace(opts.Outcome)
		if len(slugWords(outcome)) == 0 {
			return answers{}, fmt.Errorf(
				`invalid value for --outcome: %s; write a sentence with at least one word (for example odachi start --outcome "appointment list")`,
				quote(opts.Outcome))
		}
		a.outcome = outcome
	}

	if opts.Audience != "" {
		normalized, ok := normalizeAudience(opts.Audience)
		if !ok {
			return answers{}, fmt.Errorf(
				"invalid value for --audience: %s; use phone, cpu or both (for example odachi start --audience phone)",
				quote(opts.Audience))
		}
		a.audience = normalized
	}

	done, err := uniqueValues("--done", opts.Done)
	if err != nil {
		return answers{}, err
	}
	a.done = done

	exclude, err := uniqueValues("--exclude", opts.Exclude)
	if err != nil {
		return answers{}, err
	}
	for _, item := range exclude {
		if len(slugWords(item)) == 0 {
			return answers{}, fmt.Errorf(
				"invalid value for --exclude: %s; use a word or a sentence with letters (for example --exclude payments)",
				quote(item))
		}
	}
	a.exclude = exclude

	if opts.Sensitive != "" {
		normalized, ok := normalizeSensitive(opts.Sensitive)
		if !ok {
			return answers{}, fmt.Errorf(
				"invalid value for --sensitive: %s; use none, names, health or money (for example odachi start --sensitive none)",
				quote(opts.Sensitive))
		}
		a.sensitive = normalized
	}

	if opts.Deadline != "" {
		normalized, ok := normalizeDeadline(opts.Deadline)
		if !ok {
			return answers{}, fmt.Errorf(
				"invalid value for --deadline: %s; use today, week or relaxed (for example odachi start --deadline today)",
				quote(opts.Deadline))
		}
		a.deadline = normalized
	}

	learn, err := uniqueValues("--learn", opts.Learn)
	if err != nil {
		return answers{}, err
	}
	a.learn = learn

	return a, nil
}

// answersFromBrief maps a parsed brief onto the answers the artifacts need. The
// brief has no audience or deadline: those questions were answered when it was
// written, and the file is not rewritten.
func answersFromBrief(b brief.Brief) answers {
	return answers{
		outcome: b.Outcome[0],
		done:    b.DoneWhen,
		exclude: b.NonGoals,
	}
}

// uniqueValues trims and de-duplicates a repeatable flag, rejecting an empty
// value instead of writing a blank bullet.
func uniqueValues(flagName string, values []string) ([]string, error) {
	var out []string
	seen := map[string]bool{}

	for _, raw := range values {
		value := strings.TrimSpace(raw)
		if value == "" {
			return nil, fmt.Errorf(
				"invalid value for %s: it is empty; write a value and run odachi start again", flagName)
		}
		key := strings.ToLower(value)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, value)
	}
	return out, nil
}

// normalizeAudience accepts the number the human types and the flag value.
func normalizeAudience(value string) (string, bool) {
	switch normalizeAnswer(value) {
	case "1", "celular", "celu", "telefono", "movil", "phone":
		return "phone", true
	case "2", "compu", "computadora", "pc", "desktop", "cpu":
		return "cpu", true
	case "3", "los dos", "ambos", "celular y compu", "both":
		return "both", true
	default:
		return "", false
	}
}

// normalizeSensitive accepts the number the human types and the flag value.
func normalizeSensitive(value string) (string, bool) {
	switch normalizeAnswer(value) {
	case "1", "ninguno", "ninguna", "nada", "none":
		return "none", true
	case "2", "nombres", "nombre", "names":
		return "names", true
	case "3", "salud", "health":
		return "health", true
	case "4", "plata", "dinero", "money":
		return "money", true
	default:
		return "", false
	}
}

// normalizeDeadline accepts the number the human types and the flag value.
func normalizeDeadline(value string) (string, bool) {
	switch normalizeAnswer(value) {
	case "1", "hoy", "today":
		return "today", true
	case "2", "esta semana", "semana", "week":
		return "week", true
	case "3", "sin apuro", "sinapuro", "sin prisa", "relaxed":
		return "relaxed", true
	default:
		return "", false
	}
}

// normalizeAnswer lowercases, folds accents, turns dashes into spaces and
// collapses runs of whitespace, so "Esta  Semana" and "esta-semana" agree.
func normalizeAnswer(value string) string {
	folded := foldAccents(strings.ToLower(strings.TrimSpace(value)))
	folded = strings.ReplaceAll(folded, "-", " ")
	return strings.Join(strings.Fields(folded), " ")
}

// splitList splits a free-text answer on commas and semicolons: the human was
// asked to choose several, and a single line is how they answer.
func splitList(value string) []string {
	var out []string
	for _, part := range strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == ';' }) {
		if item := strings.TrimSpace(part); item != "" {
			out = append(out, item)
		}
	}
	return out
}

// appendUnique appends value unless an equivalent one is already listed.
func appendUnique(items []string, value string) []string {
	key := strings.ToLower(strings.TrimSpace(value))
	for _, item := range items {
		if strings.ToLower(strings.TrimSpace(item)) == key {
			return items
		}
	}
	return append(items, value)
}

// quote is how a message names what the human typed.
func quote(value string) string {
	return strconv.Quote(value)
}

// orNone renders an empty optional list as "none".
func orNone(items []string) string {
	if len(items) == 0 {
		return "none"
	}
	return strings.Join(items, ", ")
}
