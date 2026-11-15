# Next session

Written after 0309 and `v0.45.4`.

## Where the project is

- **0309 is open on `feat/0309-gentle-gaps`, unmerged and unreleased.** The gentle-shell audit closed: the
  attribution `tui-design` carries is pinned by a test, `ui-craft` gained the two pages nobody owned
  (`accessible-and-adaptive.md`, `grouping-and-emphasis.md`), and the PR policy question was answered -the
  free-form message with the `Risk:` line stays, the conventional regex does not come.
- **0308 is merged**: PR #624, on `main`, after 0307 (PR #623). The Tauri door carries the nine scaffold
  steps, the three conditions for not loading it, and points at `arch-project`.
- `v0.45.4` is published: annotated tag on `0ed7d2e` (the merge of release PR #621), sixteen assets
  signed and attested, and the binary on this machine was installed from that release and reports
  `odachi 0.45.4`. `v0.45.3` carried 0305; this one carries 0306.
- `main` is at that merge. 0306 landed in PR #620 and the release PR #621 renamed `Unreleased`. The
  branch of the feature was deleted on merge, as every branch should be.
- This file quotes the line the tool counts for the tree it describes:

  ```
  total: 307 features · 761 steps done of 761 · 0 open questions
  ```

  -copied from `odachi summary`; `odachi doctor` check 14 recomputes this line from the task documents
  and warns when the note quotes another count, so the number is never typed-.

## What waits for a decision

- **The home's `bootstrap` still carries the five references.** The session's plugin mirrored `arch-project`
  into `~/.agents/skills` when the folder appeared, but the first door there was not rewritten: the payload
  is canonical from here on, and `~/.agents/skills` is only refreshed by an install (`odachi setup`, or
  `pack --scope home`), which is the human's yes. Until then the home holds the same material twice.
- **The next feature is not chosen.** Before choosing, `docs/review-debt.md` is the reading list, and it
  is stale: it says "Last written: while closing 0271" and its `## Open` section says nothing is open,
  while the closes from 0291 to 0307 never wrote their rows there. Decide whether the file is rewritten
  from the store or retired; today it is a list nobody can trust.
- **The board look is unproven by a piece.** 0306 wrote the text an agent reads before producing one, and
  no board video was rendered or watched. The first piece made with it is the measurement that says
  whether the reference works.
- **One cosmetic debt, one line**: the `v0.45.1` section of `CHANGELOG.md` writes `drive` without the
  backticks. Fix it whenever the file is touched; it does not justify a pull request of its own.
- **`Unreleased` is empty** and the next feature writes under it, re-creating the heading.

- **`schema.Current` is 9**: the verdict carries `corrects`, the refutation carries `declared_at`, and
  `agents.toml` and `odd/review.toml` share the counter. A binary at `v0.43.0` or older refuses a
  verdict written by this one and says "upgrade".
- **The ratchet covers three packages through declared builders**, and all three declare every builder
  they own.
- **The gate's reasons are declared once**: one table judges the state, the words and the continuation,
  and a test reads the `Reason…` constants from `gate.go` with `go/ast`.
- **The text a session pays for is pinned by bytes** in `internal/archtest/bytes.baseline`, failing in
  both directions. The skills themselves are pinned by their trigger, their sections and the 60-line
  body; their bytes are not in that baseline.
- **A correction remembers its predecessor**, and a third round needs the human's reason.
- **A guard is proven by breaking it**: `check --run --mutate <file> --remove "<string>" --expect
  "<text>"` takes the string out in a copy, runs the command there, and requires a failure that names
  the expected text.
- The fence keeps its six active rules and no template: a non-goal that names no path stays in
  `brief.md`.
- `odachi doctor` is 15 checks and this note is what closed its only warning.

## What shipped

- **0309, patch (unreleased).** The gentle-shell audit closed: `tui-design`'s attribution is pinned by a
  test -it was never missing, it lives in the body because the installer copies `SKILL.md`, `references/`
  and `assets/` only-, and `ui-craft` gained `accessible-and-adaptive.md` and `grouping-and-emphasis.md`,
  the two pages nothing owned.

- **0308, patch (unreleased).** `arch-tauri` gained `references/tauri-scaffold.md` -the nine steps of a
  Tauri v2 scaffold in the order that keeps the boundary-, its `## When` recovered the three conditions
  for not loading it, and `tauri-bridge.md` stopped sending the tree and the ADRs to `bootstrap`.

- **0307, patch (unreleased).** The catalog gained its second level: `arch-project` carries the triage
  by project kind, the twelve birth decisions, the tree, the guard recipes, the commands per stack, the
  UI contract, the fifteen lessons, the Go + HTMX page and the ADR template, and `bootstrap` routes to
  it with no reference of its own. The five files moved; the four pieces the retired skill carried and
  nothing here did entered with a row each.

- **0306, patch (v0.45.4).** `motion-graphics` gained two references and their routing: `the-board.md`
  (the hand-drawn look as a system: paper and ink, one accent per meaning, seeded rough strokes, draw-on
  as a verb with weight by distance, the camera as the transition inside one argument, correction as the
  drama, and the traps this look sets) and `the-voice.md` (the human's recording as the clock: the two
  turns, the script they read, the word timestamps as an instrument, the cue on the stressed word, and
  the same alignment shape on HyperFrames and Remotion). `motion-principles.md` gained the pan/zoom row
  the transitions table was missing, `the-piece.md` points its times at the recording when there is one,
  the body gained one `## Never`, and the impact map counts eight references.

## 0309: what it settled, and what it leaves

- **A file-name diff reads a moved file as a lost one.** `tui-design/NOTICE.md` disappeared from the
  listing and the attribution was in the body all along, moved there on purpose. The audit's first finding
  was a false positive, and it was found by opening the file, not by trusting the diff. The test exists so
  the next rewrite cannot make the same disappearance real.
- **"Two references missing" was four sections.** Reading the source pages against the pages that exist
  showed the six states and the numbers already owned, and only the markup, the keyboard, the narrowing
  and the grouping unowned.
- **Still unported, and named:** `issue-creation` and `rdd-defect-workflow` have no counterpart in this
  catalog. `judgment-day`, `skill-registry` and `gentle-ai` are absorbed by the review quorum, the
  generated index and `drive`.

## 0308: what it settled, and what it leaves

- **A port is verified by a diff, not by memory.** The Tauri files were all present and the references
  were adapted cleanly, so the four-file gap of 0304 did not repeat; what the diff found instead was
  content: an order that lived in one section of the retired body and nowhere else, and a stop-list that
  lost two of its three conditions on the way in.
- **A level decays quietly.** `arch-tauri` said "load `bootstrap` and nothing here applies" for a project
  with no Rust core, which was true until 0307 created the level between them. Nothing failed; the
  sentence just aged. The test now pins the level, not only the rules.
- **What is still unowned:** `ANTIGRAVITY` -the human's own project name- is still written in
  `references/rust-guard-recipes.md`. It is harmless in this catalog and noise in a binary anyone
  installs.
- **A test that checks presence does not check a promise.** Two review rounds found the same shape of
  defect in the guards this feature wrote: the stop sentence was pinned word by word, so the wrong door
  passed, and the nine steps were pinned by existence, so any order passed. Both are pinned whole now
  -the sentence with its door, the steps by increasing position- and each was proven by breaking it.

## 0307: what it settled, and what it leaves

- **A level is a door, not a drawer.** The material for any stack existed and was unreachable: five
  references inside `bootstrap`, behind a trigger that asks whether a project is starting in odachi. The
  door is what changed; the text mostly moved.
- **The port was never finished.** 0304 recorded five of the eight files and neither the tree, the
  command matrix, the ADR template nor the Go page; a "complete port" that nobody diffed against the
  retired folder was a sentence. The four entered here.
- **A stack page is not a stack door.** `go-htmx.md` lives inside `arch-project` because it is twenty
  lines of choices, not a boundary; `arch-tauri` is a door because it has a bridge, an ACL and three UI
  layers. The next stack with that depth earns `arch-<stack>`.

## 0306: what it settled, and what it leaves

- **A look is a reference, not a skill.** The trigger is the same situation the skill already answers,
  and a second skill would be a second owner of one rule. The look is the positive side of
  `the-ai-look.md`, and the catalog did not have it.
- **A refused answer is not a verdict.** The first review declared all seven paths and wrote rows for
  six; the gate refused it by name, kept the answer in `…refused-<hash>.json`, and wrote no verdict. The
  second run paid two draws and covered seven of seven. The coverage rule is a row per changed path per
  lens, and a declaration is not a row.
- **The changelog entry belongs to the change commit.** The entry landed after the close, the tree moved,
  and the proof ran twice, twelve minutes each. 0305 did it right: the entry travelled in `55640cc`, the
  same commit as the code.
- **A `## Done when` bullet is one physical line.** `plan --done` refuses to rewrite a document whose
  bullet wraps, and that refusal cost another full measurement. Write the criterion short.
- **A rule can leave the body when the body already carries it.** The 60-line ceiling had no room left,
  and the sentence "say what you did not render or verify" lives verbatim in `## Output`; the deletion
  was recorded as a craft decision instead of being silent.

## The limits to remember

- **The proof names the index tree** (`git write-tree`). The order that closes a feature: the change
  commit, then `plan --done` on top of it, then stage the records and measure, then commit the records.
  The mark names the change commit's tree; the proof measures the closing one.
- **A document commit after a close leaves the local `check` red**, on purpose. CI regenerates its own
  proof from a clean checkout.
- **The undo row answers four different things**: `noticed` with the symbol the compiler named, `the copy
  did not build` for a failure that names nothing the change added, `not owed: the change adds no
  production behavior`, and `missed` for a suite that stayed green.
- **Text with backticks is not edited with a double-quoted shell string.** PowerShell reads `` `t `` and
  `` `b `` as escapes. Edit with the file tool or single quotes, and read the result back.
- **A guard that passes the first time proves nothing.** Verify every new guard by mutation.
- **Evidence answers only its own question.** Any stored answer keyed by tree needs the question in the
  key.
- **A refused review answer is kept, never edited.** A verdict is not edited to pass the gate.
- **A pin is exact, not a ceiling**, and it fails in both directions.
- **`## Unreleased` has to be re-created after every release**, and the release is **minor when the
  stored shape moves**, not when the diff is large.
- **A release is a tag on the merge of the release pull request.** Install from the release; a local
  `go build` is not published bytes. A version is never moved or reused.
- **`windows-latest` is the slow gate**: three to eleven minutes against one to four for the others.
  Merging with it pending is the shortcut the gate exists to stop.
- **PowerShell's `Add-Content` writes CRLF and a trailing newline**: `internal/archtest` fails a file
  written that way in the working copy.
- **The fence is read from the base**: a policy change lands first, in its own pull request.
- **One heavy gate per machine**: `go test ./...` takes about twelve minutes here, and `check --run` runs
  it twice when the undo row can revert production files.

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
