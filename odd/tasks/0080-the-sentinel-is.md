# the-sentinel-is
Base: 30f7e1ad6748c72928f72990a972c036526a9244

## Requirements (verbatim, from the user)
R1 · "the sentinel is pinned end to end, and the debt row leaves the file"

## Done when
- the selection adapter answers both only when git says the folder has no repository, and names a git failure instead of guessing
- init names the clone's store when it cannot resolve it, and writes nothing
- a fake git port makes both branches reachable in a test

## Steps
- [ ] T1 · the adapter takes the git port: a test drives no-repository and git-failed, and init names the store   R1

## History
L1 · "the sentinel is pinned end to end, and the debt row leaves the file"
