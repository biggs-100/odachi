# the-gate-asks

## Requirements (verbatim, from the user)
R1 · "the gate asks once per feature, not once per commit: a feature with open steps needs only its proof, and the tree that closes it needs the one verdict that judged the feature's diff"

## Done when
- the task doc records the commit the feature started from, and start writes it
- with the project mode on, a tree whose feature still has open steps is ready with its proof alone
- the tree that closes a feature needs a verdict, and the verdict must name the feature's base
- the blocked message names the runnable review with the recorded base
- --require-review still forces the gate for one tree and --no-review still turns it off for one run
- a feature with no steps, or a document with no base, keeps the old rule: the mode asks for a verdict on every tree
- tests pin the five cases and the README says the review unit is the feature

## Steps
- [x] T1 · the task doc records the commit the feature started from, and start writes it from the repository's HEAD   R1 · e427880
- [x] T2 · the gate asks for one verdict when the feature closes, and the verdict has to name the feature's base   R1 · 8d4cbae

## History
L1 · "the gate asks once per feature, not once per commit: a feature with open steps needs only its proof, and the tree that closes it needs the one verdict that judged the feature's diff"
L2 · OFICIO · "The task doc records the commit the feature started from, and the CLI reads it from HEAD" (porque the gate cannot name the review of a feature it cannot identify: a sha in the document is visible, runnable, and survives the closing commit that only carries the doc)
L3 · OFICIO · "The mode asks once per feature: an open step owes its proof, and the closing tree owes the verdict of the feature's base" (porque a review per commit re-judges the same diff and pays the same model again; the feature is the unit of work, and the base in the document is what makes the one verdict runnable and exact)
