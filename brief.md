# the-paid-answer

## What must exist
- the paid answer survives a later failure: each model group's verdict is stored as it lands, and the gate still refuses a lens nobody judged

## Done when
- a run stores each model group's answer as soon as it lands: a later group's failure leaves the earlier answers on disk
- the run that fails says how many lenses of how many were saved and which ones are missing, and names the command that completes them
- review --show calls the verdict partial when it covers only some of the lenses this tree requires
- check --require-review still blocks on a lens that no stored verdict answered
- a stored verdict still never replaces an answered lens, and a test pins that refusal
- the README says what a partial verdict is and how to finish it

## Not included (first version)
- retrying a failed group inside the same run: the human runs the command again
- running model groups in parallel
- resuming a run that was killed: the store is the checkpoint, not a journal
- changing the reviewer command, the lens prompts or the gate's rule
- prices, budgets or automatic model choice

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: how a long run checkpoints its work without ever storing a half answer.
