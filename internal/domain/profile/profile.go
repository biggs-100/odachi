// Package profile is the only place ceremony is configured. A profile may
// change how much the tool talks, how it asks, and what counts as proof. It may
// never change whether a gate exists: this struct has eight fields, and a
// reflection test fails the build if a ninth appears (see Contract B in
// odd/tasks/odachi-bootstrap.md).
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

// enumField is one configurable axis: its TOML name, its reader, and the values
// it may hold. The TOML name appears in errors because the human edits TOML.
type enumField struct {
	name    string
	allowed []string
	value   func(Profile) string
}

var enums = []enumField{
	{"brief_source", []string{"interview", "repo", "file"}, func(p Profile) string { return p.BriefSource }},
	{"questions", []string{"one-by-one", "batched", "all"}, func(p Profile) string { return p.Questions }},
	{"assumptions", []string{"never", "declare", "prefer-safe"}, func(p Profile) string { return p.Assumptions }},
	{"output", []string{"line", "detail", "json"}, func(p Profile) string { return p.Output }},
	{"proof", []string{"command", "suite-or-ci"}, func(p Profile) string { return p.Proof }},
	{"explain", []string{"why", "none"}, func(p Profile) string { return p.Explain }},
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
