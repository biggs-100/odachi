# Next session: the craft has a trust boundary

Written at the close of 0158, on v0.11.90. The chat is gone tomorrow; this file is
the state.

## What 0158 shipped

- **A skill, `security`, loaded by situation**: input from outside, a permission, a secret,
  a dependency, a log line, a failure, before a release. Thirteen rules and four prohibitions
  in 39 lines, and every claim about an outside system carries its source and version -OWASP
  Top 10:2025, ASVS v5.0.0 (2025-05-30), NIST SSDF 1.1, `govulncheck`- because a security
  claim ages like a dependency.
- **ADR 0046 says which half the gate enforces.** The fence over paths, the proof over a
  command the human chooses and the risk lens over the diff are the enforcement; the trust
  boundary, the owner of an object, the life of a secret and what failing closed means are
  judgement. No scanner, and the ADR says why: a pattern list measures the pattern.
- **The impact map stopped counting skills by hand.** Its skills row had already lost
  `design`; it now carries the whole payload and `internal/archtest/impactmap_test.go` pins it
  against `skills.Names()`, so a new skill costs one line of the map instead of a claim nobody
  measures.

What this feature did not do, on purpose: it did not touch the review lens, it added no
scanner and no doctor check, and it changed none of the seventeen skills that already shipped.

## What is open

- `docs/review-debt.md` carries the rows this close's verdict leaves.
- The queue below, in the order the evidence argues for it.

## What is measured, so nobody re-measures it

- `doctor` on this tree: fourteen checks, no problems; the home copy of the craft still
  carries seventeen of the eighteen skills until the release installs it (`odachi init` does
  it; the doctor compares what is installed, not what is missing).
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

- **the review's configured model no longer exists.** `odd/review.toml` assigns the four
  lenses to `opencode-go/mimo-v2.6-flash`, and the provider now serves `mimo-v2.6-pro`,
  `mimo-v2.5` and `mimo-v2.5-pro`. The subagent path survives because the model is chosen by
  hand at the close; `review --exec` would fail. A small feature: the file, a test that the
  model resolves, and a refusal that names the models it can see.
- **the step declares its route.** inline or delegated, and which agent, recorded in the task
  doc. The one fact gentle-ai writes per task (`Route:`) that odachi does not. Cheap.
- **bookkeeping nobody can put back.** `plan --done T2` records no tree when the worktree is
  not the commit (0147's and 0157's T2 both), and `undo --file task` is a dead end when
  several task documents have backups: it discovers all of them and then asks again with no
  way to name one. Two symptoms, one family: a record the tool writes and cannot repair.
- **a change resumable from the diff needs no document.** The light lane: no brief and no task
  doc when the work can be resumed from the diff alone. Expensive, because it changes rule 1.

## What this session does first

The model the review configures, because it is the one item on this list that is already
broken and turns a future `review --exec` into a dead end. Then the queue, in the order above.

## How to find the next real work

`docs/review-debt.md` names what a verdict left; `doctor` and `bench` find friction nobody
has named; the hygiene trio (`go vet`, `go mod tidy -diff`, `gofmt -l .`) is cheap; and
`clean` says which class of file grew back.
