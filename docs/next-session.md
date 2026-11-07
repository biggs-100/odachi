# Next session

Written at the close of 0225, `the-declared-risk-meets-the-diff` (unreleased, after v0.28.1). The
declared `Risk:` now meets the diff; the next feature gives a non-goal the condition that reopens it.

## Where the project is

- `main` at the merge of #412 (`8b94bd7`), `odachi 0.28.1` installed, `doctor` 15/15, no open feature,
  no open questions.
- 223 features and 563 steps closed; `docs/review-debt.md` has no open rows.
- 0225 shipped in two stacked PRs (#411 the parser, #412 the note) and is in the changelog's Unreleased
  section: the release is not cut yet.

## What this session measured

- **The close commit needs the proof before it exists.** The hook refused the task-doc commit with
  "the proof for this change is missing": `check --run` judges the worktree, so the proof is written
  with the close change uncommitted and the commit then carries exactly that tree.
- **A 408-line feature is two PRs.** 0225 was 387 + / 21 -; the parser and the note split at 219 and
  197 changed lines. The price: `ParseRisk` lands on `main` with tests and no consumer until #412; the
  stack accepts an intermediate state one PR would hide.
- **The verdict took seconds** with the documented command and `deepseek-v4.1-flash`: approved on the
  reliability lens, no findings, no re-review.
- **The note is silent for our own feature.** 0225 touches `internal/app/check` and
  `internal/domain/change`, which the classifier calls focused, so the new line never appears in this
  feature's own ready output; its five tests are the only witness.

## The next feature

**0226: a non-goal says when to revisit it.** A brief's "Not included" is a promise with no date, and
the list grows until nobody knows which entry still holds. The convention: each non-goal names the
condition that reopens it (a measurement, a second caller, a version), and `shape` asks for it.

- the brief's non-goal section carries the convention, and `shape` asks the question when it drafts one
- a test pins the skill's text, observed red before the text lands
- the changelog's Unreleased section names the convention

Then the two documental pieces (the verdict/proof binding in `docs/releases.md`, the harness cost in
`doctor`) can ride the next close. The behavioural evaluation of the core skills stays deferred.

## Still open from earlier sessions

- **0218: the installer names the process that holds the binary.** When `odachi.exe` is in use, say
  which process and what to do, instead of a `Copy-Item` IOException. Two installs were blocked by
  another project's gate.
- **0216: pin `all:*` with a repo test.** A regression to `*` would pass the suite for dotfile
  assets; today only a build probe covers it.
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
