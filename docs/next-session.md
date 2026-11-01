# Next session: a retired model says the line that chose it

Written at the close of 0159, on v0.11.91. The chat is gone tomorrow; this file is
the state.

## What 0159 shipped

- **The review keeps the model it had, and a failed run says where the model came from.** The
  first measurement hid `mimo-v2.6-flash` -the provider's list shows only the newest version of
  each family- and this feature briefly moved the four lenses to the pro; the full list showed
  the flash one **active at 0.14/0.28**, and the host deprecates `mimo-v2.5` in its favour, so
  `odd/review.toml` keeps the value it already had. What ships is the comment that says what to
  check when a model goes, an archtest that every lens has a judge, and a failure that names
  the lens, the model and the line that chose it: `--model`, `odd/review.toml` or the session.
- Lesson, and it cost a review finding: **a filtered list is not a measurement.** The provider's
  list shows one version per family unless every version is asked for, and a missing model in
  that answer is not a retirement.
- **A reviewer that cannot start says who it asked and where that came from.** The three
  failure paths of `--exec` -could not run, timed out, non-zero exit- name the lens, the model
  and its source: `--model`, `odd/review.toml` or the session. An offline tool cannot know
  what a provider still serves; it can make a retirement one edit instead of an investigation.
- What it did not do: no network, no provider query, no automatic retry with another model,
  and the four lens prompts are untouched. A verdict from a model nobody chose is a review
  nobody asked for.

## What is open

- `docs/review-debt.md` carries the rows this close's verdict leaves.
- The queue below.

## What is measured, so nobody re-measures it

- `doctor` on this tree: fourteen checks, no problems. v0.11.91 is installed and one
  binary is in the PATH (`%LOCALAPPDATA%\odachi\bin`); the stale copy in `~/bin` is gone.
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

- **`evidence_class` and a 2-of-3 vote on refutations.** The comparison with gentle-ai's
  lenses argued for it: a finding declares deterministic or inferential evidence, the
  deterministic ones skip the refuter, and an inferential `BLOCKER`/`CRITICAL` needs two of
  three to fall. An old refutation without votes keeps today's meaning, and the schema counter
  moves (ADR). The comparison itself: odachi's lenses carry 18 concrete patterns each against
  gentle-ai's 8-9 rules, but gentle-ai votes on refutations and sorts evidence by class.
- **Four patterns the lenses are missing**: `test.only` that lets CI pass and misallocated
  coverage (`reliability`), operational thresholds and observability for what can fail in
  production (`resilience`), and a PR that does not explain what it touched when that makes it
  unsafe to review (`readability`). Four lines in existing assets.
- **the step declares its route.** inline or delegated, and which agent, recorded in the task
  doc. Cheap.
- **bookkeeping nobody can put back.** `plan --done T2` records no tree when the worktree is
  not the commit, and `undo --file task` is a dead end when several task documents have
  backups.
- **a change resumable from the diff needs no document.** The light lane: no brief and no task
  doc when the work can be resumed from the diff alone. Expensive, because it changes rule 1.

## How to find the next real work

`docs/review-debt.md` names what a verdict left; `doctor` and `bench` find friction nobody
has named; the hygiene trio (`go vet`, `go mod tidy -diff`, `gofmt -l .`) is cheap; and
`clean` says which class of file grew back.
