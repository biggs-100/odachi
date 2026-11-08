# Next session

Written at the close of 0241, `the-parent-uses-the-composed-hand-off`, after `v0.31.0`. The chain that
opened the day is measured end to end: a step loads the craft, the text moves the artifact, the verb
composes the hand-off, and the parent uses it. One question stays open, and it is the one this session
did not measure.

## Where the project is

- `main` at the merge of #445 (`98666e7`), `odachi 0.31.0` installed, no open feature, no open
  questions. `doctor` reads 15 checks with one warning: this project's block predates `0.31.0`, which
  is the open question below, left in place as the evidence.
- 239 features and 586 steps closed; `docs/review-debt.md` has no open rows.
- `v0.31.0` is the tag a person installs: the `delegate` verb, the `skills` field in the task doc, and
  the block sentence that names the verb. `Unreleased` is empty again.
- The installed binary, the home worker and a project initialized by the release all carry the
  mechanism.

## What this session measured

- **The task's shape is the lever.** A bare symptom loaded nothing in 12 runs; a task-doc step loaded
  `drive` and `tests` in 6/6, with and without the block.
- **The text moves the artifact.** The same step without the skill: 12/12 against 6/12; the test
  first, the red and the name move from 1/3 to 3/3. Edge coverage is equal because the prompt names
  the edges.
- **The worker obeys the paths it is given.** With the section, both paths were read before the first
  edit in 3/3; without it, nothing in 3/3.
- **The parent drops the section by hand.** Four delegations, none with it.
- **A sentence is advisory; a mechanism is not.** The worker's refusal fired 2/3; the composed verb
  was used 3/3 and carried the exact path every time.
- The eight probes are recorded in `docs/skill-evaluation.md`, with the rubric, the limits and the two
  dropped hypotheses (the trigger description, the block as the lever).

## Next

- **The block a project already has** (the only open question). This repository's own `AGENTS.md`
  predates `0.31.0` and does not name `delegate`; `odachi init` rewrites it, and `doctor` compares the
  voice, not the whole block. So an upgraded project keeps the old text and the mechanism is invisible
  there. The next feature either makes `doctor` name a stale block or makes the upgrade refresh it;
  the probe after it is the same three runs on an upgraded project.
- A release is owed when the next observable change lands; `Unreleased` is empty.

## The limits to remember

- The gate reads the fence from the PR's `base.sha`: a plan that lifts a rule must merge first, and
  `gh run rerun` keeps the old base; a rebase onto the new `main` is what gives the gate the new fence.
- A probe that runs an agent is evidence, not a proof: three runs per arm, one model, one task.
- Each probe run starts in its own isolated temp folder; a run that validates a fixture can be run 1
  when nothing in the fixture changes after it.
- The installed binary is the gate: a change to a verb is invisible on this machine until a release
  installs it.
- The fence protects `skills/**`; `internal/domain/rules` and `internal/domain/taskdoc` are not
  protected, which is why the verb and the field could land without lifting a rule.
- A prompt sentence is not a mechanism: the refusal lost once in three, and the composed verb closed
  the gap.
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

The delegation, on an installed `0.31.0`:

```sh
odachi plan "..." --req R1 --route delegated --agent odachi-worker --skill tests --skill drive
odachi delegate odd/tasks/<file>.md --step T1 --surface <path> --command "go test ./..."
```

The probe, per run, on a project the release initialized:

```sh
opencode run --format json --agent odachi --model opencode-go/deepseek-v4.1-flash "Trabajá el paso T1 del plan."
```
