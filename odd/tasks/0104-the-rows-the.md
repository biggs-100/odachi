# the-rows-the
Base: cb0e6cf56e2c78decec2bfa5c90070064b4b32d2

## Requirements (verbatim, from the user)
R1 · "the rows the store carries into a review prompt are measured: how much of the prompt they are, how many name paths the candidate never touched, and what each way of pruning them would save"

## Done when
- every prompt the store holds is measured: total bytes, the carried section, its rows, and how many name a path the candidate never touched
- each of the three prunings is measured on the same prompts: only the candidate's paths, findings without passes, and a byte budget
- the measurement says what each pruning would take away, and why the gate does not lose coverage when the prompt is pruned
- the numbers, the method and the decision they open land in docs/review-debt.md

## Steps
- [x] T1 · the carried section is measured on every prompt: bytes, rows, and rows outside the candidate   R1 · tree:b6f9e64a3b7a31c78772d2fa4e7776a29d85b8ca
- [x] T2 · the three prunings are measured on the same prompts, with what each takes away   R1
- [x] T3 · the numbers and the decision they open land in the debt file   R1 · tree:7437a7509b185bcba4207f46d83f72fbe8e19fbd

## History
L1 · "the rows the store carries into a review prompt are measured: how much of the prompt they are, how many name paths the candidate never touched, and what each way of pruning them would save"
L2 · OFICIO · "the measurement reads the prompts the store already holds, and prunes only the prompt: the store keeps every row, so the merge and the coverage gate never depend on what the prompt carried" (porque a display decision cannot weaken a gate; if it did, the beam would be in the wrong place - the prompt is the reviewer's context, the store is the evidence)
