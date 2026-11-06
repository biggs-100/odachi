# Next session

Written at the close of 0218, `the-verifier-runs-focused-commands`. The delegation request now names
which commands the verifier receives.

## What shipped

- **The verifier runs the focused package.** `slices/references/delegation.md` says the request
  carries the focused package the change touches, never a second `go test ./...`, because a full
  suite beside the gate competes for the machine and makes the gate's own measurement wrong (0217).
  The verifier stays read-only: the rule changes the commands it receives, not what it may do.
- **The gate.** `TestTheVerifierRunsFocusedCommands` was red before the text landed. Proof
  `proof/0218-the-verifier-runs-focused-commands.txt`; the review approved the tree on the
  reliability lens with no findings. Released as v0.23.0 and installed.

## What was measured

- **The install hit the same contention, from the other side.** The installer failed with a raw
  `Copy-Item` IOException because another project's `odachi check --run` held
  `%LOCALAPPDATA%\odachi\bin\odachi.exe`. It installed once that gate released the file. The message
  is the defect: it should name the running process and the next action, not quote PowerShell.
- **The other project's gate is a regular guest.** PID 22540 ran the same
  `go test ./... -count=1` command found during 0217's correction; it started while this session was
  closing. One heavy gate per machine is now a rule with a measurement behind it.
- **The default of 900 covered the wait.** The install waited twelve minutes and the gate was green
  before it; the bound did its job.

## The next feature

- **The installer names the process that holds the binary.** When `odachi.exe` is in use, the
  installer says which process and what to do, instead of a `Copy-Item` IOException. The measurement
  is in this file: two installs blocked today, both by a `check --run` in flight.
- **Pin `all:*`** with a repo test, still open from 0216: a regression to `*` would pass the suite
  for dotfile assets.
- One writer per worktree; one heavy gate per machine.

## The commands that work

```sh
odachi review --base e28725a045b526a673b97407d39165aa7cf15f6d \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

`--base` is the base the task doc names, not `main`.

```sh
$env:PATH = "C:\Windows\System32;C:\Windows;C:\Windows\System32\Wbem;C:\Program Files\Go\bin;C:\Users\USER\go\bin;C:\Users\USER\AppData\Local\odachi\bin;C:\Users\USER\AppData\Roaming\npm"; odachi check --run "go test ./..."
```

No `--timeout` is needed; the default is 900. A build goes to
`%TEMP%\opencode\odachi\<feature>\`, never to the clone.
