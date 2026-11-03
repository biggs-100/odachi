# the-run-that

## What must exist
- the run that reads a draft leaves the brief written: start --brief records it as brief.md, and the shape skill stops promising a file the run never wrote

## Done when
- a draft outside brief.md becomes brief.md in the canonical shape, and the draft itself is untouched
- a draft that is already brief.md is not rewritten
- an existing brief.md beside a draft elsewhere is refused without --force, naming the flag
- the shape skill says the draft may live anywhere and the run records it
- tests pin the three paths and go test ./... is green

## Not included (first version)
- no change to the interview path
- no new flag and no new verb
- a stored proof is never rewritten, whatever the writer learns

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: por qué el artefacto lo escribe el run que lo promete.
