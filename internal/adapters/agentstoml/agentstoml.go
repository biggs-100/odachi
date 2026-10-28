// Package agentstoml decodes and renders agents.toml, in the clone's store, the project's
// choice of which clients receive the native agent.
package agentstoml

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/biggs-100/odachi/internal/domain/persona"
)

// file mirrors the one field a human writes.
type file struct {
	Agents string `toml:"agents"`
}

// Decode parses the one-key store. Unknown keys and unknown values are errors,
// because a store the tool half-understands would silently pick a client.
func Decode(data []byte) (string, error) {
	var f file
	md, err := toml.Decode(string(data), &f)
	if err != nil {
		return "", fmt.Errorf("agents: invalid TOML: %w", err)
	}

	if unknown := md.Undecoded(); len(unknown) > 0 {
		keys := make([]string, 0, len(unknown))
		for _, key := range unknown {
			keys = append(keys, key.String())
		}
		return "", fmt.Errorf("agents: unknown keys: %s; this file only configures agents", strings.Join(keys, ", "))
	}

	selection, err := persona.NormalizeSelection(f.Agents)
	if err != nil {
		return "", fmt.Errorf("agents: %w", err)
	}
	return selection, nil
}

// Render writes the canonical store. The comment is part of the artifact: the
// human who opens it must see that skipping a client never deletes its files.
func Render(selection string) ([]byte, error) {
	normalized, err := persona.NormalizeSelection(selection)
	if err != nil {
		return nil, fmt.Errorf("agents: %w", err)
	}
	return []byte("# odachi: which clients receive the native agent; changing the choice does not delete it, it only stops writing it\nagents = \"" + normalized + "\"\n"), nil
}
