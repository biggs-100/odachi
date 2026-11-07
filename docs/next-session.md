# Next session

Written at the close of 0226, `a-non-goal-says-when-it-comes-back` (unreleased, after v0.28.1). The
non-goal now carries the condition that reopens it; the two open items from earlier sessions are next.

## Where the project is

- `main` at the merge of #415 (`cdd1765`), `odachi 0.28.1` installed, `doctor` 15/15, no open feature,
  no open questions.
- 224 features and 565 steps closed; `docs/review-debt.md` has no open rows.
- 0225 and 0226 are both in the changelog's Unreleased section: the release is not cut yet.

## What this session measured

- **The plan lands before the code when the base fence protects the target.** 0225's fence protected
  `skills/**`, so 0226 needed #414 merged before #415 could pass CI: the gate reads the fence from the
  PR's base, not from the change.
- **The ruleset refuses a direct push to `main`.** The close commit of 0225 was declined by a
  repository rule, not by the hook: every landing, documents included, goes through a PR.
- **The pin was red on three strings and green after the text.** `TestShapeDatesEveryNonGoal` failed
  with `shape no dice "A non-goal says when it comes back"`: the test names the rule, the three shapes
  and the Output line, so a later edit cannot quietly drop it.
- **The implementation was 40 changed lines**; the plan was 107 against its base.
- **The interview keeps one answer per non-goal.** A second question per entry would need a new flag,
  and the fence presets already need the first word; the convention lives where an agent drafts the
  brief.

## Next

The two items still open from earlier sessions, then the documental pair:

- **0218: the installer names the process that holds the binary.** When `odachi.exe` is in use, say
  which process and what to do, instead of a `Copy-Item` IOException. Two installs were blocked by
  another project's gate.
- **0216: pin `all:*` with a repo test.** A regression to `*` would pass the suite for dotfile assets;
  today only a build probe covers it.
- **The verdict/proof binding in `docs/releases.md`** and **the harness cost in `doctor`** can ride the
  close of whichever lands next.
- The behavioural evaluation of the core skills stays deferred.
- One writer per worktree; one heavy gate per machine.

## The limits to remember

- The review counts even when it is the same model: it is a second reading, and the gate reads it.
- The gate's default timeout is 900 seconds; the suite alone takes about 250, and a cold or shared
  machine needs the bound.
- A feature that touches a path the base fence protects needs the plan PR merged first; when the
  fence allows the path, the plan rides the feature PR.
- A stacked PR retargeted to `main` keeps its green: the head commit did not move, and the gate had
  already run against the same bytes.

## The commands that work

```sh
odachi review --base <the base the task doc names> \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

`--base` is the base the task doc names, not `main`.

```sh
$env:PATH = "C:\Windows\System32;C:\Windows;C:\Windows\System32\Wbem;C:\Program Files\Go\bin;C:\Users\USER\go\bin;C:\Users\USER\AppData\Local\odachi\bin;C:\Users\USER\AppData\Roaming\npm"; odachi check --run "go test ./..."
```

No `--timeout` is needed; the default is 900. A build goes to
`%TEMP%\opencode\odachi\<feature>\`, never to the clone.
