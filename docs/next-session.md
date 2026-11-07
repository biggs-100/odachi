# Next session

Written at the close of 0230, `a-release-says-which-bytes-were-judged`, with `v0.29.0` published and
installed. The backlog of earlier sessions is empty; the harness cost is the next feature.

## Where the project is

- `main` at the release merge of #426 (`1a18088`), `odachi 0.29.0` installed from the published
  release, `doctor` 15/15, no open feature, no open questions.
- 228 features and 573 steps closed; `docs/review-debt.md` has no open rows.
- `v0.29.0` carries 0225 to 0230 and is the tag a person installs; `Unreleased` is empty until the
  next feature writes under it.

## What this session measured

- **The gate judges with the installed binary, not with the tree's code.** At 0230's close the local
  `check` still printed the false "no verdict" line after two releases' worth of source fixes: the
  binary on the PATH was 0.28.1. A binary built from the tree printed `verdict: f120eb1c` and no
  note. A change to `check` is not observable on this machine until a release, or a build.
- **A review finding can be half false and half true.** 0230's WARNING read the line as reaching a
  blocked report; the call sits after the ready gate's early return, so it cannot. The missing test
  was real, and `TestABlockedCloseSaysNothingAboutTheVerdict` now pins the absence.
- **The release is the installed gate's update path.** `v0.29.0` was cut on the merge of #426, the
  workflow ran the suite, built and signed, and the notes came from the changelog section verbatim.
  The published assets and the body were read back before installing.

## Next

- **The harness cost in `doctor`.** rsc has a `context-budget`; odachi does not say what its always-on
  costs. `doctor` can report the bytes of the block plus `SKILLS.md` and the installed skill count,
  the way it already reports the store. `internal/app/doctor/**` is protected by the current fence, so
  the plan PR lands before the line.
- The behavioural evaluation of the core skills stays deferred.
- One writer per worktree; one heavy gate per machine.

## The limits to remember

- The review counts even when it is the same model: it is a second reading, and the gate reads it.
- The gate's default timeout is 900 seconds; the suite alone takes about 250, and a cold or shared
  machine needs the bound; with the installer tests it ran about six minutes on this machine.
- A feature that touches a path the base fence protects needs the plan PR merged first; when the
  fence allows the path, the plan rides the feature PR.
- The installed binary is the gate: a change to `check` is invisible on this machine until a release
  or a local build, and CI builds the gate from the PR's base ref, not from the change.
- A store's "newest" is the file time, not the time the record declares.
- A test that pins an embed needs a file a regression makes disappear, and a plain file beside it so
  the package still builds and the red lands in the test.
- The ruleset refuses a direct push to `main`: every landing, documents included, goes through a PR.

## The commands that work

```sh
odachi review --base <the base the task doc names> \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

`--base` is the base the task doc names, not `main`. A correction after a fix is
`odachi review --since <the approved tree>` with the same `--exec` and no `--base`.

```sh
$env:PATH = "C:\Windows\System32;C:\Windows;C:\Windows\System32\Wbem;C:\Program Files\Go\bin;C:\Users\USER\go\bin;C:\Users\USER\AppData\Local\odachi\bin;C:\Users\USER\AppData\Roaming\npm"; odachi check --run "go test ./..."
```

No `--timeout` is needed; the default is 900. A build goes to
`%TEMP%\opencode\odachi\<feature>\`, never to the clone.

The release: rename `## Unreleased` to `## vX.Y.Z - <date>` in a release PR, merge it, then tag the
merge commit and push the tag (the tag publishes). On this machine:

```sh
git tag -a vX.Y.Z -m "vX.Y.Z" <merge-sha>; git push origin vX.Y.Z
$env:ODACHI_VERSION='X.Y.Z'; powershell -NoProfile -ExecutionPolicy Bypass -File scripts/install.ps1
```
