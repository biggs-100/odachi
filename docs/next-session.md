# Next session

Written while closing 0264. Nothing is owed.

## Where the project is

- `v0.38.7` is published (tag on `edd782e`); this session shipped `v0.38.6` (0263) and `v0.38.7` (0264).
- The installed binary is `0.38.7`; `doctor` is clear except this file, which this run pays.
- 262 features, 639 steps, 0 open questions; `docs/review-debt.md` has no row.
- `v0.38.7` renamed `Unreleased`: the next feature writes under it again.

## What this session did

- **0263, the lenses are measured** (`v0.38.6`): `odachi review probe` asks the four shipping prompts
  about a fixture with five planted defects and a clean arm, and scores found, low, missed and noise against
  a written key. Answers live in `.git/odachi/probe/`, never in the review store. First run: 5 of 5 planted
  defects found, 10 findings on the clean arm. Its own review found five defects -the clean arm was never
  measured, the fixture said it was an exam, `causal: false` counted twice, runs overwrote each other, the
  summary printed twice- and every one was paid.
- **The research** (`docs/review-research.md`) is the frame everything else sits in: six sources with sample
  sizes, eight failure classes, the coverage matrix, and the numbers from this clone.
- **0264, the clean arm pays** (`v0.38.7`): the ten findings the lenses named on the arm with nothing planted
  are fixed in the code -404 only for `sql.ErrNoRows`, the remote status checked, read and write errors
  propagated, a client timeout, a bounded body, a negated limit clamped, handler and cache tests- and the
  five planted defects are still found. The manifest and the seeded arm were not touched, and the fence now
  protects the manifest by path.

## The questions that wait

- **One run is one draw.** The same clean code, same model, three runs: 10 findings, then 1, then 4, with a
  different one each time. Precision is not measured until the probe repeats: `--runs N`, a median and the
  spread per lens. That is the next feature on the instrument, and it is the reason 0264 did not reach
  `clean 0`.
- **`start --force` regenerates `odd/fence.toml` and drops doctrine rules.** Three features in a row:
  `proof/**`, `internal/assets/review/*.md` and `.github/workflows/**` had to be restored by hand, and each
  time a different tool caught it (the archtest, then the review). The fix is that `start` merges the fence
  instead of rewriting it. This is the oldest open defect in the method.
- **`brief.md` still says `clean 0`** in "What must exist" and "Done when", and the reading says otherwise.
  The deviation is recorded in `odd/tasks/0264-brazo-limpio.md` and in PR #509. The human changes the
  criterion with `odachi ask --change "..." --req R1`, or leaves it as the aspiration it turned out to be.

## The limits to remember

- **The local gate does not see what the CI gate sees.** `check --run "go test ./..."` runs no `gofmt`, and
  `gofmt -l .` in CI covers `testdata` too. A file can be gofmt-clean on disk and dirty in the blob: check
  `git show HEAD:<path> | gofmt -l` before pushing a change under `testdata`.
- **The fixture's files are testdata, so nothing compiles them.** To prove the clean arm still builds and its
  tests pass, copy the four files to a temp module with a `go.mod` and run `go test` there.
- **A probe answer is not a verdict.** `.git/odachi/probe/` is separate on purpose; writing a fixture's answer
  into `reviews/` would make `check` believe a tree was judged.
- **The manifest is the exam, and the fence protects it.** A feature that "cleans" the clean arm could pass by
  deleting the key; the path rule is what makes that impossible.
- **`review --since <tree>` is the cheap correction lane**, and three rounds this session each reviewed a
  delta of a few lines. A merge waits for the delta to be approved.
- **The release chain**: branch `release/vX.Y.Z`, rename `## Unreleased` and nothing else, PR, merge, then an
  annotated tag on the merge commit. Renaming every `## Unreleased` in the file rewrites the prose too; edit
  the heading line only.
- One heavy gate per machine; `go test ./...` takes about 700 s and `check --run` writes the proof.

## The commands that work

```sh
odachi check --run "go test ./..." --timeout 1800
```

```sh
# the measurement: eight prompts (two arms, four lenses), runs nobody
odachi review probe
```

```sh
# the measurement with a reviewer; answers land in .git/odachi/probe/
odachi review probe --model opencode-go/deepseek-v4.1-flash \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

```sh
# the clean arm has no build of its own: prove it in a temp module
Copy-Item internal/probe/testdata/lenses/clean/*.go $env:TEMP\cc\; "module x`n`ngo 1.25" | Set-Content $env:TEMP\cc\go.mod; cd $env:TEMP\cc; go test ./...
```
