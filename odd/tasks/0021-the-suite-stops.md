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
- [x] T1 · the Git port reads the candidate once (tree, paths, diff) and check/review use that read   R1 · 9fa70ea
- [x] T2 · measure the suite, record the result and cut the last measured edge   R1 · 2ca0690

## History
L1 · "the suite stops being the slowest part of a commit: each verb reads the candidate once"
L2 · OFICIO · "The Git port answers the candidate once: tree, paths and diff from one throwaway index" (porque the three answers describe one worktree, and the old verbs paid the same merge-base and the same staging three times)
L3 · OFICIO · "The store's ignore is prepared once, and the read repeats only when a write can move the tree or after the reviewer" (porque a write can remove earlier verdicts from the candidate, and the reviewer is arbitrary code that may edit the tree it judged)
L4 · OFICIO · "HEAD's fork point is HEAD: the default base pays no merge-base process" (porque the same commit's fork point with itself is itself; a named base still asks, because another branch's commits must not read as deletions)
L5 · OFICIO · "The ignore probe asks both spellings in one git process" (porque git refuses -q with two pathnames but still exits 0 when any of them is ignored)
L6 · OFICIO · "The suite wall is recorded as a range, never as the best run" (porque the machine is shared, and the honest signal for this change is the work removed: processes and warm per-verb latency)
