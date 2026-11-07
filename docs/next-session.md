# Next session

Written at the close of 0229, `the-skills-embed-is-measured` (unreleased, after v0.28.1). The embed
that ships the skills is measured; the backlog the earlier sessions were carrying is empty.

## Where the project is

- `main` at the merge of #423 (`ba17cfe`), `odachi 0.28.1` installed, `doctor` 15/15, no open feature,
  no open questions.
- 227 features and 571 steps closed; `docs/review-debt.md` has no open rows.
- 0225, 0226, 0227, 0228 and 0229 are in the changelog's Unreleased section: the release is not cut.

## What this session measured

- **An embed directive is invisible without a file that a regression makes disappear.** The first
  fixture carried only hidden files; with `*` the directory had nothing embeddable and the *build*
  failed, which kept the assertion from ever running. A plain control file keeps the directory
  embeddable with either pattern and moves the red into the test, where it names each missing file.
- **`*` still walks a dot directory; what it skips are the files whose names begin with `.` or `_`.**
  The fixture proved it: with `*`, `.embed-fixture/_under.txt` and `.embed-fixture/.hidden.txt` vanish
  while the plain file stays.
- **Two test helpers counted every directory as a skill.** The fixture exposed the gap: the loader and
  `skills --check` skip a directory without a `SKILL.md`, and the expectations did not. Both now stat
  the file first.
- **A closed feature's commit and its record travel together.** The hook refused the helper fix alone:
  the proof judges the worktree, and the record must be in the tree the proof names. The fix and the
  T1/T2 record landed in one commit.

## Next

The backlog is empty; what is left is documentation and a release:

- **The verdict/proof binding in `docs/releases.md`** and **the harness cost in `doctor`**: the two
  documental pieces that have ridden every close for several sessions.
- **The release**: 0225 to 0229 are in Unreleased; a `v0.29.0` tag with its note is the natural close
  of this stretch.
- The behavioural evaluation of the core skills stays deferred.
- One writer per worktree; one heavy gate per machine.

## The limits to remember

- The review counts even when it is the same model: it is a second reading, and the gate reads it.
- The gate's default timeout is 900 seconds; the suite alone takes about 250, and a cold or shared
  machine needs the bound; with the installer tests it ran about six minutes on this machine.
- A feature that touches a path the base fence protects needs the plan PR merged first; when the
  fence allows the path, the plan rides the feature PR.
- A test that pins an embed needs a file a regression makes disappear, and a plain file beside it so
  the package still builds and the red lands in the test.
- A store's "newest" is the file time, not the time the record declares.
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
