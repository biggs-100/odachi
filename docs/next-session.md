# Next session

Written when the session closed, after `v0.40.2`, 0280 and the measurement of 0281. Nothing is owed.

## Where the project is

- `v0.40.2` is published: annotated tag on `713d6ab` (`Merge pull request #549`), sixteen assets built
  and attested, and the binary on this machine reports `odachi v0.40.2`.
- `main` is past `0281`'s merge (PR #551). 279 features, 680 steps, 0 open questions.
- **0281 shipped no release on purpose**: it measured and wrote a document, and a measurement is not
  a capability (`docs/releases.md`). The reading is `docs/review-research.md` Part 5c.
- The floor runs on every review: `lint` at `CRITICAL`, pinned in `.golangci.yml`, approved in this
  clone. A close that requests changes has a written way out: `review waive --row <fingerprint>`.
- `## Unreleased` does not exist: the next feature that changes a capability writes it above
  `## v0.40.2`.

## What shipped

- **0280, a close that requests changes says so and only a written waiver opens it** (`v0.40.2`).
  The gate prints the blocking row and its fingerprint; `--show` prints the same identity;
  `review waive --row` records the decision in the task document with reason, who and when; the
  close then passes and its line keeps saying both halves. Its own review found the defect the
  feature prevents -`BlockingAfter` returned one row, so waiving it skipped the rest- and
  `BlockingAllAfter` answers every row now.
- **0281, the probe measured across two models.** `deepseek-v4.1-flash` found all five planted
  defects and answered the clean arm with six suggestions; `muse-spark-1.3` found three of five and
  said nothing on the clean arm. Read one by one, deepseek's six are real improvements to the
  fixture, all `SUGGESTION`, none invented.

## The decision that waits

**The `clean` column lies, and fixing how it is read is the next change.** A `SUGGESTION` on the arm
with nothing planted counts the same as a fabricated `CRITICAL`, so a model that suggests
improvements is scored as noisy and a model that says nothing is scored as precise. Split the reading
by severity (a suggestion is a note, a fabricated blocking row is the noise the metric exists for),
and the cross-model numbers become comparable. The fixture itself is not the problem and it is
fenced anyway.

After that, in the research's order: heterogeneous judges for full-tier trees, and a third model from
another family for the probe, with `--runs 5` to separate the model's opinion from the draw's.
Separately, `govulncheck` is still not installed, and the check table is ready for it: installing the
tool is the decision, not a step.

**One wart, named and not fixed:** a verdict's `reviewer` field concatenates its merges and every
correction's command, so one field reads as a paragraph of names.

## The limits to remember

- **A fence change is two problems.** The base ref is the policy: touching a protected path needs the
  rule already there, and loosening a rule needs `brief.md` to move in the same change.
- **A rebase moves the proof and the verdict**, because both name content. Re-run `check --run`, and
  for a document-only delta use `review --since <tree> --because "<why>"`.
- **A verdict can stand behind more than one row.** Anything that decides about "the blocking row"
  must ask `BlockingAllAfter`: answering the first and stopping is how the second disappears.
- **A measurement is not a release.** A light feature that only writes a document still owes the
  suite at its close, and needs no verdict because the evidence asks for no lens.
- **The probe's answers live in `.git/odachi/probe/<fixture-tree>-<arm>-<lens>-<n>.json`**, and the
  summary's numbers are that run's answers. Read them one by one before trusting a count.
- **`golangci-lint run ./...` is ~1 minute on a warm cache** and runs on every review here.
- **The prompt of a feature this size is ~145 KB** against `max_prompt_kb = 120`: it splits in two
  parts, and a split costs one call per part per draw.
- **PowerShell's `Set-Content`/`Add-Content` write CRLF and the repo is LF**: after editing Go code,
  `gofmt -w` is not optional.
- **The local hook can be skipped with `--no-verify`** for a single-document change; CI runs the
  whole suite with `--fence-from`. Never for code.
- **A release renames `## Unreleased` to `## vX.Y.Z - <date>`** before the tag.
- **One heavy gate per machine**: `go test ./...` takes about 700 s here; `./cmd/...` alone about
  330 s.

## The commands that work

```sh
odachi check --run "go test ./..." --timeout 1800
```

```sh
# the review of a closing tree: four lenses, two draws each, the candidate split in two parts
odachi review --runs 2 --base <feature base> --lens reliability --model opencode-go/deepseek-v4.1-flash \
  --exec 'C:\Users\USER\AppData\Local\Temp\odachi-rev.bat {model} {prompt} C:\Users\USER\AppData\Local\Temp\review.json'
```

```sh
# a correction measures only the delta and carries the predecessor's rows
odachi review --since <tree prefix> --runs 2 --lens reliability --because "<why>" --exec '...'
```

```sh
# when the review says no: the gate prints the fingerprint, the decision is written down
odachi review --show
odachi review waive --row <fingerprint> --because "<why the tree ships with this row open>"
```

```sh
# the probe, one model at a time: four lenses on two arms, eight calls per draw
odachi review probe --model <provider/model> --exec '<command> {model} {prompt} <out.json>'
```

```sh
# the release, after the merge it names: rename Unreleased, then the tag, then the binary
git tag -a vX.Y.Z <merge commit> -m "vX.Y.Z: <the sentence the notes carry>" && git push origin vX.Y.Z
go build -trimpath -ldflags "-s -w -X main.version=vX.Y.Z" -o %LOCALAPPDATA%\odachi\bin\odachi.exe ./cmd/odachi
```
