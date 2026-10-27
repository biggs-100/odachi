// Package role reads the source of a professional role: the one file a human
// edits, from which odachi renders a selectable agent and the role's skill.
//
// A role is not a voice and not a profile. The profile is the ceremony the
// binary enforces; the persona is how the agent talks (ADR 0006); a role is what
// a professional knows and how they work. It is content, so it is parsed and
// checked like content, never guessed at.
package role

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// The eight sections a role carries, in the order a human reads them. The first
// six are the professional; the seventh is the evidence; the eighth is how the
// role proves itself.
var Sections = []string{
	"Who it is",
	"What it asks before starting",
	"How it works",
	"What it refuses",
	"How it knows it worked",
	"What it does not know",
	"Where this comes from",
	"Probe",
}

// The limits, exported so the message a human reads and the rule that refuses
// the file cannot disagree. They are the skill's limits: a role renders a skill,
// and a role that cannot become a loadable skill is not a role.
const (
	// MaxDescription is the most characters a description may have.
	MaxDescription = 200
	// TriggerPrefix is how every description starts.
	TriggerPrefix = "Trigger:"
	// NamePattern is the one slug shape a directory and a name share.
	NamePattern = `^[a-z][a-z0-9-]{2,40}$`
)

var namePattern = regexp.MustCompile(NamePattern)

// Role is one parsed source. Path is how the file is named in messages.
type Role struct {
	Name           string
	Description    string
	Trigger        string
	Path           string
	HasFrontmatter bool
	// Bodies maps each canonical section to its text, trimmed. A section the
	// file does not carry is absent, which Problems reports.
	Bodies map[string]string
}

// Problem is one violation, attributed to the file a human has to open.
type Problem struct {
	File   string
	Reason string
}

// Parse reads one role source. It is total: malformed input never panics, it
// fails validation later.
func Parse(path, text string) Role {
	r := Role{Path: path, Bodies: map[string]string{}}

	front, body, ok := splitFrontmatter(text)
	if !ok {
		return r
	}
	r.HasFrontmatter = true
	r.Name = frontmatterValue(front, "name")
	r.Description = frontmatterValue(front, "description")
	r.Trigger = strings.TrimSpace(strings.TrimPrefix(r.Description, TriggerPrefix))
	r.Bodies = sectionBodies(body)
	return r
}

// Problems reports every violation in this source, one reason each.
func (r Role) Problems(slug string) []Problem {
	problems := []Problem{}
	add := func(format string, args ...any) {
		problems = append(problems, Problem{File: r.Path, Reason: fmt.Sprintf(format, args...)})
	}

	if !r.HasFrontmatter {
		return []Problem{{File: r.Path, Reason: "there is no frontmatter: the file must start with --- and close it with ---"}}
	}

	if r.Name == "" {
		add("the frontmatter does not say name")
	} else {
		if !namePattern.MatchString(r.Name) {
			add("the name %q does not match the shape %s (lowercase letters, numbers and dashes, starting with a letter)", r.Name, NamePattern)
		}
		if r.Name != slug {
			add("the name %q does not match the file %q", r.Name, slug)
		}
	}

	if r.Description == "" {
		add("the frontmatter does not say description")
	} else {
		if !strings.HasPrefix(r.Description, TriggerPrefix) {
			add("the description does not start with %q: nobody knows when this role is the one", TriggerPrefix)
		}
		if length := utf8.RuneCountInString(r.Description); length > MaxDescription {
			add("the description has %d characters; the maximum is %d", length, MaxDescription)
		}
	}

	for _, section := range Sections {
		switch body := r.Bodies[section]; {
		case body == "":
			add("the %q section is missing or empty", section)
		case strings.Contains(body, "TODO"):
			add("the %q section contains TODO: a role ships complete or it does not ship", section)
		}
	}

	return problems
}

// ValidSlug reports whether a slug can name a role.
func ValidSlug(slug string) bool {
	return namePattern.MatchString(slug)
}

// Template renders the source `odachi role new` starts from: the eight sections
// with the question each one answers.
func Template(slug string) string {
	var b strings.Builder
	b.WriteString("---\nname: " + slug + "\n")
	b.WriteString("description: " + TriggerPrefix + " (say when this professional is the one to load)\n---\n\n")
	b.WriteString("## Who it is\n\n(one paragraph, no adjectives: what this professional is responsible for)\n\n")
	b.WriteString("## What it asks before starting\n\n- (the question a senior would ask before touching the work)\n\n")
	b.WriteString("## How it works\n\n1. (the order of the method, not the tools)\n\n")
	b.WriteString("## What it refuses\n\n- (the thing a senior would not accept from a junior)\n\n")
	b.WriteString("## How it knows it worked\n\n- (the number and where it comes from)\n\n")
	b.WriteString("## What it does not know\n\n- (the boundary: what this role must not answer)\n\n")
	b.WriteString("## Where this comes from\n\n- (three to five sources with a date)\n\n")
	b.WriteString("## Probe\n\n(task: a small request. expected: what the role must ask or refuse)\n")
	return b.String()
}

// splitFrontmatter cuts the --- block from the body. A file without both fences
// has no frontmatter at all; a truncated one is not half-accepted.
func splitFrontmatter(text string) (front, body string, ok bool) {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	lines := strings.Split(text, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return "", "", false
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return strings.Join(lines[1:i], "\n"), strings.Join(lines[i+1:], "\n"), true
		}
	}
	return "", "", false
}

// frontmatterValue reads one `key: value` line, dropping the quotes.
func frontmatterValue(front, key string) string {
	for _, line := range strings.Split(front, "\n") {
		name, value, found := strings.Cut(line, ":")
		if !found || strings.TrimSpace(name) != key {
			continue
		}
		return unquote(value)
	}
	return ""
}

// unquote drops the quotes a YAML scalar may carry.
func unquote(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 && (value[0] == '"' || value[0] == '\'') && value[len(value)-1] == value[0] {
		return strings.TrimSpace(value[1 : len(value)-1])
	}
	return value
}

// sectionBodies reads every `## <heading>` block, keyed by the heading. A
// heading that is not one of the canonical sections is ignored: the file is the
// human's, and an extra note must not break the parse.
func sectionBodies(body string) map[string]string {
	bodies := map[string]string{}
	current := ""
	var lines []string

	flush := func() {
		if current != "" {
			bodies[current] = strings.TrimSpace(strings.Join(lines, "\n"))
		}
	}

	for _, line := range strings.Split(body, "\n") {
		heading, ok := strings.CutPrefix(strings.TrimSpace(line), "## ")
		if ok {
			flush()
			current = strings.TrimSpace(heading)
			lines = nil
			continue
		}
		if current != "" {
			lines = append(lines, line)
		}
	}
	flush()
	return bodies
}
