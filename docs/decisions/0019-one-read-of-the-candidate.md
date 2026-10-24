# 0019 — One read of the candidate

Status: accepted · Date: 2026-10 · Session: v0.11.6

## Context

Feature 0019 (which took no ADR) aimed at the same target and missed: `cmd/odachi` went
from 69.5s to ~44s. Its note said the serial doctor tests dominated the tail.
Measuring instead of trusting that note showed something else: the end-to-end
verbs dominated, and each verb measured the same worktree three times.

On Windows a git process costs about 50ms. A `check` with the review gate on
paid: `TreeHash` (rev-parse, read-tree, add, write-tree), `ChangedPaths`
(merge-base, diff --name-only, ls-files), `Diff` (merge-base, read-tree, add,
diff) — about eleven processes, and two of them asked the same merge-base and
staged the same worktree twice. A review paid the same plus its own store
preparation, and `store()` prepared the store again after the reviewer ran.

The three answers were never independent: the tree, the changed paths and the
diff describe one worktree at one instant. Reading them three times was not
three facts; it was one fact paid for three times.

## Decision

- **The Git port answers the candidate once.** `Candidate(root, base)` returns
  the content tree, the changed paths and the diff, from one throwaway index:
  merge-base, read-tree, add -A, write-tree, `diff --cached --name-only`,
  `diff --cached`. The index is the worktree snapshot, so `--cached` answers
  what the old worktree diff plus `ls-files` answered, in one process.
- **The verbs share the value.** `check` uses it for the fence, the lens class
  and the proof tree; `review` uses it for the prompt, the floor, the receipt
  and the store. A second reading of the same bytes is the cost this decision
  removes.
- **The store is prepared once, and the read is repeated only when the write
  can move the tree.** The ignore probe runs once per verb. When the write of
  the store's own ignore file can remove earlier verdicts from the candidate,
  the read happens again; in a project that ran `init`, it does not.
- **After the reviewer, read again.** The reviewer command is arbitrary code; a
  reviewer that edits the tree must not have its answer filed under bytes it did
  not judge. `afterReviewer` drops the read, and the store measures the tree
  that exists now.
- **HEAD's fork point is HEAD.** The default base no longer pays a `merge-base`
  process; a named base still does, because another branch's commits must not
  read as deletions.
- **One ignore probe for both spellings.** Git refuses `-q` with more than one
  pathname but still exits 0 when any of them is ignored, so `path` and `path/`
  travel together in one process.

## Alternatives rejected

- **Cache the read across invocations.** The worktree changes between commands —
  a test writes a file, `--attach` writes a store, `--run` runs a command. A
  cache keyed by anything less than the bytes it describes is the stale-proof
  defect this tool exists to prevent.
- **A fake git in the e2e tests.** It would make the suite fast and prove
  nothing about the product's real edge. The evidence is a real repository.
- **More parallelism.** Feature 0019 measured 8-way oversubscription as worse:
  the suite is process-bound, not CPU-bound, and the machine is shared.
- **Skip tests or `-short` in the proof.** The proof is the product's evidence;
  a faster proof that runs less is not a proof.
- **Direct reads of git's object database.** It would remove the processes and
  take on a private format for a public contract. Not worth it for 50ms.

## Consequences

- A review drops from about 13 git processes to 6; a check from about 11 to 5-6
  when the fence and the gate ask for paths and diff. The isolated refute e2e
  test went from 71 to 62 git processes.
- Warm single-verb latency on this machine: `review --show` 1029ms → 351ms,
  `check` ~1009ms → ~340ms.
- The suite wall is machine-bound. Measured on the same tree: 23.8-24.6s in a
  quiet window, 25.6-36s while the machine carried background load; the target
  of 25s is met on an idle machine, not guaranteed on a shared one. The honest
  signal for the change is the work removed (processes, per-verb latency), not
  the best wall-clock run.
- No new dependency; the domain stays stdlib-only and knows nothing about git.
- `check` on a project with no fence and no review gate still computes the diff
  it does not read. It is one process, and a lazy read would reintroduce the
  divergence between verbs that this decision removes.

## Sources

- The 0019 measurement (`docs/decisions` has no entry; the task doc 0019 keeps
  the numbers) and this session's `GIT_TRACE2_EVENT` count.
- The three old readings: `TreeHash`, `ChangedPaths`, `Diff` in
  `internal/adapters/gitx`.
