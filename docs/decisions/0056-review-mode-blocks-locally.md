# 0056 - Review mode blocks locally

Status: accepted — Date: 2026-11-24 — Feature: 0361

## Context

The audit behind feature 0361 found that `mode = "on"` was advisory in practice:
agents skipped review and verification for features 0354-0360, and older probes had
already shown real misses. A close could name missing evidence and still commit
unless its caller supplied `--require-review`; the commit hook inherited that
advisory behavior. Generated rules, skills and help taught a stronger contract
than the gate enforced.

[ADR 0043](0043-the-review-never-authorises-delivery.md) correctly keeps review
evidence local and review approval separate from delivery. Its enforcement
decision nevertheless let missing local evidence pass.
[ADR 0045](0045-the-local-run-reads-the-base-ci-reads.md) also recorded a local CI
reproduction without the explicit review bypass that a clean runner needs.

## Decision

- **Mode on is a local gate.** A non-passive close without a valid verdict for the
  required lenses and changed-file coverage blocks. A partial, stale or
  incomplete verdict is not valid evidence.
- **A named risk requires independent verification.** The close blocks until the
  clone store holds a valid `odachi verify` record for that tree.
- **A strict non-passive close declares its risk.** Its commit message must carry
  a valid `Risk:` line; `Risk: none` must be written as
  `Risk: none - <written reason>` so the absence of a named risk is explained.
- **The feature diff is the candidate.** Review lenses and per-file coverage are
  read from the feature's `Base:` through the closing tree, not from an empty or
  documentation-only closing worktree.
- **The commit hook enforces the close.** It evaluates review, verification and
  the risk declaration for the commit being made.
- **The bypasses stay explicit and scoped.** Mode off makes review and
  verification optional. `--require-review` asks for a valid review of the
  current tree, including a passive tree, and refuses self-review.
  `--no-review` bypasses both review and required verification for one run
  without changing the file.
- **CI uses `--no-review` explicitly.** A clean CI runner has no clone review
  store and makes no reviewer or verifier model calls; the fence and proof still
  run. No second model vendor is mandated.
- **Review approval never authorises delivery.** Commit, push, pull request,
  merge, tag and release remain separate human decisions under ordinary
  repository policy.

## What this does not do

- It does not make local review evidence portable to CI or manufacture a verdict
  for an earlier tree.
- It does not require a second vendor or make CI pay for a model call.
- It does not turn an approved review into permission to deliver anything.

## Consequences

- Local reviewer and verifier model calls are real work, with their own cost and
  timeout, and the gate accounts for that work instead of merely mentioning it.
- `--no-review` is an explicit one-run bypass; using it is visible at the call
  site rather than hidden in CI behavior.
- CI stays deterministic and model-free while still checking the base fence and
  running the proof.
- Existing unreviewed closed trees block if they are rechecked under the new
  policy; the new gate does not retroactively create evidence for them.
