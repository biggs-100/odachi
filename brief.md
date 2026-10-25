# the-verdict-cannot

## What must exist
- the verdict cannot be from the model that wrote the tree: the proof and the verdict record who produced them, and a self-review is refused with the command that fixes it

## Done when
- a proof run by an agent session records the model it ran under, and an old proof without that line still parses
- a verdict records the model that judged, and an old verdict without it still parses
- check --require-review refuses a verdict whose model is the one in the proof, and names review --model
- the README says this is disclosure, not identity: a model name can be lied about

## Not included (first version)
- cryptographic identity or signing the verdict
- a review app or a GitHub App
- refusing a verdict written by a human with no model
- changing the review prompt or the lenses

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: what a self-review can and cannot prove.
