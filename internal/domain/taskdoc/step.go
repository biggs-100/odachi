package taskdoc

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// MaxStepText is the longest text a new step may carry, in characters. A step
// is one short line the human can read in the document; anything longer is a
// paragraph pretending to be a unit of work.
const MaxStepText = 120

// NextStep is the first unused T number. The number is max+1, never the first
// hole: an id that was used once is never reused, so a commit message that says
// "T3" always means the same step.
func (d TaskDoc) NextStep() int {
	highest := 0
	for _, s := range d.Steps {
		number, ok := stepNumber(s.ID)
		if ok && number > highest {
			highest = number
		}
	}
	return highest + 1
}

// Route says who runs a step: nobody declared it (empty Kind), the parent runs
// it (RouteInline), or another agent does and Agent names that agent
// (RouteDelegated).
type Route struct{ Kind, Agent string }

// AddStep appends a step with the next T number, optionally linked to a
// requirement the document already declares. The text is normalized to one
// canonical line (single spaces) before it is stored, so the rendered document
// parses back to the same step.
//
// An unknown requirement is an error that lists the ones the document has: the
// agent must not invent a link the user never asked for. A bad text is an error
// too, and in both cases the receiver is returned untouched.
func (d TaskDoc) AddStep(text, requirement string) (TaskDoc, error) {
	return d.AddStepBy(text, requirement, Route{})
}

// AddStepBy is AddStep with the executor the step declares. The route and the
// agent are decided before the document is written, because a step that names an
// agent and one that names none produce different lines, and fixing the line
// afterwards is how a declaration turns into a guess.
func (d TaskDoc) AddStepBy(text, requirement string, route Route) (TaskDoc, error) {
	normalized, err := canonicalStepText(text)
	if err != nil {
		return d, err
	}

	link := strings.ToUpper(strings.TrimSpace(requirement))
	if link != "" && !d.HasRequirement(link) {
		return d, fmt.Errorf("requirement %s is not in this task doc; the requirements are %s", link, requirementIDs(d.Requirements))
	}

	executor, err := canonicalRoute(route)
	if err != nil {
		return d, err
	}

	step := Step{ID: "T" + strconv.Itoa(d.NextStep()), Text: normalized, Requirement: link, Route: executor.Kind, Agent: executor.Agent}

	steps := make([]Step, len(d.Steps), len(d.Steps)+1)
	copy(steps, d.Steps)
	d.Steps = append(steps, step)
	return d, nil
}

// canonicalRoute validates the pair a step declares and returns it in the
// spelling the document writes. The contradictions are refused here, where the
// document is built, and not only at the flag: any writer gets the same rule.
func canonicalRoute(route Route) (Route, error) {
	kind := strings.ToLower(strings.TrimSpace(route.Kind))
	agent := strings.TrimSpace(route.Agent)

	switch {
	case kind == "" && agent != "":
		return Route{}, fmt.Errorf("an agent needs a route: say how the step runs with %s or %s", RouteInline, RouteDelegated)
	case kind == "":
		return Route{}, nil
	case kind != RouteInline && kind != RouteDelegated:
		return Route{}, fmt.Errorf("the route %q is not one I know: the routes are %s and %s", route.Kind, RouteInline, RouteDelegated)
	case kind == RouteInline && agent != "":
		return Route{}, fmt.Errorf("an %s step has no agent: the parent runs it, so drop the agent or say %s", RouteInline, RouteDelegated)
	case kind == RouteDelegated && agent == "":
		return Route{}, fmt.Errorf("a %s step names the agent that runs it: pass --agent <name>", RouteDelegated)
	}
	if strings.ContainsAny(agent, " /·\t\r\n") {
		return Route{}, fmt.Errorf("the agent %q cannot be written in a step line: it has no spaces, no \"/\" and no \"·\"", agent)
	}
	return Route{Kind: kind, Agent: agent}, nil
}

// HasRequirement reports whether the document declares the id, ignoring case
// because an agent typing `--req r1` means R1.
func (d TaskDoc) HasRequirement(id string) bool {
	for _, r := range d.Requirements {
		if strings.EqualFold(r.ID, id) {
			return true
		}
	}
	return false
}

// Reopen unchecks every done step that cites the requirement and drops its mark.
// The change line already carries the human's words; the mark was evidence for
// the step as it was, and it does not survive a change that invalidates it. An
// empty requirement reopens nothing: without a link there is no step to blame.
func (d TaskDoc) Reopen(requirement string) (TaskDoc, []string) {
	link := strings.ToUpper(strings.TrimSpace(requirement))
	if link == "" {
		return d, nil
	}

	steps := make([]Step, len(d.Steps))
	copy(steps, d.Steps)

	reopened := []string{}
	for i, s := range steps {
		if !s.Done || !strings.EqualFold(s.Requirement, link) {
			continue
		}
		steps[i].Done = false
		steps[i].Commit = ""
		reopened = append(reopened, s.ID)
	}

	d.Steps = steps
	return d, reopened
}

// canonicalStepText validates and normalizes a step's text: one line, no
// leading dash (the render adds the checkbox), at most MaxStepText characters.
func canonicalStepText(text string) (string, error) {
	if strings.ContainsAny(text, "\r\n") {
		return "", fmt.Errorf("a step must be one line, without line breaks")
	}

	normalized := strings.Join(strings.Fields(text), " ")
	switch {
	case normalized == "":
		return "", fmt.Errorf(`the step has no text: write odachi plan "<text>"`)
	case strings.HasPrefix(normalized, "-"):
		return "", fmt.Errorf(`a step does not start with a dash: odachi plan writes "- [ ] T<n>" itself; write the text without the dash`)
	case utf8.RuneCountInString(normalized) > MaxStepText:
		return "", fmt.Errorf("the step has %d characters and the maximum is %d; summarize it in one line",
			utf8.RuneCountInString(normalized), MaxStepText)
	}

	return normalized, nil
}

// stepNumber reads the number of a T id.
func stepNumber(id string) (int, bool) {
	upper := strings.ToUpper(strings.TrimSpace(id))
	if !strings.HasPrefix(upper, "T") {
		return 0, false
	}
	number, err := strconv.Atoi(strings.TrimPrefix(upper, "T"))
	if err != nil {
		return 0, false
	}
	return number, true
}

// requirementIDs lists the requirements a document has, for an error that tells
// the agent what it can link to without opening the file.
func requirementIDs(requirements []Requirement) string {
	if len(requirements) == 0 {
		return "there are no requirements in this task doc"
	}
	ids := make([]string, 0, len(requirements))
	for _, r := range requirements {
		ids = append(ids, r.ID)
	}
	return strings.Join(ids, ", ")
}
