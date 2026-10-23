# 0010 · Review mode and risk: RDD's result without RDD's machinery

Status: accepted · Date: 2026-10 · Session: v0.11.1

## Context

Odachi already had the core of a receipt-driven review: a proof tied to the tree
hash, a verdict stored as `.odachi/reviews/<tree>.json`, four lenses selected by
the changed paths, a refuter that needs a citation, and the optional
`check --require-review` gate. What it did not have was RDD's proportionality and
its human-owned switch:

- `check --require-review` asked for a verdict even when the candidate was only
  documentation, where there is no lens to answer.
- The switch was a flag on every command; the pre-commit hook ran plain
  `check`, so a project could not say "review is on here" once and have the gate
  obey it.
- The sensitive signals stopped at auth, money, data shape and deployment, and
  did not see public contracts or concurrency.

RDD solves the same problem with an assessor, consent envelopes, receipts with
lineage, transactions, leases and a review lifecycle. Those are delivery
machinery for a large team; odachi's candidate is already exact (the tree hash),
the verdict is already the receipt (the file name), and a corrected candidate is
a new tree, not a correction budget.

## Decision

- **Passive candidates ask for no review.** `check --require-review` computes the
  required lenses before reading the verdict. If the candidate asks for none,
  there is nothing to require and the gate continues. A malformed or absent
  verdict is not consulted for a passive candidate.
- **The switch is a project file, and the human owns it.** `odachi review mode
  status|on|off` reads or writes `.odachi/review.toml` (`mode = "on"|"off"`). A
  missing file means `off`, preserving the opt-in default. The file is versioned:
  it is a gate decision, not generated evidence.
- **`check` reads the switch; flags override it for one run.**
  `--require-review` forces on, `--no-review` forces off, and passing both is an
  error. A malformed switch blocks with the file named instead of quietly
  meaning off.
- **Risk is still path-based and volume-blind, but wider.** The full tier now
  includes public contracts (`api`, `contract`, `openapi`, `proto`, `graphql`,
  `sdk`), concurrency (`concurrency`, `race`, `mutex`, `lock`, `async`, `worker`,
  `queue`) and delivery (`ci`, `.github`, `deploy`, `release`, `goreleaser`,
  `terraform`, `helm`, `k8s`, `pipeline`). No LLM assessor, no line counts.
- **No RDD lifecycle was copied.** There is no consent envelope, no
  start/status transition, no transaction or lease, no lineage service, no
  `review_due` per commit and no orchestrator. The tree hash is the candidate;
  one verdict per tree is the receipt; a correction is a new tree.

## Consequences

- A docs-only change no longer pays for a review it cannot use, even when the
  project mode is `on`. A code change still asks for `reliability`; a contract,
  concurrency or delivery change asks for the four lenses.
- The pre-commit hook needs no change: it already runs `odachi check --quiet`,
  and `check` now sees the project switch. A project that wants review turns it
  on once; a project that does not leaves the file absent.
- The switch lives beside the other gate settings, not inside the frozen profile.
  `review mode on|off` backs up an existing file before replacing it; `status`
  and an unchanged `on`/`off` write nothing.
- The gate still never fabricates approval: without a verdict for this tree, a
  non-passive candidate with the mode on exits 2 and names `odachi review`.
- The remaining difference from RDD is operational, not evidentiary: no
  transactional store, no multi-reviewer lineage, no automatic risk assessor.
  Those are needed for automated delivery at scale, not for deciding whether
  this candidate has a design defect.
