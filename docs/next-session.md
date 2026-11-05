# Next session

Written at the close of 0206, `the-grammar-reaches-every-skill`. Every skill the
curriculum ships now says what it returns, and `odachi skills --check` refuses one
that does not.

## What shipped

- **The contract belongs to every skill.** `internal/domain/skill/skill.go` asks for a
  non-empty `## Output` before it looks at references, so a skill with no depth owes the
  same section as a router. The refusal names the skill and the section. ADR 0050 carries
  an amendment that says the staged migration landed.
- **The seventeen bodies.** `acceptance`, `bootstrap`, `commits`, `data`, `drive`,
  `go-tests`, `pull-request`, `role-craft`, `security`, `shape`, `ship`, `skill-craft`,
  `skill-improve`, `slices`, `tests`, `triage` and `writing` declare what they hand back.
  `drive` and `slices` were at the sixty-line cap, so each moved one block word for word:
  `drive`'s block codes and review policy to `references/the-gate.md`, and `slices`'
  delegation policy to `references/delegation.md`.
- **The scaffold.** `odachi skill new` writes `## Output` with the other sections, so the
  template passes `--check` unchanged, and `skill-craft` teaches the rule that now holds.
- **The installation.** A binary built from this tree refreshed `~/.agents/skills`: 17
  skills replaced, the second run wrote nothing, and doctor's check 9 compares 19 with no
  drift.
- **The gate.** The proof is `proof/0206-the-grammar-reaches-every-skill.txt`; the review
  approved tree `a99ec3a` with no findings on the four lenses.

## What the review left

Nothing. The four rows of 0205 are still open because this feature changed none of the
behaviour they name: `internal/app/pack/pack.go:461` (a new reference counted as a backup),
`internal/domain/skill/skill.go:202` (the gates check is a substring, not a parsed table),
`internal/app/doctor/doctor.go:882` (the missing-reference drift branch has no test) and
`internal/adapters/gitx/gitx.go:220` (any `Head` failure read as an empty history).

## The next feature

- **The four debt rows above.** They are small, they live in code the last two features
  touched, and the ledger has carried them long enough.
- **The gate is slower than its default timeout.** `go test ./...` takes ~411s on this
  machine, `check`'s default is 120s, and every proof here needs `--timeout 900`. A default
  that fits the suite, or a suite split that keeps the gate honest: not a shorter command.
- **One writer per worktree; one heavy gate per machine.** Two agents in one checkout race
  the index and the tree a proof certifies. Parallel lines belong in separate clones.
- **The PATH a session inherits.** This session's host process carried a 13.8 KB `PATH`
  (the installer tests before 0202 left their temp dirs in it); Windows truncates a long
  environment for children, so `findstr`, `ping` and `start` did not resolve and
  `internal/adapters/shellcmd` failed for an environment reason, not a code one. The
  machine's own PATH is clean (2.3 KB). If a gate fails that way, set a short `$env:PATH`
  before `odachi check --run` and say so in the close.

## The command that works

```sh
odachi review --base eac2079a2075e7a879bad670eed4994be98fc349 \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

`--base` is the base the task doc names, not `main`: with `main` the review judges a
candidate that excludes the plan commit and says nothing changed.

```sh
$env:PATH = "C:\Windows\System32;C:\Windows;C:\Windows\System32\Wbem;C:\Program Files\Go\bin;C:\Users\USER\AppData\Local\odachi\bin;C:\Users\USER\AppData\Roaming\npm"; odachi check --run "go test ./..." --timeout 900
```

The proof is the real suite; the short PATH only repairs what this session inherited.
