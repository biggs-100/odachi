# Next session: the lenses name what they missed

Written at the close of 0162, on v0.11.95. The chat is gone tomorrow; this file is
the state.

## What 0162 shipped

Four defects the comparison with gentle-ai's lenses found missing, in the prompts that look at
a diff:

- **reliability** now names a focused or skipped test -`.only`, a disabled sibling- that makes
  CI pass by shrinking the suite, and a test that proves one function through a whole flow,
  which costs more than the behaviour it proves.
- **resilience** asks for a threshold the project chose -without a number, "it is fine" is a
  guess- and for something in production that would show an expected failure: a counter, an
  event, an alert.
- **readability** names behaviour the diff changes and the brief never asked for: the two
  artefacts a reader needs are the diff and the document, and the lens holds both.

The four prompts stay within the format's bounds (twelve to twenty rules, each a line a reviewer
can act on), no number was borrowed from another project, and the engine -the prompt builder and
the refutation rule- was fenced out of the feature on purpose.

## The policy that had to land first, and why

The fence protected the four lens prompts because 0161 -a feature about who judges- wrote that
rule there, and a fence rule lives in the tree until the next feature rewrites it. CI reads the
fence from the **base ref**, so the feature that improves a lens was refused on all three
platforms while the local gate passed: the worktree fence was already the new one.

The repair is the landing order 0154 wrote down, done again (PR #248): a policy-only commit
-brief and `odd/fence.toml`, no task doc- merged first, then the feature rebased and force-pushed.
The fence now protects `proof/**` and the review's own question, and names no lens asset.

Two facts for the next time, so they are not paid again:

- **`odachi check --ci` answers this before the push.** It resolves the base from the remote and
  reads the fence there; it was run after the rebase, when it said `fence: origin/main` and
  `ready`. It belongs before the first push of any feature that touches a fenced path.
- **`start` numbers a task doc from what the branch sees.** Creating a policy feature on `main`
  while a feature branch holds the next number collides (two 0162s). The precedent commit
  (138c783) lands a policy with no task doc, which also avoids the collision.

## What is open

- `docs/review-debt.md` carries the rows this close's verdict leaves.
- The queue below.

## What is measured, so nobody re-measures it

- `doctor` on this tree: fourteen checks, no problems. v0.11.95 is installed and one binary is
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
