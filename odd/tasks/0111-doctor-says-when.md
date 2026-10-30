# doctor-says-when
Base: 4d7b8a3a35f6998b0a89f4431f1772912f0d3e09

## Requirements (verbatim, from the user)
R1 · "doctor says when the rules block is older than the binary: a project that upgrades odachi learns that AGENTS.md carries text this binary would not write, and the repair is one command"

## Done when
- a project whose block matches what this binary writes is ok, as it is today
- a project whose block differs gets a warning that names odachi init, and it is a warning, not a blocker
- line endings and trailing spaces do not make a block stale, so a Windows checkout is not a defect
- a test drives each answer, and one of them fails without the change

## Steps
- [x] T1 · doctor compares the block it would write with the one AGENTS.md carries   R1 · tree:45136516837c80f4221d2d1c9f6493431022f41b
- [x] T2 · the changelog says a warning, not a blocker   R1 · tree:f23a6629a798c7fcf82f8dcba923d96284761339

## History
L1 · "doctor says when the rules block is older than the binary: a project that upgrades odachi learns that AGENTS.md carries text this binary would not write, and the repair is one command"
