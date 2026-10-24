# the-suite-stops

## Requirements (verbatim, from the user)
R1 · "the suite stops being the slowest part of a commit: each verb reads the candidate once"

## Done when
- go test ./cmd/odachi finishes under 25s on this machine (28.7s before)
- check and review read the worktree once: one merge-base and one throwaway index per run
- the tree hash, the changed paths and the diff stay byte-identical to before, pinned by tests
- the full suite stays green, race-clean on ./internal/..., and gofmt/vet clean
- the proof still runs once per tree

## Steps
- [ ] T1 · the Git port reads the candidate once (tree, paths, diff) and check/review use that read   R1
- [ ] T2 · measure the suite, record the result and cut the last measured edge   R1

## History
L1 · "the suite stops being the slowest part of a commit: each verb reads the candidate once"
