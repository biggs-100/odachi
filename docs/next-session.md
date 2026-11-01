# Next session: the block closes itself, and the one decision that waits

Written at the close of 0146, on v0.11.85 plus one feature. The chat is gone
tomorrow; this file is the state.

## What the last three features shipped

- **0144 (v0.11.85)**: the review never authorises delivery. With `mode = on`, a close
  without a verdict is ready and names the review it did not get; only a causal
  `BLOCKER` in the verdict that exists stops a ship, and `--require-review` is the one
  refusal (ADR 0043).
- **0145 (v0.11.85 base)**: `.github/workflows/gate.yml` runs `check` without
  `--no-review`, because the base it is built from names the missing review instead of
  refusing the tree. The flag waited for the release that reads the new rule.
- **0146 (unreleased)**: the block carries the close —the phrase a human uses, the file
  to rewrite, what goes in it— and `doctor` gained a fourteenth check that warns when
  `docs/next-session.md` does not name the newest feature. This file is the first one
  that check reads.

## What is open

- `docs/review-debt.md` has no open rows, and `odd/tasks` has no open questions.
- The feature that closed today owes its proof and its review; both are named below.

## What is measured, so nobody re-measures it

- `doctor` on this tree: fourteen checks. The binary on the PATH is v0.11.85 and still
  says thirteen until the next release is installed — a fact, not a defect.
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

## How to find the next real work

`docs/review-debt.md` names what a verdict left; `doctor` and `bench` find friction nobody
has named; the hygiene trio (`go vet`, `go mod tidy -diff`, `gofmt -l .`) is cheap; and
`clean` says which class of file grew back.
