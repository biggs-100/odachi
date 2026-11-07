# Next session

Written at the close of 0231, `the-harness-says-what-it-costs`, after `v0.29.0`. The backlog is
empty; the behavioural evaluation of the skills is the one thing still deferred.

## Where the project is

- `main` at the merge of #429 (`c04db37`), `odachi 0.29.0` installed, `doctor` 15/15, no open
  feature, no open questions.
- 229 features and 575 steps closed; `docs/review-debt.md` has no open rows.
- `v0.29.0` is the tag a person installs. `Unreleased` carries one entry, 0231 (the harness cost), so
  the installed binary does not print the new `doctor` line until the next release. The project's own
  rule says to cap at two or three entries before publishing, so it can wait for the next feature.

## What this session measured

- **The always-on cost is a number now.** `doctor`'s skills line names the bytes of the odachi block in
  `AGENTS.md` and the bytes of `SKILLS.md`, beside the installed count it already printed. It is the
  first measurement of what an agent reads on every turn, and it needed no sixteenth check.
- **A formatter with two callers belongs to neither verb.** `clean` and `doctor` both print sizes, so
  the format moved to `internal/domain/units`: two verbs cannot print the same number two ways.
- **The review found the branches the brief promised and no test covered**: the unreadable block and
  the file with no block. Both were paid with tests, and the debt list records the row as closed.
- **The installed binary lags `main` until a release.** A live `doctor` on this branch still prints the
  old line; the built binary and the tests print the new one. 0230 wrote the limit, and this session
  met it again from the other side.

## Next

- **The behavioural evaluation of the core skills** —does an agent do what a skill says—, deferred
  since the rsc-harness analysis. It is a different kind of work: a probe per skill, not a unit test.
- **A release** when another entry lands in `Unreleased`; one entry is under the project's own cap.
- One writer per worktree; one heavy gate per machine.

## The limits to remember

- The review counts even when it is the same model: it is a second reading, and the gate reads it.
- The gate's default timeout is 900 seconds; the suite alone takes about 250, and a cold or shared
  machine needs the bound; with the installer tests it ran about six minutes on this machine.
- A feature that touches a path the base fence protects needs the plan PR merged first; when the
  fence allows the path, the plan rides the feature PR.
- The installed binary is the gate: a change to `check` or `doctor` is invisible on this machine until
  a release or a local build, and CI builds the gate from the PR's base ref, not from the change.
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
