# a-step-marked

## What must exist
- a step marked with --commit names the bytes of that commit, so the mark survives a rebase too

## Done when
- plan --done T1 --commit <sha> stores the tree of that commit when the commit exists
- a commit the repository cannot read keeps the sha as given, because that is all that is known
- the message names what was stored, and progress resolves it to the landed commit
- tests pin the canonical mark, the fallback and the resolution

## Not included (first version)
- rewriting the old task docs
- a new flag
- changing the default: the tree of HEAD stays
- the panel itself

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: what makes a mark durable.
