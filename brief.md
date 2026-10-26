# the-gate-json

## What must exist
- the gate JSON names every reason, and its keys, in a test

## Done when
- every reason the gate can print appears as a literal in a test, so renaming one breaks the suite
- the keys are pinned for a ready tree and for a blocked one, with next only where it exists
- the README names check --json as the gates machine contract

## Not included (first version)
- the human lines of check
- the exit codes
- progress --json
- the panel itself
- a schema version

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: how a contract stops being a convention.
