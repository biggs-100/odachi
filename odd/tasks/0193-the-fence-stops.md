# the-fence-stops
Base: ab26c957d18a3b105105b2f0b4e37258a5d74238

## Requirements (verbatim, from the user)
R1 · "the fence stops protecting the review's model policy, so a change to [lenses] can land"

## Done when
- odd/fence.toml has no rule that protects odd/review.toml
- brief.md no longer carries the retention non-goal
- odd/fence.toml still protects proof/ (the archtest stays green)
- odachi check is green
- go test ./... is green

## Steps
- [ ] T1 · the fence no longer protects odd/review.toml and still protects proof/**   R1 · route:inline

## History
L1 · "the fence stops protecting the review's model policy, so a change to [lenses] can land"
