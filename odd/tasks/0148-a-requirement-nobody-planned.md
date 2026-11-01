# a-requirement-nobody-planned
Base: 27a6050dcaf5e15e2edc9554ec05acad883f9769

## Requirements (verbatim, from the user)
R1 · "a requirement with no step that names it stops the close and the line names it, because a requirement nobody planned is a requirement nobody built"

## Done when
- `check` refuses a close when a requirement has no step, and the refusal names the requirement and what to do
- a feature whose steps cover every requirement still closes, and its proof is written exactly as before
- the tests pin both sides: the uncovered requirement and the covered one

## Steps
- [ ] T1 · a requirement with no step stops the close, and the refusal names it   R1
- [ ] T2 · the changelog and the impact map say what the gate reads now   R1

## History
L1 · "a requirement with no step that names it stops the close and the line names it, because a requirement nobody planned is a requirement nobody built"
