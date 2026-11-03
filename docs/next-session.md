# Next session: T2 of the trailer, and a gate that must not guess

Written at the close of a session that shipped five features and two releases. **v0.13.1 is
published, signed and installed**; the worktree is on `feat/0181-the-diff-speaks`, the open
feature, with its plan on local `main` and its first step on the branch.

## What shipped

- **0176, v0.13.0**: `design.json` names a screenshot per artboard, `design list` and the canvas
  name the ones without it, and the design skill makes it part of done.
- **0177, v0.13.1**: the design skill got back the four instructions 0176's compression lost.
- **0178**: the installer's refusal without cosign has its test.
- **0179**: the fence stopped protecting `.github/**`, so the next feature could touch the workflow.
- **0180**: the release body is printed from the changelog by `scripts/release-notes.sh`, the
  workflow edits the release with it, and a tag with no section is refused.

## The lesson this session paid for

**A skill that judges well still needs the read-back to leave an artifact.** The blind experiment
preferred the good artboard twice (8 vs 3) and named 9 and 10 of the 12 tells with measured
evidence, yet the earlier artboard had shipped with its panel flush to the edge because the
read-back never ran. A rule nobody can see is not a step; the fix was the screenshot in the
manifest, not another rule.

## What was measured, so nobody re-measures it

- **The screen test**: two artboards, same content, shuffled screenshots, three blind judges. 2/2
  preferred the good one, 8/10 vs 3/10, 9/12 and 10/12 tells named. Report:
  `.git/odachi/scratch/design-test/experiment/report.md`.
- **`start --brief` writes no `brief.md`**: the code passes `writeBrief: false`, so the draft must
  already be `brief.md` when the run starts, or the tree keeps the previous feature's brief next to
  the new task doc. The shape skill says otherwise, and that sentence is wrong.
- **The brief parser is line-based**: a wrapped bullet is cut at the first newline. One item per
  line, however long.
- **A bash test skips on this machine** (WSL bash) and runs in CI, like the installer's tests.
- **`check --run` wants `--timeout 1500`** for a cold suite; the cached one closes in minutes.

## What is open: 0181, the diff speaks for itself

The plan is `odd/tasks/0181-the-diff-speaks.md`; T1 is done.

- **T1 (done)**: `internal/domain/change` reads `Odachi-Change:` from commit messages and refuses
  an empty trailer or two different lines; `gitx.Messages` answers the range since a base.
- **T2 (next)**: the gate judges a document-less change. It needs the identity (a trailer with no
  open document means this change is the feature), the base (the merge-base with the remote's
  default branch, the same ref `check --ci` reads), and the proof name (a slug of the line). The
  delicate part: a project with a closed document and no trailer must keep today's answer.
- **T3**: `odachi review` builds the prompt with the trailer line where a light document's line
  goes, and says the diff is the record.
- **T4**: `skills/drive/SKILL.md` and the changelog name the lane.
- Then the proof, the review and the PR. The plan commit is on local `main` and travels in the PR.

## What is already true, so it is not re-decided

- **The fence protects `proof/**` only**; `.github/**` left the fence in 0179.
- **The release body is the changelog**, printed by the workflow (0180).
- **A release is a tag someone can install**, and a patch ships as soon as it is green.
- **A stored proof is never rewritten.**
- **`docs/review-debt.md` is empty.**
- **The installed binary is 0.13.1.**

## How to find the next real work

`odachi summary` names 0181 and its steps, and the task doc names the next one. The ledger is empty,
so the task doc is the map today.
