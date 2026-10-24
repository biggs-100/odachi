# fast-proof

## What must exist
- the suite and the proof loop stop being the slowest part of a commit

## Done when
- go test ./cmd/odachi finishes under 25 seconds on this machine (it was ~70)
- the full suite stays green, race-clean on ./internal/..., and gofmt/vet clean
- the workflow gets the proof by running the suite once per tree, not twice
- go test ./... and odachi check are green

## Not included (first version)
- skipping tests or using -short in the proof
- caching a proof across different trees or different commands
- changing the gates or the proof format

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: no rush.
- Wants to learn: where the time in a test suite actually goes.
