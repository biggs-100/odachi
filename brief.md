# the-suite-stops

## What must exist
- the suite stops being the slowest part of a commit: each verb reads the candidate once

## Done when
- go test ./cmd/odachi finishes under 25s on this machine (28.7s before)
- check and review read the worktree once: one merge-base and one throwaway index per run
- the tree hash, the changed paths and the diff stay byte-identical to before, pinned by tests
- the full suite stays green, race-clean on ./internal/..., and gofmt/vet clean
- the proof still runs once per tree

## Not included (first version)
- skipping tests or using -short in the proof
- weakening doctor, fence or coverage checks, or deleting tests
- a new dependency or a custom test runner
- changing the proof or verdict formats
- caching a proof across trees
- tuning GOMAXPROCS or -parallel to win the benchmark

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: no rush.
- Wants to learn: where the time of a test suite actually goes.
