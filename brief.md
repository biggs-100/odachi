# progress-can-be

## What must exist
- progress can be read by a machine: the same report as data, not only as lines

## Done when
- the report is computed once as data: feature, steps with done and commit, proof state and review state
- the human lines are rendered from that data, so the two cannot disagree
- progress --json prints it as one object, and a project with no feature says so in a field
- the state names are stable and pinned by a test
- the human report reads exactly as today

## Not included (first version)
- the TUI plugin and its npm packaging: another project
- new facts the report does not already know
- changing the human wording or the exit code

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: how one shape serves two readers.
