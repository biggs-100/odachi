# one-command-refreshes
Base: 9b0742b3b2d7e2f344690e0671997ec7f14a4e4d

## Requirements (verbatim, from the user)
R1 · "one command refreshes the machine: odachi setup reports what it would change before it changes it, writes nothing under --dry-run, and names what it skipped, so an upgraded binary leaves a current home in one run"

## Done when
- --dry-run changes no byte anywhere and names every file it would write
- a second setup after the first reports every destination unchanged, and the same run twice is a no-op
- the report names the clients that are not installed here and what it left alone, so nothing is silently skipped

## Steps
- [x] T1 · setup --dry-run writes no byte and names every file it would write   R1 · tree:705a115e1ae8c16551246f6d69a1bb241bc5828a
- [x] T2 · the report names the clients that are not here and what it left alone   R1 · tree:58b3df7d52e74afb65844c9fdd2db4d76fc79b4d
- [ ] T3 · the docs and the changelog say how to refresh a machine   R1

## History
L1 · "one command refreshes the machine: odachi setup reports what it would change before it changes it, writes nothing under --dry-run, and names what it skipped, so an upgraded binary leaves a current home in one run"
