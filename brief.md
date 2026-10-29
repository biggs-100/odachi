# the-review-prompt

## What must exist
- the review prompt carries what the review can act on: every finding, and the clean rows of the paths the candidate changed, and the same questions are answered after the cut

## Done when
- a prompt built for a candidate carries every finding and only the clean rows of the paths that changed
- the same three probes that answered 0101 are answered on a pruned prompt of the same tree, and the answers and times are recorded
- the suite pins the filter: findings stay, passes of untouched paths go
- the numbers and the decision land in the debt file, and the changelog says what changed

## Not included (first version)
- no change to the verdict shape
- no change to the proof format, and the stored proof is never rewritten

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: whether less memory changes what a reviewer answers.
