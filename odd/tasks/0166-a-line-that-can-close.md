# a-line-that-can-close
Base: e9de1aa37fb32db56bb7f4c7b237305963743816

## Requirements (verbatim, from the user)
R1 · "a change that fits one line is written as one line: `odachi start --light "<requirement>"` opens a feature with a task doc that carries the quoted requirement and one step with the same words, no brief, and the project's fence and voice untouched"

## Done when
- `odachi start --light "<requirement>"` writes one task doc with the line `Light: yes`, R1 quoted and T1 named with the same words, and writes no brief
- the same run does not rewrite `odd/fence.toml` or `odd/profile.toml`
- `odachi check` on a light feature does not ask for a brief, and still refuses a fence violation, an open question and a missing or stale proof
- a light feature closes like any other: one step done means the proof is owed and the review is asked for its tree
- `odachi plan "<text>"` on a light feature is refused, saying that a second step is the signal to leave the lane
- a document without `Light:` behaves exactly as today, and `Light: yes` outside the preamble is prose the gate never reads
- `drive` and the changelog name the lane and say when not to use it

## Steps
- [ ] T1 · the task doc carries the lane it was written for, in its preamble   R1
- [ ] T2 · start --light writes one line and leaves the fence and the voice alone   R1
- [ ] T3 · check does not ask a light feature for a brief, and asks the same of everything else   R1
- [ ] T4 · plan refuses a second step on a light feature   R1
- [ ] T5 · drive and the changelog name the lane and when not to use it   R1

## History
L1 · "a change that fits one line is written as one line: `odachi start --light "<requirement>"` opens a feature with a task doc that carries the quoted requirement and one step with the same words, no brief, and the project's fence and voice untouched"
