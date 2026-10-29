# the-review-prompt
Base: c0ca30e44c53a3515b8a815fa7a1b24deb88785b

## Requirements (verbatim, from the user)
R1 · "the review prompt carries what the review can act on: every finding, and the clean rows of the paths the candidate changed, and the same questions are answered after the cut"

## Done when
- a prompt built for a candidate carries every finding and only the clean rows of the paths that changed
- the same three probes that answered 0101 are answered on a pruned prompt of the same tree, and the answers and times are recorded
- the suite pins the filter: findings stay, passes of untouched paths go
- the numbers and the decision land in the debt file, and the changelog says what changed

## Steps
- [x] T1 · the pruned prompt answers the probes of 0101, measured on the same tree   R1
- [ ] T2 · a prompt carries every finding and the clean rows of the changed paths   R1
- [ ] T3 · the debt file and the changelog say what the cut changes   R1

## History
L1 · "the review prompt carries what the review can act on: every finding, and the clean rows of the paths the candidate changed, and the same questions are answered after the cut"
