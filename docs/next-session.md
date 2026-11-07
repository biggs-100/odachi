# Next session

Written at the close of 0224, `the-gate-names-what-the-change-adds` (v0.28.1). The overengineering
investigation is closed in its two halves; the rsc-harness analysis is documented and its first
improvement waits for the next session.

## Where the project is

- `main` at the merge of the 0224 close, `odachi 0.28.1` installed, `doctor` 15/15, no open feature,
  no open questions.
- 222 features and 560 steps closed; `docs/review-debt.md` has no open rows.
- The design, cognitive and PR-system comparisons with gentle-shell and gentle-ai are paid.
- The overengineering investigation is paid: the rule in `shape`, the review questions in
  `pull-request`, the measurement in `check` (v0.28.0 and v0.28.1).

## What this session measured

- **The rsc-harness analysis.** Read-only, 274 skills, 162 tests, v4.0.5. The full analysis and the
  four improvements live in `docs/learning-from-rsc-harness.md`; the short version is below.
- **A note that always fires is noise.** The gate's new line first named the task doc in every
  feature; the filter for `odd/` and `proof/` is what makes it mean "scaffolding".
- **The gate test was written after the code.** The RED was observed by switching the line off; it is
  weaker than test-first, and the next gate feature starts with the test.
- **The skill bodies are near the cap.** `pull-request` is at 59 lines; the next rule for it goes to
  a reference.

## The next feature

**0225: the declared `Risk:` meets the diff.** Odachi classifies a diff with
`internal/domain/review/tier.go` (auth, money, contracts, concurrency, delivery), but the `Risk:`
line in the commit body is the agent's word and nothing contrasts it. The check already reads the
commit messages for the `Odachi-Change:` trailer; it can read `Risk:` too.

- A parser for `Risk:` in `internal/domain/change`, next to the trailer, with its refusals.
- A note in the ready output, never a block: "the review classifies this change as full (it touches
  <paths>), and the commits declare Risk: none: name the risk or say why it is none."
- Tests: a full-tier diff with `Risk: none` warns; a full-tier diff with a named risk does not; a
  passive diff stays silent; the verdict never moves.

Then **0226: a non-goal says when to revisit it** (a convention in the brief and `shape`, pinned).

The two documental pieces (the verdict/proof binding in `docs/releases.md`, the harness cost in
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

## The commands that work

```sh
odachi review --base 9a211be952e250bfc7adc0acdbe496c88143ff1f \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

`--base` is the base the task doc names, not `main`.

```sh
$env:PATH = "C:\Windows\System32;C:\Windows;C:\Windows\System32\Wbem;C:\Program Files\Go\bin;C:\Users\USER\go\bin;C:\Users\USER\AppData\Local\odachi\bin;C:\Users\USER\AppData\Roaming\npm"; odachi check --run "go test ./..."
```

No `--timeout` is needed; the default is 900. A build goes to
`%TEMP%\opencode\odachi\<feature>\`, never to the clone.
