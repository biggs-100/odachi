# closing-a-feature
Base: c1398cef52c22969cba04a1cbfbf206c1068de2c

## Requirements (verbatim, from the user)
R1 · "closing a feature is reviewed and publishing it is not: the verdict is owed by the commit that marks the last step done, and a later documents-only tree does not ask for it again"

## Done when
- with the mode on, the commit that closes a feature owes the verdict of the feature's base
- a documents-only tree after the close is ready with its proof, as long as the task doc did not change
- a tree with code after the close still owes a verdict: the gate does not go blind
- a feature with no steps or no base keeps the old per-tree rule
- --require-review still forces one tree and --no-review still turns the mode off for one run
- tests pin the four cases and the README says the review is owed by the close, not by the tail

## Steps
- [x] T1 · the gate asks only when the task doc closes the feature, or when the tree carries code after the close   R1 · 451712c
- [ ] T2 · the README says the review is owed by the close, and the cases are pinned by tests   R1

## History
L1 · "closing a feature is reviewed and publishing it is not: the verdict is owed by the commit that marks the last step done, and a later documents-only tree does not ask for it again"
L2 · OFICIO · "The close is the change that marked the last step, read from HEAD; a documents-only tree after it does not reopen the gate" (porque a release note does not change the code the verdict judged, and asking again charges the same bytes twice; code after the close still owes, so the gate never goes blind)
L3 · OFICIO · "The README says a note after the close is ready with its proof, and code after the close owes a verdict" (porque the rule has two halves and a human who only reads the first will ship code unreviewed; the second half is what keeps the gate honest)
