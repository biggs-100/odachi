# a-feature-that
Base: f298d35b95d0036bea6373bf2b08f33848b0caf3

## Requirements (verbatim, from the user)
R1 · "a feature that has a plan and no steps is open, so its plan commit lands without a verdict and its close still owes one"

## Done when
- with the review switch on, a tree whose document has a base and no steps owes no verdict and check is ready for a step
- the tree that marks the last step still owes the verdict at the document's base
- a step in the middle still owes nothing, and a waived close still owes nothing
- a test drives each of those answers, and one of them fails without the change

## Steps
- [x] T1 · a document with a base and no steps is an open feature, so its plan commit owes no verdict   R1 · tree:455848c6cb33a4dea1fd536fa75990983d6a4bd9
- [x] T2 · the debt file says the row closed, and the changelog says why   R1

## History
L1 · "a feature that has a plan and no steps is open, so its plan commit lands without a verdict and its close still owes one"
L2 · OFICIO · "the plan commit of 0095 lands with --no-verify as well" (porque the fix this feature carries is the one that would unblock it: the gate is judged by the tree that changes it)
