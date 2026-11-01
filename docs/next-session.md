# Next session: a second reading is not a second opinion

Written at the close of 0161, on v0.11.93. The chat is gone tomorrow; this file is
the state.

## What 0161 shipped

- **The identity is the model, not its setting.** The gate compared the whole model string, so
  `provider/name#high` and `provider/name` read as two judges -and a review by the same model at
  another effort passed as a second opinion it never was. `ModelIdentity` cuts the variant and
  the case, and the sentence a human reads names the effort each side ran with.
- **A self-review is named, not refused, by default.** It already was -every review refusal but
  a causal finding is softened (0043)- and nothing pinned it; there is a test now, and the
  sentence says what it is: *a second reading, not a second opinion*. `--require-review` still
  refuses.
- **A waiver is a decision on the document.** `odachi review waive --because "<why>"` writes
  `Review: waived` into the current feature's task doc and the reason into its history, through
  the same write guard as the other verbs. No reason, no waiver; the proof and the fence stay.
- ADR 0048 says why: the lane exists for a project that owns one model, and the strong guarantee
  is something a project asks for instead of something the default decides.

## What is open

- `docs/review-debt.md` carries the rows this close's verdict leaves.
- The queue below.

## What is measured, so nobody re-measures it

- `doctor` on this tree: fourteen checks, no problems. v0.11.93 is installed and one binary is
  in the PATH (`%LOCALAPPDATA%\odachi\bin`).
- `check --run` needs `--timeout 400` for this suite on this machine.
- `opencode run` as `--exec` still stalls to the 600s timeout here: the review path that works
  is `review --base <sha>` into `.git/odachi/scratch/`, a subagent answers it with a model that
  is not the proof's, and `review --base <sha> --attach <answer.json>` files the verdict.
- `bench`: 11 journeys and 29 steps, no dead ends.
- Cost, measured 2026-11 for the reviewer: `mimo-v2.6-flash` 0.14/0.28 per million tokens,
  `deepseek-v4.1-flash` 0.15/0.60, `mimo-v2.6-pro` 0.435/0.87. The project configures the first
  for the four lenses.

## What is already true, so it is not re-decided

- **Ownership is declared, never inferred** (ADR 0038), and a failed read is not an absence.
- **The review gate lives in the clone** (ADR 0042). CI cannot verify a review.
- **Four artifacts are the memory**: brief, task doc, proof and profile.
- **The changelog goes back to `## Unreleased`**; the release PR renames it and the tag
  publishes.
- **Every irreversible action names its target and waits for the human.**

## The one decision that waits, and is not a polish

**CI cannot enforce the review.** ADR 0042 says what would change it: a trusted job that
runs `odachi review` with its own model and key, or a receipt signed outside the tree.

## Queued

- **Four patterns the lenses are missing**, from the comparison with gentle-ai: `test.only` that
  lets CI pass and misallocated coverage (`reliability`), operational thresholds and
  observability for what can fail in production (`resilience`), and a PR that does not explain
  what it touched when that makes it unsafe to review (`readability`). Four lines in existing
  assets.
- **the step declares its route.** inline or delegated, and which agent, recorded in the task
  doc. Cheap.
- **bookkeeping nobody can put back.** `plan --done T2` records no tree when the worktree is not
  the commit, and `undo --file task` is a dead end when several task documents have backups.
- **a change resumable from the diff needs no document.** The light lane: no brief and no task
  doc when the work can be resumed from the diff alone. Expensive, because it changes rule 1.

## How to find the next real work

`docs/review-debt.md` names what a verdict left; `doctor` and `bench` find friction nobody has
named; the hygiene trio (`go vet`, `go mod tidy -diff`, `gofmt -l .`) is cheap; and `clean`
says which class of file grew back.
