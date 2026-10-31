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

// The four answers to "who decided where the craft goes", spelled once so a
// message and a test name the same thing. Two verbs weigh the same facts -`init`
// when it installs and `role` when it writes a role's skill- and they must not be
// able to disagree about the same project.
const (
	// SkillsFromFlag is a run that named the scope: the decision is the flag's.
	SkillsFromFlag = "flag"
	// SkillsFromStore is a project that already chose: the store's answer stands.
	SkillsFromStore = "store"
	// SkillsFromAdoption is the inference of a repository that already lives and
	// carries no copy of the craft: the machine is where it stays.
	SkillsFromAdoption = "adoption"
	// SkillsFromDefault is a project that never chose and is starting now.
	SkillsFromDefault = "default"
)

// SkillsScope is the three facts a scope is decided from: the choice the project
// recorded, the flag this run was given, and the adoption inference. The answer
// carries who decided it, because a human who reads "from the store" learns the
// flag is remembered and one who reads "the default" learns nothing was chosen.
type SkillsScope struct {
	// Stored is the store's skills key, or empty when the project never chose.
	Stored string
	// Flag is what this run asked for, or empty.
	Flag string
	// Adopted says this is a repository that already lives, that nobody chose a
	// scope for, and that carries no copy of the craft here.
	Adopted bool
}

// Resolve answers where the craft lives and who decided it, in the order the
// facts weigh: the flag, the store, the adoption, the default.
func (s SkillsScope) Resolve() (scope, source string) {
	switch {
	case strings.TrimSpace(s.Flag) != "":
		return strings.ToLower(strings.TrimSpace(s.Flag)), SkillsFromFlag
	case s.Stored != "":
		return s.Stored, SkillsFromStore
	case s.Adopted:
		return SkillsHome, SkillsFromAdoption
	default:
		return SkillsProject, SkillsFromDefault
	}
}

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
