# Next session

Written when the session closed, after `v0.39.1` and 0277 shipped. Nothing is owed.

## Where the project is

- `v0.39.1` is published: annotated tag on `40d20d8` (`Merge pull request #539`), sixteen assets built and
  attested by the release workflow, and the binary installed on this machine reports `odachi v0.39.1`.
- `main` is `ed95cb1` (`Merge pull request #541`), the note that records the field reading.
- 275 features, 666 steps, 0 open questions; `docs/review-debt.md` holds no open row.
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
- **0276, the count counts answers.** Found by reviewing the shipped candidate again: a split prompt
  counted one draw per part (`2` for `--runs 1`, `6` for three), the handover printed `--runs` without
  `--exec` -a command the verb refuses-, and the refusal named a count that pays nothing. The review of
  that fix found one more real defect and one false CRITICAL, the last one closed with a test and a
  recorded dismissal.
- **0277 records the field reading** in `docs/review-research.md` Part 5b: the four verdicts, the
  CRITICAL a full second read found where the first had approved, and what the numbers do not prove.

## The decision that waits

**What the review does next: run the deterministic tools.** Part 8's first item has not moved: a linter
with a real ruleset plus `govulncheck` in the gate, with the result as a floor row the review must
acknowledge. It is the one class of defect where a model should not be the judge, and the cheapest of
the remaining items. The other candidates, in the order the research ranks them: repeat the probe across
models (Part 8's third item is half kept), a standing blocker visible in the gate's own line, and
heterogeneous judges for full-tier trees -the format already supports one model per lens, and the
decision is what two opinions cost-.

**One wart, named and not fixed:** a verdict's `reviewer` field concatenates its merges, so a verdict
written by six landings reads `odachi-reviewer + … + odachi-reviewer` six times. The identity the gate
compares is `models`, which is right; the field is a list of names, and deduping it is a small change
that nobody has asked for.

## The limits to remember

- **A finding about the record costs a full cycle, and so does one about the code.** 0276 took four
  proofs (one failed on `TestEveryGoFileIsGofmtClean`) and two review rounds.
- **A correction is not a second reading.** The verdict that closed 0275 measured only the delta and
  found nothing; the next full read of the merged tree found a CRITICAL. Use `--runs` on the full
  candidate, not only on corrections.
- **The paperwork lane does not cover a close.** A feature that writes only documents still owes the
  suite on the tree that closes it; only the documents written *after* the close ride without one.
- **`--runs` multiplies the cost.** One group per model answers all its lenses in one call, so four
  lenses cost one call per draw; a split candidate costs one call per part per draw.
- **PowerShell's `Add-Content` writes CRLF and the repo is LF**: after appending Go code, `gofmt -w` is
  not optional, and `TestEveryGoFileIsGofmtClean` is what says so.
- **The console is not UTF-8** (0271), one tree one verdict (0268), a docs-only change owes no new proof
  *after* the close.
- **One heavy gate per machine**: `go test ./...` takes about 700 s here and three to five minutes on CI.

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
