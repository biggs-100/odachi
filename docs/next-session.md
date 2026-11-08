# Next session

Written at the close of 0240, `a-delegated-step-is-handed-with-delegate`, after `v0.30.0`. The seven
probes closed the question that opened the day; the mechanism that answers them is merged and not yet
released.

## Where the project is

- `main` at the merge of #442 (`7a72068`), `odachi 0.30.0` installed, `doctor` 15/15, no open feature,
  no open questions.
- 238 features and 585 steps closed; `docs/review-debt.md` has no open rows.
- `v0.30.0` is the tag a person installs: it carries the worker's refusal. `Unreleased` carries one
  entry, 0240 -the `delegate` verb and the `skills:` field in the task doc-, and it is a contract
  change, so the next version is minor: `v0.31.0`.
- The installed binary does not know `odachi delegate` yet: the verb exists in `main`, not on this
  machine.

## What this session measured

- **The first two probes were a null for the right reason.** A bare symptom in a bare folder loaded
  nothing in 12 runs, with and without the skill installed.
- **The chain fires on a step.** Delivered as a task-doc step, `drive` and `tests` loaded in 6/6, with
  the block and without it: the block is not the lever.
- **The text moves the artifact.** The same step with the skill removed: 12/12 rubric points against
  6/12; the test first, the red and the name move from 1/3 to 3/3. Edge coverage is equal because the
  requirement names the edges.
- **The worker obeys the paths it is given.** With `## Skills to load before work`, both paths were
  read before the first edit in 3/3; without it, no `SKILL.md` in 3/3, though every skill was
  installed. It reads the files, not the client's list.
- **The parent drops the section.** Four delegations in three runs, none with the section or a path,
  though the parent had loaded `drive` and `tests` itself.
- **The refusal is advisory.** On `v0.30.0` it fired in 2/3; the retry carried the section 2/2; one
  worker reasoned around it, loaded the skills itself and wrote down why.
- **The mechanism exists (0240).** `odachi plan ... --skill` records the craft in the step line;
  `odachi delegate` prints the hand-off with the section and the exact paths, or an explicit none;
  the block names the verb.

## Next

- **Release `v0.31.0`** (approved, not done): rename `Unreleased` to `## v0.31.0 - <date>` in a release
  PR, merge, tag the merge commit, push the tag, then install with `ODACHI_VERSION=0.31.0`.
- **The eighth probe** (approved, not done): three runs of the door on a delegated step with the
  installed `delegate`, and read whether the parent composes the hand-off with it and passes the paths.
  If it does, the chain is closed; if it does not, the sentence in the block is the next lever.
- `docs/next-session.md` is current again.

## The limits to remember

- The gate reads the fence from the PR's `base.sha`: a plan that lifts a rule must merge first, and
  `gh run rerun` keeps the old base; a rebase onto the new `main` is what gives the gate the new fence.
- A probe that runs an agent is evidence, not a proof: three runs per arm, one model, one task.
- Each probe run starts in its own isolated temp folder; a run that validates a fixture can be run 1
  when nothing in the fixture changes after it.
- The installed binary is the gate: a change to a verb is invisible on this machine until a release
  installs it.
- The fence protects `skills/**`; `internal/domain/rules` is not protected, which is why the block can
  name a verb without lifting a rule.
- A prompt sentence is not a mechanism: the refusal lost once in three. Prefer the path the parent
  cannot skip.
- One writer per worktree; one heavy gate per machine; the suite takes about 700s locally.

## The commands that work

```sh
odachi review --base <the base the task doc names> \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

```sh
$env:PATH = "C:\Windows\System32;C:\Windows;C:\Windows\System32\Wbem;C:\Program Files\Go\bin;C:\Users\USER\go\bin;C:\Users\USER\AppData\Local\odachi\bin;C:\Users\USER\AppData\Roaming\npm"; odachi check --run "go test ./..." --timeout 1800
```

The release, after the release PR merges:

```sh
git tag -a vX.Y.Z -m "vX.Y.Z" <merge-sha>; git push origin vX.Y.Z
$env:ODACHI_VERSION='X.Y.Z'; powershell -NoProfile -ExecutionPolicy Bypass -File scripts/install.ps1
```

The delegation, once installed:

```sh
odachi plan "..." --req R1 --route delegated --agent odachi-worker --skill tests --skill drive
odachi delegate odd/tasks/<file>.md --step T1 --surface <path> --command "go test ./..."
```

The eighth probe, per run, in a fresh copy of the 0237 fixture:

```sh
opencode run --format json --agent odachi --model opencode-go/deepseek-v4.1-flash "Trabajá el paso T1 del plan."
```

Read the `subagent` call's prompt: it carries the section when the parent used the verb.
