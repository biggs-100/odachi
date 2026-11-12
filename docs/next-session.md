# Next session

Written after `v0.40.3`, 0282 and the release that published it.

## Where the project is

- `v0.40.3` is published: annotated tag on `5ec461e` (`Merge pull request #555`), sixteen
  assets built and attested, and the binary on this machine reports `odachi v0.40.3`.
- `main` is past `0282`'s merge (PR #554). 280 features, 684 steps, 0 open questions.
- **`main` is protected now:** every change enters through a pull request, and the release
  rename is no exception. A fence change is still two problems -the policy lands in `main`
  before the code that touches it, or CI reads the old rule from the base ref-, and #553
  was that first PR.
- The probe's table prints `seeded noise`, `seeded notes`, `clean noise` and
  `clean notes`, and the summary line carries the same four.
- `## Unreleased` exists again as the worklog of the next feature that changes a
  capability, and `scripts/release-notes.sh` refuses a tag whose section is missing.
- `odachi doctor` is clean: 15 checks, no warnings.

## What shipped

- **0282, the probe's reading separates a suggestion from an invented row** (`v0.40.3`).
  `probe.TallyOf` reads the findings nobody planted by what they claim: a suggestion is a
  note, a warning or above is noise. One function serves both arms. The table and the
  summary name both numbers on both arms, so a model that improves the fixture is no
  longer scored below one that stays quiet.
  The stored answers of 0281 were replayed through the new reading -not re-asked, not
  recounted by hand- and reproduce the old totals exactly (`noise 5 · clean 6` and
  `noise 1 · clean 0`): the split adds a meaning and not a measurement. Neither model
  invented a defect on the arm with nothing planted; deepseek's six findings there are
  six suggestions, and both models' seeded noise is made only of claims.
- **The changelog got its sections back.** Found while writing 0282's line: `v0.40.1` and
  `v0.40.2` were tagged with their entries still under `## Unreleased`, so
  `scripts/release-notes.sh` refused both tags and both release workflows failed (runs
  `43282780214` and `43293820137`). `v0.40.3`'s workflow published the notes.

## The decision that waits

**The research document still says the opposite of what it measured.** Part 8's item 3
carries the annotation "Repeating the probe across models is still open", and Part 5c is
that repetition, so the annotation is false. Correcting it is a document change, and it
should travel with the next thing that touches `docs/review-research.md`.

After that, in the research's order: a third model from another family for the probe, with
`--runs 5` to separate the model's opinion from the draw's; and heterogeneous judges for
full-tier trees. Separately, `govulncheck` is still not installed, and installing the tool
is the human's decision, not a step.

**One wart, named and not fixed:** a verdict's `reviewer` field concatenates its merges and
every correction's command, so one field reads as a paragraph of names.

## The limits to remember

- **A fence rule can name its own reopening.** `back when the probe's arm or its scoring
  changes for a reason this feature records` is a condition: when a feature is that reason,
  the rule is narrowed in its own PR and the brief moves in the same commit. The code that
  touches the freed path waits for that PR to land.
- **A measurement's vocabulary is part of the measurement.** Two numbers that answer
  different questions must not print as one: it is the difference between "this reviewer
  has an opinion" and "this reviewer invents".
- **A rebase does not move a proof or a verdict when the content does not change.** A tree
  id is content: the five commits were rebased onto the merge of their own opening and the
  proof and the verdict still named `0e44c60`.
- **A release renames `## Unreleased` before the tag, and the rename is a PR**
  (`release/vX.Y.Z`), with the tag on that merge. A tag with no section publishes no notes:
  the workflow fails and the assets still go up, which is how two releases shipped blank.
- **A verdict can stand behind more than one row.** Anything that decides about "the
  blocking row" must ask `BlockingAllAfter`: answering the first and stopping is how the
  second disappears.
- **The probe's answers live in `.git/odachi/probe/<fixture-tree>-<arm>-<lens>-<n>.json`**,
  and the name does not carry the model: the `model` field inside does. A replay deck is
  the way to re-read a measured run without paying for it again.
- **`golangci-lint run ./...` is ~1 minute on a warm cache** and runs on every review here.
- **PowerShell's `Set-Content`/`Add-Content` write CRLF and the repo is LF**: after editing
  Go code, `gofmt -w` is not optional, and a document edited with `-replace` needs its line
  endings checked.
- **The local hook can be skipped with `--no-verify`** for a single-document change; CI runs
  the whole suite with `--fence-from`. Never for code.
- **One heavy gate per machine**: `go test ./...` takes about 700 s here; `./cmd/...` alone
  about 330 s. CI's gate is faster than a cold local run.

## The commands that work

```sh
odachi check --run "go test ./..." --timeout 1800
```

```sh
# the review of a closing tree: the lenses the evidence asks for, two draws each
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
# re-read a measured run without paying for it: a deck that answers each arm and lens
odachi review probe --exec 'replay.bat {prompt}'
```

```sh
# the release: a branch that renames Unreleased, a PR, then the tag on its merge
git switch -c release/vX.Y.Z && git commit -am "release vX.Y.Z" && gh pr create --base main
git tag -a vX.Y.Z <merge commit> -m "vX.Y.Z: <the sentence the notes carry>" && git push origin vX.Y.Z
go build -trimpath -ldflags "-s -w -X main.version=vX.Y.Z" -o %LOCALAPPDATA%\odachi\bin\odachi.exe ./cmd/odachi
```
