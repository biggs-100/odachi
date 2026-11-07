# Next session

Written at the close of 0224, `the-gate-names-what-the-change-adds`. The overengineering
investigation is closed: the rule lives in `shape`, the review questions in `pull-request`, and the
gate names what a change adds.

## What shipped

- **The gate names the scaffolding.** The ready output lists the new files (capped) and the
  dependency files the diff adds, in one line and without moving the verdict. The new files come
  from the diff the candidate already read; the dependency files from a named list of manifests and
  lockfiles. The plan and the proof are filtered out: every feature adds its task doc, and a note
  that always fires is a note nobody reads.
- **What it does not do.** It does not parse abstractions (not machine-detectable), and it never
  blocks: the gate reports, the PR names the reason, the reviewer asks the three questions.
- **The gate.** `TestTheReadyOutputNamesWhatTheChangeAdds` and
  `TestTheReadyOutputIsSilentWhenTheChangeAddsNothing` cover it; the first was observed red by
  switching the line off. Proof `proof/0224-the-gate-names-what-the-change-adds.txt`; the review
  approved the tree on the reliability lens with no findings. Released as v0.28.1 and installed.

## What was measured

- **A note that always fires is noise.** The first cut would have named the task doc in every
  feature; the filter for `odd/` and `proof/` is what makes the line mean "scaffolding".
- **The diff already had the answer.** New files are `--- /dev/null` followed by `+++ b/<path>` in
  the patch the candidate read; a second git call would have been the overengineering the feature
  warns about.
- **The tests were written after the code.** The RED was observed by switching the line off, which
  is weaker than writing the test first; the next gate feature starts with the test.

## The next feature

- **The installer names the process that holds the binary**, open from 0218: when `odachi.exe` is in
  use, say which process and what to do, instead of a `Copy-Item` IOException.
- **Pin `all:*`** with a repo test, open from 0216: a regression to `*` would pass the suite for
  dotfile assets.
- One writer per worktree; one heavy gate per machine.

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
