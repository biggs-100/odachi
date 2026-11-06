# Next session

Written at the close of 0217, `the-gate-fits-the-suite`. The gate's default bound now fits the
suite this machine actually runs.

## What shipped

- **The default is 900 seconds.** Measured on 2026-11-06: `go test ./...` takes about 700 seconds;
  `cmd/odachi` alone takes 207, but 686 under `./...` because it competes with `gitx` (113) and
  `shellcmd` (26). The 600-second default killed a green suite. The constant, the test and the
  comment name the measurement; the help text and `docs/odachi-bootstrap.md` say 900.
- **The gate itself did not move.** Verdicts, reasons, exit codes, proof format and `--timeout`
  semantics are untouched; only the default bound changed.
- **The gate.** Proof `proof/0217-the-gate-fits-the-suite.txt`, written by the tree's binary with no
  `--timeout`; the review approved the tree on the reliability lens with no findings. Released as
  v0.22.1 and installed.

## What was measured

- **The suite is 3.3x slower under `./...` than alone.** `cmd/odachi` 207 s alone, 686 s under
  `./...`, with `gitx` at 113 s and `shellcmd` at 26 s in parallel. The wall time is about 700 s.
- **A timeout proof is empty for a reason.** `go test` prints a package only when it finishes; at
  600 s none had, so the failed proof carried no output. It was not the adapter losing bytes.
- **The number is the default, not a waiver.** `--timeout` still tightens the bound; the gate ran
  green without it.

## The next feature

- **Make the suite faster.** The contention is the real cost: `cmd/odachi` spends 686 s under
  `./...` against 207 alone. Test isolation, a shared harness, or `-p` are the candidates; the
  measurement is in the constant's comment and this file.
- **Pin `all:*`** with a repo test, still open from 0216: a regression to `*` would pass the suite
  for dotfile assets.
- One writer per worktree; one heavy gate per machine.

## The commands that work

```sh
odachi review --base 1212e58f0ee1307a6b55e8ede06d1b33224db531 \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

`--base` is the base the task doc names, not `main`.

```sh
$env:PATH = "C:\Windows\System32;C:\Windows;C:\Windows\System32\Wbem;C:\Program Files\Go\bin;C:\Users\USER\go\bin;C:\Users\USER\AppData\Local\odachi\bin;C:\Users\USER\AppData\Roaming\npm"; odachi check --run "go test ./..."
```

No `--timeout` is needed now; the default is 900. A build goes to
`%TEMP%\opencode\odachi\<feature>\`, never to the clone.
