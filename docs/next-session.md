# Next session

Written at the close of 0232, `the-first-behavioural-measurement`, after `v0.29.0`. The skills' effect
has its first measurement; the second probe —a task where a test is owed— is the next question.

## Where the project is

- `main` at the merge of #431 (`47237a6`), `odachi 0.29.0` installed, `doctor` 15/15, no open feature,
  no open questions.
- 230 features and 576 steps closed; `docs/review-debt.md` has no open rows.
- `v0.29.0` is the tag a person installs. `Unreleased` carries two entries, 0231 (the harness cost) and
  0232 (the measurement); the project's own cap is two or three before publishing, so the next
  feature or a release are both near.

## What this session measured

- **The instrument works, and it was validated before it measured.** `opencode run --format json`
  names every tool call in order. A canary skill that answers `BANANA` proved the transcript shows a
  load, in a project copy and in `~/.agents/skills`; the canary was removed and the home verified.
- **OpenCode reads the home copy of the skills**, not only the project's. The "without" arm therefore
  moves `~/.agents/skills/tests` out and back, measured at 59 → 58 → 59 installed with `doctor`. A
  temporary home does not work: the model stops resolving (`provider.no-route`).
- **The first result is a null, and it is informative.** Six runs (three per arm) on a one-line fix:
  none loaded `tests`, none wrote a test; both arms fixed the cause with an edit and verified with
  `go run .`. The skill changed nothing because it was never consulted, and the task may be below the
  threshold the skill itself names ("verification is proportional").
- **The suite took 704s today** and the default 900 timed out twice under load; the close proof ran
  with `--timeout 1800`. CI stays in the minutes, so the bound is a local-machine concern.

## Next

- **The second probe: a task where a test is clearly owed.** A new function with edges (empty, one,
  many, an error), or a defect whose shape hides in an edge. Run the same two arms. If the skill is
  still not loaded, the finding is about the trigger and the description, and that is the next thing
  to edit; the document already names this as the reopen condition.
- **A release** when the third entry lands, or now if the two are what a person should install.
- One writer per worktree; one heavy gate per machine.

## The limits to remember

- The review counts even when it is the same model: it is a second reading, and the gate reads it.
- The suite's local bound: about 700s on a loaded machine; `--timeout 1800` for a close proof, and the
  default 900 is the CI bound, not a local one.
- A feature that touches a path the base fence protects needs the plan PR merged first; when the
  fence allows the path, the plan rides the feature PR.
- The installed binary is the gate: a change to `check` or `doctor` is invisible on this machine until
  a release or a local build, and CI builds the gate from the PR's base ref, not from the change.
- A store's "newest" is the file time, not the time the record declares.
- A probe that runs an agent is not a test: it costs tokens, it varies between runs, and its evidence
  is the transcript; three runs per arm is the minimum, not a proof.
- The ruleset refuses a direct push to `main`: every landing, documents included, goes through a PR.

## The commands that work

```sh
odachi review --base <the base the task doc names> \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

`--base` is the base the task doc names, not `main`. A correction after a fix is
`odachi review --since <the approved tree>` with the same `--exec` and no `--base`.

```sh
$env:PATH = "C:\Windows\System32;C:\Windows;C:\Windows\System32\Wbem;C:\Program Files\Go\bin;C:\Users\USER\go\bin;C:\Users\USER\AppData\Local\odachi\bin;C:\Users\USER\AppData\Roaming\npm"; odachi check --run "go test ./..." --timeout 1800
```

A build goes to `%TEMP%\opencode\odachi\<feature>\`, never to the clone.

The probe, per arm, in a fresh copy of the fixture:

```sh
opencode run --format json --agent build --model opencode-go/deepseek-v4.1-flash "<the request>"
```

The without arm moves the skill and puts it back, with the count at both ends:

```sh
Move-Item ~/.agents/skills/tests $TEMP/opencode/odachi/<feature>/tests-off; odachi doctor   # 58
Move-Item $TEMP/opencode/odachi/<feature>/tests-off ~/.agents/skills/tests; odachi doctor   # 59
```

The release: rename `## Unreleased` to `## vX.Y.Z - <date>` in a release PR, merge it, then tag the
merge commit and push the tag (the tag publishes). On this machine:

```sh
git tag -a vX.Y.Z -m "vX.Y.Z" <merge-sha>; git push origin vX.Y.Z
$env:ODACHI_VERSION='X.Y.Z'; powershell -NoProfile -ExecutionPolicy Bypass -File scripts/install.ps1
```
