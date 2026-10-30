// Package agentstoml decodes and renders agents.toml, in the clone's store, the project's
// choice of which clients receive the native agent and whether it carries them
// at all.
package agentstoml

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/schema"
)

// file mirrors the fields a human writes. ProjectAgents and Skills are optional:
// their absence is the project that never chose, which keeps the mirror in the
// tree and the craft in the project.
type file struct {
	// Schema is the version that wrote this store. It is not a human choice: a
	// store written before the key existed carries none and reads as the first
	// schema, and a value above this binary's is refused as newer, not broken.
	Schema        int    `toml:"schema"`
	Agents        string `toml:"agents"`
	ProjectAgents string `toml:"project_agents"`
	Skills        string `toml:"skills"`
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

	// The version is asked before the unknown keys are named: a store from a newer
	// odachi carries both, and the key that makes it newer is not a typo for the
	// human to fix.
	if _, err := schema.Read(f.Schema); err != nil {
		return persona.Store{}, fmt.Errorf("agents: %w", err)
	}

	if unknown := md.Undecoded(); len(unknown) > 0 {
		keys := make([]string, 0, len(unknown))
		for _, key := range unknown {
			keys = append(keys, key.String())
		}
		return persona.Store{}, fmt.Errorf("agents: unknown keys: %s; this file holds the client choice, where the agent lives and where the skills live", strings.Join(keys, ", "))
	}

	selection, err := persona.NormalizeSelection(f.Agents)
	if err != nil {
		return persona.Store{}, fmt.Errorf("agents: %w", err)
	}
	projectAgents, err := persona.NormalizeProjectAgents(f.ProjectAgents)
	if err != nil {
		return persona.Store{}, fmt.Errorf("agents: %w", err)
	}
	skills, err := persona.NormalizeSkills(f.Skills)
	if err != nil {
		return persona.Store{}, fmt.Errorf("agents: %w", err)
	}
	return persona.Store{Agents: selection, ProjectAgents: projectAgents, Skills: skills}, nil
}

// Render writes the canonical store. The comments are part of the artifact: the
// human who opens it must see that skipping a client never deletes its files,
// and that the project choice decides where the agent lives while the skills
// choice decides where the craft lives.
func Render(store persona.Store) ([]byte, error) {
	normalized, err := persona.NormalizeSelection(store.Agents)
	if err != nil {
		return nil, fmt.Errorf("agents: %w", err)
	}
	projectAgents, err := persona.NormalizeProjectAgents(store.ProjectAgents)
	if err != nil {
		return nil, fmt.Errorf("agents: %w", err)
	}
	skills, err := persona.NormalizeSkills(store.Skills)
	if err != nil {
		return nil, fmt.Errorf("agents: %w", err)
	}

	text := "# odachi: the schema that wrote this store; an older binary says upgrade instead of calling it broken\n" +
		"schema = " + strconv.Itoa(schema.Current) + "\n" +
		"# odachi: which clients receive the native agent; changing the choice does not delete it, it only stops writing it\n" +
		"agents = \"" + normalized + "\"\n"
	if projectAgents != "" {
		text += "# odachi: project_agents = none keeps the agent in the home; write puts the project's own copy back\n" +
			"project_agents = \"" + projectAgents + "\"\n"
	}
	if skills != "" {
		text += "# odachi: skills = home keeps one copy of the craft per machine; project puts this project's own copy back\n" +
			"skills = \"" + skills + "\"\n"
	}
	return []byte(text), nil
}
