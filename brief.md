# the-gate-asks

## What must exist
- the gate asks once per feature, not once per commit: a feature with open steps needs only its proof, and the tree that closes it needs the one verdict that judged the feature's diff

## Done when
- the task doc records the commit the feature started from, and start writes it
- with the project mode on, a tree whose feature still has open steps is ready with its proof alone
- the tree that closes a feature needs a verdict, and the verdict must name the feature's base
- the blocked message names the runnable review with the recorded base
- --require-review still forces the gate for one tree and --no-review still turns it off for one run
- a feature with no steps, or a document with no base, keeps the old rule: the mode asks for a verdict on every tree
- tests pin the five cases and the README says the review unit is the feature

## Not included (first version)
- changing the lens selection, the verdict format or the review store
- discovering the base from a remote or from the default branch
- reviewing every step as its own unit
- changing CI or the release workflow
- a journal, CAS or resume machinery
- rewriting the task docs of features already closed

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: how to make a gate cost once per unit of work without making it blind.
