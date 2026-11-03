# Next session: one decision, and an empty ledger

Written at the close of the session that shipped 0182 through 0186 and published v0.14.1 through v0.14.5.
**v0.14.5 is published, signed and installed**; `main` is clean; `doctor` is 14/14 once this note lands.
Nothing is open in `docs/review-debt.md`; what is left is the cleanup policy.

## What shipped

- **0182, v0.14.1**: the review asks the document the gate asks -the newest one, whatever base it
  judges-. A ref compared to a recorded commit never matched, so a trailer stood in for an open
  feature's vision and a light feature was judged against "there is no brief.md yet".
- **0183, v0.14.2**: the default branch answers without a branch name. The gate asks the repository for
  the local branch the base names and whether HEAD is one of its commits; CI's detached pull request
  keeps the lane working.
- **0184, v0.14.3**: `review --attach` refuses a model that is not an identifier
  (`provider/name#variant`) before anything is written.
- **0185, v0.14.4**: the close owes its verdict from the feature's diff, not from when its document was
  committed. A range of documents owes nothing; a post-verdict tail that only moved documents is
  paperwork through `lastVerdictTree`.
- **0186, v0.14.5**: `start --brief <draft>` records the draft as `brief.md` in the canonical shape; a
  draft that is already `brief.md` is never rewritten, and an existing one beside a draft elsewhere
  needs `--force`, which backs up the previous bytes.

## The lessons this session paid for

**A question asked in two places drifts, even when both quote the same rule.** 0182's row named a
rebase; the default path was the bug every time. 0183's first rule also named the wrong fact (a
branch's upstream is not the default branch: `git checkout -b feat origin/main` inherits it).

**A proof that runs where the machine has a global git identity does not measure CI.** 0183's first CI
run failed on all three platforms; run the suite with `GIT_CONFIG_GLOBAL` and `GIT_CONFIG_SYSTEM`
pointing at empty files to see the CI shape.

**A refusal is a command a person can type.** 0186's first messages said `odachi start --force again`
(which takes the interview path) and `--brief again` (which the flag package rejects). The reviewer
found it; every blocked path now prints the exact command with its operand.

**A refusal and a probe are two different failures.** A stat or a comparison that fails is not "no
brief.md": the run blocks with the reason (0125), and 0186 added the port method (`SameFile`, answered
by `os.SameFile`) because two spellings can name one file on Windows and macOS.

**Pay a warning before the verdict, not after.** 0186 took four review rounds and each of the first
three found something real -the untested `--force` path, the string comparison, the wrong remediation
commands and the untested `SameFile` error-. Paying them before attaching kept the closing tree the
only tree judged, and the fourth round was clean.

## What was measured, so nobody re-measures it

- **The light lane's vision** on the default base was broken before 0182, measured in
  `.git/odachi/scratch/light-base-probe`.
- **`review --attach` with a display-name model** was stored and poisoned the next `--since`; fixed in
  0184, with the reproduction as a red test.
- **`git checkout -b feat origin/main` records that upstream** for the feature branch
  (`branch.autoSetupMerge`), verified with `for-each-ref`.
- **macOS runners failed four times with no steps and a 15m cancel** (capacity, not code);
  `gh run rerun --failed` passed.
- **A docs edit after the verdict moves the gate**: it asks for a verdict of the new tree. Run the
  proof with the closing documents staged.

## Queued

1. **the cleanup policy has two decisions and a nudge** (queued 2026-11-03, after comparing gentle-ai
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
- **The release body is the changelog**, printed by the workflow (proven again on v0.14.5).
- **A change with no document names itself in its commit** (0181).
- **The review asks the newest document** (0182), and **the default branch is a fact of the commits**
  (0183).
- **An attached verdict's model is an identifier** (0184), and **a close owes its verdict from the
  diff** (0185).
- **The run records the brief it read** (0186).
- **The reviewer is a different model than the proof's**, by project configuration.
- **`main` is protected**: four checks, everything through a PR.
- **The installed binary is 0.14.5**, and the home carries the updated agents and skills.

## How to find the next real work

`docs/review-debt.md` has nothing open; the cleanup policy is the map today. `odachi summary` says 0186
is closed and nothing newer exists.
