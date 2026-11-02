# Next session: two rows that are both about a name

Written at the close of 0164, on v0.11.96. The chat is gone tomorrow; this file is
the state.

## What 0164 shipped

The two bookkeeping dead ends a close can leave, closed at the verb that made them:

- **A step that closed with no tree takes the mark afterwards.** `odachi plan --done T2 --commit
  <sha>` records it on a step that is already done and carries no mark; a step that already
  carries one is still a no-op, and with no `--commit` the message names the flag. The mark 0163
  lost to a dirty worktree could not be written by any command before this.
- **`undo` names the document it restores.** `--artifact odd/tasks/0164-....md`, or the bare file
  name `--list` prints, restores that one and backs up what it replaces. With several documents
  of one kind the refusal now names `--artifact`; before, `--file task` demanded `--file`, the
  flag the run already had.

Both come from the same defect: the record exists, and the verb cannot be pointed at it. The
review approved with one warning left in `docs/review-debt.md`, and the 0163 row is still there
too: both are about a name that is read wrong.

## What is measured, so nobody re-measures it

- `go test ./...` is green on this tree; the `cmd` package alone takes ~100s, so `check --run`
  wants `--timeout` of 400 or more here.
- The order that closes a feature: write the closing documents first, run `check --run` on that
  tree, then commit. The hook refuses a close whose proof belongs to an earlier tree. `proof/` is
  gitignored, so it never moves the tree it certifies, and a release note written *after* the
  close is paperwork: `check` stays ready and says which tree the proof belongs to.
- The review path that works: `odachi review --base <sha>` writes the prompt into
  `.git/odachi/scratch/`, a subagent reads that file and answers with the JSON, and
  `odachi review --base <sha> --attach <answer.json>` files the verdict. `--exec` with
  `opencode run` still stalls.
- `doctor` on this tree: 14 checks, no problems, after installing v0.11.96. The home copy of a
  skill only catches up with a binary built from the branch that changed it.
- The four lenses run on `opencode-go/mimo-v2.6-flash`, 0.14/0.28 per million tokens.

## What is open

- `docs/review-debt.md` carries two rows, and they are one feature: a step tail with two commit
  parts drops the first, and a bare `--artifact` name skips the `--file` contradiction and makes
  the refusal lie. Both are "a name is read as something it is not".
- The queue below.

## What is already true, so it is not re-decided

- **Ownership is declared, never inferred** (ADR 0038), and a failed read is not an absence.
- **The review gate lives in the clone** (ADR 0042). CI cannot verify a review.
- **Four artifacts are the memory**: brief, task doc, proof and profile.
- **The changelog goes back to `## Unreleased`**; the release PR renames it and the tag publishes.
- **Every irreversible action names its target and waits for the human.**
- **A gate judges; it does not repair.** A record is repaired by the verb that made it.

## The one decision that waits, and is not a polish

**CI cannot enforce the review.** ADR 0042 says what would change it: a trusted job that runs
`odachi review` with its own model and key, or a receipt signed outside the tree.

## Queued

- **the two rows, as one feature.** A tail with two commit parts should stay text instead of
  keeping the last, and a bare `--artifact` name should answer the kind question so the
  contradiction is refused with a message that is true. Cheap and already written down.
- **a change resumable from the diff needs no document.** The light lane: no brief and no task
  doc when the work can be resumed from the diff alone. Expensive, because it changes rule 1.

## How to find the next real work

`docs/review-debt.md` names what a verdict left; `doctor` and `bench` find friction nobody has
named; the hygiene trio (`go vet`, `go mod tidy -diff`, `gofmt -l .`) is cheap; and `clean` says
which class of file grew back.
