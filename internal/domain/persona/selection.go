package persona

import (
	"fmt"
	"strings"
)

// SelectionPath is where a project records which clients receive the native
// agent. The file is optional: no file means both clients, the behavior every
// project had before the switch existed.
const SelectionPath = ".odachi/agents.toml"

// The three values `odachi init --agents` and the store know.
const (
	SelectionBoth     = "both"
	SelectionOpenCode = "opencode"
	SelectionClaude   = "claude"
)

// NormalizeSelection accepts the three values the switch knows. Matching is case
// insensitive and ignores surrounding space because a human writes the file;
// anything else is refused instead of guessed.
func NormalizeSelection(value string) (string, error) {
	switch normalized := strings.ToLower(strings.TrimSpace(value)); normalized {
	case SelectionBoth, SelectionOpenCode, SelectionClaude:
		return normalized, nil
	default:
		return "", fmt.Errorf("selección de clientes %q inválida: usá both, opencode o claude", strings.TrimSpace(value))
	}
}

// FilesForSelection returns the primary mirror of each client the selection
// names, in the table's order. It is the filtered view the verbs that refresh an
// existing agent read, so the same decision governs init and the refreshes.
func FilesForSelection(block, selection string) ([]File, error) {
	_, clients, err := ResolveSelection(selection)
	if err != nil {
		return nil, err
	}
	files := make([]File, 0, len(clients))
	for _, client := range clients {
		if file, ok := FilesFor(client.Agent, block); ok {
			files = append(files, file)
		}
	}
	return files, nil
}

// ResolveSelection turns a stored or typed value into the clients it names.
// Empty is the zero value and means both, the same as no file.
func ResolveSelection(value string) (string, []Client, error) {
	if strings.TrimSpace(value) == "" {
		value = SelectionBoth
	}
	normalized, err := NormalizeSelection(value)
	if err != nil {
		return "", nil, err
	}
	if normalized == SelectionBoth {
		return normalized, Clients(), nil
	}

	selected := make([]Client, 0, 1)
	for _, client := range clients {
		if client.Agent == normalized {
			selected = append(selected, client)
		}
	}
	return normalized, selected, nil
}
