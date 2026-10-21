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

// The canonical Spanish headers, exactly as the human writes them (Contract A
// in odd/tasks/odachi-bootstrap.md). The user reads this file, so the headers
// are Spanish; the code that reads them is English.
const (
	headerName      = "# "
	headerOutcome   = "## Qué tiene que existir"
	headerDoneWhen  = "## Listo cuando"
	headerNonGoals  = "## NO incluye (primera versión)"
	headerDecisions = "## Decisiones tomadas"
)

// chapter names the brief part a section header feeds, if any. Unknown
// sections return "" and their content is ignored.
func chapter(header string) string {
	switch header {
	case headerOutcome:
		return "outcome"
	case headerDoneWhen:
		return "donewhen"
	case headerNonGoals:
		return "nongoals"
	case headerDecisions:
		return "decisions"
	default:
		return ""
	}
}

// Parse reads a brief.md and returns its parts. Bullet lines outside a known
// section, empty bullets, and unknown headers are ignored.
func Parse(md string) Brief {
	var b Brief
	section := ""

	for _, raw := range strings.Split(md, "\n") {
		line := strings.TrimSpace(strings.TrimSuffix(raw, "\r"))
		if line == "" {
			continue
		}

		// A header always ends the previous section, known or not.
		if strings.HasPrefix(line, "## ") {
			section = chapter(line)
			continue
		}
		if strings.HasPrefix(line, headerName) {
			b.Name = strings.TrimSpace(strings.TrimPrefix(line, headerName))
			section = ""
			continue
		}
		if strings.HasPrefix(line, "#") {
			// Deeper or malformed header: not part of any brief section.
			section = ""
			continue
		}

		item, ok := bullet(line)
		if !ok {
			continue
		}
		switch section {
		case "outcome":
			b.Outcome = append(b.Outcome, item)
		case "donewhen":
			b.DoneWhen = append(b.DoneWhen, item)
		case "nongoals":
			b.NonGoals = append(b.NonGoals, item)
		case "decisions":
			b.Decisions = append(b.Decisions, item)
		}
	}

	return b
}

// Render returns the canonical brief.md: the title when the brief has a name,
// then the four Canonical Spanish sections in the order a human reads them.
// Parse(Render(b)) round-trips, so Render is the only way the tool writes this
// artifact. Empty items are skipped: an empty bullet carries nothing and Parse
// would drop it anyway.
func Render(b Brief) string {
	var out strings.Builder

	if name := strings.TrimSpace(b.Name); name != "" {
		out.WriteString(headerName + name + "\n")
	}

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

// Missing returns the Spanish labels of the required parts that are empty, in
// the order the canonical brief declares them. An empty result means the brief
// is complete enough to work from.
func (b Brief) Missing() []string {
	var missing []string
	if len(b.Outcome) == 0 {
		missing = append(missing, "qué tiene que existir")
	}
	if len(b.DoneWhen) == 0 {
		missing = append(missing, "listo cuando")
	}
	if len(b.NonGoals) == 0 {
		missing = append(missing, "no incluye")
	}
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
