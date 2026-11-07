# Next session

Written at the close of 0227, `the-installer-names-the-process-that-holds-the-binary` (unreleased,
after v0.28.1). The Windows installer now names the process that holds the binary; the last open
item, the `all:*` pin, is next.

## Where the project is

- `main` at the merge of #417 (`7611ea2`), `odachi 0.28.1` installed, `doctor` 15/15, no open feature,
  no open questions.
- 225 features and 567 steps closed; `docs/review-debt.md` has no open rows: 0227 closed its last one
  with the test it was missing.
- 0225, 0226 and 0227 are in the changelog's Unreleased section: the release is not cut yet.

## What this session measured

- **A review can find the same defect twice.** The first verdict caught the holder search comparing
  spellings (`C:/tools/odachi` against the process's `C:\...`), the second the untested no-holder
  branch. Both were paid with tests; the third verdict carried the second row because the fix is a
  test and the script's bytes did not move, and the debt list is where it closed.
- **CI's Windows runner is short-named.** The gate failed on `RUNNER~1`: the installer prints the
  canonical path and the test compared the short spelling. `GetLongPathName` in the test asks the
  same question the script asks; the red was the environment, not the change (0214).
- **PowerShell 5.1 reads a BOM-less `.ps1` as ANSI behind `-File`.** The accents in the installer's
  message arrive as mojibake in the test; the documented `irm | iex` path reads the UTF-8 bytes
  correctly. The installer tests assert ASCII only, on purpose.
- **The Windows harness can hold a file two ways.** A sleeper whose own image is the target (the real
  case: another project's pre-commit hook) and a handle opened without sharing from a process odachi
  cannot name (an indexer, an antivirus).

## Next

The last open item, then the documental pair:

- **0216: pin `all:*` with a repo test.** A regression to `*` would pass the suite for dotfile assets;
  today only a build probe covers it.
- **The verdict/proof binding in `docs/releases.md`** and **the harness cost in `doctor`** can ride the
  close of whichever lands next.
- The behavioural evaluation of the core skills stays deferred.
- One writer per worktree; one heavy gate per machine.

## The limits to remember

- The review counts even when it is the same model: it is a second reading, and the gate reads it.
- The gate's default timeout is 900 seconds; the suite alone takes about 250, and a cold or shared
  machine needs the bound; with the installer tests it ran about six minutes on this machine.
- A feature that touches a path the base fence protects needs the plan PR merged first; when the
  fence allows the path, the plan rides the feature PR.
- A test that compares a path compares what Windows reports: ask `GetLongPathName`, not the folder's
  spelling.
- The ruleset refuses a direct push to `main`: every landing, documents included, goes through a PR.

## The commands that work

```sh
odachi review --base <the base the task doc names> \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

`--base` is the base the task doc names, not `main`. A correction after a fix is
`odachi review --since <the approved tree>` with the same `--exec` and no `--base`: the stored
verdict names it.

```sh
$env:PATH = "C:\Windows\System32;C:\Windows;C:\Windows\System32\Wbem;C:\Program Files\Go\bin;C:\Users\USER\go\bin;C:\Users\USER\AppData\Local\odachi\bin;C:\Users\USER\AppData\Roaming\npm"; odachi check --run "go test ./..."
```

No `--timeout` is needed; the default is 900. A build goes to
`%TEMP%\opencode\odachi\<feature>\`, never to the clone.
