# Next session

Written at the close of 0228, `the-newest-verdict-is-the-store-s-newest` (unreleased, after v0.28.1).
The gate now reads the verdict the store wrote last; the last open item, the `all:*` pin, is next.

## Where the project is

- `main` at the merge of #420 (`d62fdf2`), `odachi 0.28.1` installed, `doctor` 15/15, no open feature,
  no open questions.
- 226 features and 569 steps closed; `docs/review-debt.md` has no open rows.
- 0225, 0226, 0227 and 0228 are in the changelog's Unreleased section: the release is not cut yet.

## What this session measured

- **A gate that reads the reviewer's clock reads the wrong tree.** `lastVerdictTree` ordered verdicts
  by the `at` the answer declares. 0227's correction said a date where its predecessors said a time,
  lost to the tree it corrected, and `check` named an existing verdict as missing. The fix asks the
  file's modification time —the fact `clean` and `review.NewestTrees` already order the store by— and
  the false line is gone from `main`.
- **The fixture needed the port, not a workaround.** The fake store gained `ModTime`, and it has to
  answer the slashed and the OS spelling the `Read` already accepted; without that it refused a file
  the read had just found, and a test failed for the fixture's reason, not the code's.
- **The red was observed on the right symptom.** With the old comparison the test printed
  `the feature is closed and its tree has no verdict` about a tree the store had judged; with the
  file time the note vanished and the gate stayed ready.
- **A review can find the same defect twice, and CI has its own.** 0227 earned three rounds —the
  spelling of a path and the untested branch— and a red from the Windows runner's short names.

## Next

The last open item, then the documental pair:

- **0216: pin `all:*` with a repo test.** A regression to `*` would pass the suite for dotfile assets;
  today only a build probe covers it. `skills/**` is protected by the current fence, so the plan PR
  lands before the test.
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
- A store's "newest" is the file time, not the time the record declares: a reviewer can write a date
  where a clock was expected.
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
