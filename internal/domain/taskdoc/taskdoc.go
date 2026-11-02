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
//
// Route says who runs the step and Agent names the executor when it is not the
// parent. Both are declarations: the tool never infers that a step without a
// route was run inline, because a document written before the field existed
// says nothing instead of saying the parent did it.
type Step struct {
	ID, Text, Requirement, Commit string
	Route, Agent                  string
	Done                          bool
}

// The two routes a step can declare.
const (
	// RouteInline is the parent that runs the step itself.
	RouteInline = "inline"
	// RouteDelegated is another agent that runs it, named in Agent.
	RouteDelegated = "delegated"
)

// routePrefix opens the tail part that says who runs the step.
const routePrefix = "route:"

// TaskDoc is the parsed task document.
type TaskDoc struct {
	Title string
	// Base is the commit this feature started from: the gate asks for the review
	// of that diff, not of the commit that closes the feature.
	Base string
	// ReviewWaived is the human's decision that this feature lands without a
	// review verdict. The gate names it instead of hiding it; the proof and the
	// fence still apply.
	ReviewWaived bool
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

	// reviewWaivedLine is the human's decision that this feature lands without a
	// review verdict. It lives in the preamble, where the gate reads it, and it
	// is the only ceremony a small change may skip: the proof and the fence stay.
	reviewWaivedLine = "Review: waived"

	headerRequirementsES = "## Requisitos (textual, del usuario)"
	headerDoneWhenES     = "## Listo cuando"
	headerStepsES        = "## Pasos"
	headerHistoryES      = "## Historial"
)

// checkboxPrefix opens a step line: "- [x] " or "- [ ] ".
const checkboxPrefix = "- ["

// basePrefix opens the one preamble line Parse reads besides the title: the
// commit the feature started from.
const basePrefix = "Base: "

// Separator between an id and its text, and between a requirement and its
// commit. The human reads these documents, so the glyph is the one the
// canonical example uses.
const sep = " · "

var (
	// idPattern is R1, R2, ... or T1, T2, ...: a letter run then a number.
	idPattern = regexp.MustCompile(`^[A-Za-z]+\d+$`)

	// gapPattern is the wide gap that separates step text from its
	// requirement/route/commit tail. The canonical form uses three spaces so a
	// " · " inside the text stays part of the text.
	gapPattern = regexp.MustCompile(`[ \t]{2,}`)

	// tailSepPattern splits the tail into its parts: the requirement, the route
	// and the commit, in that order. Tolerant of a hand-written line that has no
	// spaces around the glyph, because a document is read by people first.
	tailSepPattern = regexp.MustCompile(`\s*·\s*`)
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
	// preamble is the only place the base line is read: after the title and
	// before the first section. `current == ""` is not enough, because an unknown
	// header or a `###` note also clears it, and a note later in the document
	// would become the commit the closing gate must name.
	preamble := true

	for _, raw := range strings.Split(md, "\n") {
		line := strings.TrimSpace(strings.TrimSuffix(raw, "\r"))
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "## ") {
			current = section(line)
			preamble = false
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

		// The base lives in the preamble, between the title and the first
		// section. Only that position is read: a "Base: " inside the history is
		// prose the human wrote, not a field.
		if preamble && strings.HasPrefix(line, basePrefix) {
			d.Base = strings.TrimSpace(strings.TrimPrefix(line, basePrefix))
			continue
		}

		// The waiver lives in the same preamble for the same reason: a line in
		// the history that says "Review: waived" is prose, not a decision the
		// gate must obey.
		if preamble && line == reviewWaivedLine {
			d.ReviewWaived = true
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
	if d.Base != "" {
		b.WriteString(basePrefix + d.Base + "\n")
	}
	if d.ReviewWaived {
		b.WriteString(reviewWaivedLine + "\n")
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
//   - [x] T1 · text   R1 · tree:4a1c9f
//   - [ ] T2 · text   R1 · route:delegated/odachi-worker
//   - [ ] T3 · text   · route:inline
//
// The tail is the requirement, then the route, then the commit, so a reader
// finds who ran the step next to the bytes it produced. A commit or a route
// without a requirement link has no first slot, so it renders after a lone "·"
// and Parse reads it back into the same Step.
func renderStep(s Step) string {
	box := "[ ]"
	if s.Done {
		box = "[x]"
	}
	line := fmt.Sprintf("- %s %s%s%s", box, s.ID, sep, s.Text)

	tail := make([]string, 0, 3)
	if s.Requirement != "" {
		tail = append(tail, s.Requirement)
	}
	if route := renderRoute(s); route != "" {
		tail = append(tail, route)
	}
	if s.Commit != "" {
		tail = append(tail, s.Commit)
	}

	switch {
	case len(tail) == 0:
		return line
	case s.Requirement == "":
		return line + "   · " + strings.Join(tail, sep)
	default:
		return line + "   " + strings.Join(tail, sep)
	}
}

// renderRoute is the one spelling of a route: "route:inline", or
// "route:delegated/<agent>" when another agent runs the step.
func renderRoute(s Step) string {
	switch {
	case s.Route == "":
		return ""
	case s.Agent != "":
		return routePrefix + s.Route + "/" + s.Agent
	default:
		return routePrefix + s.Route
	}
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

// parseStep reads one `- [x] T1 · text   R1 · route:inline · tree:abc` line. The
// requirement, route and commit tail is optional; the checkbox decides Done.
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
		if parsed, ok := parseTail(last); ok {
			step.Requirement, step.Route, step.Agent, step.Commit = parsed.requirement, parsed.route, parsed.agent, parsed.commit
			step.Text = strings.Join(segments[:len(segments)-1], "  ")
			return step, true
		}
	}
	step.Text = tail
	return step, true
}

// tail is the three facts a step line carries after its text, in order.
type tail struct{ requirement, route, agent, commit string }

// parseTail reads the last segment of a step line when it is the canonical tail
// and reports false when it is not: a segment that is neither an id nor a tail
// opened by a lone "·" is text, exactly as it was before routes existed.
//
// A route is "route:<word>" or "route:delegated/<agent>". A word the tool does
// not know is kept as it was written, so the document is never rewritten with a
// meaning its author did not give it.
func parseTail(segment string) (tail, bool) {
	trimmed := strings.TrimSpace(segment)
	parts := tailSepPattern.Split(trimmed, -1)
	kept := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			kept = append(kept, part)
		}
	}
	if len(kept) == 0 {
		return tail{}, false
	}

	// A tail is either a requirement id or an old commit-only tail, which the
	// canonical form writes after a single "·": "   · 4a1c9f". The dot says the
	// first part is not a requirement, even when it looks like one: a sha whose
	// letters run into digits ("cafe123") is exactly what idPattern matches.
	leadingDot := strings.HasPrefix(trimmed, "·")
	if !leadingDot && !idPattern.MatchString(kept[0]) {
		return tail{}, false
	}

	var parsed tail
	if !leadingDot && idPattern.MatchString(kept[0]) {
		parsed.requirement = kept[0]
		kept = kept[1:]
	}
	for _, part := range kept {
		if spec, isRoute := strings.CutPrefix(part, routePrefix); isRoute {
			parsed.route, parsed.agent = splitRoute(spec)
			continue
		}
		parsed.commit = part
	}
	return parsed, true
}

// splitRoute reads "inline" or "delegated/odachi-worker" into its two halves.
func splitRoute(spec string) (route, agent string) {
	route, agent, _ = strings.Cut(spec, "/")
	return strings.TrimSpace(route), strings.TrimSpace(agent)
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
