# a-closing-tree
Base: acf703d72c8d74778a567aec7548118027098a2a

## Requirements (verbatim, from the user)
R1 · "a closing tree that asks for no lens owes no verdict: the gate asks the same classifier the review asks, and a feature whose whole diff is documents closes without paying for a review"

## Done when
- a feature whose diff from its base is documents only closes ready, with no verdict and no --lens forced
- a feature whose diff carries code still owes the verdict at its base
- code after the close still owes, and a waived close still owes nothing
- a test drives each answer, and one of them fails without the change

## Steps
- [x] T1 · the gate asks the feature's own diff, so a documents-only close owes no verdict   R1 · tree:9c81aea399cef87b301c11db51f71ae24bd3fe74
- [x] T2 · the debt row leaves the file and the changelog says why   R1 · tree:b397c1298383f455a682179f6f10503bb0e5c8a1

## History
L1 · "a closing tree that asks for no lens owes no verdict: the gate asks the same classifier the review asks, and a feature whose whole diff is documents closes without paying for a review"
L2 · OFICIO · "the gate asks the feature's diff, base to tree, not the tree's own uncommitted change: the verdict it demands is the one that judges everything the feature merges, so a close that only marks a step must not be measured as if the code never happened" (porque the proof lane already reads the feature's range; asking the same range for the review is what makes check and review answer one question)
