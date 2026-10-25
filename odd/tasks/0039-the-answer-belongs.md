# the-answer-belongs

## Requirements (verbatim, from the user)
R1 · "the answer belongs to the bytes it judged: every group is filed under the tree the prompt named, and a reviewer that changes the worktree stops the run instead of refiling it"

## Done when
- every group's verdict is filed under the tree the prompt named, even when a reviewer left a different worktree
- a reviewer that changed the worktree stops the run: it names the tree the answer belongs to, what landed, and that the new bytes ask for their own review
- check and review --show keep naming the tree they compare, and the gate's rule does not change
- a test drives a reviewer that changes the tree mid-run: one file for the named tree, every lens in it, and no file for the tree the reviewer left
- the README says the run is bound to one candidate.

## Steps
- [ ] T1 · a moved worktree stops the run and says where the answer lives and that the new bytes ask for their own review   R1
- [ ] T2 · a moved worktree stops the run, says where the answer lives, and the new bytes ask for their own review   R1

## History
L1 · "the answer belongs to the bytes it judged: every group is filed under the tree the prompt named, and a reviewer that changes the worktree stops the run instead of refiling it"
L2 · OFICIO · "The run holds the candidate it named and files every group under it, instead of re-reading the worktree before each write" (porque a reviewer's edits must not move an answer out from under the tree that was judged; gentle-ai freezes the candidate, and odachi only needs one value, read once)
