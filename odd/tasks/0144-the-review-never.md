# the-review-never
Base: 35dd843d1a59d9869040c35a1d63ffee5c1f6f3b

## Requirements (verbatim, from the user)
R1 · "the review never authorises delivery: with the project's switch on, check names what the review says and only a finding it made stops a ship"

## Done when
- with mode on, a close without a verdict is ready and names the review that is missing
- a causal BLOCKER from a review that exists still blocks until it is refuted
- the hook prints the reminder on a quiet run, and CI stops passing --no-review
- the ADR says the review does not authorise delivery, and the help and the README say it

## Steps
- [ ] T1 · with mode on, a close without a verdict is ready and names the missing review   R1
- [ ] T2 · a causal blocker from a review that exists still blocks, and the hook prints the note   R1
- [ ] T3 · CI stops passing --no-review, and the ADR, help and README say the review does not authorise delivery   R1

## History
L1 · "the review never authorises delivery: with the project's switch on, check names what the review says and only a finding it made stops a ship"
