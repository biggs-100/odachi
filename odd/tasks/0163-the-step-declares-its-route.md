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
- [ ] T1 · the step line carries a route and an agent, and parse and render return them unchanged   R1
- [ ] T2 · plan refuses a delegated route with no agent and an inline route that names one   R1
- [ ] T3 · the command line takes --route and --agent, and the help names both   R1
- [ ] T4 · the drive skill sends a delegated step to the flag and its pinned test follows   R1
- [ ] T5 · the changelog names the route a step declares   R1

## History
L1 · "a step says who executes it -the parent inline, or a named agent delegated- and the task doc keeps that declaration on the step, so a reader knows who did the work without opening a chat"
