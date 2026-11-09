# a-project-writes-nothing
Base: a16fe652f4b1e7bd0f8ca6125769a491266dd26a

## Requirements (verbatim, from the user)
R1 · "no path writes a project agent file: the app's project half is deleted, and a stored `project_agents` value only names what an older release did"

## Done when
- neither `odachi init` nor the app's `Run` carries an option or a branch that writes a project agent file: `NoProjectAgents`, `resolveProjectAgents` and `writePersonas` with its helpers are gone
- a stored `project_agents = "write"` is still read, does nothing, and is not rewritten; `doctor` keeps naming a project agent file an older release left, with its path
- the tests that pinned the project mirror are deleted and the ones that pin the home door still pass
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- Unreleased names the change: a contract change, minor

## Steps
- [x] T1 · delete the project half: the option, the resolver and the writers, with their tests   R1 · route:inline · tree:b9c1356056e6087d66208d838fbf71948821725c
- [x] T2 · a stored project_agents stays readable and inert, and doctor keeps naming a project agent file a release left   R1 · route:inline · tree:b9c1356056e6087d66208d838fbf71948821725c
- [x] T3 · Unreleased and the impact map name the deletion as a contract change, minor   R1 · route:inline · tree:b9c1356056e6087d66208d838fbf71948821725c

## History
L1 · "no path writes a project agent file: the app's project half is deleted, and a stored `project_agents` value only names what an older release did"
