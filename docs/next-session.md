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

- **The 686-second reading was contaminated, and the correction is the lesson.** `cmd/odachi` alone
  took 207 s on 2026-11-06 (the `-v` run); the 686 s came from a run where another heavy suite was
  alive — its per-test times were *faster*, which only external load explains. Another project's
  `odachi check --run "... go test ./... -count=1"` was found running on this machine. The default
  stays 900 as a bound for a shared machine; the v0.22.1 note's 700 s is the contended number, not
  the suite's. The released note is not edited: the tag already published it.
- **The 600 s timeout in 0216 was the same mistake.** The gate ran while the verifier's own full
  suite was alive: one heavy gate per machine is a measurement rule, not a preference.
- **A verifier runs the focused commands, not a second full gate.** The prompt handed
  `go test ./...` to `odachi-verify`; the delegation reference says the request carries the
  verification commands, and the request should name the package, not the world.
- **A timeout proof is empty for a reason.** `go test` prints a package only when it finishes; at
  600 s none had, so the failed proof carried no output. It was not the adapter losing bytes.

## The next feature

- **The verifier's commands are focused.** `slices/references/delegation.md` says the request
  carries the verification commands; make it say the focused package, never a second full gate, and
  pin it in `skills/embed_test.go`.
- **Pin `all:*`** with a repo test, still open from 0216: a regression to `*` would pass the suite
  for dotfile assets.
- One writer per worktree; one heavy gate per machine — now measured, not a note.

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
