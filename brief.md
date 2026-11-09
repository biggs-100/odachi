# a-project-writes-nothing

## What must exist
- no path writes a project agent file: the app's project half is deleted, and a stored `project_agents` value only names what an older release did

## Done when
- neither `odachi init` nor the app's `Run` carries an option or a branch that writes a project agent file: `NoProjectAgents`, `resolveProjectAgents` and `writePersonas` with its helpers are gone
- a stored `project_agents = "write"` is still read, does nothing, and is not rewritten; `doctor` keeps naming a project agent file an older release left, with its path
- the tests that pinned the project mirror are deleted and the ones that pin the home door still pass
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- Unreleased names the change: a contract change, minor

## Not included (first version)
- no change to the bundled skill texts: the 19 keep their words - back when the lane needs a sentence they do not already say
- no change to the review or the classifier - back when a verdict has to read something the diff cannot show
- no change to odachi bench and no new verb - back when a bench journey crosses a commit, or a human asks for a verb
- no change to the stored proofs: never touched - never; they are the evidence
- no change to the design canvas, its token or the installers - back when an install path changes
- no change to the release workflow - back when a release must run a command the workflow does not run today
- no deletion of an agent file a project already has - back when a human asks for a migration verb; the file is named, never removed
- no change to `setup`, which already writes the home only - back when the home half moves

## Decisions taken
- It is used from the computer, by a person maintaining the tool.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: whether deleting the last project writer leaves the store's key harmless.
- A stored `project_agents = "write"` stays readable and inert; the clon's file is not rewritten by init.
