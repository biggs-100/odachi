# a-review-prompt

## What must exist
- a review prompt and its answer have an owner, so neither lands in the tree nor in a scroll

## Done when
- odachi review writes the prompt into the clone's store beside the verdict and names the path, so an agent never improvises one
- odachi review --exec reads the answer from the {verdict} file the command was given, and stdout becomes diagnosis
- without {verdict} the answer is still read from stdout, so no existing reviewer breaks
- a test drives both capture paths and the prompt path, observed red first

## Not included (first version)
- no new verb and no new flag
- no change to where the verdicts live
- no new field in the verdict schema

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: how to capture a child process's answer without trusting its stdout.
