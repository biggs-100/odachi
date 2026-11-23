// Package profile is the only place ceremony is configured. A profile may
// change how much the tool talks, how it asks, and what counts as proof. It may
// never change whether a gate exists: this struct has eight fields, and a
// reflection test fails the build if a ninth appears (see Contract B in
// docs/odachi-bootstrap.md).
package profile

import (
	"fmt"
	"slices"
	"strings"
)

// Profile mirrors odd/profile.toml one field per line. An empty field means
// "inherit from Extends".
type Profile struct {
	Name        string
	Extends     string
	BriefSource string
	Questions   string
	Assumptions string
	Output      string
	Proof       string
	Explain     string
}

// Facts is what the tool can observe before asking anything.
type Facts struct {
	HasBrief bool
	Machine  bool
}

// enumField is one configurable axis: its TOML name, its reader, its writer, and
// the values it may hold. The TOML name appears in errors because the human edits
// TOML.
type enumField struct {
	name    string
	allowed []string
	value   func(Profile) string
	assign  func(*Profile, string)
}

var enums = []enumField{
	{"brief_source", []string{"interview", "repo", "file"}, func(p Profile) string { return p.BriefSource }, func(p *Profile, v string) { p.BriefSource = v }},
	{"questions", []string{"one-by-one", "batched", "all"}, func(p Profile) string { return p.Questions }, func(p *Profile, v string) { p.Questions = v }},
	{"assumptions", []string{"never", "declare", "prefer-safe"}, func(p Profile) string { return p.Assumptions }, func(p *Profile, v string) { p.Assumptions = v }},
	{"output", []string{"line", "detail", "json"}, func(p Profile) string { return p.Output }, func(p *Profile, v string) { p.Output = v }},
	{"proof", []string{"command", "suite-or-ci"}, func(p Profile) string { return p.Proof }, func(p *Profile, v string) { p.Proof = v }},
	{"explain", []string{"why", "none"}, func(p Profile) string { return p.Explain }, func(p *Profile, v string) { p.Explain = v }},
}

// Axis is one configurable field of a profile: its TOML name, the values it may
// hold, and what the profile in force says today. The editor reads this list, so
// the axes live here and not in a screen (0336).
type Axis struct {
	Name    string
	Allowed []string
	Value   string
}

// Axes lists the configurable fields in the order odd/profile.toml prints them.
func Axes(p Profile) []Axis {
	axes := make([]Axis, 0, len(enums))
	for _, field := range enums {
		axes = append(axes, Axis{Name: field.name, Allowed: slices.Clone(field.allowed), Value: field.value(p)})
	}
	return axes
}

// Set names one axis and gives it a value, refusing an axis nobody declares and a
// value the axis does not allow. The rest of the profile is untouched: an editor
// changes one field and the file keeps what the human already chose (0336).
func Set(p Profile, name, value string) (Profile, error) {
	for _, field := range enums {
		if field.name != name {
			continue
		}
		if value != "" && !slices.Contains(field.allowed, value) {
			return Profile{}, fmt.Errorf("%s: %q is not an allowed value; allowed values: %s", name, value, strings.Join(field.allowed, ", "))
		}
		field.assign(&p, value)
		if err := Validate(p); err != nil {
			return Profile{}, err
		}
		return p, nil
	}
	return Profile{}, fmt.Errorf("unknown field: %q; the fields are %s", name, strings.Join(fieldNames(), ", "))
}

// fieldNames lists the configurable fields, sorted, for an error a human can act
// on: the name that did not exist leads to the names that do (0336).
func fieldNames() []string {
	names := make([]string, 0, len(enums))
	for _, field := range enums {
		names = append(names, field.name)
	}
	slices.Sort(names)
	return names
}

// Validate rejects an unknown value, naming the field, the reject and the
// allowed values. Empty is valid: it means inherited.
func Validate(p Profile) error {
	for _, field := range enums {
		value := field.value(p)
		if value == "" {
			continue
		}
		if !slices.Contains(field.allowed, value) {
			return fmt.Errorf("%s: %q is not an allowed value; allowed values: %s", field.name, value, strings.Join(field.allowed, ", "))
		}
	}
	return nil
}

// Resolve walks the extends chain and returns the profile with every inherited
// field filled in. The child wins over the parent; an unknown name and a cycle
// are errors.
func Resolve(known map[string]Profile, name string) (Profile, error) {
	if _, ok := known[name]; !ok {
		return Profile{}, fmt.Errorf("unknown profile: %q; available profiles: %s", name, strings.Join(names(known), ", "))
	}

	var chain []Profile
	var path []string
	seen := map[string]bool{}
	current := name

	for {
		if seen[current] {
			return Profile{}, fmt.Errorf("cycle between profiles: %s", strings.Join(append(path, current), " -> "))
		}
		seen[current] = true
		path = append(path, current)

		p, ok := known[current]
		if !ok {
			return Profile{}, fmt.Errorf("profile %q extends %q, which does not exist", path[len(path)-2], current)
		}
		chain = append(chain, p)

		if p.Extends == "" {
			break
		}
		current = p.Extends
	}

	// Root first, so each child overwrites what it inherited.
	resolved := Profile{}
	for i := len(chain) - 1; i >= 0; i-- {
		merge(&resolved, chain[i])
	}
	resolved.Name = name

	if err := Validate(resolved); err != nil {
		return Profile{}, err
	}
	return resolved, nil
}

// Infer picks a profile from what the tool can see, never from what it assumes.
//
// A brief is the one signal that the human already defined a project; a git
// repository is not, because `odachi init` creates one before `start` decides the
// profile. Inferring senior from `.git` turned every new beginner project into a
// senior one — exactly backwards. Senior is earned with `odachi level --to`, or
// inferred from a brief the human brought.
func Infer(f Facts) string {
	switch {
	case f.Machine:
		return "ci"
	case f.HasBrief:
		return "senior"
	default:
		return "novato"
	}
}

// Builtin is the shipped chain: novato -> senior -> ci. Every value here only
// changes ceremony; none of them can switch a gate off.
func Builtin() map[string]Profile {
	return map[string]Profile{
		"novato": {
			Name:        "novato",
			Extends:     "",
			BriefSource: "interview",
			Questions:   "one-by-one",
			Assumptions: "never",
			Output:      "line",
			Proof:       "command",
			Explain:     "why",
		},
		"senior": {
			Name:        "senior",
			Extends:     "novato",
			BriefSource: "repo",
			Questions:   "batched",
			Assumptions: "declare",
			Output:      "detail",
			Proof:       "suite-or-ci",
			Explain:     "none",
		},
		"ci": {
			Name:        "ci",
			Extends:     "senior",
			BriefSource: "file",
			Questions:   "all",
			Assumptions: "never",
			Output:      "json",
			Proof:       "suite-or-ci",
			Explain:     "none",
		},
		// office-peru speaks to a public office worker who writes documents,
		// research and reports, not to a developer. It talks more, asks one
		// question at a time and keeps command proof because the gate reads a
		// command's output; the Peruvian legal knowledge lives in its skill (0354).
		"office-peru": {
			Name:        "office-peru",
			Extends:     "novato",
			BriefSource: "interview",
			Questions:   "one-by-one",
			Assumptions: "never",
			Output:      "detail",
			Proof:       "command",
			Explain:     "why",
		},
	}
}

// merge copies every non-empty field of src into dst.
func merge(dst *Profile, src Profile) {
	setIfNotEmpty(&dst.Name, src.Name)
	setIfNotEmpty(&dst.Extends, src.Extends)
	setIfNotEmpty(&dst.BriefSource, src.BriefSource)
	setIfNotEmpty(&dst.Questions, src.Questions)
	setIfNotEmpty(&dst.Assumptions, src.Assumptions)
	setIfNotEmpty(&dst.Output, src.Output)
	setIfNotEmpty(&dst.Proof, src.Proof)
	setIfNotEmpty(&dst.Explain, src.Explain)
}

func setIfNotEmpty(dst *string, value string) {
	if value != "" {
		*dst = value
	}
}

// names returns the known profile names sorted, for stable error messages.
func names(known map[string]Profile) []string {
	out := make([]string, 0, len(known))
	for name := range known {
		out = append(out, name)
	}
	slices.Sort(out)
	return out
}
