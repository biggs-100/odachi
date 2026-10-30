# one-command-refreshes
Base: b464b2b

## Requirements (verbatim, from the user)
R1 · "one command refreshes the machine: odachi setup reports what it would change before it changes it, writes nothing under --dry-run, and names what it skipped, so an upgraded binary leaves a current home in one run"

## Done when
- --dry-run changes no byte anywhere and names every file it would write
- a second setup after the first reports every destination unchanged, and the same run twice is a no-op
- the report names the clients that are not installed here and what it left alone, so nothing is silently skipped

## Steps
- [x] T1 · setup --dry-run writes no byte and names every file it would write   R1 · tree:f076b5f025153d1b6efed010a437895ba14fb249
- [x] T2 · the report names the clients that are not here and what it left alone   R1 · tree:0f1449f39008de35b93e0c9b27052b767e40099b
- [x] T3 · the docs and the changelog say how to refresh a machine   R1 · tree:19737abbb480f26277588fce0f4ab24a6b429a0e

## History
L1 · "one command refreshes the machine: odachi setup reports what it would change before it changes it, writes nothing under --dry-run, and names what it skipped, so an upgraded binary leaves a current home in one run"
