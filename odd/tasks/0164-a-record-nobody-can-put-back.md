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
- [x] T1 · a done step with no mark accepts its tree afterwards, and a marked step stays a no-op   R1 · tree:24c83056eafe0e659e0645011bd28d144107779c
- [x] T2 · plan names the flag that records the tree it could not record   R1 · tree:510c961fdf2b082b74400e6092fc820825684ea6
- [x] T3 · undo names the artifact it restores with --artifact   R1 · tree:f017d081025eadc637c36cde269791daad301487
- [x] T4 · several task documents with backups name --artifact instead of a dead end   R1 · tree:2c116329de6b812540f22ba62b040d4c0560fa92
- [x] T5 · the changelog names the two repairs   R1 · tree:a4f30e6f8cc6fb61afa32bb3c830fbde7f426d20

## History
L1 · "the two bookkeeping dead ends a close can leave: a step marked done with no tree can be given its tree afterwards, and `undo --file task` can name which task document it restores, so neither record stays unreachable"
L2 · OFICIO · "a bare file name skips the --file/--artifact contradiction, and the refusal then says there is no backup of a document the run itself hid (review WARNING on undo.go:152); the next feature pays it" (porque the verdict approved this tree and a warning earns no round; the fix is to let a bare name answer the kind question, and that is a change to the same match the tests already cover, so it lands with them)
