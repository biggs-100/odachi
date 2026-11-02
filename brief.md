# the-review-of

## What must exist
- the review of a light feature is judged against that feature's own line, not against the previous feature's brief

## Done when
- odachi review --base <sha> on a light feature writes a prompt whose vision section quotes that feature's requirement and names the task document it came from
- a normal feature's prompt carries brief.md exactly as before
- a light document that is already closed still names its own line, so the closing tree is judged against it
- a task document that cannot be read is named as such: the review never falls back to a brief it did not read
- no lens, verdict or store entry changes: the same tree still gets the same lenses and the same JSON

## Not included (first version)
- no change to internal/domain/review/**: the lane never chooses lenses and never changes the shape of a verdict
- no new flag on review: the feature is found, not named by hand
- no change to start --light: the lane keeps writing one task doc and nothing else
- no redefinition of brief.md for a normal feature

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: why the context a prompt carries is part of its evidence.
