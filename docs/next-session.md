# Next session

Written after `v0.39.0` shipped. Nothing is owed.

## Where the project is

- `v0.39.0` is published: annotated tag on `adf4f69` (`Merge pull request #536`), assets built and
  attested by the release workflow -which ran the suite before goreleaser-, and the binary installed on
  this machine reports `odachi v0.39.0`. `main` is that merge commit.
- 273 features, 661 steps, 0 open questions; `docs/review-debt.md` holds no open row.
- The closing tree carries `proof/0275-lo-que-una-tirada-no-ve.txt` and one verdict
  (`.git/odachi/reviews/bbe8d2ff….json`, four lenses), written as a correction of `dfca16e5`.
- `## Unreleased` does not exist: the next feature writes it again, above `## v0.39.0`.

## What shipped

- **0275, a verdict can be paid by more than one draw.** `--runs N` (1 to 10) asks each lens N times and
  files the union as one verdict. A defect is identified by its file, bytes, lens, place and severity
  -never by the sentence that describes it-, so two draws that name it pay one row and the first draw's
  words survive. A lens paid by one draw is brought to N by a later run that pays only the missing
  answers: the union only adds, so drawing again can never turn a `changes_requested` into an
  `approved`. A draw that does not answer leaves the ones that did, names the draw it lost and the
  top-up that pays the rest. The verdict carries the count per lens (`draws`, schema 5), `--show` prints
  it, and an absent field reads as one draw so every verdict already in a store keeps its meaning.
- **The review found the defect the design had.** A lens already paid to N was asked again whenever it
  shared a call with a lens that still owed one -a group answers all its lenses in one call- and the
  file ended at N+1, past the count the run was given. Fixed in `06d2f21` by nesting the rounds: draw
  two carries only the lenses that owe a second answer. The fix changed the bytes the row named, so the
  correction retired the row by itself.
- **The release is a minor, not the patch the brief said.** `docs/releases.md` calls a stored-record
  change a minor; the brief was not rewritten, and the divergence is recorded in the changelog entry.

## The questions that wait

- **None open.** The next feature writes the new `## Unreleased`.
- **The instrument is in the tool and not yet measured on a real candidate.** The probe measured
  variance on a fixture -10, 1 and 4 findings on the same clean code-; the next measurement is the same
  tree reviewed with `--runs 1` and `--runs 3`, and it is evidence, not a feature.

## The limits to remember

- **The verdict that closed 0275 has schema 4 and no `draws`.** It was written by `v0.38.14`, before the
  field existed: the first record that carries the count is the next review of any tree.
- **`--runs` multiplies the cost.** Four lenses times N calls, each up to the 600 s `--exec` timeout; the
  ceiling of ten is the probe's own, and past it an experiment is designed somewhere else.
- **A round carries only the lenses that owe it.** The `draws` count is exact only because of that:
  asking a group without it asks a finished lens again and counts it.
- **`CHANGELOG.md` carries `Unreleased` twice** -the heading and the prose about it- so the rename is
  line-based, and writing the file with Windows line endings rewrites it; the hook normalizes to LF.
- **One heavy gate per machine**: `go test ./...` takes about 700 s here and three to four minutes on CI.
- The console is not UTF-8 (0271), one tree one verdict and a wider base covers a later feature (0268),
  and a docs-only change owes no new proof.

## The commands that work

```sh
odachi check --run "go test ./..." --timeout 1800
```

```sh
# a verdict paid by several draws, for the lens the risk line names
odachi review --runs 3 --base <feature base> --lens reliability --model opencode-go/deepseek-v4.1-flash \
  --exec 'C:\Users\USER\AppData\Local\Temp\odachi-rev.bat {model} {prompt} C:\Users\USER\AppData\Local\Temp\review.json'
```

```sh
# a correction measures only the delta and carries the predecessor's rows
odachi review --since <tree prefix> --lens reliability --exec '...'
```

```sh
# the release, after the merge it names: the tag, then the binary
git tag -a vX.Y.Z <merge commit> -m "vX.Y.Z: <the sentence the notes carry>" && git push origin vX.Y.Z
go build -trimpath -ldflags "-s -w -X main.version=vX.Y.Z" -o %LOCALAPPDATA%\odachi\bin\odachi.exe ./cmd/odachi
```
