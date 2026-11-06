# Next session

Written at the close of 0213, `the-block-keeps-notes-not-builds`. The door every agent reads now
separates a note from a build: the clone keeps notes, and a build dies in the run that made it.

## What shipped

- **The block names two classes.** `.git/odachi/scratch/` is for a note a later turn has to read
  (a plan, a draft), and `clean` empties it. A build artifact and a one-use transport (a binary,
  a PR body, a prompt answer) live in the temp folder and die with the run. The golden block test
  pins the paragraph.
- **This project's block carries it.** A binary built on this tree rewrote `AGENTS.md`; a second
  `init` wrote nothing, `setup` left the thin door alone, and `doctor` check 11 is green.
- **The gate.** Proof `proof/0213-the-block-keeps-notes-not-builds.txt`; the review approved the
  tree on the reliability lens with no findings.

## What the review left

Nothing. `docs/review-debt.md` has no open rows.

## The next feature

- **Proportional verification (0214).** The last ODD gap from the comparison: a passive change is
  verified by reading, a medium one keeps the writer's own verification, and only a named
  high-risk item adds an independent verifier; plus two pieces odachi lacks, the comparison
  against the base ("did it already fail?") and the named known-environmental failures. It fits
  `slices/references/delegation.md` and `tests`.
- **The readback after publishing.** From the same comparison: `gh pr create` and `gh issue
  create` are never verified by reading the result back; a small rule in `pull-request`/`writing`
  would close it.
- **`assets/` in the skill contract**, only when a skill needs a template or a schema.
- One writer per worktree; one heavy gate per machine.

## The command that works

```sh
odachi review --base 30ef920dc4e3e406698fcafcba0e7ee30ada0b61 \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

`--base` is the base the task doc names, not `main`.

```sh
$env:PATH = "C:\Windows\System32;C:\Windows;C:\Windows\System32\Wbem;C:\Program Files\Go\bin;C:\Users\USER\go\bin;C:\Users\USER\AppData\Local\odachi\bin;C:\Users\USER\AppData\Roaming\npm"; odachi check --run "go test ./..."
```

The default fits the suite. The short PATH only repairs what the session inherits, and a build
goes to `%TEMP%\opencode\odachi\<feature>\` now, never to the clone.
