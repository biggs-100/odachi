// Package layout names every path odachi writes inside the project.
//
// There used to be a `.odachi/` folder that meant "the tool's", and it mixed two
// owners: the fence and the profile are the team's decisions — they travel with
// the repository and a teammate obeys them — while verdicts, versions and the
// client selection are the clone's and live in the store.
//
// So the worktree keeps what the team reads, next to the plan: `odd/`. The store
// keeps what only this clone produced. One spelling each, because a path written
// twice is a path that drifts.
package layout

import "path/filepath"

const (
	// Odd is the folder the team reads: the plan and the rules.
	Odd = "odd"
	// Tasks is the plan: one document per feature.
	Tasks = "odd/tasks"
	// Roles is the team's knowledge: one document per professional role.
	Roles = "odd/roles"

	// Brief is the vision of the feature being built.
	Brief = "brief.md"
	// Fence is what the project refuses to do.
	Fence = "odd/fence.toml"
	// Profile is the ceremony and the voice: how the agent talks to the human.
	Profile = "odd/profile.toml"
	// Review is the review switch and the lens table.
	Review = "odd/review.toml"
)

// LegacyDir is the folder older releases used for everything: the rules, the
// store and the lock. It is read only to point a project at today's layout.
const LegacyDir = ".odachi"

// LegacyRules is the tracked half of the old folder, in the order a human reads
// the move: the profile, the fence, the review table, and the roles folder.
// Verdicts, versions and the client choice were never the team's and are adopted
// by `odachi store move`.
var LegacyRules = []string{Profile, Fence, Review, Roles}

// LegacyMoves is the git commands that move one old rule into today's layout.
// present are the old paths that still exist (`.odachi/fence.toml`, ...). The
// tool prints these and never runs them: a tracked file is moved by git and
// committed by the human, so a migration cannot rewrite somebody's history
// behind their back.
func LegacyMoves(present []string) []string {
	in := make(map[string]bool, len(present))
	for _, path := range present {
		in[path] = true
	}

	moves := make([]string, 0, len(present))
	for _, to := range LegacyRules {
		from := filepath.ToSlash(filepath.Join(LegacyDir, filepath.Base(to)))
		if in[from] {
			moves = append(moves, "git mv "+from+" "+to)
		}
	}
	return moves
}
