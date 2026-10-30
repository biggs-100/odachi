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
// agent, whether the project carries them at all, and where the craft skills
// live. The three answers share one file because they are one decision - where
// the agent and the skills it reads come from - and because the next plain init
// has to honour all three.
type Store struct {
	Agents        string
	ProjectAgents string
	// Skills is where the craft lives: persona.SkillsProject, persona.SkillsHome,
	// or the empty string for a store that never chose, which keeps the project.
	Skills string
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
