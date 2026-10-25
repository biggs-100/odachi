# closing-a-feature

## What must exist
- closing a feature is reviewed and publishing it is not: the verdict is owed by the commit that marks the last step done, and a later documents-only tree does not ask for it again

## Done when
- with the mode on, the commit that closes a feature owes the verdict of the feature's base
- a documents-only tree after the close is ready with its proof, as long as the task doc did not change
- a tree with code after the close still owes a verdict: the gate does not go blind
- a feature with no steps or no base keeps the old per-tree rule
- --require-review still forces one tree and --no-review still turns the mode off for one run
- tests pin the four cases and the README says the review is owed by the close, not by the tail

## Not included (first version)
- searching the store for a verdict of another tree
- a shared or remote store
- changing CI or the release workflow
- rewriting closed task docs
- changing the lens selection

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: how to tell a unit of work from the paperwork that follows it.
