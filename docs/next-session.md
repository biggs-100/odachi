# Next session

Written at the close of 0205, `the-skill-contract`. A skill is now a folder: its `SKILL.md`
routes and the depth travels beside it in `references/`, installed and checked by the binary.

## What shipped

- **The skill folder.** `skills/<name>/references/*.md` ship with the skill: `odachi init` and
  `odachi pack` install every file, `odachi doctor` compares the folder file by file, and
  `odachi skills --check` refuses a path the body names and the folder does not have, a file
  nobody names in `## References`, a reference that is not a `.md`, a router whose gates omit a
  reference, and an empty `## Output`. The 60-line body cap stays.
- **The grammar.** A skill with more than one reference routes them in `## Decision gates` (one
  row per situation, the path in backticks); a skill with references says what it returns in
  `## Output` and lists them in `## References`. `skill-craft` carries the rule and ADR 0050
  records the target and the migration debt.
- **The carriers.** `ui-craft` is the first router: the order and the first move stay in the
  body, and the catalogue, the numbers, the motion and the last read are four references.
  `design` declares its output and keeps its 60 lines by moving the folder, manifest and
  comments formats to `references/files.md`.
- **The doctor debt.** The design-plugin review-debt row is paid: the door note always writes a
  sentence, so the empty branch was unreachable; the dead check is gone.
- **The gate.** The proof is `proof/0205-the-skill-contract.txt`; the review is approved for
  tree `23faa45` with 2 WARNING and 2 SUGGESTION, all causal, all in `docs/review-debt.md`.

## What the review left

Four rows in `docs/review-debt.md` are open: the old `Messages`/`Head` row, plus
`internal/app/pack/pack.go:461` (a new reference is counted as a backup although none ran),
`internal/domain/skill/skill.go:202` (the gates check is a substring, not a parsed table), and
`internal/app/doctor/doctor.go:882` (the missing-reference branch has no test). The next feature
that touches those paths pays them.

## The next feature

- **The migration of the other sixteen skills.** Nineteen skills exist; 0205 touched three
  (`ui-craft`, `design`, `skill-craft`). The other sixteen do not carry `## Output` yet, and
  `--check` does not demand it from a skill with no references. The follow-up adds the output
  contract where it earns its place, and splits into references the bodies that exceed the cap.
  The ADR 0050 names this debt.
- **The gate is slower than its default timeout.** `go test ./...` takes ~411s on this machine
  (`cmd/odachi` is the bulk and is not cacheable because its tests touch the environment), while
  `check`'s default timeout is 120s. The proof needed `--timeout 900`. A feature of its own: a
  default that fits the suite, or a suite split that keeps the gate honest. Do not "fix" it with a
  shorter command: the proof is the real suite.
- **One writer per worktree; one heavy gate per machine.** Two agents in the same checkout race the
  index and the tree a proof certifies. Two gates on one machine fight for the Go cache and the
  CPU: the second looks hung. Parallel lines belong in separate clones or worktrees. gentle-ai is
  the reference: `docs/opencode-profiles.md` there says background jobs are non-durable and must
  not be parallel writers in one worktree, and its review binds a frozen tree and worktree.

## The command that works

The shell adapter that runs `review --exec` does not resolve `opencode` from PATH on this machine;
the reviewer needs the host executable. The invocation that worked:

```sh
odachi review --base 5332d6f6ee2b696de540f0c428d99b40fa126fb3 \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

`--base` is the base the gate names (the task doc's `Base:`), not `main`: with `main` the review
judges a candidate that excludes the plan commit and says nothing changed. The proof:

```sh
odachi check --run "go test ./..." --timeout 900
```

The default 120s is too short for this suite.
