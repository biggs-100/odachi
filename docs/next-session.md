# Next session: the block closes itself, and the one decision that waits

Written at the close of 0149, on v0.11.86 plus two features. The chat is gone
tomorrow; this file is the state.

## What the last six features shipped

- **0144 (v0.11.85)**: the review never authorises delivery. With `mode = on`, a close
  without a verdict is ready and names the review it did not get; only a causal
  `BLOCKER` in the verdict that exists stops a ship, and `--require-review` is the one
  refusal (ADR 0043).
- **0145 (v0.11.85 base)**: `.github/workflows/gate.yml` runs `check` without
  `--no-review`, because the base it is built from names the missing review instead of
  refusing the tree. The flag waited for the release that reads the new rule.
- **0146 (v0.11.86)**: the block carries the close —the phrase a human uses, the file
  to rewrite, what goes in it— and `doctor` gained a fourteenth check that warns when
  `docs/next-session.md` does not name the newest feature. This file is the first one
  that check reads.
- **0147 (v0.11.86)**: the check asks one question —does the note name the newest
  feature?— and stopped reading the highest number the note mentions. A bare year is four
  digits too, and the review of 0146 found the warning that never cleared. The fix
  deleted the reading instead of adding a second exception.
- **0148 (unreleased)**: the gate reads the requirements. At the close, a requirement no
  step names is a `NeedsWork` that names it and prints the plan command; mid-feature the
  list stays empty on purpose, because a requirement a change added waits for its step.
- **0149 (unreleased)**: the two rows the review of 0148 left are closed. The plural
  sentence and the open lane are pinned, and both pins were observed red before they
  landed; `docs/review-debt.md` is empty again.

## What is open

- `docs/review-debt.md` has no open rows, and `odd/tasks` has no open questions.
- Nothing: both features closed with their proof and their verdict.

## What is measured, so nobody re-measures it

- `doctor` on this tree: fourteen checks, no problems. v0.11.86 is installed, and this
  repository's own block carries the closing paragraph, so the warning that named the old
  block is gone.
- `opencode run` as `--exec` still stalls to the 600s timeout on this machine: the review
  path that works is `review --base <sha>` into `.git/odachi/scratch/`, a subagent answers
  it, and `review --base <sha> --attach <answer.json>` files the verdict.
- `check --run` needs `--timeout 400` for this suite on this machine.
- `bench`: 11 journeys and 29 steps, no dead ends.

## What is already true, so it is not re-decided

- **Ownership is declared, never inferred** (ADR 0038), and a failed read is not an absence.
- **The review gate lives in the clone** (ADR 0042). CI cannot verify a review, and no
  receipt signed outside the tree exists.
- **Four artifacts are the memory**: brief, task doc, proof and profile. A file nothing
  writes and nothing reads ages by itself — this one now has a reader.
- **The changelog goes back to `## Unreleased`**; the release PR renames it and the tag
  publishes.
- **Every irreversible action names its target and waits for the human.**

## The one decision that waits, and is not a polish

**CI cannot enforce the review.** ADR 0042 says what would change it: a trusted job that
runs `odachi review` with its own model and key, or a receipt signed outside the tree.
Both are features. Until one exists, the review is a discipline held at the close, where
the store is — and the CI gate naming it is not the same as checking it.

## Queued, in the order the comparison with gentle-ai's ODD argued

Read at the start of a session, not now: each one becomes a feature, with its own brief, when
it starts.

- **the proof names what it proves.** `check --run` would record which requirements the closing
  feature declared, so evidence maps to a requirement instead of to a whole tree. It needs the
  fence correction first: `internal/domain/proof/**` and `proof/**` are protected by "no change
  to the proof format", and that rule is the human's to change.
- **the step declares its route.** inline or delegated, and which agent, recorded in the task
  doc. It is the one fact gentle-ai writes per task (`Route:`) that odachi does not. Cheap.
- **a change resumable from the diff needs no document.** The light lane: no brief and no task
  doc when the work can be resumed from the diff alone. Expensive, because it changes rule 1 and
  the floor `check` enforces.
- **a step done without a commit is a step nobody can find.** `plan --done T2` records no tree
  when the worktree is not the commit and says so in passing; 0147's T2 closed that way.

## How to find the next real work

`docs/review-debt.md` names what a verdict left; `doctor` and `bench` find friction nobody
has named; the hygiene trio (`go vet`, `go mod tidy -diff`, `gofmt -l .`) is cheap; and
`clean` says which class of file grew back.
