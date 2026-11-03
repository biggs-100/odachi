# the-default-branch

## What must exist
- the default branch answers without a branch name: the local branch that follows the base, and whether HEAD is one of its commits, so a detached checkout is refused or judged for the same reason a named one is

## Done when
- a trailer whose HEAD is detached at a commit of the local default branch is refused by name, not judged
- the CI-shaped checkout, with a detached HEAD and no local default branch, still reads the trailer: the document-less lane keeps working on a pull request
- a detached HEAD on a feature branch keeps the trailer as the vision
- when git cannot name the branch the base follows, the gate blocks with the reason instead of dropping the error
- tests pin the four cases in gitx and the refusal in the lane, and go test ./... is green

## Not included (first version)
- no change to the trailer's proofs, review or naming
- no new flag and no new verb
- a stored proof is never rewritten, whatever the writer learns

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: por qué la rama por defecto es un hecho de los commits, no de una cadena.
