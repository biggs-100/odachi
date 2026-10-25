# the-review-asks
Base: e14bef2ce523e52c2461f106c1d7cddfe20626e5

## Requirements (verbatim, from the user)
R1 · "the review asks for an answer, not for work: the prompt tells the reviewer to read and not to run anything, and a candidate with no lens says so instead of claiming a verdict exists"

## Done when
- the review prompt carries the rule: answer from the prompt, do not run commands, do not edit files
- review --exec on a documents-only candidate with no verdict says the change asks for no lens and names --lens all
- a tree whose lenses are all answered still says this tree already has a verdict
- a reviewer that answers with something that is not a verdict keeps its own output as the diagnosis
- the README explains the reviewer's side of the contract in one paragraph

## Steps
- [ ] T1 · the review prompt says answer from the prompt, do not run commands, do not edit files   R1
- [ ] T2 · a candidate with no lens says so and names --lens all, and only an answered tree says it has a verdict   R1

## History
L1 · "the review asks for an answer, not for work: the prompt tells the reviewer to read and not to run anything, and a candidate with no lens says so instead of claiming a verdict exists"
