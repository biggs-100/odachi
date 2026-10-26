# adopting-odachi-in
Base: 5eca09489ed3ac93d13342ca2636ca255250febb

## Requirements (verbatim, from the user)
R1 · "adopting odachi in a real repo does not over-review what init wrote, does not leave a stale proof unexplained, and installs a reviewer that cannot change the tree"

## Done when
- the paths init writes are plumbing, not candidate: a plan commit that only adds them asks for no lens
- a stale proof names the paths that differ between the proof's tree and this one
- init installs a reviewer agent that cannot edit files or run commands, in both clients
- the README says how to run the review with that agent
- real code still asks for its lenses, so the gate does not go blind
- tests pin the passive plan, the named stale paths and the installed agent

## Steps
- [ ] T1 · the paths init writes are plumbing, not candidate, so the plan commit asks for no lens   R1
- [ ] T2 · a stale proof names the paths that differ, instead of only saying the code moved   R1
- [ ] T3 · init installs a read-only reviewer agent and the README says how to run it   R1

## History
L1 · "adopting odachi in a real repo does not over-review what init wrote, does not leave a stale proof unexplained, and installs a reviewer that cannot change the tree"
