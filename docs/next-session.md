# Next session: the two rows the trailer lane left

Written at the close of the session that shipped 0181 and v0.14.0. **v0.14.0 is published, signed and
installed**; `main` is clean at `4580acf`; `doctor` is 14/14. The next work is the two rows the
re-review of 0181 left, and two holes found while using the tool.

## What shipped

- **0181, v0.14.0**: a change can name itself in its commit. `Odachi-Change: <line>` in a commit of
  the branch is the vision when no document is open; `check` measures the change from the merge-base
  with the default branch, names the proof after the line, and refuses a trailer on the default
  branch, an empty one or two different ones; `review` judges the same line and says the diff is the
  record. The light lane stays.
- **The release body wrote itself.** v0.14.0 is the first release whose notes came from the workflow
  step of 0180, printed from the changelog: the body is in the release and nobody edited it by hand.

## The lesson this session paid for

**A lane that reads the diff must ask the same question the gate asks.** The first review found a
CRITICAL: `review` read the trailer before it asked whether a document was open, so a documented
feature could be judged against a line `check` ignored. The fix was one question, asked the same way
in both verbs, with two tests. The re-review found the mirror case: the review filters documents by
the candidate's base while the gate takes the newest document overall. That one is not paid; it is
in the ledger, and it is the same defect one level down.

## What was measured, so nobody re-measures it

- **The lane works end to end.** In `.git/odachi/scratch/lane-smoke/repo`: no document, no brief,
  `Odachi-Change: bring back the wait state` in the commit, `check --run "echo ok"` ready with
  `proof/change-bring-back-the-wait-state.txt`, and the same trailer on `main` refused by name. The
  installed v0.14.0 answers the same.
- **`check` without `--require-review` forgot the verdict.** The task doc was committed closed
  before the verdict, so the mode's `closedBefore` read the close as done and stopped asking; the
  flag caught it. The close commit belongs after the verdict, or the mode should anchor on the base.
- **`start --brief` writes no `brief.md`** (`writeBrief: false`): the draft must already be that file.
  The shape skill says otherwise. Still open.
- **A proof must be run with the close documents staged**, or the pre-commit hook refuses it as
  stale: stage, `check --run`, then commit.

## Queued

1. **the review filters documents by base, the gate by the newest** (review-debt): a rebased branch
   or a second open document lets the two verbs disagree. Fix: ask the newest document, no filter,
   with a test at a different base.
2. **the detached HEAD lets a trailer on the default branch through** (review-debt): `Branch()`
   fails, the CLI drops the error, and the refusal needs a name. Decide what "on the default branch"
   means without one, and keep CI's detached checkout working.
3. **the gate mode forgets a verdict the close commit preceded**: decide whether the close commit
   must come after the verdict, or `closedBefore` anchors on the document's base.
4. **`start --brief` and the shape skill disagree**: the tool reads the draft and writes no brief;
   either it writes `brief.md` or the skill says to place the draft first.

## What is already true, so it is not re-decided

- **The fence protects `proof/**` only**; `.github/**` left the fence in 0179.
- **The release body is the changelog**, printed by the workflow (0180, proven on v0.14.0).
- **A change with no document names itself in its commit** (0181); the light lane still writes one.
- **A stored proof is never rewritten.**
- **The installed binary is 0.14.0**, and the home carries the updated skills, `Odachi-Change`
  included.

## How to find the next real work

`docs/review-debt.md` names the two rows; this note names the other two holes. `odachi summary` says
0181 is closed and nothing newer exists, so the ledger is the map today.
