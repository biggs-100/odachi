# the-review-s
Base: 7d7aef8d68d44cadb19aa95ee1af619eb0506b1c

## Requirements (verbatim, from the user)
R1 · "the review's size wall is measured on this project's own prompts: how close a candidate runs to the limit, and whether a reviewer that closed inside it read what it was handed"

## Done when
- the prompts the store holds are measured and written down: how many, the biggest, and how many run within 10 KB of the limit
- the measurement says which files may come from a printed form, so an over-limit file is not read as proof that an over-limit prompt was sent
- one probe on a prompt that closed inside the limit is run with the project's model, and its answer is recorded whatever it says
- the numbers, the method and what they do not prove land in docs/review-debt.md

## Steps
- [x] T1 · the prompts the store holds are measured: count, biggest, and how close to the limit   R1 · tree:eabb53c4ee39afcb97292cf3e0614084c8571d99
- [x] T2 · one probe on a prompt that closed inside the limit is run and its answer recorded   R1
- [x] T3 · the numbers, the method and what they do not prove land in the debt file   R1 · tree:698e3491c5ebf7d04db2a4dcb95ad016dc937234

## History
L1 · "the review's size wall is measured on this project's own prompts: how close a candidate runs to the limit, and whether a reviewer that closed inside it read what it was handed"
L2 · OFICIO · "the probe uses the project's model through a plain run, not the odachi-reviewer agent, because that agent answers a JSON verdict and a probe question is not a review" (porque the measurement is about reading a large prompt, and the agent's task would change the conditions; the note says it measures reading, not reviewing)
