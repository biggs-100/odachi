# Next session

Written at the close of 0223, `the-smallest-change`. The overengineering investigation became two
rules: one before the brief, one in the PR.

## What shipped

- **`shape` asks for the smallest change.** Before declaring, the change names what it reuses, what
  it adds (files, dependencies, abstractions) and what it deletes; an abstraction needs a second
  caller; a pattern waits for three uses for the same reason; the props describe what is rendered,
  not the situation; a wrong abstraction is inlined back first.
- **`pull-request` carries the list.** Every PR names its new files, dependencies and abstractions
  with their reason and the code it reuses, and the reviewer asks what to delete, what was reused
  and which requirement each line serves.
- **The gate.** `TestShapeAsksForTheSmallestChange` and `TestPullRequestNamesWhatItAdds` were red
  before the text landed. Proof `proof/0223-the-smallest-change.txt`; the review approved the tree on
  the reliability lens with no findings. Released as v0.28.0 and installed.

## What the investigation measured

- **The shape is duplication and reinvention, not volume.** Block duplication +81% (40.3 → 73.0 per
  million changed lines, 2023–2026), moved/refactored code 21% → 3.8%, function connectivity −35%
  (GitClear, 623M changes); 40% of developers say AI produces unnecessary or duplicative code and
  96% doubt its reliability (Sonar, 1,100 devs); 66% spend more time fixing "almost right" code
  (Stack Overflow, 49k devs).
- **The fix is instruction, not tooling.** An ACL ARR 2026 submission ("Don't Over-Engineer It")
  measures agents adding unrequested tests and documentation on SWE-bench and reports a single
  instruction reverses it. The paper is not peer-reviewed yet and its PDF is behind OpenReview's
  browser check; the abstract is what we have.
- **The rule of three has a test.** Same reason means one decision would change every copy at once;
  the props smell test separates a sound abstraction from a merged situation; the recovery is
  inline-first (Stack Overflow, November 2026).
- **`pull-request` is at 59 lines.** The next rule for it moves depth to a reference; `shape` is at
  49.

## The next feature

- **The gate names what the change adds.** `check` reports the new files, dependencies and
  abstractions in the diff as a warning, not a blocker, so the reviewer sees the scaffolding before
  reading the code. It is a code feature in `internal/app/check`, with its own test.
- **The installer names the process that holds the binary**, open from 0218.
- **Pin `all:*`** with a repo test, open from 0216.
- One writer per worktree; one heavy gate per machine.

## The commands that work

```sh
odachi review --base da3353df799a470576a65b267c7c35871899b432 \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

`--base` is the base the task doc names, not `main`.

```sh
$env:PATH = "C:\Windows\System32;C:\Windows;C:\Windows\System32\Wbem;C:\Program Files\Go\bin;C:\Users\USER\go\bin;C:\Users\USER\AppData\Local\odachi\bin;C:\Users\USER\AppData\Roaming\npm"; odachi check --run "go test ./..."
```

No `--timeout` is needed; the default is 900. A build goes to
`%TEMP%\opencode\odachi\<feature>\`, never to the clone.
