# a-record-nobody-can-put-back

## What must exist
- the two bookkeeping dead ends a close can leave: a step marked done with no tree can be given its tree afterwards, and `undo --file task` can name which task document it restores, so neither record stays unreachable

## Done when
- `odachi plan --done T2 --commit <sha>` on a step that is already done and carries no mark records the tree and says it did
- the same call on a step that already carries a mark is still a no-op that writes nothing
- `odachi plan --done T2` with no `--commit` on a step with no mark records nothing and says how to record it
- `odachi undo --file task --artifact odd/tasks/0163-....md` restores that document, and a path odachi keeps no version of is refused with the ones it does have
- several task documents with backups stop being a dead end: the refusal lists them and names `--artifact`
- a document restored with `--artifact` backs up what it replaces, as the plain undo does

## Not included (first version)
- no automatic repair: `check` never writes a tree a human did not ask for
- no change to how a step is first marked, and no guessing when the worktree is dirty
- no new artifact kind, no new backup location, no `--collect` change
- no git: `undo` still never reverts code

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: why a record must be repairable by naming it, not by re-running the verb.
