// Package agentstoml decodes and renders .odachi/agents.toml, the project's
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
		return "", fmt.Errorf("agents: TOML inválido: %w", err)
	}

	if unknown := md.Undecoded(); len(unknown) > 0 {
		keys := make([]string, 0, len(unknown))
		for _, key := range unknown {
			keys = append(keys, key.String())
		}
		return "", fmt.Errorf("agents: claves desconocidas: %s; el archivo solo configura agents", strings.Join(keys, ", "))
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
	return []byte("# odachi: qué clientes reciben el agente nativo; cambiar la elección no lo borra, solo deja de escribirlo\nagents = \"" + normalized + "\"\n"), nil
}
