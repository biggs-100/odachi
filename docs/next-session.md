# Next session

Written at the close of 0221, `writing-measures-the-review-path`. The cognitive comparison with
gentle-ai is paid: the review path and the commands that measure it are in the pair.

## What shipped

- **The document names the review path.** `the-doc.md` has `## Commands` with
  `git diff --name-only -- '*.md'`, and a review checklist: what to review first, what is
  intentionally out of scope, the previous and the next PR, one decision per section and the
  acceptance checklist.
- **`pull-request` measures the budget.** The 400-line budget is checked with
  `gh pr view <PR> --json additions,deletions,changedFiles`, not estimated by eye.
- **A dense document triggers `writing`.** Its `## When` now names a document that feels long,
  dense or hard to scan, not only one that has to make someone act.
- **The gate.** The two pinned tests were red before the text landed. Proof
  `proof/0221-writing-measures-the-review-path.txt`; the review approved the tree with all four
  lenses (risk, readability, reliability, resilience) and no findings. Released as v0.26.0 and
  installed.

## What was measured

- **A skills-only change can earn the four lenses.** The review classifier treated the changed
  `skills/` paths as shipped text and ran risk, readability, reliability and resilience instead of
  the floor; it approved 4 of 4. That is stronger than expected, not a defect.
- **The plan-first flow held again.** The base fence protected `writing` and `pull-request`, so the
  plan landed first (#395) and the feature followed (#396).
- **The install was clean.** The release installer replaced two skills and left 19 matched.

## The next feature

- **The installer names the process that holds the binary**, open from 0218: when `odachi.exe` is in
  use, say which process and what to do, instead of a `Copy-Item` IOException.
- **Pin `all:*`** with a repo test, open from 0216: a regression to `*` would pass the suite for
  dotfile assets.
- One writer per worktree; one heavy gate per machine.

## The commands that work

```sh
odachi review --base 60c4b4e0123b2ab441a85aeb4bacaf776c0e37f3 \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

`--base` is the base the task doc names, not `main`.

```sh
$env:PATH = "C:\Windows\System32;C:\Windows;C:\Windows\System32\Wbem;C:\Program Files\Go\bin;C:\Users\USER\go\bin;C:\Users\USER\AppData\Local\odachi\bin;C:\Users\USER\AppData\Roaming\npm"; odachi check --run "go test ./..."
```

No `--timeout` is needed; the default is 900. A build goes to
`%TEMP%\opencode\odachi\<feature>\`, never to the clone.
