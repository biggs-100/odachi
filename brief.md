# the-sentinel-is

## What must exist
- the sentinel is pinned end to end, and the debt row leaves the file

## Done when
- the selection adapter answers both only when git says the folder has no repository, and names a git failure instead of guessing
- init names the clone's store when it cannot resolve it, and writes nothing
- a fake git port makes both branches reachable in a test

## Not included (first version)
- no change to where the store lives
- no new verb and no new flag

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: how a test reaches an error branch without breaking the machine.
