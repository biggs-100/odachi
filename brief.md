# the-ledger-names

## What must exist
- the ledger names the commit a step landed in, and that commit exists on the branch, without manual correction

## Done when
- plan --done T1 without --commit records the tree of HEAD in the step's mark
- progress resolves that mark to the commit that carries the tree on this branch
- a mark whose bytes are not on the branch says so instead of naming a commit that does not exist
- a task doc that carries a commit keeps working, and progress resolves it when its bytes landed elsewhere
- tests pin the mark, the resolution and the missing case

## Not included (first version)
- rewriting the old task docs
- changing the repository's merge strategy
- resolving a squashed step into the several commits it became
- a new verb, or a flag beyond the optional --commit
- the panel itself

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: why a commit is not the identity of a change.
