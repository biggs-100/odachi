# Next session

Written with 0242 open, `an-agent-lives-only-in-the-home`, after `v0.31.0`. The chain the eight probes
measured is closed; the product decision that follows from it -one agent, in the home- is opened and
waits for its first step.

## Where the project is

- `main` at the merge of #447 (`97fc704`), the plan for 0242: the brief, a fence narrowed to the
  review and the classifier, and three steps planned. No code yet.
- `odachi 0.31.0` installed; 240 features, 586 of 589 steps done, no open questions in the plan.
- `v0.31.0` is the tag a person installs: the `delegate` verb and the `skills` field. `Unreleased` is
  empty.
- `doctor` reads 15 checks with two warnings: this project's block predates `0.31.0` (the open
  question below), and `docs/next-session.md` was stale (this file).

## What this session did

- **Eight probes**, recorded in `docs/skill-evaluation.md`: a bare task loads nothing (12 runs); a
  task-doc step loads `drive` and `tests` (6/6); the text moves the artifact (12/12 against 6/12);
  the worker obeys the paths it is given (3/3 and 0/3); the parent drops the section by hand (4
  delegations, none); the worker's refusal catches 2 in 3; the composed verb was used 3 in 3 with the
  exact path.
- **`v0.30.0`**: the worker returns `status: blocked` without the skills section.
- **`v0.31.0`**: `odachi plan --skill` records the craft in the step; `odachi delegate` composes the
  hand-off; the block names the verb.
- **0242 opened**: the agent lives only in the home. The three decisions taken with the human: init
  stops writing and `doctor` names an older project's agent (never deletes it); `skill new` without
  `--scope` fails; the bundled craft defaults to the home.
- **T1 started and reverted**: forcing global flipped 26 tests in 17 files -init, scope, store,
  adoption, doctor, role-. The rewrite is a block of its own, not the end of a long session.

## Next: the first block of 0242

- **T1 - init.** The CLI is small: `runInit` forces `Global` and `NoProjectAgents`, and the help
  says the agent lives in the home. The domain changes where absence lands: `persona.SkillsScope`
  resolves the default to `SkillsHome`, and `effectiveSkills` in `internal/app/init` follows. Then
  the test rewrite, by group: the tests that assert `.opencode/agents`/`.claude` in the project, the
  ones that assert the skills default is project, the store (`project_agents = none`,
  `skills = home`), and adoption and role.
  `TestMain` in `cmd/odachi/init_skills_test.go` already isolates the home in a temp directory, so the
  suite cannot write the real home.
- **T2 - `odachi skill new <slug> --scope project|home`.** Today it writes `skills/<slug>`, the
  repo's curriculum. A user skill goes to `.agents/skills/<slug>/SKILL.md` or
  `~/.agents/skills/<slug>/SKILL.md`; the template is the same, and the registry outside the repo is
  not touched.
- **T3 - `doctor` names a project agent an older release left**, keeping its 15 checks. The fence no
  longer protects it: the plan for 0242 narrowed the rule.

## The open question that survives

- **The block a project already has.** This repository's `AGENTS.md` predates `0.31.0` and does not
  name `delegate`. 0242's `init` rewrites the block, so a run of init fixes it, but nothing tells a
  project that its block is old; `doctor` names agent files, not a stale block. The next feature
  decides: `doctor` names a stale block, or the upgrade refreshes it. `doctor`'s warning 4 is this,
  left in place as the evidence.

## The limits to remember

- The gate reads the fence from the PR's `base.sha`: a plan that narrows a rule merges first; a rerun
  keeps the old base, and a rebase onto the new `main` gives the gate the new fence.
- A probe that runs an agent is evidence, not a proof: three runs per arm, one model, one task.
- Each probe run starts in its own isolated temp folder; a run that validates a fixture can be run 1
  when nothing changes after it.
- The installed binary is the gate: a verb change is invisible until a release installs it.
- The fence protects `skills/**` and the review and the classifier; doctor was deliberately released
  for 0242.
- The cmd package's `TestMain` isolates the home; run it for any change to init.
- One writer per worktree; one heavy gate per machine; the suite takes about 700s locally.

## The commands that work

```sh
odachi review --base <the base the task doc names> \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

```sh
$env:PATH = "C:\Windows\System32;C:\Windows;C:\Windows\System32\Wbem;C:\Program Files\Go\bin;C:\Users\USER\go\bin;C:\Users\USER\AppData\Local\odachi\bin;C:\Users\USER\AppData\Roaming\npm"; odachi check --run "go test ./..." --timeout 1800
```

The packages T1 touches first:

```sh
go test ./internal/app/init/ ./internal/app/adoption/ ./internal/domain/persona/ ./cmd/odachi/
```

The delegation, on `0.31.0`:

```sh
odachi plan "..." --req R1 --route delegated --agent odachi-worker --skill tests --skill drive
odachi delegate odd/tasks/<file>.md --step T1 --surface <path> --command "go test ./..."
```
