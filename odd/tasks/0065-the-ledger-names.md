# the-ledger-names
Base: 2388cd513ab21a71f4d5c9db70e70351bd7d0a3e

## Requirements (verbatim, from the user)
R1 · "the ledger names the commit a step landed in, and that commit exists on the branch, without manual correction"

## Done when
- plan --done T1 without --commit records the tree of HEAD in the step's mark
- progress resolves that mark to the commit that carries the tree on this branch
- a mark whose bytes are not on the branch says so instead of naming a commit that does not exist
- a task doc that carries a commit keeps working, and progress resolves it when its bytes landed elsewhere
- tests pin the mark, the resolution and the missing case

## Steps
- [x] T1 · plan --done records the tree of HEAD when no commit was given, and a repository with no commit says so   R1 · 6c0847c
- [x] T2 · progress resolves a tree mark, and a commit whose bytes landed elsewhere, to the commit on this branch   R1 · tree:5055fcf605aaf853f750b93dc95f6f1fdc21569b

## History
L1 · "the ledger names the commit a step landed in, and that commit exists on the branch, without manual correction"
