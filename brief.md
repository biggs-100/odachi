# the-review-s

## What must exist
- the review's size wall is measured on this project's own prompts: how close a candidate runs to the limit, and whether a reviewer that closed inside it read what it was handed

## Done when
- the prompts the store holds are measured and written down: how many, the biggest, and how many run within 10 KB of the limit
- the measurement says which files may come from a printed form, so an over-limit file is not read as proof that an over-limit prompt was sent
- one probe on a prompt that closed inside the limit is run with the project's model, and its answer is recorded whatever it says
- the numbers, the method and what they do not prove land in docs/review-debt.md

## Not included (first version)
- no new verb and no new flag
- no change to the verdict shape
- no change to the proof format, and the stored proof is never rewritten

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: this week.
- Wants to learn: how to measure a model's reading without trusting its silence.
