# Next session: the row the lane left, and the holes around it

Written at the close of the session that shipped 0182 and v0.14.1. **v0.14.1 is published, signed and
installed**; `main` is clean; `doctor` is 13/14 until this note names 0182. The next work is the row
0181 left in the ledger, and the three holes found while using the tool.

## What shipped

- **0182, v0.14.1**: the review asks the document the gate asks. `odachi review` filtered documents by
  the base its candidate was judged against, and a document records a commit while the default base is
  `HEAD`: the filter found nothing, so an `Odachi-Change:` trailer stood in for an open feature's
  vision, and a light feature was judged against "there is no brief.md yet" unless the human named its
  hash. Both lookups now read `gateDocument`: the newest document, the one `check` takes. A correction
  keeps its predecessor's base.
- **The docs commits rode the feature's PR.** `main` is protected by four required checks, so a plain
  push is refused; the two commits the last close left travelled with 0182 (PR #291).

## The lesson this session paid for

**A question asked in two places drifts, even when both places quote the same rule.** 0181's fix made
the review ask "is a document open" using the base its comment already promised; the gate asked the
same question of the newest document. The row named a rebase. The measurement found the default path:
the ref and the commit never match, so the review read the trailer on every branch and the light lane
was never its own vision. The repair was one function both lookups call, not two corrections.

## What was measured, so nobody re-measures it

- **The light lane's vision was broken by default.** In `.git/odachi/scratch/light-base-probe`:
  `odachi start --light "the pay button waits for the answer"`, a change, `odachi review` → the prompt
  said "(there is no brief.md yet: run odachi start)". With `--base <the document's hash>` it carried
  the line. After 0182's build, the same command carries the line. The probe repo is still there.
- **A correction keeps its own base.** `--since` measures from the predecessor's tree; the newest
  document may be a later feature, so the vision still comes from the predecessor's base. Pinned by
  `TestACorrectionAsksTheDocumentItsPredecessorJudged`.
- **`review --show` without `--base` says "documents only" about a closed feature.** It judges HEAD →
  worktree; the gate judges the feature's range. The gate's own hint, `odachi review --base <doc
  base>`, is the command that asks the same question.
- **A docs edit after the proof makes the proof stale by tree.** `check` refused it, then the second
  `check --run` was mostly cached. Run the proof with the close documents staged.

## Queued

1. **the detached HEAD lets a trailer on the default branch through** (review-debt): `Branch()` fails,
   the CLI drops the error, and the refusal needs a name. Decide what "on the default branch" means
   without a branch name, and keep CI's detached checkout working.
2. **the gate mode forgets a verdict the close commit preceded**: decide whether the close commit must
   come after the verdict, or `closedBefore` anchors on the document's base.
3. **`start --brief` and the shape skill disagree**: the tool reads the draft and writes no brief;
   either it writes `brief.md` or the skill says to place the draft first.
4. **the cleanup policy has two decisions and a nudge** (queued 2026-11-03, after comparing gentle-ai
   and gentle-shell):
   - prune the review prompts and scaffolding at the moment the verdict is stored. `keep_prompts = 0`
     already says they are disposable; today only `clean --apply` does it. gentle-shell does exactly
     this at write time (`historyMaxTasks`, default 200) and gentle-ai prunes its backups after every
     install (`Prune(root, 5)`, pinned never).
   - a `doctor` check or a close line that names the store's growth, like gentle-ai's maintenance loop,
     but without deleting: the human keeps the `--apply`.
   - what stays manual forever: `proof/` and `odd/tasks`. Bounding them is a decision about which
     evidence is forgotten, and that is an ADR, not a `clean` rule.
   - measured 2026-11-03 on this clone: 21.7 MB in 1187 files, 20.8 MB recoverable, of which 18.3 MB
     is the agent's scratch. That scratch now also holds `light-base-probe` and the 0182 build; move
     anything to keep to `docs/` before `clean --apply`.

## What is already true, so it is not re-decided

- **The fence protects `proof/**` only**; `start` regenerates it from the brief, so a non-goal that
  must survive has to be in the brief's list.
- **The release body is the changelog**, printed by the workflow (proven again on v0.14.1).
- **A change with no document names itself in its commit** (0181); the light lane still writes one, and
  its line is now the vision on the default base (0182).
- **A stored proof is never rewritten.**
- **The review asks the newest document, whatever base it judges** (0182).
- **`main` is protected**: four checks, everything through a PR.
- **The installed binary is 0.14.1**, and the home carries the updated agents and skills.

## How to find the next real work

`docs/review-debt.md` names the detached-HEAD row; this note names the other three. `odachi summary`
says 0182 is closed and nothing newer exists, so the ledger is the map today.
