# the-gate-s

## What must exist
- the gate's waiver answer is driven by a test that calls ReviewOwed directly, so the branch cannot be lost or inverted

## Done when
- check.ReviewOwed answers nothing owed for a closed feature whose document records the waiver
- the same document without the waiver is owed at the feature's base
- the test fails when the waiver branch is removed, verified by mutation

## Not included (first version)
- changing when the gate asks: the rule of 0061 stays
- a new flag, a new mode or a new message
- the other branches of ReviewOwed, already pinned by 0051
- the panel itself

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: how a characterization test proves it can fail.
