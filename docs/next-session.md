# Next session

Written at the close of 0222, `the-chain-says-where-it-stands`. The PR-system comparison with
gentle-ai is paid: the chain and the rollback boundary are in the pair.

## What shipped

- **A chained PR says where it stands.** `pull-request` names the branch prefixes (`feat/`, `fix/`,
  `docs/`, `chore/`, `refactor/` and the project's lanes), the chain context (tracker, position,
  base, depends on, next PR, what it leaves out), the draft tracker for a feature-branch chain, the
  `📍` diagram, and the polluted-diff rule: retarget or rebase until only this unit shows.
- **A commit says what a revert would take with it.** `commits` now names the files or the behavior
  that cannot be removed alone, so each commit can be reverted by itself.
- **The gate.** `TestPullRequestCarriesTheChain` and `TestCommitsNameTheRollback` were red before the
  text landed. Proof `proof/0222-the-chain-says-where-it-stands.txt`; the review approved the tree on
  the reliability lens with no findings. Released as v0.27.0 and installed.

## What was measured

- **The bodies are close to the cap.** `pull-request` is at 55 lines and `commits` at 57; the next
  rule for either skill should move depth to a reference instead of adding a line.
- **The lens count varies with the path.** The 0221 change (writing) earned all four lenses; the
  0222 change (pull-request and commits) earned the reliability floor. The classifier is path-based,
  not a measure of the change's importance.
- **The install was clean.** The release installer found the 19 skills already equal: the tree's
  binary had installed them before the gate.

## The next feature

- **The installer names the process that holds the binary**, open from 0218: when `odachi.exe` is in
  use, say which process and what to do, instead of a `Copy-Item` IOException.
- **Pin `all:*`** with a repo test, open from 0216: a regression to `*` would pass the suite for
  dotfile assets.
- **The next skill rule moves depth to a reference:** `pull-request` and `commits` are near the
  60-line cap.
- One writer per worktree; one heavy gate per machine.

## The commands that work

```sh
odachi review --base 0cddda8d773faf10248340217369abb4fa9b7a3c \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

`--base` is the base the task doc names, not `main`.

```sh
$env:PATH = "C:\Windows\System32;C:\Windows;C:\Windows\System32\Wbem;C:\Program Files\Go\bin;C:\Users\USER\go\bin;C:\Users\USER\AppData\Local\odachi\bin;C:\Users\USER\AppData\Roaming\npm"; odachi check --run "go test ./..."
```

No `--timeout` is needed; the default is 900. A build goes to
`%TEMP%\opencode\odachi\<feature>\`, never to the clone.
