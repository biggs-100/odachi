# the-default-branch
Base: 8e9d26d1c61ac23cc726b4c15f6af8e6760b12b8

## Requirements (verbatim, from the user)
R1 · "the default branch answers without a branch name: the local branch that follows the base, and whether HEAD is one of its commits, so a detached checkout is refused or judged for the same reason a named one is"

## Done when
- a trailer whose HEAD is detached at a commit of the local default branch is refused by name, not judged
- the CI-shaped checkout, with a detached HEAD and no local default branch, still reads the trailer: the document-less lane keeps working on a pull request
- a detached HEAD on a feature branch keeps the trailer as the vision
- when git cannot name the branch the base follows, the gate blocks with the reason instead of dropping the error
- tests pin the four cases in gitx and the refusal in the lane, and go test ./... is green

## Steps
- [x] T1 · gitx answers whether HEAD is on the branch that follows the base, from the commits, not from a name   R1 · route:inline · tree:69f551d13b41ac8373f3e69cdd64f34b65f12c0b
- [x] T2 · the gate asks that question and blocks when it cannot answer; the lane pins the detached refusal   R1 · route:inline · tree:130fc3ea5bf3663b789a661469bf508701ba2804
- [x] T3 · the ledger row closes and the changelog names the rule   R1 · route:inline · tree:4a1c405856e68c0ea2883330fc4f687eebd8125e

## History
L1 · "the default branch answers without a branch name: the local branch that follows the base, and whether HEAD is one of its commits, so a detached checkout is refused or judged for the same reason a named one is"
L2 · OFICIO · "the fence keeps the proof rule the last feature carried: proof/ is not this feature's to touch" (porque start regenerates the fence from this brief, and a standing protection dropped by omission is the defect 0179 paid; the rule protects a path, not a reason)
