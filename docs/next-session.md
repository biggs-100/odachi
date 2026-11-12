# Next session

Written after `v0.39.1` shipped. Nothing is owed.

## Where the project is

- `v0.39.1` is published: annotated tag on `40d20d8` (`Merge pull request #539`), sixteen assets built and
  attested by the release workflow, and the binary installed on this machine reports `odachi v0.39.1`.
  `main` is that merge commit.
- `v0.39.0` shipped the instrument and `v0.39.1` pays the four defects the next two draws found in it.
- 274 features, 665 steps, 0 open questions; `docs/review-debt.md` holds no open row.
- The closing tree of 0276 carries `proof/0276-el-conteo-cuenta-respuestas.txt` and one verdict
  (`.git/odachi/reviews/bd559098….json`, two draws), written as a correction of `721b02cc`.
- `## Unreleased` does not exist: the next feature writes it again, above `## v0.39.1`.

## What shipped

- **0275, a verdict can be paid by more than one draw.** `--runs N` (1 to 10) asks each lens N times and
  files the union as one verdict. A defect is identified by its file, bytes, lens, place and severity
  -never by the sentence that describes it-, so two draws that name it pay one row and the first draw's
  words survive. A lens paid by one draw is brought to N later, paying only the missing answers: the
  union only adds, so drawing again cannot turn a `changes_requested` into an `approved`. The verdict
  carries the count per lens (`draws`, schema 5), `--show` prints it, and an absent field reads as one
  draw. A draw that does not answer leaves the ones that did and names the draw it lost.
- **0276, the count counts answers.** Found by reviewing the shipped candidate twice: a split prompt
  counted one draw per part (`2` for `--runs 1`, `6` for three), the handover printed `--runs` without
  `--exec` -a command the verb refuses-, and the refusal named a count that pays nothing. The review of
  that fix found one more real defect (the refusal named one more than the largest count, spending
  several while promising one) and one false CRITICAL, closed with a test and a recorded dismissal.
- **The instrument paid for itself on day one.** A second draw over the tree the first had approved found
  three defects, one of them a CRITICAL in the code that shipped hours before.

## The questions that wait

- **None open.** The next feature writes the new `## Unreleased`.
- **One wart, named and not fixed:** a verdict's `reviewer` field concatenates its merges, so a verdict
  written by six landings reads `odachi-reviewer + … + odachi-reviewer` six times. The identity the gate
  compares is `models`, which is right; the field is a list of names, and deduping it is a small change
  that nobody has asked for.

## The limits to remember

- **A finding about the record costs a full cycle**, and a finding about the *code* costs one too: 0276
  took three proofs (one failed on `TestEveryGoFileIsGofmtClean`) and two review rounds.
- **`--runs` multiplies the cost.** One group per model answers all its lenses in one call, so four lenses
  cost one call per draw; a split candidate costs one call per part per draw.
- **The console is not UTF-8** (0271), one tree one verdict (0268), a docs-only change owes no new proof.
- **PowerShell's `Add-Content` writes CRLF and the repo is LF**: after appending Go code, `gofmt -w` is
  not optional, and `TestEveryGoFileIsGofmtClean` is what says so.
- **One heavy gate per machine**: `go test ./...` takes about 700 s here and three to five minutes on CI;
  the first proof of a feature is rarely the last.

## The commands that work

```sh
odachi check --run "go test ./..." --timeout 1800
```

```sh
# several draws for the lens the risk line names
odachi review --runs 2 --base <feature base> --lens reliability --model opencode-go/deepseek-v4.1-flash \
  --exec 'C:\Users\USER\AppData\Local\Temp\odachi-rev.bat {model} {prompt} C:\Users\USER\AppData\Local\Temp\review.json'
```

```sh
# a correction measures only the delta and carries the predecessor's rows
odachi review --since <tree prefix> --runs 2 --lens reliability --exec '...'
```

```sh
# the release, after the merge it names: the tag, then the binary
git tag -a vX.Y.Z <merge commit> -m "vX.Y.Z: <the sentence the notes carry>" && git push origin vX.Y.Z
go build -trimpath -ldflags "-s -w -X main.version=vX.Y.Z" -o %LOCALAPPDATA%\odachi\bin\odachi.exe ./cmd/odachi
```
