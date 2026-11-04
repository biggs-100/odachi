# the-fence-stops

## What must exist
- the fence stops protecting the review's model policy, so a change to [lenses] can land

## Done when
- odd/fence.toml has no rule that protects odd/review.toml
- brief.md no longer carries the retention non-goal
- odd/fence.toml still protects proof/ (the archtest stays green)
- odachi check is green
- go test ./... is green

## Not included (first version)
- no change to what proof/ and odd/tasks are: never touched
- no change to the review's lenses or the model it names
- no new verb and no new flag

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: why the fence is rebuilt from the brief's non-goals, and why a rule must name paths instead of a phrase.
