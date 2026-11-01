# Next session: a block falls by evidence

Written at the close of 0160, on v0.11.92. The chat is gone tomorrow; this file is
the state.

## What 0160 shipped

- **A finding declares its evidence: cited or argued.** `cited` means the reviewer quoted the
  line of the diff that shows the defect; `argued` means it reasoned without one. The field is
  optional and **absence reads as argued**, so every verdict written before this release keeps
  the strict meaning, and an invented class is refused by name.
- **The class decides the quorum.** A cited finding falls to one cited refutation; an argued one
  needs two; a `stands` is never a vote to remove, and `refuted` without a citation is no vote
  at all. Three witnesses is the ceiling, because a block decided by a crowd is a block nobody
  can read.
- **Each answer keeps its own file.** The first is still `<tree>.refuted.json`; the next two are
  numbered. No stored answer is rewritten -0044's promise about evidence applies to refutations
  too- and a tree refuted before the numbering is still read as round 1.
- **The gate says how many it has**: the finding's id, the cited refutations it has and the ones
  it needs. The verdict schema moved to 4, so an older binary says "upgrade" instead of "unknown
  key". ADR 0047 names the rule and its limit: the class is the reviewer's account, and the tool
  cannot verify that a citation quotes the line it names.

## What is open

- `docs/review-debt.md` carries the rows this close's verdict leaves.
- The queue below.

## What is measured, so nobody re-measures it

- `doctor` on this tree: fourteen checks, no problems. v0.11.92 is installed and one binary is
  in the PATH (`%LOCALAPPDATA%\odachi\bin`).
- `check --run` needs `--timeout 400` for this suite on this machine.
- `opencode run` as `--exec` still stalls to the 600s timeout here: the review path that works
  is `review --base <sha>` into `.git/odachi/scratch/`, a subagent answers it with a model that
  is not the proof's, and `review --base <sha> --attach <answer.json>` files the verdict.
- `bench`: 11 journeys and 29 steps, no dead ends.

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
