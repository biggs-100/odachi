# Next session: a lane that lands, and the machine that runs it

Written at the close of 0166, on **v0.11.99** -released, installed and green on this machine:
`doctor` 14 checks, no problems, and `main` clean and in sync with `origin`. The chat is gone
tomorrow; this file is the state.

## What 0166 shipped

A change that fits one line is written as one line:

- `odachi start --light "<requirement>"` opens a feature with a task doc that carries the
  requirement quoted, one step with the same words and `Light: yes`, and nothing else: no brief,
  and `odd/fence.toml` and `odd/profile.toml` stay exactly as they were.
- `check` does not ask a light feature for a brief -the one line is the vision- and asks for the
  fence, the open questions and the proof exactly as before. The lane closes like any feature,
  because the one step is what the gate counts, and `plan` refuses a second step: a change that
  needs one needs a brief.
- `Light: yes` lives in the preamble, where `Base` and `Review: waived` already live, so the
  artifact carries the shape and no command line has to remember it. The write guard now knows
  all three preamble lines; before, a document with `Review: waived` could not be rewritten at
  all, and 0166 paid that too.

The review approved with one warning left in `docs/review-debt.md`: two of the lane's refusals
have no test.

## The landing order, paid again

The fence on `main` protected `internal/app/check/**`, and this feature changes what `check`
asks. CI reads the fence from the **base ref**, so the feature was refused on all three
platforms while the local gate passed. The repair is the one 0154 wrote: a **policy-only** commit
-`brief.md` and `odd/fence.toml`, no task doc- landed first (PR #256), and the feature rebased on
top. `check --ci` names it before the push, and it is worth running before the first push of any
feature that touches a fenced path.

## What is measured, so nobody re-measures it

- **The whole suite takes ~360 s on this machine**, and `check --run` needs `--timeout` of 1500 to
  be safe. The `cmd/odachi` package alone is ~94 s; the proof output's number is that package, not
  the total.
- `proof/` is gitignored, so the proof never moves the tree it certifies. The order that closes a
  feature: write the closing documents, run `check --run` on that tree, then commit.
- **Killing `odachi` mid-proof leaves `go test` running on Windows.** That orphan holds the Go
  build cache lock, and every later `go test` waits on it without printing anything, so
  `check --run` reports a timeout with empty output. Before blaming the suite, look for a stray
  `go test` process and kill it.
- The review path that works: `odachi review --base <sha>` writes the prompt into
  `.git/odachi/scratch/`, a subagent reads that file and answers with the JSON, and
  `odachi review --base <sha> --attach <answer.json>` files the verdict.
- `doctor`: 14 checks, no problems. The four lenses run on `opencode-go/mimo-v2.6-flash`.

## What is open

- `docs/review-debt.md` carries one row: the light lane's untested refusals.
- A friction this close found and nobody has named in the queue: **`check --run` reports a hung
  command as a timeout with empty output**, and the human cannot tell "slow" from "blocked on a
  lock". Naming the child process, or writing the output as it arrives, would turn a 17-minute
  mystery into one line.

## What is already true, so it is not re-decided

- **Ownership is declared, never inferred** (ADR 0038), and a failed read is not an absence.
- **The review gate lives in the clone** (ADR 0042). CI cannot verify a review.
- **Four artifacts are the memory**: brief, task doc, proof and profile.
- **The changelog goes back to `## Unreleased`**; the release PR renames it and the tag publishes.
- **Every irreversible action names its target and waits for the human.**
- **A gate judges; it does not repair.** A record is repaired by the verb that made it.
- **What a parser cannot read, it refuses.** Tolerating a line is not the same as guessing it.
- **A fence rule names a file the feature does not touch.** A policy that a feature needs lands
  alone, before it.

## The one decision that waits, and is not a polish

**CI cannot enforce the review.** ADR 0042 says what would change it: a trusted job that runs
`odachi review` with its own model and key, or a receipt signed outside the tree.

## Queued

- **the row, paid.** Three tests, one per refusal of the light lane.
- **a proof that says why it is waiting.** `check --run` should name the command it launched and
  what it is blocked on, or stream what arrives, so a lock is not read as a slow suite.
- **a change resumable from the diff needs no document** is now partly built: the light lane is
  the cheap half. What is left is the half without a task doc, and nothing has measured it.

## How to find the next real work

`docs/review-debt.md` names what a verdict left; `doctor` and `bench` find friction nobody has
named; the hygiene trio (`go vet`, `go mod tidy -diff`, `gofmt -l .`) is cheap; and `clean` says
which class of file grew back.
