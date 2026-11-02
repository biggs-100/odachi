# the-step-declares-its-route

## What must exist
- a step says who executes it -the parent inline, or a named agent delegated- and the task doc keeps that declaration on the step, so a reader knows who did the work without opening a chat

## Done when
- `odachi plan "<text>" --req R1 --route inline` adds a step whose line names the route, and `--route delegated --agent odachi-worker` names the agent too
- a delegated route with no agent, and an inline route that names one, are refused with a message that says what to pass instead
- a step written before this change parses unchanged: a missing route reads as not declared, never as inline
- render then parse returns the same route and agent, and a done step keeps the route beside its tree
- the drive skill sends a delegated step to the flag instead of a craft note, and the skill's pinned test follows the new sentence

## Not included (first version)
- no gate change: `check`, `review` and the fence neither read nor judge a route
- the tool never checks that the named agent did the work: it records a declaration, like ownership (ADR 0038)
- no new verb, no store entry, no `summary`/`progress` change: the route lives on the step line and nowhere else
- no route on a requirement, and no second agent in the same field: one step, one executor

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: why a declaration nobody verifies is still worth writing down.
