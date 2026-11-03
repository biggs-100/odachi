# Next session: two decisions, and an empty ledger

Written at the close of the session that shipped 0182, 0183, 0184 and 0185 and published v0.14.1 through
v0.14.4. **v0.14.4 is published, signed and installed**; `main` is clean; `doctor` is 14/14 once this
note lands. Nothing is open in `docs/review-debt.md`; what is left is two decisions.

## What shipped

- **0182, v0.14.1**: the review asks the document the gate asks. A document records a commit and the
  review was asked about a ref (`--base HEAD`), so the base filter found nothing: a trailer stood in for
  an open feature's vision, and a light feature was judged against "there is no brief.md yet".
- **0183, v0.14.2**: the default branch answers without a branch name. The gate asks the repository for
  the local branch the base names (`origin/main` names `main`) and whether HEAD is one of its commits.
  CI's detached pull request keeps the lane working; a repository git cannot read blocks with the reason.
- **0184, v0.14.3**: `review --attach` refuses a model that is not an identifier
  (`provider/name#variant`) before anything is written, so a later `--since` correction can merge it.
- **0185, v0.14.4**: the review question anchors on the feature's diff. A close committed before its
  verdict is named instead of read as answered; a range of documents owes nothing; a tree after a
  verdict that only moved documents stays paperwork through `lastVerdictTree`.

## The lessons this session paid for

**A question asked in two places drifts, even when both quote the same rule.** 0182's row named a
rebase; the default path was the bug every time.

**A proof that runs where the machine has a global git identity does not measure CI.** 0183's first CI
run failed on all three platforms: the fixtures commit inside a clone, `git clone` copies no local
config, and this machine's global `user.name` hid it. Run the suite with `GIT_CONFIG_GLOBAL` and
`GIT_CONFIG_SYSTEM` pointing at empty files to see the CI shape.

**The reviewer must be a different model than the proof.** The gate refuses a verdict named by the
model that wrote the proof; the second opinion is the lens model `odd/review.toml` assigns
(`mimo-v2.6-flash` for reliability). A verdict cannot be replaced, and with the wrong one stored the
documented way out is removing its file from `.git/odachi/reviews/` by hand (0081).

**An expectation in a test can encode the defect.** 0185 changed the review question, and the e2e
`progress` test failed because it expected "not due" on a release tail *without a verdict* -the old
rule, written as a fixture-. When the rule changes, the test that pinned the old answer is part of the
change, not an obstacle to it.

## What was measured, so nobody re-measures it

- **The close-committed-first case** now names the verdict: `TestTheModeNamesAVerdictWhoseCloseCameFirst`
  was observed red before the fix.
- **`closedBefore` still answers the proof lane** (the release-note paperwork), and only there; the
  review lane asks `featurePassive` and `lastVerdictTree`.
- **`review --attach` with a display-name model** was stored and poisoned the next `--since`; fixed in
  0184, and the reproduction is a red test.
- **A docs edit after the verdict moves the gate**: it asks for a verdict of the new tree. Run the
  proof with the closing documents staged.
- **macOS runners failed twice with no steps and a 15m cancel** (capacity, not code);
  `gh run rerun --failed` passed.

## Queued

1. **`start --brief` and the shape skill disagree**: the tool reads the draft and writes no brief;
   either it writes `brief.md` from the draft (with the `--force` backup) or the skill says to place the
   draft at `brief.md` first. The first leaves the artifact the run promised.
2. **the cleanup policy has two decisions and a nudge** (queued 2026-11-03, after comparing gentle-ai
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
     is the agent's scratch. That scratch now holds `light-base-probe`, `probe-upstream` and the 0182
     build; move anything to keep to `docs/` before `clean --apply`.

## What is already true, so it is not re-decided

- **The fence protects `proof/**` only**; `start` regenerates it from the brief, so a rule that must
  survive is written by hand before the plan commit.
- **The release body is the changelog**, printed by the workflow (proven again on v0.14.4).
- **A change with no document names itself in its commit** (0181).
- **The review asks the newest document, whatever base it judges** (0182).
- **The default branch is a fact of the commits** (0183).
- **An attached verdict's model is an identifier, checked before the write** (0184).
- **A close owes its verdict from the feature's diff, not from the commit order** (0185).
- **The reviewer is a different model than the proof's**, by project configuration.
- **`main` is protected**: four checks, everything through a PR.
- **The installed binary is 0.14.4**, and the home carries the updated agents and skills.

## How to find the next real work

`docs/review-debt.md` has nothing open; the two items above are the map today. `odachi summary` says
0185 is closed and nothing newer exists.
