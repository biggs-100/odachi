# the-review-asks

## What must exist
- the review asks for an answer, not for work: the prompt tells the reviewer to read and not to run anything, and a candidate with no lens says so instead of claiming a verdict exists

## Done when
- the review prompt carries the rule: answer from the prompt, do not run commands, do not edit files
- review --exec on a documents-only candidate with no verdict says the change asks for no lens and names --lens all
- a tree whose lenses are all answered still says this tree already has a verdict
- a reviewer that answers with something that is not a verdict keeps its own output as the diagnosis
- the README explains the reviewer's side of the contract in one paragraph

## Not included (first version)
- changing the reviewer binary, its tools or its flags
- parsing a JSON out of a transcript
- retrying the reviewer automatically
- changing the lens vocabulary or the four prompts
- the refutation flow

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: how to write a contract for something you do not control.
