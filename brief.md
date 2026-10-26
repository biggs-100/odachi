# adopting-odachi-in

## What must exist
- adopting odachi in a real repo does not over-review what init wrote, does not leave a stale proof unexplained, and installs a reviewer that cannot change the tree

## Done when
- the files init installs (agent files and skills) are plumbing, not candidate: a tree that only changes them asks for no lens
- a stale proof names the paths that differ between the proof's tree and this one
- init installs a reviewer agent that cannot edit files or run commands, in both clients
- the README says how to run the review with that agent
- real code still asks for its lenses, so the gate does not go blind
- tests pin the passive plan, the named stale paths and the installed agent

## Not included (first version)
- reading ODD task docs without a number: a separate product decision
- changing the slug rule
- a shared or remote review store
- changing the lens vocabulary
- running reviews automatically
- the transaction store gentle-ai has

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: how to adopt a tool in a repo without charging the repo for the tool's own files.
