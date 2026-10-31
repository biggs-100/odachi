// Package adoption owns the one inference odachi draws about a project that
// already lives: a repository with a commit and no copy of the craft here keeps
// the craft in the machine, and one with no native agent file of ours keeps its
// agent there too. The two halves are inferred separately, because a project that
// carries its copy can still keep the agent in the machine.
//
// Three verbs live on this: `init` writes the project, `role` writes a role's
// skill, and `doctor` reads the shape and says what it is. They must not answer
// differently about the same project, so the answer is here and they ask it.
//
// The facts are gathered by the caller: every verb has its own ports, and its own
// words for a read that failed. What the facts mean together is this package.
package adoption

import "github.com/biggs-100/odachi/internal/domain/persona"

// Facts is what the inference reads. None of it is a choice: the two stored keys
// are the project's, and the three booleans are what the disk answers today.
type Facts struct {
	// Lives says this repository already has a commit: it is a project that
	// already lives, and adopting odachi there is not the same job as starting one.
	Lives bool
	// CraftHere says the project carries the copy of the craft (.agents/skills).
	// The copy, not the folder: `.agents/` belongs to other tools too.
	CraftHere bool
	// AgentHere says the project carries the native agent file a client reads. The
	// file, not the client's folder.
	AgentHere bool
	// Stored is the clone's store, read and never guessed.
	Stored persona.Store
}

// Craft says the craft stays in the machine: the project already lives, nobody
// chose a scope, and no copy is here. An inference does not move a project that
// already carries the craft.
func (f Facts) Craft() bool {
	return f.Lives && f.Stored.Skills == "" && !f.CraftHere
}

// Agent says the same for the native agent, on its own facts: a project that keeps
// its copy of the craft can still keep the agent in the machine, and a recorded
// choice stands over the inference.
func (f Facts) Agent() bool {
	return f.Lives && f.Stored.ProjectAgents == "" && !f.AgentHere
}

// Scope is where the craft lives and who decided it, with the flag this run was
// given. It defers to persona: the four answers weigh the same there as they do in
// every message that names them.
func (f Facts) Scope(flag string) (scope, source string) {
	return persona.SkillsScope{Stored: f.Stored.Skills, Flag: flag, Adopted: f.Craft()}.Resolve()
}

// Reader is the facts as questions a verb answers with its own ports. A verb wires
// only what its job needs; a question left nil answers false, which is what a
// project with no history and no copy looks like.
type Reader struct {
	Lives     func() (bool, error)
	CraftHere func() (bool, error)
	AgentHere func() (bool, error)
	Stored    func() (persona.Store, error)
}

// Read asks the questions in order and stops at the first one that fails: a read
// that failed is not a fact that is false, and a footprint decided from it installs
// the wrong thing. The error is the caller's own, so the refusal names its path.
func Read(r Reader) (Facts, error) {
	var facts Facts
	var err error

	if r.Lives != nil {
		if facts.Lives, err = r.Lives(); err != nil {
			return Facts{}, err
		}
	}
	if r.CraftHere != nil {
		if facts.CraftHere, err = r.CraftHere(); err != nil {
			return Facts{}, err
		}
	}
	if r.AgentHere != nil {
		if facts.AgentHere, err = r.AgentHere(); err != nil {
			return Facts{}, err
		}
	}
	if r.Stored != nil {
		if facts.Stored, err = r.Stored(); err != nil {
			return Facts{}, err
		}
	}
	return facts, nil
}
