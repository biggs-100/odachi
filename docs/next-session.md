# Next session

Written at the close of 0207, `the-ledger-of-0205-is-paid`; 0208, `the-corrupt-ref`, closed the
last row with one test. The ledger is empty, and the next candidates are the gate's default
timeout and the PATH a session inherits.

## What shipped

- **`pack` counts what happened.** A skill that only gained a file is `written` (JSON
  `written`), never "backed up and written": `installSkill` records whether `files.Backup`
  ran. Two tests pin the real run and the dry run, both observed red first.
- **`skills --check` reads the grammar.** `## Decision gates` is a data row and
  `## References` is a bullet; the table header and the separator route nothing, and the
  refusal says which shape it wanted. Three tests were red against the old `Contains`.
- **The doctor's drift has both halves.** `TestTheDoctorNamesAReferenceTheInstalledCopyLacks`
  seeds a source reference the installed copy lacks; observed red by disabling the branch.
- **`gitx.Messages` keeps its errors.** Only an unborn HEAD is an empty history:
  `symbolic-ref -q HEAD` tells the two apart, and the callers block with "could not read the
  commits since `<base>`". The empty-repository test stayed green.
- **The gate.** Proof `proof/0207-the-ledger-of-0205-is-paid.txt`; review approved tree
  `bcfbc7d` on the reliability lens.

## What the review left

One WARNING was recorded in `docs/review-debt.md`: it said `symbolic-ref` answers for a corrupt
ref too, so the corrupt half of the last row was still swallowed. Measured false on git
2.52.0.windows.1 -a corrupt ref makes `symbolic-ref` exit 128, and a valid sha of a missing
object fails at `git log`- and 0208 pinned the shape with `TestMessagesRefusesACorruptRef`,
observed red by restoring the old swallow. The ledger is empty.

## The next feature

- **The corrupt-ref test.** One test in `gitx`, so the warning cannot return without a red. It
  is the only row open.
- **The gate is slower than its default timeout.** `go test ./...` takes ~411s on this machine,
  `check`'s default is 120s, and every proof here needs `--timeout 900`. A default that fits
  the suite, or a suite split that keeps the gate honest: not a shorter command.
- **The PATH a session inherits.** This host carried a 13.8 KB `PATH`; Windows truncates it
  for children, so `findstr`, `ping` and `start` did not resolve, `internal/adapters/shellcmd`
  failed for an environment reason, and the v0.18.0 installer could not find `cosign` (it
  lives in `C:\Users\USER\go\bin`). A doctor row that names the length and the repair would
  have saved both diagnoses.
- **One writer per worktree; one heavy gate per machine.** Parallel lines belong in separate
  clones.

## The command that works

```sh
odachi review --base 132f51788ebc58f20b8ff149faea1006854eb37d \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

`--base` is the base the task doc names, not `main`: with `main` the review judges a candidate
that excludes the plan commit and says nothing changed.

```sh
$env:PATH = "C:\Windows\System32;C:\Windows;C:\Windows\System32\Wbem;C:\Program Files\Go\bin;C:\Users\USER\go\bin;C:\Users\USER\AppData\Local\odachi\bin;C:\Users\USER\AppData\Roaming\npm"; odachi check --run "go test ./..." --timeout 900
```

The proof is the real suite; the short PATH only repairs what the session inherits.
