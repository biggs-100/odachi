# the-gate-s
Base: be5d13c8cf7b1f24b398589d48529df077087445

## Requirements (verbatim, from the user)
R1 · "the gate's waiver answer is driven by a test that calls ReviewOwed directly, so the branch cannot be lost or inverted"

## Done when
- check.ReviewOwed answers nothing owed for a closed feature whose document records the waiver
- the same document without the waiver is owed at the feature's base
- the test fails when the waiver branch is removed, verified by mutation

## Steps
- [x] T1 · check.ReviewOwed is called with a closed waived document, and with the same document without the waiver   R1 · 0416636

## History
L1 · "the gate's waiver answer is driven by a test that calls ReviewOwed directly, so the branch cannot be lost or inverted"
