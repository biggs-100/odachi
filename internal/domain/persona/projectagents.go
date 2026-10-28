package persona

import (
	"fmt"
	"strings"
)

// The two values the project-agents key knows: whether the native agent files
// live in the project, or only in the home. Absence is not a third value a human
// types: it means the project never chose, and the default stays the mirror in
// the tree, the behavior every project had before the key existed.
const (
	ProjectAgentsWrite = "write"
	ProjectAgentsNone  = "none"
)

// Store is what the clone's agents.toml holds: which clients receive the native
// agent, and whether the project carries them at all. Both answers live in the
// same file because they are the same decision - where the agent comes from.
type Store struct {
	Agents        string
	ProjectAgents string
}

// NormalizeProjectAgents accepts the two values the key knows, and the empty
// string because absence is the default. Matching is case insensitive and
// ignores surrounding space; anything else is refused instead of guessed.
func NormalizeProjectAgents(value string) (string, error) {
	switch normalized := strings.ToLower(strings.TrimSpace(value)); normalized {
	case "":
		return "", nil
	case ProjectAgentsWrite, ProjectAgentsNone:
		return normalized, nil
	default:
		return "", fmt.Errorf("invalid project agents %q: use write or none", strings.TrimSpace(value))
	}
}
