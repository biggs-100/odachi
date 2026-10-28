# the-review-refuses

## What must exist
- the review refuses to start a reviewer on a prompt bigger than the project allows, and says the size, the limit and the ways out

## Done when
- odd/review.toml carries the size limit, and a file without it means no limit
- review --exec refuses before starting anything when the prompt is over the limit, naming the size, the limit and the two ways out
- the printed form still prints, because it spends nothing, and says the size on stderr
- a test drives both answers, observed red first

## Not included (first version)
- no new verb and no new flag: the limit is a key in odd/review.toml
- no change to how the prompt is built
- no model chosen by the tool

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: how a known limit becomes a refusal before the cost, not after it.
