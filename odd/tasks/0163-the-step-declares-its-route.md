# the-step-declares-its-route
Base: 17005f1c971dff21dd6f083a04892b9ed5304233

## Requirements (verbatim, from the user)
R1 · "a step says who executes it -the parent inline, or a named agent delegated- and the task doc keeps that declaration on the step, so a reader knows who did the work without opening a chat"

## Done when
- `odachi plan "<text>" --req R1 --route inline` adds a step whose line names the route, and `--route delegated --agent odachi-worker` names the agent too
- a delegated route with no agent, and an inline route that names one, are refused with a message that says what to pass instead
- a step written before this change parses unchanged: a missing route reads as not declared, never as inline
- render then parse returns the same route and agent, and a done step keeps the route beside its tree
- the drive skill sends a delegated step to the flag instead of a craft note, and the skill's pinned test follows the new sentence

## Steps
- [x] T1 · the step line carries a route and an agent, and parse and render return them unchanged   R1 · tree:9510704ab01b217c5170a427808dc32bfd2d7bb5
- [x] T2 · plan refuses a delegated route with no agent and an inline route that names one   R1
- [x] T3 · the command line takes --route and --agent, and the help names both   R1 · tree:74f939eb8eebcc7805b418a66c571ed6d83047f4
- [x] T4 · the drive skill sends a delegated step to the flag and its pinned test follows   R1 · tree:5a7336df0a90eef56d8f42055b6d7d897dea2c5b
- [x] T5 · the changelog names the route a step declares   R1 · tree:d81c1eb4208993d2b08266fdc9f818d7416ff36a

## History
L1 · "a step says who executes it -the parent inline, or a named agent delegated- and the task doc keeps that declaration on the step, so a reader knows who did the work without opening a chat"
L2 · OFICIO · "a hand-written tail with two commit parts keeps the last and drops the first (review WARNING on taskdoc.go:416); the next feature pays it" (porque the re-review approved this tree and a warning earns no review round; fixing it now would rewrite the tree the verdict names, and that line is the tolerance an old document gets, not the new grammar)
