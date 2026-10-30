package persona

import (
	"fmt"
	"strings"
)

// The two values the skills key knows: whether the craft lives in the project or
// once per machine in the home. The store keeps the answer because it outlives
// one run: a project that put its skills in the home must not get a project copy
// rebuilt by the next plain init.
const (
	// SkillsProject is the project's own copy, the answer every project had
	// before the key existed.
	SkillsProject = "project"
	// SkillsHome is one copy per machine, in the human's home.
	SkillsHome = "home"
)

// NormalizeSkills accepts the two values the key knows, and the empty string
// because absence is the default the tool always had: the project. Matching is
// case insensitive and ignores surrounding space; anything else is refused
// instead of guessed, so a typo cannot move the skills without saying so.
func NormalizeSkills(value string) (string, error) {
	switch normalized := strings.ToLower(strings.TrimSpace(value)); normalized {
	case "":
		return "", nil
	case SkillsProject, SkillsHome:
		return normalized, nil
	default:
		return "", fmt.Errorf("invalid skills scope %q: use project or home", strings.TrimSpace(value))
	}
}
