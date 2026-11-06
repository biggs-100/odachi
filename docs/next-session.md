# Next session

Written at the close of 0210, `the-environment-a-command-inherits`. The doctor's fifteenth
check names the PATH a command inherits, so a proof that fails because a binary disappeared has
a line that explains why.

## What shipped

- **0210, the environment.** `odachi doctor` gained check 15: it measures the PATH this process
  hands to a child, warns above the 8191 characters cmd can pass (counted in characters, not
  bytes), and on a platform that does not cut the environment it says which one it measured.
  `CheckCount`, the usage and `docs/impact-map.md` say fifteen. The configured-project fixture
  sets a short PATH so the check does not read the machine's own.
- **0207-0209, earlier in the session.** The four review-debt rows paid (pack's report, the
  gates table, the doctor's missing drift, gitx's Head); the corrupt ref pinned; and the gate's
  default timeout at 600 seconds.
- **The gates.** `proof/0210-the-environment-a-command-inherits.txt`; the review of the first
  tree found the platform not named and bytes counted instead of characters, both paid in the
  branch, and the correction's verdict refuted a false BLOCKER (the `strconv` import was in the
  tree, not in the delta).

## What the review left

Nothing open. The two findings of 0210 were paid before the close, the blocker was refuted with
the file and the proof, and `docs/review-debt.md` has no open rows.

## The next feature

- **One writer per worktree; one heavy gate per machine.** Parallel lines belong in separate
  clones, and two gates on one machine fight for the Go cache.
- **The store is 79 MB.** `odachi clean --apply` would free 77.9 MB, and the lifetimes in
  `odd/review.toml` are the switch: `keep_prompts = 0` already, `keep_verdicts = 10`. The
  decision to run it is the human's (0187).
- Nothing else is named: the ledger is empty and the last session's candidates are closed.

## The command that works

```sh
odachi review --base fd32dc1d539ad7c0d69efeb2e68024c98acccbae \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

`--base` is the base the task doc names, not `main`: with `main` the review judges a candidate
that excludes the plan commit and says nothing changed.

```sh
$env:PATH = "C:\Windows\System32;C:\Windows;C:\Windows\System32\Wbem;C:\Program Files\Go\bin;C:\Users\USER\go\bin;C:\Users\USER\AppData\Local\odachi\bin;C:\Users\USER\AppData\Roaming\npm"; odachi check --run "go test ./..."
```

The default now fits the suite; the short PATH only repairs what the session inherits, and the
doctor's check 15 says so when it does not.
