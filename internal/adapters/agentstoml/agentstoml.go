// Package agentstoml decodes and renders agents.toml, in the clone's store, the project's
// choice of which clients receive the native agent and whether it carries them
// at all.
package agentstoml

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/biggs-100/odachi/internal/domain/persona"
)

// file mirrors the fields a human writes. ProjectAgents is optional: its absence
// is the project that never chose, which keeps the mirror in the tree.
type file struct {
	Agents        string `toml:"agents"`
	ProjectAgents string `toml:"project_agents"`
}

// Decode parses the store. Unknown keys and unknown values are errors, because a
// store the tool half-understands would silently pick a client or silently put
// the agent back in the tree.
func Decode(data []byte) (persona.Store, error) {
	var f file
	md, err := toml.Decode(string(data), &f)
	if err != nil {
		return persona.Store{}, fmt.Errorf("agents: invalid TOML: %w", err)
	}

	if unknown := md.Undecoded(); len(unknown) > 0 {
		keys := make([]string, 0, len(unknown))
		for _, key := range unknown {
			keys = append(keys, key.String())
		}
		return persona.Store{}, fmt.Errorf("agents: unknown keys: %s; this file only configures agents", strings.Join(keys, ", "))
	}

	selection, err := persona.NormalizeSelection(f.Agents)
	if err != nil {
		return persona.Store{}, fmt.Errorf("agents: %w", err)
	}
	projectAgents, err := persona.NormalizeProjectAgents(f.ProjectAgents)
	if err != nil {
		return persona.Store{}, fmt.Errorf("agents: %w", err)
	}
	return persona.Store{Agents: selection, ProjectAgents: projectAgents}, nil
}

// Render writes the canonical store. The comments are part of the artifact: the
// human who opens it must see that skipping a client never deletes its files,
// and that the project choice decides where the agent lives.
func Render(store persona.Store) ([]byte, error) {
	normalized, err := persona.NormalizeSelection(store.Agents)
	if err != nil {
		return nil, fmt.Errorf("agents: %w", err)
	}
	projectAgents, err := persona.NormalizeProjectAgents(store.ProjectAgents)
	if err != nil {
		return nil, fmt.Errorf("agents: %w", err)
	}

	text := "# odachi: which clients receive the native agent; changing the choice does not delete it, it only stops writing it\n" +
		"agents = \"" + normalized + "\"\n"
	if projectAgents != "" {
		text += "# odachi: project_agents = none keeps the agent in the home; write puts the project's own copy back\n" +
			"project_agents = \"" + projectAgents + "\"\n"
	}
	return []byte(text), nil
}
