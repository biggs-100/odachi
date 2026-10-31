# odachi-clean-reports

## What must exist
- odachi clean reports every class of artifact in a clone with its owner, its size and what it would do, and applies retention only to the classes odachi itself wrote

## Done when
- odachi clean prints every class with its owner, its file count and its size, and names a path it could not read instead of counting it as empty
- odachi clean --apply leaves the newest review records and the newest backups and deletes nothing outside the store: git's files, another tool's files, proof/ and odd/ stay byte-identical
- odachi clean --apply --legacy removes the old agent scratch files in .git/ after naming them, and --apply without --legacy leaves them
- the block says where an ad-hoc note goes, and odachi clean empties that folder

## Not included (first version)
- no change to the proof format, and the stored proof is never rewritten
- no change to the bundled skills

## Decisions taken
- It is used from the phone and the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: por que borrar exige un dueno declarado y una vida util por clase.
