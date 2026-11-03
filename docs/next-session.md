# Next session: three decisions, and an empty ledger

Written at the close of the session that shipped 0182, 0183 and 0184 and published v0.14.1, v0.14.2 and
v0.14.3. **v0.14.3 is published, signed and installed**; `main` is clean; `doctor` is 14/14 once this
note lands. Nothing is open in `docs/review-debt.md`; what is left is three decisions.

## What shipped

- **0182, v0.14.1**: the review asks the document the gate asks. A document records a commit and the
  review was asked about a ref (`--base HEAD`), so the base filter found nothing: a trailer stood in for
  an open feature's vision, and a light feature was judged against "there is no brief.md yet". Both
  lookups ask `gateDocument`; a correction keeps its predecessor's base.
- **0183, v0.14.2**: the default branch answers without a branch name. The gate asks the repository for
  the local branch the base names (`origin/main` names `main`) and whether HEAD is one of its commits.
  CI's detached pull request keeps the lane working; a repository git cannot read blocks with the reason.
- **0184, v0.14.3**: `review --attach` refuses a model that is not an identifier (`provider/name#variant`)
  before anything is written, so a later `--since` correction can merge the verdict it judged.

## The lessons this session paid for

**A question asked in two places drifts, even when both quote the same rule.** 0182's row named a
rebase; the default path was the bug every time.

**A proof that runs where the machine has a global git identity does not measure CI.** 0183's first CI
run failed on all three platforms: the fixtures commit inside a clone, `git clone` copies no local
config, and this machine's global `user.name` hid it. Run the suite with `GIT_CONFIG_GLOBAL` and
`GIT_CONFIG_SYSTEM` pointing at empty files to see the CI shape.

**The reviewer must be a different model than the proof.** 0184's first verdict named the model that
wrote the proof (`deepseek-v4.1-flash`); the gate refused it as "a second reading, not a second
opinion". The second opinion is the lens model `odd/review.toml` assigns -`mimo-v2.6-flash` for
reliability-, and it is the one that was stored.
- A verdict cannot be replaced: with the wrong one already stored, the documented way out is removing
  the file in the clone's `.git/odachi/reviews/` by hand (0081, and the store's `Accepted` rows).

## What was measured, so nobody re-measures it

- **`review --attach` accepted `"model": "DeepSeek V4.1 Flash"`**, sealed it, and the next
  `review --attach <next> --since <that tree>` refused with `a model is an identifier`. Fixed by 0184;
  the reproduction is now the red test.
- **The light lane's vision on the default base was broken end to end** before 0182, measured in
  `.git/odachi/scratch/light-base-probe` with a built binary.
- **`git checkout -b feat origin/main` records `origin/main` as the feature's upstream**, which broke
  0183's first rule; the review found it as a causal CRITICAL.
- **A docs edit after the verdict moves the gate**: it asks for a verdict of the new tree, and a
  docs-only commit after a close is not free once the feature carries code. Run the proof with the
  closing documents staged.

## Queued

1. **the gate mode forgets a verdict the close commit preceded**: decide whether the close commit must
   come after the verdict, or `closedBefore` anchors on the document's base.
2. **`start --brief` and the shape skill disagree**: the tool reads the draft and writes no brief;
   either it writes `brief.md` or the skill says to place the draft first.
3. **the cleanup policy has two decisions and a nudge** (queued 2026-11-03, after comparing gentle-ai
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
- **The release body is the changelog**, printed by the workflow (proven again on v0.14.3).
- **A change with no document names itself in its commit** (0181).
- **The review asks the newest document, whatever base it judges** (0182).
- **The default branch is a fact of the commits** (0183).
- **An attached verdict's model is an identifier, checked before the write** (0184).
- **The reviewer is a different model than the proof's**, by project configuration.
- **`main` is protected**: four checks, everything through a PR.
- **The installed binary is 0.14.3**, and the home carries the updated agents and skills.

## How to find the next real work

`docs/review-debt.md` has nothing open; the three items above are the map today. `odachi summary` says
0184 is closed and nothing newer exists.
