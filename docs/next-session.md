# Next session

Written when the session closed, after `v0.40.2` and 0280 shipped. Nothing is owed.

## Where the project is

- `v0.40.2` is published: annotated tag on `713d6ab` (`Merge pull request #549`), sixteen assets
  built and attested, and the binary on this machine reports `odachi v0.40.2`.
- `main` is `713d6ab`. 278 features, 679 steps, 0 open questions.
- The closing tree of 0280 carries `proof/0280-un-cierre-que.txt` and one verdict
  (`.git/odachi/reviews/12961c63….json`), a correction that approved the fix with the answered row
  carried as `prior`.
- **The floor runs on every review**: `lint` at `CRITICAL`, pinned in `.golangci.yml`, approved in
  this clone.
- `## Unreleased` does not exist: the next feature writes it again, above `## v0.40.2`.

## What shipped

- **0280, a close that requests changes now has a written way out.** The gate prints the row it
  stands behind and the row's fingerprint; `--show` prints the same identity;
  `odachi review waive --row <fingerprint> --because "<why>"` records the decision in the task
  document with the reason, who signed it and when. The close then goes through and its line keeps
  saying both halves. The decision answers one row: another tree, changed bytes or a new finding ask
  again.
- **The review of that tree found the defect the feature prevents.** `BlockingAfter` returned one
  row, so waiving it let the close through without examining the rest of the verdict;
  `BlockingAllAfter` answers every row now, the refusal counts the rest and lists the fingerprints
  already answered, and two cases pin it.
- This is the first feature closed since the floor existed, and the first whose verdict was
  `changes_requested` before the fix.

## The decision that waits

**What the review does next is the probe, not the gate.** Part 8's standing-blocker item is now
kept. What is left in the research's order: repeat the probe across models (the fixture is one
model's reading, and Part 5b is one model's field measurement), heterogeneous judges for full-tier
trees, and a clean arm that earns its name. Separately, `govulncheck` is still not installed, and
the check table is ready for it: installing the tool is the decision.

**Two warts, named and not fixed:** a verdict's `reviewer` field concatenates its merges and every
correction's command, so one field reads as a paragraph of names; and the waiver record is prose in
the history —it ports to a keyed record the day something other than the gate reads it—.

## The limits to remember

- **A fence change is two problems.** The base ref is the policy: touching a protected path needs
  the rule already there, and loosening a rule needs `brief.md` to move in the same change. Order:
  the rule and the brief first, the code after.
- **A rebase moves the proof and the verdict**, because both name content. Re-run `check --run`, and
  for a document-only delta use `review --since <tree> --because "<why>"`.
- **A verdict can stand behind more than one row.** Anything that decides about "the blocking row"
  has to ask `BlockingAllAfter`: answering the first and stopping is how the second disappears.
- **`golangci-lint run ./...` is ~1 minute on a warm cache** and runs on every review here. A
  teammate without the binary gets a `CRITICAL` row saying the command could not run.
- **The prompt of a feature this size is ~145 KB** against `max_prompt_kb = 120`, so it splits in
  two parts and a split costs one call per part per draw.
- **PowerShell's `Set-Content`/`Add-Content` write CRLF and the repo is LF**: after editing Go code,
  `gofmt -w` is not optional.
- **The local hook can be skipped with `--no-verify`**, and the one place it is right is a change
  that is a single document: CI runs the whole suite on it. Never for code.
- **A release renames `## Unreleased` to `## vX.Y.Z - <date>`** before the tag.
- **The console is not UTF-8** (0271), one tree one verdict (0268), a docs-only change owes no new
  proof *after* the close.
- **One heavy gate per machine**: `go test ./...` takes about 700 s here and three to five minutes on
  CI; `./cmd/...` alone is about 330 s.

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
# the release, after the merge it names: rename Unreleased, then the tag, then the binary
git tag -a vX.Y.Z <merge commit> -m "vX.Y.Z: <the sentence the notes carry>" && git push origin vX.Y.Z
go build -trimpath -ldflags "-s -w -X main.version=vX.Y.Z" -o %LOCALAPPDATA%\odachi\bin\odachi.exe ./cmd/odachi
```
