# 0016 · Review receipts by content: findings per blob with a deterministic floor

Status: accepted (implemented in T1–T6; step commits in `odd/tasks/0016-review-verdicts-carry.md`) · Date: 2026-10 · Session: v0.11.2

## Context

ADR 0003 made the verdict a receipt named after the tree, and ADR 0010 made the
review proportional (path-based risk, human-owned switch). Both stand. What
they left slow is the fix loop: the verdict is all-or-nothing per tree hash, so
correcting one file in an otherwise untouched candidate invalidates every
finding — including the findings about files that did not change a byte. A
working session that fixes a reviewer's finding pays a full re-review, and an
agent under that pressure is tempted to commit unreviewed fixes.

External evidence, gathered before deciding:

- Nix and Bazel answer the same question ("revalidate only what changed") with
  content addressing, not environment addressing; Bazel re-runs only what its
  dependency graph proves changed and skips tests that did not.
- Three independent LLM review tools (Bito, prism-reviewer, CodeRabbit-class)
  converged on incremental review with a per-finding fingerprint; one reports
  ~90% token savings on push updates, with no loss of coverage.
- DiffGate and Revet run a deterministic floor (path/AST rules) before any LLM,
  scope findings to the changed lines, and escalate a finding to a gate only
  when the evidence proves the risk is reachable ("earn the right to block").
- Claude Code review and inline.dev discard findings the reviewer cannot defend
  from the code it read, and keep the discard reason.
- A promptfoo fix (PR #7192) showed a wrongly scoped diff (two-dot instead of
  merge-base) produces false positives at scale: the candidate must be defined
  correctly before any lens runs.
- Meta's review research found satisfaction tracks the P75 of review time and
  that diff size barely predicts latency; idle waiting dominates. For odachi
  the P75 is the fix loop: finding → corrected tree → next valid commit.

None of this changes what a verdict may claim. It changes what a verdict may
carry over when the tree it names is gone.

## Decision

- **Findings gain content identity.** Every finding is keyed by
  `(path, blob hash)` — the blob id git already owns — plus lens, status
  (`open`/`dismissed`) and, when dismissed, the reason. Blob hashes are
  derived from git objects and must never disagree with them: they are cache
  fields, never a second source of identity.
- **The verdict file still names the tree.** `.odachi/reviews/<tree>.json`
  keeps ADR 0003 storage untouched: the approval *is* the file name. What
  changes is its interior. "One tree, one verdict" stays; what a new verdict
  may inherit are per-finding facts from a previous verdict whose blobs still
  match — never an approval.
- **Potency is per blob, not per tree.** On a new candidate: a finding whose
  blob is identical stays potent — open findings remain open (they keep
  failing the gate; they never excuse anything), dismissed findings remain
  dismissed with their reason. Only changed blobs re-run the lenses that
  apply to their path class. The review-shy fix loop becomes: re-review the
  two files you touched, inherit the rest.
- **Lens selection stays path-based; scope becomes per file.** ADR 0010's
  tiering (code → `reliability`; contracts, concurrency, delivery → the four
  lenses) is unchanged, but assignment can differ per file inside one
  candidate: `init.go` may answer `reliability` while `api.go` answers all
  four.
- **Coverage is the gate's arithmetic.** For every changed blob, each
  path-class lens must have a current answer — a finding, a `No findings.`
  pass, or a dismissal with reason. A blob without coverage is an exit 2
  naming `odachi review`, exactly the missing-verdict path today. An old
  verdict without per-blob keys grants no coverage and is reported as not
  computable, not silently shuffled in: an approval the tool cannot prove is
  an approval it must not use.
- **A deterministic floor runs before any lens.** Path/AST rules on the
  changed lines — no LLM, milliseconds, reproducible. The floor (a) detects
  the path class that selects the lenses, deterministically instead of by
  filename alone, and (b) hands the lens its line set: lens findings must cite
  a flagged line or its immediate context. That is "earned blocking" in the
  lenses' vocabulary: a finding without a defensible citation goes down a
  tier instead of up.
- **Dismissals are first-class and reusable.** A dismissal stores its reason
  and a fingerprint (path, line neighborhood, lens, content signature); the
  same dismissed defect is never re-raised in later candidates. The verdict
  consequently says what it looked at and why it discarded — a review that
  cannot defend its silence does not stay silent, it records why.
- **Full-candidate diffs use merge-base.** The two-dot diff surfaces other
  branches' changes as deletions and manufactures false positives; the
  merge-base diff is the candidate, as promptfoo learned in production.

## Alternatives rejected

- **Review at the merge point instead of per commit.** Faster per commit, but
  it lets sensitive candidates pass several commits unread — speed bought by
  weakening the gate. The layering stays: proof per commit (hard, automatic),
  review per candidate (opt-in, human switch).
- **A correction budget** — RDD's machinery ADR 0010 rejected; per-blob
  potency achieves the same savings with exact identity instead of a
  bookkeeping allowance.
- **Multi-model consensus** (several models vote to cut false positives):
  models share training bias, so agreement correlates error more than it
  removes it, and it triples cost. The validation floor is the cheaper,
  deterministic version of the same goal.
- **Signing and trust ladders** (receipts signed by protected refs): delivery
  machinery a single-repo CLI does not need; `.odachi/reviews/` stays out of
  the tree per ADR 0003, and the human-owned switch per ADR 0010 is the
  authority story.

## Consequences

- The fix loop stops paying for what it did not touch: the P75 it optimizes is
  finding → next valid commit, and that loop shrinks to the changed files.
- Old findings persist across corrections, which makes the review cumulative
  inside a feature instead of amnesiac: the fourth candidate does not
  re-litigate the first three.
- The verdict format gains fields; verdicts written before this decision carry
  the same tree-name validity but cannot serve coverage and force one fresh
  review — a one-time cost per old tree, not a migration.
- The review remains opt-in and human-owned (ADR 0010); nothing here changes
  the switch, only the price of using it.

## Sources

- promptfoo PR #7192: merge-base vs two-dot diff false positives.
- DiffGate (srbsa/diffgate) and Revet: deterministic floor, diff-scoped
  findings, earned blocking.
- prism-reviewer (vyoman-labs): per-finding content fingerprints; Bito:
  incremental review; inline.dev: suppressed findings keep their reason.
- Claude Code Code Review: validation pass discards undefended findings.
- Nix and Bazel docs: content addressing and incremental correctness.
- Meta engineering (NudgeBot, Time In Review): review latency research,
  P75 satisfaction correlation.
