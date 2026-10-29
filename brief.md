# the-rows-the

## What must exist
- the rows the store carries into a review prompt are measured: how much of the prompt they are, how many name paths the candidate never touched, and what each way of pruning them would save

## Done when
- every prompt the store holds is measured: total bytes, the carried section, its rows, and how many name a path the candidate never touched
- each of the three prunings is measured on the same prompts: only the candidate's paths, findings without passes, and a byte budget
- the measurement says what each pruning would take away, and why the gate does not lose coverage when the prompt is pruned
- the numbers, the method and the decision they open land in docs/review-debt.md

## Not included (first version)
- no new verb and no new flag
- no change to the verdict shape
- no change to the proof format, and the stored proof is never rewritten

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: this week.
- Wants to learn: whether a review prompt should carry the whole memory or only the part it can act on.
