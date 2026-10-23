// Package taskdoc parses odd/tasks/NNNN-<slug>.md: the textual requirements,
// the observable done-when list, the ordered steps with their commits, and the
// chronological history. The agent writes it, the human reads it.
//
// Parse is total: malformed input never panics. Render is the canonical form,
// and Parse(Render(d)) == d for any document Parse produced.
package taskdoc

import (
	"fmt"
	"regexp"
	"strings"
)

// Requirement is one textual requirement, quoted from the user.
type Requirement struct{ ID, Text string }

// Step is one ordered unit of work. Requirement and Commit are empty when the
// agent has not linked them yet.
type Step struct {
	ID, Text, Requirement, Commit string
	Done                          bool
}

// TaskDoc is the parsed task document.
type TaskDoc struct {
	Title        string
	Requirements []Requirement
	DoneWhen     []string
	Steps        []Step
	History      []string
}

// The canonical English headers. The artifacts are English; the agent chat is
// the human's language. The Spanish headers stay as read-only aliases so a
// task document written before the language migration still parses.
const (
	headerTitle        = "# "
	headerRequirements = "## Requirements (verbatim, from the user)"
	headerDoneWhen     = "## Done when"
	headerSteps        = "## Steps"
	headerHistory      = "## History"

	headerRequirementsES = "## Requisitos (textual, del usuario)"
	headerDoneWhenES     = "## Listo cuando"
	headerStepsES        = "## Pasos"
	headerHistoryES      = "## Historial"
)

// checkboxPrefix opens a step line: "- [x] " or "- [ ] ".
const checkboxPrefix = "- ["

// Separator between an id and its text, and between a requirement and its
// commit. The human reads these documents, so the glyph is the one the
// canonical example uses.
const sep = " · "

var (
	// idPattern is R1, R2, ... or T1, T2, ...: a letter run then a number.
	idPattern = regexp.MustCompile(`^[A-Za-z]+\d+$`)

	// gapPattern is the wide gap that separates step text from its
	// requirement/commit tail. The canonical form uses three spaces so a
	// " · " inside the text stays part of the text.
	gapPattern = regexp.MustCompile(`[ \t]{2,}`)

	// tailPattern is "R1" or "R1 · 4a1c9f".
	tailPattern = regexp.MustCompile(`^([A-Za-z]+\d+)(?:\s*·\s*(\S+))?$`)

	// commitOnlyPattern is the fallback for a step recorded with a commit
	// but no requirement link: "· 4a1c9f".
	commitOnlyPattern = regexp.MustCompile(`^·\s*(\S+)$`)
)

// section names the task-doc part a header feeds, or "" for an unknown header.
func section(header string) string {
	switch header {
	case headerRequirements, headerRequirementsES:
		return "requirements"
	case headerDoneWhen, headerDoneWhenES:
		return "donewhen"
	case headerSteps, headerStepsES:
		return "steps"
	case headerHistory, headerHistoryES:
		return "history"
	default:
		return ""
	}
}

// Parse reads a task document. Content outside the canonical sections, and
// lines that do not match a canonical shape, are ignored rather than fatal:
// an agent appending a note must not break progress tracking.
func Parse(md string) TaskDoc {
	var d TaskDoc
	current := ""

	for _, raw := range strings.Split(md, "\n") {
		line := strings.TrimSpace(strings.TrimSuffix(raw, "\r"))
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "## ") {
			current = section(line)
			continue
		}
		if strings.HasPrefix(line, headerTitle) {
			d.Title = strings.TrimSpace(strings.TrimPrefix(line, headerTitle))
			current = ""
			continue
		}
		if strings.HasPrefix(line, "#") {
			current = ""
			continue
		}

		switch current {
		case "requirements":
			if r, ok := parseRequirement(line); ok {
				d.Requirements = append(d.Requirements, r)
			}
		case "donewhen":
			if item, ok := bullet(line); ok {
				d.DoneWhen = append(d.DoneWhen, item)
			}
		case "steps":
			if s, ok := parseStep(line); ok {
				d.Steps = append(d.Steps, s)
			}
		case "history":
			// History lines are free text with a user-chosen id prefix; keep
			// them verbatim so nothing the human said is lost or reworded.
			d.History = append(d.History, line)
		}
	}

	return d
}

// Next returns the first step that is not done, in document order.
func (d TaskDoc) Next() (Step, bool) {
	for _, s := range d.Steps {
		if !s.Done {
			return s, true
		}
	}
	return Step{}, false
}

// Progress returns how many steps are done and how many exist.
func (d TaskDoc) Progress() (done, total int) {
	for _, s := range d.Steps {
		if s.Done {
			done++
		}
	}
	return done, len(d.Steps)
}

// Record marks the step with the given id as done and records its commit. An
// unknown id is an error and leaves the receiver untouched. The receiver is a
// value, and the steps slice is cloned, so the caller's document is never
// mutated by surprise.
func (d TaskDoc) Record(id, commit string) (TaskDoc, error) {
	for i, s := range d.Steps {
		if s.ID != id {
			continue
		}
		steps := make([]Step, len(d.Steps))
		copy(steps, d.Steps)
		steps[i].Done = true
		steps[i].Commit = commit
		d.Steps = steps
		return d, nil
	}
	return d, fmt.Errorf("unknown step: %q; steps available: %s", id, stepIDs(d.Steps))
}

// Render returns the canonical form. Parse(Render(d)) equals d for every
// document Parse produced.
func (d TaskDoc) Render() string {
	var b strings.Builder

	if d.Title != "" {
		b.WriteString(headerTitle + d.Title + "\n")
	}

	b.WriteString("\n" + headerRequirements + "\n")
	for _, r := range d.Requirements {
		fmt.Fprintf(&b, "%s%s\"%s\"\n", r.ID, sep, r.Text)
	}

	b.WriteString("\n" + headerDoneWhen + "\n")
	for _, item := range d.DoneWhen {
		b.WriteString("- " + item + "\n")
	}

	b.WriteString("\n" + headerSteps + "\n")
	for _, s := range d.Steps {
		b.WriteString(renderStep(s) + "\n")
	}

	b.WriteString("\n" + headerHistory + "\n")
	for _, h := range d.History {
		b.WriteString(h + "\n")
	}

	return b.String()
}

// renderStep writes one canonical step line:
//
//   - [x] T1 · text   R1 · 4a1c9f
//   - [ ] T2 · text   R2
//
// A commit without a requirement link has no canonical slot, so it renders in
// the tail as "· commit"; Parse reads it back into the same Step.
func renderStep(s Step) string {
	box := "[ ]"
	if s.Done {
		box = "[x]"
	}
	line := fmt.Sprintf("- %s %s%s%s", box, s.ID, sep, s.Text)
	switch {
	case s.Requirement != "":
		line += "   " + s.Requirement
		if s.Commit != "" {
			line += sep + s.Commit
		}
	case s.Commit != "":
		line += "   · " + s.Commit
	}
	return line
}

// parseRequirement reads one `R1 · "text"` line. The quotes are the human's
// mark that the text is the user's own words; they are not part of the text.
func parseRequirement(line string) (Requirement, bool) {
	id, rest, ok := cut(line)
	if !ok || !idPattern.MatchString(id) {
		return Requirement{}, false
	}
	text := strings.TrimSpace(rest)
	if !strings.HasPrefix(text, `"`) {
		return Requirement{ID: id, Text: text}, true
	}
	// Tolerate an unterminated quote rather than dropping what the user said.
	text = strings.TrimSuffix(strings.TrimPrefix(text, `"`), `"`)
	return Requirement{ID: id, Text: strings.TrimSpace(text)}, true
}

// parseStep reads one `- [x] T1 · text   R1 · commit` line. The requirement
// and commit tail is optional; the checkbox decides Done.
func parseStep(line string) (Step, bool) {
	if !strings.HasPrefix(line, checkboxPrefix) {
		return Step{}, false
	}
	close := strings.Index(line, "]")
	if close < 0 {
		return Step{}, false
	}
	done := strings.EqualFold(strings.TrimSpace(line[len(checkboxPrefix):close]), "x")

	rest := line[close+1:]
	id, tail, ok := cut(strings.TrimSpace(rest))
	if !ok || !idPattern.MatchString(id) {
		return Step{}, false
	}

	step := Step{ID: id, Done: done}
	// Split the tail on the wide gap; keep the text trimmed on the right so
	// the canonical three-space gap does not become part of it.
	tail = strings.TrimRight(tail, " \t")
	segments := gapPattern.Split(tail, -1)
	if len(segments) > 1 {
		last := segments[len(segments)-1]
		if m := tailPattern.FindStringSubmatch(last); m != nil {
			step.Requirement = m[1]
			step.Commit = m[2]
			step.Text = strings.Join(segments[:len(segments)-1], "  ")
			return step, true
		}
		if m := commitOnlyPattern.FindStringSubmatch(last); m != nil {
			step.Commit = m[1]
			step.Text = strings.Join(segments[:len(segments)-1], "  ")
			return step, true
		}
	}
	step.Text = tail
	return step, true
}

// cut splits a line at the first id/text separator and trims the id.
func cut(line string) (string, string, bool) {
	i := strings.Index(line, sep)
	if i < 0 {
		return "", "", false
	}
	return strings.TrimSpace(line[:i]), line[i+len(sep):], true
}

// bullet returns the text of a "- ..." line; an empty bullet is not a bullet.
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

// stepIDs lists the known ids for an error message, so the caller can fix it
// without opening the file. Empty means the document declares no steps.
func stepIDs(steps []Step) string {
	if len(steps) == 0 {
		return "none"
	}
	ids := make([]string, 0, len(steps))
	for _, s := range steps {
		ids = append(ids, s.ID)
	}
	return strings.Join(ids, ", ")
}
