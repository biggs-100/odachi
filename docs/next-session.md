# Next session: the same gate runs before the push

Written at the close of 0157, on v0.11.90. The chat is gone tomorrow; this file is
the state.

## What 0157 shipped

- **`check --ci` is the workflow's rule as one flag.** It resolves the base from the clone's
  remote -`origin/HEAD`, then `origin/main`, then `origin/master`- and reads both the fence
  and the compared paths from that ref, so a branch that loosened its own copy of the policy
  no longer passes locally and fails in CI. It takes no `--base` or `--fence-from` beside it:
  two names for the base would be resolved by precedence, which is the silence the flag
  exists to remove.
- **It never fetches, and it says which ref it read.** The git edge is offline on purpose, so
  freshness cannot be measured here; the report gains `fence: <ref>` and the proof keeps it in
  its `# fence` header. A stale `origin/main` is a green about an older tree, and
  `git fetch origin` is what makes it current. A clone with no base blocks naming the command
  that would create it.
- **ADR 0045** and the README's CI section say the same thing in their own words. The workflow
  keeps `--fence-from "$BASE" --base "$BASE"`: its base is the PR's base commit, which is not
  necessarily the remote's default branch.

Two lessons paid on the way:

- **`start` rewrites the fence with the brief, so a permanent rule can vanish.** This feature's
  fence lost `proof/**` and nothing in the gate noticed -a new brief licenses a new fence-;
  the archtest that pins the evidence boundary is what caught it. The rule is back.
- **`plan --done T2` records no tree from a dirty worktree**, and the mark is what dirtied it:
  the next step can never equal its commit. `plan --done T2 --commit <sha>` is the spelling
  that works, and it is what this feature used after T2 lost its tree once.

## What is open

- `docs/review-debt.md` carries the rows this close's verdict leaves; read it before choosing
  the next feature.
- The queue below is unchanged except where noted.

## What is measured, so nobody re-measures it

- `doctor` on this tree: fourteen checks, no problems. v0.11.90 is installed.
- `check --run` needs `--timeout 400` for this suite on this machine.
- `opencode run` as `--exec` still stalls to the 600s timeout here: the review path that works
  is `review --base <sha>` into `.git/odachi/scratch/`, a subagent answers it, and
  `review --base <sha> --attach <answer.json>` files the verdict.
- `bench`: 11 journeys and 29 steps, no dead ends.
- `undo --file task` is a dead end when several task documents have backups: it discovers all
  of them, then asks again with no way to name one. Not fixed, named here.

## What is already true, so it is not re-decided

- **Ownership is declared, never inferred** (ADR 0038), and a failed read is not an absence.
- **The review gate lives in the clone** (ADR 0042). CI cannot verify a review, and no
  receipt signed outside the tree exists.
- **Four artifacts are the memory**: brief, task doc, proof and profile. A file nothing
  writes and nothing reads ages by itself - this one now has a reader.
- **The changelog goes back to `## Unreleased`**; the release PR renames it and the tag
  publishes.
- **Every irreversible action names its target and waits for the human.**

## The one decision that waits, and is not a polish

**CI cannot enforce the review.** ADR 0042 says what would change it: a trusted job that
runs `odachi review` with its own model and key, or a receipt signed outside the tree.
Both are features. Until one exists, the review is a discipline held at the close, where
the store is - and the CI gate naming it is not the same as checking it.

## Queued

Read at the start of a session, not now: each one becomes a feature, with its own brief, when
it starts.

- **the step declares its route.** inline or delegated, and which agent, recorded in the task
  doc. It is the one fact gentle-ai writes per task (`Route:`) that odachi does not. Cheap.
- **a change resumable from the diff needs no document.** The light lane: no brief and no task
  doc when the work can be resumed from the diff alone. Expensive, because it changes rule 1
  and the floor `check` enforces.
- **a step done without a commit is a step nobody can find.** `plan --done T2` records no tree
  when the worktree is not the commit; 0147's T2 and 0157's T2 both closed that way. The
  spelling that works is `--commit <sha>`, and the tool does not say so at the moment of the
  mistake. `undo` cannot repair it either (see the dead end above): the family is bookkeeping
  nobody can put back.
- **the proof names what it proves shipped in 0151** (`# covers R1`), so it leaves this queue.

## What this session does first

**Research: professional practice, and application security.** What a fullstack developer
must weigh before shipping -input handling, authz and authn, secrets, dependencies, logging
and personal data, migrations, backups, least privilege, failure modes- gathered from real
sources (OWASP at least) and turned into **a skill or an ADR**, never a document that rots.
The split it has to make explicit: what the gate can enforce (a fence over secrets, a proof
that runs the suite) and what is a checklist the skill teaches.

## How to find the next real work

`docs/review-debt.md` names what a verdict left; `doctor` and `bench` find friction nobody
has named; the hygiene trio (`go vet`, `go mod tidy -diff`, `gofmt -l .`) is cheap; and
`clean` says which class of file grew back.
