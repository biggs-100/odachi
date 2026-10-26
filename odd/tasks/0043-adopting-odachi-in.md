# adopting-odachi-in
Base: 5eca09489ed3ac93d13342ca2636ca255250febb

## Requirements (verbatim, from the user)
R1 · "adopting odachi in a real repo does not over-review what init wrote, does not leave a stale proof unexplained, and installs a reviewer that cannot change the tree"

## Done when
- the files init installs (agent files and skills) are plumbing, not candidate: a tree that only changes them asks for no lens
- a stale proof names the paths that differ between the proof's tree and this one
- init installs a reviewer agent that cannot edit files or run commands, in both clients
- the README says how to run the review with that agent
- real code still asks for its lenses, so the gate does not go blind
- tests pin the passive plan, the named stale paths and the installed agent

## Steps
- [x] T1 · the paths init writes are plumbing, not candidate, so the plan commit asks for no lens   R1 · 02f5dc6
- [x] T2 · a stale proof names the paths that differ, instead of only saying the code moved   R1 · 63fc5b6
- [ ] T3 · init installs a read-only reviewer agent and the README says how to run it   R1

## History
L1 · "adopting odachi in a real repo does not over-review what init wrote, does not leave a stale proof unexplained, and installs a reviewer that cannot change the tree"
L2 · OFICIO · "The rows of the files init installs are not read for content signals, and their paths are plumbing" (porque a skill that mentions secrets is the tool's own text, not a security change in the adopting repo; the adoption must not cost four lenses before the first line of code)
L3 · OFICIO · "The proof gate names the paths that differ between the proof's tree and this one, and never blocks on the diff" (porque a stray backup, an editor and a real edit look identical in one line, and the generic message makes the human audit the whole tree; if git cannot read the old tree, the generic message stands)
L4 · OFICIO · "init installs odachi-reviewer, which cannot read, run or edit, and the README runs the review with it" (porque a prompt rule is a request; a client that can deny the tools makes it a capability, and the reviewer never needs the live tree because the candidate is frozen in the prompt)
