# a-step-marked
Base: deae25a9fe783cf4623c22fc4a065fb4b9808702

## Requirements (verbatim, from the user)
R1 · "a step marked with --commit names the bytes of that commit, so the mark survives a rebase too"

## Done when
- plan --done T1 --commit <sha> stores the tree of that commit when the commit exists
- a commit the repository cannot read keeps the sha as given, because that is all that is known
- the message names what was stored, and progress resolves it to the landed commit
- tests pin the canonical mark, the fallback and the resolution

## Steps
- [ ] T1 · plan --done --commit records the tree of the commit when it can read it, and the sha when it cannot   R1

## History
L1 · "a step marked with --commit names the bytes of that commit, so the mark survives a rebase too"
