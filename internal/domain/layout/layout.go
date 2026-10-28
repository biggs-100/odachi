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
