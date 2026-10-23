// Package profiletoml decodes .odachi/profile.toml into a profile.Profile.
//
// This is the only place TOML meets the domain: the adapter owns the file
// format's key names, and the domain owns the eight fields and their allowed
// values. Decoding is strict on purpose. A misspelled key must be an error, not
// a silently ignored line, because a human who writes "gates_disabled = true"
// must be told the tool will never do that.
package profiletoml

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/biggs-100/odachi/internal/domain/profile"
)

// file mirrors profile.Profile one field per line, with the TOML key names the
// human writes. The adapter keeps this lower-case bridge so the domain struct
// stays format-agnostic.
type file struct {
	Name        string `toml:"name"`
	Extends     string `toml:"extends"`
	BriefSource string `toml:"brief_source"`
	Questions   string `toml:"questions"`
	Assumptions string `toml:"assumptions"`
	Output      string `toml:"output"`
	Proof       string `toml:"proof"`
	Explain     string `toml:"explain"`
}

// Decode parses the file bytes. Unknown keys are an error naming every
// offending key; the decoded profile is then validated by the domain, which
// stays the single source of truth for the allowed values.
func Decode(data []byte) (profile.Profile, error) {
	var f file
	md, err := toml.Decode(string(data), &f)
	if err != nil {
		return profile.Profile{}, fmt.Errorf("profile: invalid TOML: %w", err)
	}

	if unknown := md.Undecoded(); len(unknown) > 0 {
		keys := make([]string, 0, len(unknown))
		for _, key := range unknown {
			keys = append(keys, key.String())
		}
		return profile.Profile{}, fmt.Errorf(
			"profile: unknown keys: %s; a profile can only configure %s",
			strings.Join(keys, ", "), strings.Join(knownKeys(), ", "))
	}

	p := profile.Profile{
		Name:        f.Name,
		Extends:     f.Extends,
		BriefSource: f.BriefSource,
		Questions:   f.Questions,
		Assumptions: f.Assumptions,
		Output:      f.Output,
		Proof:       f.Proof,
		Explain:     f.Explain,
	}
	if err := profile.Validate(p); err != nil {
		return profile.Profile{}, fmt.Errorf("profile: %w", err)
	}
	return p, nil
}

// Load reads path and delegates to Decode.
func Load(path string) (profile.Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return profile.Profile{}, fmt.Errorf("profile: could not read %s: %w", path, err)
	}
	return Decode(data)
}

// knownKeys lists the configurable axes in canonical order, for the error
// message a human reads after a typo.
func knownKeys() []string {
	return []string{
		"name", "extends", "brief_source", "questions",
		"assumptions", "output", "proof", "explain",
	}
}
