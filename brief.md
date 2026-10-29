# a-run-that

## What must exist
- a run that asks for a lens the evidence did not require writes it beside the verdict, so a later reader sees the scope was raised and a replay cannot drop it

## Done when
- a run whose --lens added a lens the evidence did not ask writes reviews/<tree>.scope.json beside the verdict, naming what the evidence asked, what was added and the flag as it was given
- a run that added nothing writes no record, and a second run for the same tree does not rewrite the store
- --show and the prompt say the scope was raised, and the record is read by tree, so a verdict never inherits another tree's scope
- a test drives each answer, and one of them fails without the change

## Not included (first version)
- no new verb and no new flag
- no change to the verdict shape
- no change to the proof format, and the stored proof is never rewritten

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: this week.
- Wants to learn: whether a decision made before a run belongs in the artifact the run produces.
