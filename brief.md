# the-answer-belongs

## What must exist
- the answer belongs to the bytes it judged: every group is filed under the tree the prompt named, and a reviewer that changes the worktree stops the run instead of refiling it

## Done when
- every group's verdict is filed under the tree the prompt named, even when a reviewer left a different worktree
- a reviewer that changed the worktree stops the run: it names the tree the answer belongs to, what landed, and that the new bytes ask for their own review
- check and review --show keep naming the tree they compare, and the gate's rule does not change
- a test drives a reviewer that changes the tree mid-run: one file for the named tree, every lens in it, and no file for the tree the reviewer left
- the README says the run is bound to one candidate.

## Not included (first version)
- copying the worktree or a sandbox per reviewer
- a transaction store with replay, CAS or locks
- resuming a killed process
- committing, restoring or reverting the reviewer's edits
- changing the reviewer command, the lens prompts or the gate's rule

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: how to hold a candidate still while something else can edit it.
