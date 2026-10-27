# the-plan-grows
Base: 5cd431e02b67b8303d6197f2a3157f64e014c957

## Requirements (verbatim, from the user)
R1 · "the plan grows and the vision does not: when the human changes a requirement, the words land in the history and the steps that claimed it reopen"

## Done when
- odachi ask --change "<words>" --req R1 appends the words verbatim to the history and reopens every done step that cites R1, dropping its mark
- a change with no --req is refused, because the vision changes with a new feature and not with a reopened step
- drive teaches the rule, the ask help names the form, and the tests pin the line, the reopen and the refusal
- the impact map records the third follow-up as implemented

## Steps
- [x] T1 · the task doc records a change line and reopens the done steps that cite the requirement, with tests   R1 · tree:acdbcd614af49f2e4eb06983099d661a335802ab
- [x] T2 · ask --change writes it and refuses a change with no requirement, and the CLI names the form   R1 · tree:915ab2bda7c28c8903796407ba5ab2aada64a9e0
- [x] T3 · drive teaches the rule and the impact map records the third follow-up   R1 · tree:5be5563ebd8c4d3994e649437ddf247474f19c3c

## History
L1 · "the plan grows and the vision does not: when the human changes a requirement, the words land in the history and the steps that claimed it reopen"
L2 · OFICIO · "the change is recorded as L<n> · CAMBIO R1 · "<words>", one more record shape in the history" (porque the history is the only log and it is append-only; naming the requirement in the record is what makes the reopen derivable instead of guessed)
L3 · OFICIO · "reopening a step drops its tree mark" (porque the mark is evidence for the old step; keeping it would claim work the change invalidated)
