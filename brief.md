# a-feature-that

## What must exist
- a feature that has a plan and no steps is open, so its plan commit lands without a verdict and its close still owes one

## Done when
- with the review switch on, a tree whose document has a base and no steps owes no verdict and check is ready for a step
- the tree that marks the last step still owes the verdict at the document's base
- a step in the middle still owes nothing, and a waived close still owes nothing
- a test drives each of those answers, and one of them fails without the change

## Not included (first version)
- no change to what a closed feature owes
- no new verb and no new flag
- no change to the verdict or the proof format

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: where a gate is allowed to say no before the work exists.
