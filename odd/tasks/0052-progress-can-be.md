# progress-can-be
Base: 1787d6e4a84daaed644ba353e24f09ee6a6add57

## Requirements (verbatim, from the user)
R1 · "progress can be read by a machine: the same report as data, not only as lines"

## Done when
- the report is computed once as data: feature, steps with done and commit, proof state and review state
- the human lines are rendered from that data, so the two cannot disagree
- progress --json prints it as one object, and a project with no feature says so in a field
- the state names are stable and pinned by a test
- the human report reads exactly as today

## Steps
- [x] T1 · the report is computed as data and the human lines are rendered from it   R1 · fd388a8
- [x] T2 · progress --json prints that state as one object   R1 · 72e678b

## History
L1 · "progress can be read by a machine: the same report as data, not only as lines"
