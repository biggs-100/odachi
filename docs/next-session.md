# Next session: the bookkeeping a step leaves behind

Written at the close of 0163, on v0.11.95. The chat is gone tomorrow; this file is
the state.

## What 0163 shipped

A step now says who runs it, on its own line in the task doc:

- `odachi plan "<text>" --route inline` when the parent runs it, and `--route delegated
  --agent odachi-worker` when another agent does. The route sits beside the requirement and
  the tree, so the document says who did the work instead of a chat nobody keeps.
- A contradiction is refused before the line exists: a delegated step with no agent, an
  inline step with one, and an agent with no route. The message names both valid calls.
- A step written before the change reads as **not declared**, never as inline. The tool
  records a declaration and never checks it: `check`, `review` and the fence did not learn
  to read a route, and the fence protected them from learning it.
- `drive` sends a delegated step to the flag instead of the craft note it used to ask for.

The first review found a real regression and the tree that stands paid it: a step closed with
a bare `--commit` was written `   · cafe123`, and a sha whose letters run into digits matched
the requirement pattern, so the mark parsed as a requirement. The lone dot now says the first
part is not a requirement, and that old spelling round-trips. The re-review approved
(`3348e25`) with one row left in `docs/review-debt.md`: a hand-written tail with **two**
commit parts keeps the last and drops the first.

## What is measured, so nobody re-measures it

- `go test ./...` on this tree: green, and it is what `proof/0163-*.txt` holds. The `cmd`
  package alone takes ~100s; give `check --run` a `--timeout` of 400 or more on this machine.
- `proof/` is in `.gitignore`, so the proof does not move the tree it certifies. The order
  that closes a feature is: write the closing documents first (`docs/next-session.md`,
  `docs/review-debt.md`, the task doc's last step), run `check --run` on that tree, then
  commit. The hook refuses the close while the proof belongs to an earlier tree, because the
  tree that closes is the one that owes the evidence; a release note written *after* the
  close is paperwork and needs no new proof.
- The review path that works here: `odachi review --base <sha>` writes the prompt into
  `.git/odachi/scratch/`, a subagent reads that file and answers with the JSON, and
  `odachi review --base <sha> --attach <answer.json>` files the verdict. `--exec` with
  `opencode run` still stalls to the timeout; do not spend the time again.
- `doctor`: 14 checks, no problems, 2 warnings -`~/.agents/skills/drive` is stale until
  `odachi init --skills home` runs, and this file is stale until it names 0163, which it now
  does.
- Cost for the four lenses: `mimo-v2.6-flash` 0.14/0.28 per million tokens, the model
  `odd/review.toml` configures. This review ran on it.

## The fact that cost a step's mark

`odachi plan --done T2` records the tree of the step's commit only when the worktree is that
commit. The task doc itself is a working-tree file, so after the first `--done` the worktree
is dirty and every later step records no tree at all -T2 of 0163 has none. Until that is
fixed, close a step as `odachi plan --done T3 --commit HEAD`; the commit's bytes are the mark
either way.

## What is open

- `docs/review-debt.md` carries the row this close's verdict leaves (the two-part tail).
- The queue below.

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

- **bookkeeping nobody can put back.** `plan --done T2` records no tree when the worktree is
  not the commit (0163 paid this), and `undo --file task` is a dead end when several task
  documents have backups. Now first, because it bit a step of the last feature.
- **a tail with two commit parts keeps only the last.** The row 0163's re-review left; the
  second non-route part should be refused the way an unknown tail is, so the line stays text.
- **a change resumable from the diff needs no document.** The light lane: no brief and no
  task doc when the work can be resumed from the diff alone. Expensive, because it changes
  rule 1.

## How to find the next real work

`docs/review-debt.md` names what a verdict left; `doctor` and `bench` find friction nobody
has named; the hygiene trio (`go vet`, `go mod tidy -diff`, `gofmt -l .`) is cheap; and
`clean` says which class of file grew back.
