# Next session

Written after `v0.38.14` shipped. Nothing is owed.

## Where the project is

- `v0.38.14` is published: annotated tag on `d8c0d91` (`Merge pull request #533`), sixteen assets, and the
  binary installed on this machine reports `odachi v0.38.14`. `main` is that merge commit.
- The patch it names came in with #532: the gate refuses an artifact whose bytes are not UTF-8, one saved
  with a BOM, and one carrying the sequences a UTF-8 file leaves when a console reads it as CP850 (0274).
- 272 features, 656 steps, 0 open questions; `docs/review-debt.md` holds no open row.
- The closing tree carries `proof/0274-lo-que-la-compuerta-ve.txt` and one verdict
  (`.git/odachi/reviews/ab7974e0….json`, reliability, base `0a84a94`).
- `## Unreleased` does not exist: the next feature writes it again, above `## v0.38.14`.

## What shipped

- **0274, a record that does not decode is refused.** The brief corrupted on 0271 was valid UTF-8 -the bytes
  of one encoding read by another- so no decoder and no gate saw it. The gate now names the byte where a file
  stops decoding, refuses a BOM, and refuses the CP850 markers in the brief and the TOMLs, never in a task
  document: prose may have to name them.
- **A rule was planned and dropped.** Comparing every brief block with its task document's requirement
  looked like the non-heuristic detector; the tests showed it blocks valid states -a brief before its first
  document, a block whose document was removed- and the record accumulates. The craft note keeps why.

## The questions that wait

- **None open.** The next feature writes the new `## Unreleased`.
- **One approval is one draw**, and today's own feature is the example: a verdict approved a tree, and the
  next review found the record promising what the code did not do. The instrument (N draws, the union of
  what they found) has not been designed yet, and it is the next big thing.

## The limits to remember

- **A finding about the record costs a full cycle.** The review of 0274 found, twice, that the brief promised
  or described something the code did not do; each fix costs a proof (700 s) and a verdict. Write the
  brief's non-goals and decisions after the design is settled, or expect to pay twice.
- **The store's rows are keyed by bytes**: a WARNING whose claim a record edit fixed stays potent until the
  file it names changes (0265, 0271).
- **A run queued with zero jobs is GitHub's**: `gh run view <id> --json jobs` says `total=0`, and closing
  and reopening the pull request triggers a fresh run.
- **`CHANGELOG.md` carries `## Unreleased` twice** -the heading and the prose about it- so the rename is
  line-based.
- **The console is not UTF-8** (0271), **one tree, one verdict** and a wider base covers a later feature
  (0268), **a docs-only change owes no new proof**.
- **One heavy gate per machine**: `go test ./...` takes about 700 s and `check --run` writes the proof.

## The commands that work

```sh
odachi check --run "go test ./..." --timeout 1800
```

```sh
# a verdict for the tree, with the answer kept beside it
odachi review --base <feature base> --lens reliability --model opencode-go/deepseek-v4.1-flash \
  --exec 'C:\Users\USER\AppData\Local\Temp\odachi-rev.bat {model} {prompt} C:\Users\USER\AppData\Local\Temp\review.json'
```

```sh
# a correction takes the prefix the listing shows
odachi review --since bbf385ca --lens reliability --exec '...'
```

```sh
# the release, after the merge it names: heading only, an annotated tag, the binary rebuilt
go build -trimpath -ldflags "-s -w -X main.version=vX.Y.Z" -o %LOCALAPPDATA%\odachi\bin\odachi.exe ./cmd/odachi
```
