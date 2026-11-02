# a-line-that-can-close

## What must exist
- a change that fits one line is written as one line: `odachi start --light "<requirement>"` opens a feature with a task doc that carries the quoted requirement and one step with the same words, no brief, and the project's fence and voice untouched

## Done when
- `odachi start --light "<requirement>"` writes one task doc with the line `Light: yes`, R1 quoted and T1 named with the same words, and writes no brief
- the same run does not rewrite `odd/fence.toml` or `odd/profile.toml`
- `odachi check` on a light feature does not ask for a brief, and still refuses a fence violation, an open question and a missing or stale proof
- a light feature closes like any other: one step done means the proof is owed and the review is asked for its tree
- `odachi plan "<text>"` on a light feature is refused, saying that a second step is the signal to leave the lane
- a document without `Light:` behaves exactly as today, and `Light: yes` outside the preamble is prose the gate never reads
- `drive` and the changelog name the lane and say when not to use it

## Not included (first version)
- no change to review, proof or fence semantics: a light feature owes the same evidence
- no `--light` on `check`, `plan` or any other verb: the document carries the shape, the command line does not
- no light lane for a document that already has steps or a brief, and no second shape of light document
- no new verb and no change to how a normal feature is written

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: why a lane is a rule the artifact declares, not a flag the run carries.
