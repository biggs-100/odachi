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
- [x] T1 · the fence no longer protects odd/review.toml and still protects proof/**   R1 · route:inline · tree:d1d53adb7998c8d480fbe5833c26d58b1a04d6b0

## History
L1 · "the fence stops protecting the review's model policy, so a change to [lenses] can land"
L2 · OFICIO · "the retention rule is dropped, not narrowed to cleanup.go" (porque a fence protects files, not sections: odd/review.toml holds both the retention numbers and the model policy, so any rule over that file refuses a change to [lenses])
