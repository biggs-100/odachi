// Package brief parses brief.md, the human-owned vision artifact: what has to
// exist, when it is done, what is explicitly out of scope, and which durable
// decisions are already taken.
//
// Parse is total: any input, however malformed, yields a Brief and never
// panics. Unknown sections are ignored so the human can write notes in the
// same file without breaking the tool.
package brief

import "strings"

// Brief is the parsed brief.md. Every list holds one entry per bullet line.
type Brief struct {
	Name      string
	Outcome   []string
	DoneWhen  []string
	NonGoals  []string
	Decisions []string
}

// The canonical English headers. The artifacts are English; the agent chat is
// the human's language. The Spanish headers stay as read-only aliases so a
// brief written before the language migration still parses.
const (
	headerName      = "# "
	headerOutcome   = "## What must exist"
	headerDoneWhen  = "## Done when"
	headerNonGoals  = "## Not included (first version)"
	headerDecisions = "## Decisions taken"

	headerOutcomeES   = "## Qué tiene que existir"
	headerDoneWhenES  = "## Listo cuando"
	headerNonGoalsES  = "## NO incluye (primera versión)"
	headerDecisionsES = "## Decisiones tomadas"
)

// chapter names the brief part a section header feeds, if any. Unknown
// sections return "" and their content is ignored.
func chapter(header string) string {
	switch header {
	case headerOutcome, headerOutcomeES:
		return "outcome"
	case headerDoneWhen, headerDoneWhenES:
		return "donewhen"
	case headerNonGoals, headerNonGoalsES:
		return "nongoals"
	case headerDecisions, headerDecisionsES:
		return "decisions"
	default:
		return ""
	}
}

// Parse reads a brief.md and returns its parts. Bullet lines outside a known
// section, empty bullets, and unknown headers are ignored. A bullet is read as
// a list item, so a line that follows it without a blank line in between
// continues it: the human drafts at 80 columns and the tool keeps the sentence
// whole (0256).
func Parse(md string) Brief {
	var b Brief
	section := ""
	// open is the list the last bullet was appended to. A following line joins
	// that bullet until a blank line, a header or another bullet ends it.
	var open *[]string

	for _, raw := range strings.Split(md, "\n") {
		line := strings.TrimSpace(strings.TrimSuffix(raw, "\r"))
		if line == "" {
			open = nil
			continue
		}

		// A header always ends the previous section, known or not.
		if strings.HasPrefix(line, "## ") {
			section = chapter(line)
			open = nil
			continue
		}
		if strings.HasPrefix(line, headerName) {
			b.Name = strings.TrimSpace(strings.TrimPrefix(line, headerName))
			section = ""
			open = nil
			continue
		}
		if strings.HasPrefix(line, "#") {
			// Deeper or malformed header: not part of any brief section.
			section = ""
			open = nil
			continue
		}

		if !strings.HasPrefix(line, "-") {
			// Prose under a bullet is that bullet's own continuation.
			if open != nil && len(*open) > 0 {
				(*open)[len(*open)-1] += " " + line
			}
			continue
		}

		item, ok := bullet(line)
		if !ok {
			// An empty bullet ends the one above it and carries nothing.
			open = nil
			continue
		}
		switch section {
		case "outcome":
			b.Outcome = append(b.Outcome, item)
			open = &b.Outcome
		case "donewhen":
			b.DoneWhen = append(b.DoneWhen, item)
			open = &b.DoneWhen
		case "nongoals":
			b.NonGoals = append(b.NonGoals, item)
			open = &b.NonGoals
		case "decisions":
			b.Decisions = append(b.Decisions, item)
			open = &b.Decisions
		default:
			open = nil
		}
	}

	return b
}

// Render returns the canonical brief.md: the title when the brief has a name,
// then the four canonical English sections in the order a human reads them.
// Parse(Render(b)) round-trips, so Render is the only way the tool writes this
// artifact. Empty items are skipped: an empty bullet carries nothing and Parse
// would drop it anyway.
func Render(b Brief) string {
	var out strings.Builder

	if name := strings.TrimSpace(b.Name); name != "" {
		out.WriteString(headerName + name + "\n\n")
	}

	out.WriteString(RenderFeature(b))

	return out.String()
}

// RenderFeature writes the four canonical sections without the title line: the
// block a second feature appends to a brief that already exists (0271). The
// sections are the ones Parse already reads, repeated, so the reader and the
// writer live on one grammar.
func RenderFeature(b Brief) string {
	var out strings.Builder

	sections := []struct {
		header string
		items  []string
	}{
		{headerOutcome, b.Outcome},
		{headerDoneWhen, b.DoneWhen},
		{headerNonGoals, b.NonGoals},
		{headerDecisions, b.Decisions},
	}

	for _, section := range sections {
		if out.Len() > 0 {
			out.WriteString("\n")
		}
		out.WriteString(section.header + "\n")
		for _, item := range section.items {
			if strings.TrimSpace(item) == "" {
				continue
			}
			out.WriteString("- " + item + "\n")
		}
	}

	return out.String()
}

// Missing returns the English labels of the required parts that are empty, in
// the order the canonical brief declares them. An empty result means the brief
// is complete enough to work from.
func (b Brief) Missing() []string {
	var missing []string
	if len(b.Outcome) == 0 {
		missing = append(missing, "what must exist")
	}
	if len(b.DoneWhen) == 0 {
		missing = append(missing, "done when")
	}
	// Non-goals are deliberately not required. The interview lets the human pick
	// none, and forcing one invents a rule they never meant; a false fence
	// violation teaches the wrong lesson. When the list is empty, fence.toml
	// explains what a rule is for instead of pretending to protect something.
	return missing
}

// bullet returns the text of a "- ..." line. An empty bullet is not a bullet:
// it carries no content the agent could act on.
func bullet(line string) (string, bool) {
	if !strings.HasPrefix(line, "-") {
		return "", false
	}
	item := strings.TrimSpace(strings.TrimPrefix(line, "-"))
	if item == "" {
		return "", false
	}
	return item, true
}
