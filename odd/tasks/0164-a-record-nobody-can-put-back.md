# a-record-nobody-can-put-back
Base: 4546f59331d3975605a2c8a64e8637b1750c6b33

## Requirements (verbatim, from the user)
R1 · "the two bookkeeping dead ends a close can leave: a step marked done with no tree can be given its tree afterwards, and `undo --file task` can name which task document it restores, so neither record stays unreachable"

## Done when
- `odachi plan --done T2 --commit <sha>` on a step that is already done and carries no mark records the tree and says it did
- the same call on a step that already carries a mark is still a no-op that writes nothing
- `odachi plan --done T2` with no `--commit` on a step with no mark records nothing and says how to record it
- `odachi undo --file task --artifact odd/tasks/0163-....md` restores that document, and a path odachi keeps no version of is refused with the ones it does have
- several task documents with backups stop being a dead end: the refusal lists them and names `--artifact`
- a document restored with `--artifact` backs up what it replaces, as the plain undo does

## Steps
- [ ] T1 · a done step with no mark accepts its tree afterwards, and a marked step stays a no-op   R1
- [ ] T2 · plan names the flag that records the tree it could not record   R1
- [ ] T3 · undo names the artifact it restores with --artifact   R1
- [ ] T4 · several task documents with backups name --artifact instead of a dead end   R1
- [ ] T5 · the changelog names the two repairs   R1

## History
L1 · "the two bookkeeping dead ends a close can leave: a step marked done with no tree can be given its tree afterwards, and `undo --file task` can name which task document it restores, so neither record stays unreachable"
