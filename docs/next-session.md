# Next session: the holes using the tool found

Written at the close of the session that shipped 0182, 0183, v0.14.1 and v0.14.2. **v0.14.2 is
published, signed and installed**; `main` is clean; `doctor` is 14/14 once this note lands. The review
debt of 0181 is empty, and the next work is one defect found by using the tool and three decisions.

## What shipped

- **0182, v0.14.1**: the review asks the document the gate asks. `odachi review` filtered documents by
  the base its candidate was judged against, and a document records a commit while the default base is
  `HEAD`: an `Odachi-Change:` trailer stood in for an open feature's vision, and a light feature was
  judged against "there is no brief.md yet". Both lookups ask `gateDocument`; a correction keeps its
  predecessor's base.
- **0183, v0.14.2**: the default branch answers without a branch name. The trailer's default-branch
  refusal compared the checked-out branch's name with the base's; a detached HEAD has none, the CLI
  dropped the error, and the trailer was judged. The gate asks the repository for the local branch the
  base names (`origin/main` names `main`) and whether HEAD is one of its commits. CI's detached pull
  request keeps the lane working; a repository git cannot read blocks with the reason.
- **The two docs commits from 0181's close rode 0182's PR** (PR #291). `main` is protected by four
  checks; nothing lands without a PR.

## The lessons this session paid for

**A question asked in two places drifts, even when both places quote the same rule.** 0182's row named
a rebase; the measurement found the default path -a ref compared to a commit never matches-, which is
why the light lane was broken for every `odachi review` without `--base`.

**A proof that runs where the machine has a global git identity does not measure CI.** 0183's first CI
run failed on all three platforms: the new fixtures commit inside a clone, `git clone` copies no local
config, and this machine's global `user.name` made the suite pass. The clone now records its own
identity, and the suite was run with `GIT_CONFIG_GLOBAL` and `GIT_CONFIG_SYSTEM` pointing at empty
files -the CI shape- before the proof.
- To repeat: create an empty file, set both variables to it, run `go test ./...`.

**The review can find what CI cannot, and the other way round.** 0183's first verdict found a causal
CRITICAL -any branch whose upstream was the base looked like the default one, and
`git checkout -b feat origin/main` gives a feature branch exactly that upstream-. CI found the test
fixture's dependence on the machine. Neither reader substitutes for the other.

## What was measured, so nobody re-measures it

- **`review --attach` accepts a model that is not an identifier, and the next correction dies.** The
  first 0183 verdict named `DeepSeek V4.1 Flash`; attach sealed it, and
  `odachi review --attach <corrected> --since <that tree>` refused with
  `the lens "reliability": model "DeepSeek V4.1 Flash" contains ' ': a model is an identifier`.
  The row below carries the reproduction. The way out for that correction was a full review of the
  closing tree; the predecessor's `model` cannot be edited (a verdict is not rewritten).
- **A correction of a tree whose verdict names a real identifier works.** 0183's test-only fix after CI
  was corrected from `b97ad7c8` with a 9 KB delta prompt and approved.
- **The light lane's vision on the default base was broken end to end**, measured in
  `.git/odachi/scratch/light-base-probe` before and after 0182's build.
- **`git checkout -b feat origin/main` records `origin/main` as the feature's upstream**
  (`branch.autoSetupMerge`), verified with `for-each-ref`.

## Queued

1. **an attached verdict's model is free text, and a correction cannot merge it**: validate `model` at
   attach with the same `NormalizeModel` the lens table uses (refuse before writing, naming the shape),
   or stop validating it when a correction merges the predecessor's rows. The first keeps `{model}`
   substitutable; the prompt asks for "the model that judged", which is what a reviewer writes.
   Reproduction: attach a verdict whose `model` is `DeepSeek V4.1 Flash`, then `review --attach <next>
   --since <that tree>`.
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
     is the agent's scratch. That scratch now also holds `light-base-probe`, `probe-upstream` and the
     0182 build; move anything to keep to `docs/` before `clean --apply`.

## What is already true, so it is not re-decided

- **The fence protects `proof/**` only**; `start` regenerates it from the brief, so a rule that must
  survive is written by hand before the plan commit.
- **The release body is the changelog**, printed by the workflow (proven again on v0.14.2).
- **A change with no document names itself in its commit** (0181).
- **The review asks the newest document, whatever base it judges** (0182).
- **The default branch is a fact of the commits** (0183): a detached checkout is refused or judged for
  the same reason a named one is, and CI's detached checkout still reads the lane.
- **`main` is protected**: four checks, everything through a PR.
- **The installed binary is 0.14.2**, and the home carries the updated agents and skills.

## How to find the next real work

`docs/review-debt.md` has nothing open; the four items above are the map today. `odachi summary` says
0183 is closed and nothing newer exists.
