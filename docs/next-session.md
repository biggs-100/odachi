# Next session: the ledger is empty, and what a name taught

Written at the close of 0165, on v0.11.97. The chat is gone tomorrow; this file is
the state.

## What 0165 shipped

The two rows 0163 and 0164 left, paid in the place where the name is read:

- **A step tail with two commit parts stays text.** `R1 · tree:abc · note` used to keep the last
  part as the commit and lose the first on the next canonical write. Two parts in one slot are
  ambiguous, so the parser reads no tail at all and the line round-trips unchanged.
- **`--artifact` reads a bare name.** The name `--list` prints answers which kind it is, so
  `--file brief --artifact 0002-x.md` is refused as the contradiction it is instead of claiming
  there is no backup of a document the run itself hid. A name outside the four is refused with
  the four paths odachi keeps, not with the versions this run happened to look at.

The review approved with no findings, on the first round, and `docs/review-debt.md` is empty.

## The lesson this feature paid for

A parser that tolerates must also **refuse** what it cannot read. Both rows were the same shape:
a name that carried one meaning was read as another, and the tool answered with a statement that
was false -a line rewritten, a document reported missing. The fix is not "guess better": it is to
say which reading applies, and to answer with what the tool keeps when none does.

## What is measured, so nobody re-measures it

- `go test ./...` is green on this tree; the `cmd` package alone takes ~100s, so `check --run`
  wants `--timeout` of 400 or more here.
- The order that closes a feature: write the closing documents first, run `check --run` on that
  tree, then commit. The hook refuses a close whose proof belongs to an earlier tree. `proof/` is
  gitignored, so it never moves the tree it certifies, and a release note written *after* the
  close is paperwork: `check` stays ready and names the tree the proof belongs to.
- The review path that works: `odachi review --base <sha>` writes the prompt into
  `.git/odachi/scratch/`, a subagent reads that file and answers with the JSON, and
  `odachi review --base <sha> --attach <answer.json>` files the verdict. `--exec` with
  `opencode run` still stalls.
- `doctor` on this tree: 14 checks, no problems, on v0.11.97 installed. A copy of a skill in the
  home only catches up with a binary built from the branch that changed it.
- The four lenses run on `opencode-go/mimo-v2.6-flash`, 0.14/0.28 per million tokens.

## What is open

- `docs/review-debt.md` is empty: the last two verdicts left nothing open.
- The queue below, which is now one item.

## What is already true, so it is not re-decided

- **Ownership is declared, never inferred** (ADR 0038), and a failed read is not an absence.
- **The review gate lives in the clone** (ADR 0042). CI cannot verify a review.
- **Four artifacts are the memory**: brief, task doc, proof and profile.
- **The changelog goes back to `## Unreleased`**; the release PR renames it and the tag publishes.
- **Every irreversible action names its target and waits for the human.**
- **A gate judges; it does not repair.** A record is repaired by the verb that made it.
- **What a parser cannot read, it refuses.** Tolerating a line is not the same as guessing it.

## The one decision that waits, and is not a polish

**CI cannot enforce the review.** ADR 0042 says what would change it: a trusted job that runs
`odachi review` with its own model and key, or a receipt signed outside the tree.

## Queued

- **a change resumable from the diff needs no document.** The light lane: no brief and no task
  doc when the work can be resumed from the diff alone. Expensive, because it changes rule 1, and
  it is the only item left that nobody has measured.

## How to find the next real work

The ledger is empty, so the next work comes from measurement instead of a verdict: `doctor` and
`bench` find friction nobody has named, the hygiene trio (`go vet`, `go mod tidy -diff`,
`gofmt -l .`) is cheap, and `clean` says which class of file grew back.
