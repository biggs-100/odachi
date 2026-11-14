# Next session

Written after 0291 to 0298, the three releases that carried them, and the study of gentle-shell that
produced 0295 and 0297.

## Where the project is

- `v0.44.1` is published: annotated tag on `5936639` (the merge of release PR #595), sixteen assets
  signed and attested, and the binary on this machine was installed from that release and reports
  `odachi 0.44.1`. Three releases today: `v0.43.0` carried 0291 to 0293, `v0.44.0` carried 0296, and
  `v0.44.1` carried 0297.
- The branch `feat/0298-una-medicion` carries five commits over `1733e6c`, the merge this feature
  started from, and it quotes the line the tool counts for the tree it closes:

  ```
  total: 296 features · 730 steps done of 730 · 0 open questions
  ```

  -copied from `odachi summary`; `odachi doctor` check 14 recomputes this line from the task documents
  and warns when the note quotes another count, so the number is never typed-.
- **`schema.Current` is 9**: the verdict carries `corrects`, the refutation carries `declared_at`, and
  `agents.toml` and `odd/review.toml` share the counter. A binary at `v0.43.0` or older refuses a
  verdict written by this one and says "upgrade".
- **The ratchet covers three packages through declared builders**, and all three declare every builder
  they own.
- **The gate's reasons are declared once**: one table judges the state, the words and the continuation,
  and a test reads the `Reason…` constants from `gate.go` with `go/ast`.
- **The text a session pays for is pinned by bytes** in `internal/archtest/bytes.baseline`, failing in
  both directions.
- **A correction remembers its predecessor**, and a third round needs the human's reason.
- **A guard is proven by breaking it**: `check --run --mutate <file> --remove "<string>" --expect
  "<text>"` takes the string out in a copy, runs the command there, and requires a failure that names
  the expected text. The evidence goes to `proof/<feature>.mutated.txt`.
- The fence keeps its six active rules; none protects `internal/domain/gate` since 0290, and the
  measurements this project published are protected since 0298.
- `odachi doctor` was 15 checks with no problems at `v0.44.1`; since 0298 it also asks whether the
  numbers in this file came from the tool.

## What shipped

- **0291, minor (v0.43.0).** A refutation's `at` is the instant the clone sealed it, the refuter's
  declared hour lives in `declared_at`, and `--show` prints one line per witness. The refutation
  declares its schema. The first review named three real rows and all three were fixed.
- **0292, patch (v0.43.0).** `refusal.Walk` takes the builders a package declares; `review`'s refusals
  entered the mechanism, and the dead end it could then see pays.
- **0293, patch (v0.43.0).** The gate's reasons are one table; a reason in the source with no row fails
  the suite.
- **0294, no entry.** `check` declares its `refused` helper too: the last asymmetry closed.
- **0295, no entry.** `internal/domain/budget` pins the bytes of what an agent reads, in both
  directions, with the numbers measured by the test.
- **0296, minor (v0.44.0).** A verdict remembers the tree it corrects, the chain can be walked, and a
  third correction is refused with the flag that continues it. The walk stops at a link the store no
  longer holds.
- **0297, patch (v0.44.1).** A guard is proven by breaking it, and the row answers only the question it
  was asked. The first verdict named five rows and all five were answered.
- **0298, no release yet.** A measurement has a shape: the new `measure` skill carries the criterion
  before the run, the instrument, the verdict table, the excluded runs and what the data does not show.
  And the handover stopped claiming numbers: it quotes the totals line `odachi summary` prints, check 14
  recomputes it from the task documents, and a note with no such line is a warning instead of a pass.

## 0298: what it settled, and what it leaves

**The shape lives in a skill**, because the agent is who takes the number: `skills/measure/SKILL.md`
names the five sections, the two sentences that keep a byte count from becoming a cost claim, and where
the record goes -a document under `docs/`, named by its question-. The measurements this project already
published are not rewritten; the fence protects them from here on.

**A claim is evidence or it is a sentence.** `docs/next-session.md` carries the line the tool counts, on
its own line, prefixed `total: `; `odachi doctor` reads every such line and asks whether one of them is
the line it computes. Two limits came with it: the line moves every time a step closes, so the note is a
tree's claim and not a period's, and the note that lags its own close warns until it is rewritten; and the
step count still says "1 steps done of 1", because this feature moved the renderer and an amount of
wording is its own change.

## The limits to remember

- **The proof names the index tree** (`git write-tree`): a close writes its records, stages them,
  measures, and only then commits. The commit hook refuses any commit after the task doc closes until a
  proof exists for that tree.
- **A document commit after a close leaves the local `check` red**, on purpose. CI regenerates its own
  proof from a clean checkout.
- **The undo row answers three different things**: `the copy did not build`, `not owed: the change adds
  no production behavior`, and a stored verdict for a suite that stayed green.
- **A guard that passes the first time proves nothing.** Every new guard this session was verified by
  mutation, and two passed by luck before their fix.
- **Evidence answers only its own question.** 0297's first verdict found a false green because the stored
  row matched by tree alone: a new question reused old evidence. Any stored answer keyed by tree needs
  the question in the key.
- **A refused review answer is kept, never edited.** When the reviewer misses a row, the tool writes
  `…refused-<hash>.json` and names how to attach a corrected answer; a verdict is not edited to pass the
  gate.
- **A pin is exact, not a ceiling**, and it fails in both directions.
- **A walk that counts stops at a hole**: a pruned store forgets trees, and counting past a missing link
  would refuse on a fact nobody has.
- **`## Unreleased` has to be re-created after every release**, and the release is **minor when the
  stored shape moves**, not when the diff is large.
- **Measure the debt before opening the feature.** The gate's note said ten reasons and two pull
  requests; the tree had eleven and needed one.
- **A release is a tag on the merge of the release pull request.** Install from the release; a local
  `go build` is not published bytes. A version is never moved or reused.
- **`windows-latest` is the slow gate**: seven to eleven minutes against one to three for the others.
  Merging with it pending is the shortcut the gate exists to stop.
- **PowerShell's `Add-Content` writes CRLF and a trailing newline**: `internal/archtest` fails a file
  written that way in the working copy. `gofmt -w` after writing Go through a here-string, and write a
  here-string with `Set-Content` to a variable, not to `Add-Content` with a bare tab.
- **The fence is read from the base**: a policy change lands first, in its own pull request.
- **One heavy gate per machine**: `go test ./...` takes about 700 s, and `check --run` runs it twice when
  the undo row can revert the production files.

## The commands that work

```sh
odachi check --run "go test ./..." --timeout 1800
```

```sh
# prove a guard by breaking it: the string out, the failure has to name the text
odachi check --run "go test ./..." --mutate <file> --remove "<string>" --expect "<text>"
```

```sh
# the review of a closing tree: the lenses the evidence asks for, two draws each
odachi review --runs 2 --base <feature base> --exec 'C:\Users\USER\AppData\Local\Temp\odachi-rev.bat {model} {prompt} C:\Users\USER\AppData\Local\Temp\review.json'
```

```sh
# a correction measures the delta, carries the predecessor's rows, takes no --base
odachi review --since <tree prefix> --runs 2 --because "<why>" --exec '<the same command>'
```

```sh
# the release: a branch that renames Unreleased, a PR, then the tag on its merge
git switch -c release/vX.Y.Z && git commit -am "release vX.Y.Z" && gh pr create --base main
git tag -a vX.Y.Z <merge commit> -m "vX.Y.Z: <the sentence the notes carry>" && git push origin vX.Y.Z
irm https://github.com/biggs-100/odachi/releases/download/vX.Y.Z/install.ps1 | iex
odachi --version
```
