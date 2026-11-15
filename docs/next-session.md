# Next session

Written after 0291 to 0305, the ten releases that carried them, and the study of gentle-shell that
produced 0295 and 0297.

## Where the project is

- `v0.45.3` is published: annotated tag on `9273a87` (the merge of release PR #617), sixteen assets
  signed and attested, and the binary on this machine was installed from that release and reports
  `odachi 0.45.3`. Ten releases today: `v0.43.0` carried 0291 to 0293, `v0.44.0` carried 0296,
  `v0.44.1` carried 0297, `v0.44.2` carried 0298, `v0.44.3` carried 0299, `v0.44.4` carried 0300,
  `v0.45.0` carried 0301, `v0.45.1` carried 0302, `v0.45.2` carried 0303 and 0304, and `v0.45.3`
  carried 0305.
- `main` is at that merge. 0303 landed in PR #612, 0304 in PR #613 and 0305 in PR #616, and the release
  PR renamed `Unreleased` after them. This file quotes the line the tool counts for the tree it
  describes:

  ```
  total: 303 features · 751 steps done of 751 · 0 open questions
  ```

  -copied from `odachi summary`; `odachi doctor` check 14 recomputes this line from the task documents
  and warns when the note quotes another count, so the number is never typed-.

## What waits for a decision

- **The next feature is not chosen.** The three debts the last session left are closed: the store is
  clean (1.9 MB, nothing left to free), the count says "1 step done of 1", and the changelog is current.
  What is not decided is what comes next; `docs/review-debt.md` is the list to read before choosing.
- **One cosmetic debt, one line**: the v0.45.1 section of `CHANGELOG.md` writes `bootstrap` and `drive`
  without the backticks around `drive`. The published release notes already have them. Fix it whenever
  the file is touched; it does not justify a pull request of its own.
- **`Unreleased` is empty** and the next feature writes under it, re-creating the heading.

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
- The fence keeps its six active rules and, since 0299, no template: a non-goal that names no path stays
  in `brief.md`, which already carries it verbatim, and `start` counts those instead of copying them.
  None protects `internal/domain/gate` since 0290; the measurements this project published are protected
  since 0298.
- `odachi doctor` is 15 checks with no problems at `v0.44.3`, and since 0298 check 14 asks whether the
  numbers in this file came from the tool instead of only whether it names the newest feature.

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
- **0298, patch (v0.44.2).** A measurement has a shape: the new `measure` skill carries the criterion
  before the run, the instrument, the verdict table, the excluded runs and what the data does not show.
  And the handover stopped claiming numbers: it quotes the totals line `odachi summary` prints, check 14
  recomputes it from the task documents, and a note with no such line is a warning instead of a pass.
- **0299, patch (v0.44.3).** A fence rule protects files, and the rest of the sentence stays in the
  brief. `odd/fence.toml` went from 939 lines to 29 -ninety-one templates to none, the same six rules-,
  and opening a feature says three lines instead of ninety-four.
- **0300, patch (v0.44.4).** A copy that does not build names what the change took away: the row that
  reverts a change answers `noticed` with the symbol the compiler named, instead of "could not be
  judged" -which was the answer for the two features before it-.
- **0301, minor (v0.45.0).** The risk a commit declares asks for its lens: the table `commits`
  published is the table the gate reads, a declared risk joins the demand per path and at the lens
  level, and a verdict without that lens does not close.

- **0302, patch (v0.45.1).** The catalog carries a Tauri v2 foundation: the `tauri-foundation` skill
  names the boundary a desktop project keeps crossing, with its bridge templates, purity guards,
  workspace layout and three-layer UI verification.
- **0303, patch (v0.45.2).** The skill is called `arch-tauri`: the name says what it is and what it is
  for, and leaves room for the family. The folder, the test, the index row, the impact map, the byte pin
  and the home copy moved with it.
- **0304, patch (v0.45.2).** The foundation enters through the door that already exists: the triage, the
  birth decisions, the guard recipes, the UI contract and the fifteen lessons are `bootstrap`'s
  references, and the rules `drive`, `tests` and `commits` own are delegated instead of repeated.
- **0305, patch (v0.45.3).** The count says one step when there is one. The totals line pluralizes the
  step count now, and the sentence the handover quotes and the doctor recomputes come from one renderer.

## 0303 and 0304: the catalog has three levels, and each one owns its question

1. **`bootstrap`** - how a project starts in odachi: the interview, the artifacts, the first proof.
2. **`bootstrap/references/`** - the architecture of any stack: what kind of project this is, the twelve
   decisions before the first file, the boundary guard, the UI contract, and the lessons that paid for
   them. Its decision table routes each file; nothing here is a second door.
3. **`arch-tauri`** - the boundary of one stack: the domain crate never imports Tauri, commands are thin,
   the contract is generated, the UI is verified in three layers. A project with no Rust core never
   loads it.

**A copy that carries half a skill is a copy that lies.** `bootstrap` grew five references, and the
doctor's fixture seeded only its `SKILL.md`, so check 9 read the installed copy as drifted. The closing
suite caught it, the fixture now copies the whole folder, and the helper that returned only the body is
gone. It is the same defect the feature fixes at the human level.

## 0302: what it settled, and what it leaves

**A port is not a copy.** Two of its four pieces needed a new home: the guard recipes were an asset and
are read, so they became a reference; and the three links into the other skill now say what they mean
here -layout and ADRs belong to `bootstrap` and `drive`, and the UI contract is stated instead of
pointed at a file this catalog does not have-.

**The catalog has rules a folder in the home cannot satisfy.** A skill here needs at most 60 body lines,
a `## Never`, a `## Output`, and every reference routed from a row of `## Decision gates`; its row moves
the byte pin, and the impact map refuses a folder it does not name. The port failed three of those on
the first try, which is the checks doing their job.

**The originals moved, they were not deleted.** `ai-project-foundation` and `tauri-project-foundation`
now live under `~/.agents/skills-retired/`, out of the path the clients read: the payload is the
canonical copy from here on. The foundation material still waits as its own feature -its triage table,
birth checklist, guard recipes, UI doctrine and fifteen lessons, as references of `bootstrap`, with the
rules `drive`, `tests` and `commits` already own pointing at their owner-.

## 0301: what it settled, and what it leaves

**A declaration that changes nothing is a sentence.** `Risk: contract` named a lens and a verifier in
the skill and nothing read it: 0299 declared `contract` on three commits and its verdict closed with
`required: none`. The mapping now lives in `internal/domain/review/risk.go`, the gate adds those lenses
to what the paths ask -never instead- and the refusal names `odachi review --lens <lens>`. A change that
asks for no lens at all keeps asking nothing: a document is not where a lens looks.

**A rule about the gate takes effect with the release that carries it.** 0300 and 0301 both closed with
the installed binary, which predates their change, so neither answered to its own rule: 0300's undo row
had to be regenerated with `go run`, and 0301 closed with the lens its paths asked for and not the one
its four `Risk: contract` lines asked for. Expect the first feature after a release to be the first one
the new rule judges.

**A record can lie by one line.** The T4 mark of 0301 was staged before the proof and the branch was
pushed without it, so `main` held a feature whose changelog said released and whose plan said one step
open. Reading `main` after the merge is what caught it; the fix rides in this pull request.

## 0300: what it settled, and what it leaves

**A red is evidence when it names the change.** The row reverts production files and keeps the tests,
so a change whose own tests call what it added leaves a copy that cannot compile. The names now come
from the diff's added side, and only the compiler's diagnostics are read: the first version searched
the whole output and this feature's own tree proved it too wide -it named a test that had failed in
another package-. What blocks did not move: only a suite that stayed green.

**A feature that changes the checker needs its own binary for its own evidence.** The close ran with
the installed release, which predates the change, so its own undo row still said "could not be judged".
The evidence was regenerated with `go run ./cmd/odachi check --run "go test ./..."` -the binary at the
tree- and the stored file is what the gate reads afterwards. Expect this for any feature whose behavior
is the tool's behavior.

## 0299: what it settled, and what it leaves

**A file that accumulates because nobody counts it.** The fence re-rendered one commented template per
phrase the brief had ever carried, so it grew with every feature and every run said one warning per
template. The fix was not a cleanup: the renderer stopped writing what the brief owns, and the next
`start` healed the file. The same shape appeared twice more in one session -407 stale branch references
in the remote and a handover claiming a count from six features ago- and the answer was the same:
measure it, then name it.

**A guard is proven by breaking it, and a mutation that stays green is the finding.** The first mutation
of 0299's new test passed with the guarded sentence taken out: the test asked for the number and the file
and left the sentence that carries them free. The assertion was fixed, the second mutation failed naming
the expected text, and the correction kept its own commit.

**The habit**: `gh pr merge --delete-branch`. Each feature opens a branch and nothing closed it: the
remote held 407 stale references and the clone 232 merged branches, now one. A branch per feature that
nobody deletes is the same disease as the templates.

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
- **The undo row answers four different things**: `noticed` with the symbol the compiler named (`the copy did not build, and it names X`, since 0300), `the copy did not build` for a failure that names nothing the change added, `not owed: the change adds no production behavior`, and `missed` for a suite that stayed green.
- **Text with backticks is not edited with a double-quoted shell string.** PowerShell reads
  `` `t `` and `` `b `` as escapes: a replacement written that way left a tab where
  `` `tauri-foundation `` began and a backspace where `` `bootstrap `` did, and the corruption shipped
  in v0.45.1's notes. Edit with the file tool or single quotes, and read the result back.
- **A guard that passes the first time proves nothing.** Every new guard this session was verified by
  mutation, and two passed by luck before their fix.
- **A mark names a tree, and a tree nobody committed is a mark that lies.** `plan --done` before the
  commit recorded the opening tree for 0305; `undo --artifact` walked back the backups, the doc came
  back from HEAD, and the mark was written after the commit. Mark last.
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
