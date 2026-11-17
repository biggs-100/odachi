# Next session

Written after 0322 and `v0.48.5`.

## Where the project is

- **0322 is open on `feat/0322-lineas-normativas`, unreleased, patch, stacked on the unpushed 0321 branch.**
  Every normative line of the four lenses is pinned by its words in `internal/archtest/normative.baseline`, and a
  sweep names a rule lost or added in either direction. T1 and T2 are done; T3 closes with this note. Deliver 0321
  first: once 0321 is merged, the 0322 pull request shows only its own commits against `main`.
- **0321 is open on `feat/0321-reloj-de-silencio`, unreleased, patch.** `RunInput` measures first-output and idle
  silence for a reviewer; any byte re-arms the clock, and the proof path keeps only its total deadline.
- **0320 shipped in `v0.48.5`**: PR #646 merged on `c6a6ccb`, release PR #647 on `55a8816`, tag pushed, 16 signed
  and attested assets. The reviewer reads a frozen snapshot with bounded tools; the decision and its evidence are
  in `docs/reviewer-context-research.md`.
- **0319 is merged and released**: PR #643, and `v0.48.4` carries it (tag on `e5837e4`). A verdict that says
  `changes_requested` and carries no row now stops the close, and the door is written: `odachi review waive
  --because "<why>"` leaves the decision in the task document.
- **0318 is merged and released**: PR #641, carried by `v0.48.3`. The verification demand asks the feature's
  diff, so a close that only carries documents is still asked for the record its code owes.
- **0317 is merged and released**: PR #639, carried by `v0.48.2`. A verification a clean checkout cannot
  carry is named, and `--require-review` is what refuses it.
- **0316 is merged and released**: PR #637, carried by `v0.48.1`. Doctor's check 17 counts the distinct models
  that judge the lenses; on this repository it answers `1 of 4`.
- **0315 is merged and released**: PR #635, and `v0.48.0` carries it. A risk asks for a verifier, the record
  is sealed by the tree, and the gate names it.
- **`v0.48.4` is published and installed**: tag on `e5837e4`, sixteen assets signed and attested, and this
  machine runs `odachi 0.48.4`.
- `main` is at the merge of the release PR #644, and the branches were deleted on merge, as every branch
  should be.
- This file quotes the line the tool counts for the tree it describes:

  ```
  total: 320 features · 804 steps done of 804 · 0 open questions
  ```

  -copied from `odachi summary`; `odachi doctor` check 14 recomputes this line from the task documents
  and warns when the note quotes another count, so the number is never typed-.

## What waits for a decision

- **The reviewer-context decision was taken (research debt 4).** The human delegated the choice; the page in
  `docs/reviewer-context-research.md` records the three options and their costs. The choice: a frozen tree with
  bounded reads, no bash and no live worktree; `max_prompt_kb` unchanged at 120; and the planted-defect fixture
  with a clean arm, an access canary and the prompt saved. What is not done is the **measurement**: no
  before-and-after was run on this repository, so the gain is still a claim, not a number.
- **The clean arm is the next instrument job (debt 1).** The fixture's clean arm still holds findings a careful
  person would not approve; until it is clean, the precision the probe reports is not worth reading.
- **A second vendor for the four lenses** remains a decision: doctor's check 17 says `1 of 4` here, and the line
  warns without moving the exit code.
- **Two lessons from the gentle-shell comparison are still open**, in the order proposed: the typed failure with
  mutation and replay state, and the lock that only takes over a proven-dead owner. 0321 paid the silence clock and
  0322 paid the union sweep over pinned text.
- **What is left of the research besides it**: debt 1 is the fixture's clean arm, debt 2 a deterministic
  analyzer, debt 6 test adequacy.
- **The ledger**: three rules still say `none` -the question that is answered and not implemented, the
  irreversible action that waits, and the voice that belongs to the profile-.
- **The board look is unproven by a piece.** 0306 wrote the text an agent reads before producing one, and no
  board video was rendered or watched.
- **`schema.Current` is 10**: the check remedy belongs to it, and `agents.toml`, `odd/review.toml` and
  `odd/rules.toml` share the counter. The verify record has its own counter (`verify.Schema = 1`).
- **The ratchet covers three packages through declared builders**, and all three declare every builder they
  own.
- **The gate's reasons are declared once**: one table judges the state, the words, the continuation and the
  verification, and a test reads the `Reason…` constants from `gate.go` with `go/ast`.
- **The text a session pays for is pinned by bytes** in `internal/archtest/bytes.baseline`, failing in both
  directions.
- **A correction remembers its predecessor**, and a third round needs the human's reason.
- **A guard is proven by breaking it**: `check --run --mutate <file> --remove "<string>" --expect
  "<text>"`.
- The fence keeps its six active rules and no template: a non-goal that names no path stays in `brief.md`.
- `odachi doctor` is 17 checks: 16 were the install, 17 is the quorum.
- **The verifier's independence is a role, not a comparison.** The record names who verified and refuses a
  record that names nobody, but it does not compare that name with the proof's builder the way the review
  does (`ReasonReviewSelf`). Now that a line can report the quorum, the comparison is one check when a
  project runs two vendors; on this machine it would still close the queue's first item against its second.
- **Three of the catalog's fifteen ledger rows say `none`**: the question that is answered and not
  implemented, the irreversible action that waits, and the voice that belongs to the profile.
- **The board look is unproven by a piece.** 0306 wrote the text an agent reads before producing one, and no
  board video was rendered or watched.
- **`schema.Current` is 10**: the check remedy belongs to it, and `agents.toml`, `odd/review.toml` and
  `odd/rules.toml` share the counter. The verify record has its own counter (`verify.Schema = 1`) on purpose.
- **The ratchet covers three packages through declared builders**, and all three declare every builder they
  own.
- **The gate's reasons are declared once**: one table judges the state, the words, the continuation and the
  verification, and a test reads the `Reason…` constants from `gate.go` with `go/ast`.
- **The text a session pays for is pinned by bytes** in `internal/archtest/bytes.baseline`, failing in both
  directions. The skills themselves are pinned by their trigger, their sections and the body.
- **A correction remembers its predecessor**, and a third round needs the human's reason.
- **A guard is proven by breaking it**: `check --run --mutate <file> --remove "<string>" --expect
  "<text>"`. 0315's demand was proven by taking the gather out of `check.go` and watching five tests fail.
- **An open PR merges the head the remote has, not the one in your worktree.** 0317's commits were made after
  the branch was pushed; the merge took `345ae7d`, and `v0.48.1` went out without the repair its body claimed.
  Before any merge, `gh pr view <n> --json headRefOid` has to equal local `HEAD`, and a branch with an open PR
  is pushed after every commit -not once, at the beginning.
- **A squash leaves the feature's base outside the history, and CI reads it.** 0317's task document pointed at
  `345ae7d`; after the squash, `main` only had `90bb26e` with the same tree, and the gate answered `bad object`
  on a clean checkout. A base is a commit the remote has, and a squash merge is the moment to check that.
- **An evidence a clean checkout cannot carry is not a CI block.** The verdict learned it in ADR 0043 and the
  verification repeated the mistake in 0315; the repair is a note plus a flag, and the proof stays the thing
  CI reproduces. The next time a record is added to the gate, ask first where it lives.
- The fence keeps its six active rules and no template: a non-goal that names no path stays in `brief.md`.
- `odachi doctor` is 17 checks: 16 were the install, 17 is the quorum.

## What shipped

- **0322, patch (unreleased).** Every normative lens line is pinned by its words, in both directions.
- **0321, patch (unreleased).** The reviewer's clock measures silence; the proof keeps its total deadline.
- **0320, patch (v0.48.5).** The reviewer reads a frozen candidate; the prompt grows by one path; the probe
  keeps its prompts and checks access with a canary.
- **0319, patch (unreleased).** A verdict that asks for changes with no row stops the close, and the written
  waiver is the door.
- **0318, patch (v0.48.3).** The demand asks the feature's diff: a close of documents is asked for the record
  its code owes.
- **0317, patch (v0.48.2).** A verification a clean checkout cannot carry is named, and `--require-review` is
  what refuses it.
- **0316, patch (v0.48.1).** Doctor's check 17 counts the models that judge the lenses and names the one they
  share, with the fix and never a failing exit code.
- **0315, minor (v0.48.0).** A risk asks for a verifier and the gate reads the record: `odachi verify
  --exec "<command>"` composes the question from the project's own words, seals the answer under the tree,
  and `check` blocks a close whose record is missing, unreadable, unsigned or refuted.
- **0314, minor (v0.47.0).** The gate refuses a requirement that is not the brief's words. `odd/rules.toml`
  gains `gate:<reason>` as its fourth kind.
- **0313, patch (v0.47.0).** `odd/review.toml` declares schema 10 and its two check rows name their remedy.

## 0322: what it settled, and what it leaves

- **A count is not an identity.** `bytes.baseline` says the four lenses are still the same size; it cannot say the
  same rules are still there. gentle-shell's disposition sweep against a frozen fixture does, and the second file
  here is the odachi-sized version of it: 117 lines, one per surface and physical line, with `# why:` comments
  reserved for deliberate changes.
- **The sweep is a multiset in both directions.** A rule deleted, a rule rewritten to the same length and a rule
  added without permission are all named with the surface, the exact line and the file to edit. A duplicate is
  named too, because a count is compared, not just presence.
- **No lens was touched.** The 0295 pin measures the bytes that a session pays for; this one measures the words a
  reviewer obeys. The fence still protects `internal/assets/review/*.md` — the mutation used to prove the guard was
  restored before any commit.
- **What it leaves.** Only the four lens prompts are swept; the inherited rules live in `prompt.go` and the two
  voice blocks are still covered by the byte pin alone. The stack must be delivered in order, 0321 before 0322.

## 0321: what it settled, and what it leaves

- **A total deadline could not answer the question.** It killed a model that was slow but still streaming, and
  it waited the whole bound for one that had stopped. gentle-shell's split -first-output and idle, re-armed by
  any delta- is the mechanism behind the fix.
- **The reloj lives only where a model answers.** `RunInput` owns it because `review --exec` and `verify --exec`
  are model calls. `Run` and `RunStream` are the proof path, and a silent `go test` must not be cut at two minutes;
  a guard test pins that boundary.
- **The contract did not move.** Callers still read the same boolean `timedOut`; the reason behind it changed, and
  saying which reason is the separate typed-failure lesson below.
- **What it leaves.** `verify --exec` shares `RunInput`, so a verifier that works silently for more than two minutes now stops early -the first 0315 field run spent half an hour on `go test ./...` and answered nothing, which is the case this catches, but a legitimate silent probe would be caught too; watch the first real run-. Three lessons from the same comparison are still unimplemented: the union sweep that
  catches a same-size rewrite which drops a pinned rule, the failure schema with `mutation_outcome` and
  `replayability`, and the lock that does not steal by time alone. The review of the closing tree has not been run.

## 0320: what it settled, and what it leaves

- **The question was a false dichotomy.** "Frozen diff in the prompt or reviewer reading the live tree" left out
  the form the evidence supports: freeze the whole tree, then read that tree with bounded tools. `gitx.Archive`
  already exported a tree into a throwaway directory for the undo row, so no new snapshot mechanism was built.
- **The verdict still names the bytes it judged.** The snapshot is made from `cand.tree`, not from the worktree,
  and the run refuses before starting a reviewer if the export fails. A worktree that moves while the model
  answers cannot make it approve a tree different from the one in the file name.
- **The prompt carries a path, not the tree.** One section names the directory; the six rules no longer ban all
  reads, only commands and edits. The growth is bounded by a test and the 120 KB limit is pinned in `archtest`,
  so a future change cannot buy a bigger prompt to hide a bigger candidate.
- **The canary is reported, not enforced.** `odachi review probe` asks for a marker inside the candidate and one
  beside it, and prints `attached and scoped`, `not attached`, `too wide` or `no answer`. A failed canary does
  not block a measurement: the probe reports, and the human reads the line. It costs one extra call per run.
- **What it leaves.** No before-and-after number on this repository; the clean arm is still dirty; the two
  vendors question is still open; and the subagent files in the home still carry `read: deny` -`odachi setup`
  writes what the installed binary knows, so the new permissions reach the home only after a release carrying
  this binary is installed and setup runs again-. A stored `--exec` prompt now contains a temporary path that is removed
  when the run ends -it is the question, not a replayable tree-.

## 0319: what it settled, and what it leaves

- **The artifact is evidence, and one of its halves was not read.** The gate read the rows; the sentence is the
  other half, and a verdict attached with `changes_requested` and no row closed as ready.
- **The line between "a sentence" and "an argument".** A verdict with rows keeps the rule it had -blocking,
  refuted or waived-, because a sentence with an argument under it has already been answered by somebody. The
  new stop fires only where there is nothing to argue with: the sentence alone.
- **The door was already built.** `review waive --because` writes the human's decision into the task document,
  and the review gate stops asking. No new flag, no new record.
- **What it leaves.** The same sentence with non-causal rows still closes as ready. That is deliberate: those
  rows are arguments the gate can act on, and refusing them would re-open every refutation 0292 built.

## 0318: what it settled, and what it leaves

- **A gate question has an owner, and it is the feature.** The candidate answers "what is staged now"; the
  feature answers "what did this work change". The proof and the review already asked the second; the
  verification asked the first, and the difference was a close that never saw a record.
- **The classifier is shared, not copied.** `featurePassive` is the function the review gate calls, so the two
  cannot disagree about whether the diff deserves a lens.
- **The tests pin both sides.** A close of documents whose feature wrote code names the record; a feature that
  only wrote documents asks for nothing. The second is 0301's rule, now asked of the right diff.
- **What it leaves.** A close of documents still needs the record to name the closing tree -the verifier has to
  run on the tree that closes-. That is the cost of evidence addressed by content, and it is the human's call
  whether to run it before or after the last document commit.

## 0317: what it settled, and what it leaves

- **The gate and the checkout answer different questions.** A verdict and a verification live in the clone;
  a clean checkout has neither. Demanding one there is not rigour, it is an unclosable pull request: three
  platforms answered `verification_missing` on the first real PR of 0315, and the rule the feature had just
  won blocked the delivery of everything after it.
- **The name and the block are two decisions.** The rule stays visible -a ready line says the record is
  missing or refuted, with its command- and the refusal is the human's: `--require-review`. That is the
  sentence ADR 0043 already wrote for the verdict, applied to the second opinion.
- **The proof is the evidence a checkout can reproduce.** The suite runs there, the fence is read from the
  base, and the opinion of a model costs money and is not deterministic. Evidence that costs money cannot be
  a condition of every CI run, and the same reasoning answers the one-model question: the gate never spends
  the budget of a person who has one vendor.
- **What it leaves.** The demand is read from the candidate the close carries, so a close that only moves
  documents asks for nothing -the rule 0301 gave the lenses-. A feature whose code was committed step by step
  would never be asked unless its closing commit carried code. The fix is to ask the question of the feature's
  whole diff, the way the review gate does with `theCloseOwes`, and it is the next reachable thing.

## 0316: what it settled, and what it leaves

- **A number can be honest without deciding.** The fear was that a line saying "the quorum is one" would
  turn into the tool asking for a second vendor. It is a warning, the exit code does not move, and the
  decision stays where it belongs. What the line removes is the silence, not the choice.
- **The count is spelling-blind.** `Quorum` trims and lowercases, because two rows typed differently are one
  judge; counting them twice would invent a quorum the project does not have.
- **An absent table is not a quorum of one.** A project that runs no review, or one with no lens table, gets
  a line that says so instead of a warning about a number nobody wrote.
- **The opening was committed with the first step, again.** `start` and `plan` wrote the brief, the fence,
  the task document and its three steps, and the first commit swept them in because the opening was not
  committed on its own. The next feature commits the opening first; it is one command and it keeps the
  feature's own commits readable.
- **The first four commits landed on `main` locally.** No branch was created at the opening, and the branch
  policy would have refused the push anyway. They moved with `git switch -c feat/0316-the-quorum` and
  `git branch -f main origin/main`: the trees the marks name do not move, so the record stayed valid.
- **What it leaves.** Nothing to buy and nothing to enforce: the line reports, and the day a project puts a
  second vendor in a lens row, check 17 says `2 of 4` and stops warning.

## 0315: what it settled, and what it leaves

- **The tool produced the copy again, and this time it had to invent one.** 0314 compared two sentences the
  tool already wrote. The verifier had no copy: the record is new, and the shape it copies is the verdict's
  -content-addressed by tree, sealed with the clone's clock, refusing an author it cannot name-. The
  pattern is not "reuse what exists" but "copy the shape that already survived review".
- **Two reasons, not one, because the sentence differs.** A missing record sends the human to run the
  verifier; a record that answered `refuted` stops the close carrying the note that showed the requirement
  failing. Calling the second one missing would send the human to produce what is already there.
- **The demand sits in the gate and not in the review.** The review is softened into a note when the
  project's switch is on (ADR 0043); the verification is a block like the proof, because the requirement is
  "la puerta no cierra sin él". It is `NeedsWork`, not a decision: the agent can run the verifier.
- **A document-only close owes nothing, and that is 0301's rule reused.** A risk is a declaration about the
  code; a close that moved documents has no bytes a probe would catch.
- **One old test grew a record.** 0301's lens test closes a feature whose commits declare `security`, and
  the verification now fires before the review; the fixture seeds a verified record so the tree lacks
  exactly one thing. The test stayed about the lens.
- **A wrong order of paragraphs is the cheapest bug.** The first draft of the gate sentence joined risk
  names into a clause that read "the commits declared the record cannot be read": the field carries
  sentences, and the sentence has to say what is missing before it lists its entries.
- **The first field run found what the tests could not, twice.** The installed verifier answered in prose
  while the verb requires one JSON object, so its prompt gained the sentence that a JSON prompt is answered
  with that object and nothing else, and the home was refreshed. And the verifier spent its whole timeout on
  `go test ./...`, which is the gate's work: the prompt now asks for focused probes, and the second run
  answered in about twenty minutes with the packages it ran written in the record's notes.
- **The record that closed this feature was independent by name.** It says `hy4-flash-free (odachi-verify)`,
  a different model from the proof's builder, and the gate did not have to compare them: the comparison is
  what the vendor line makes possible, not what this feature promised.
- **What it leaves.** The comparison of names (see the queue), and one measurement: nobody has run a
  verifier that refutes. The first refutation is what says whether "refuted" is the answer a verifier
  reaches.

## The limits to remember

- **The note's totals are written after the last step is marked, not before.** `plan --done` rewrites the
  task document, so the number the note quotes is the one the tool counts once every step is closed.
- **A mark records the tree the worktree had when the mark was written, which is the closing one and not the
  change commit's.** `plan --done T5` said "the tree was not recorded: the worktree is not the commit" and
  then wrote the worktree's tree anyway; the second call named it. The step's record is honest about the
  bytes that carried the mark, and the proof is what speaks for the close.
- **A schema bump needs the tree's binary ahead of the installed one.** The commit hook runs `odachi` from
  the PATH, and an installed binary older than the file refuses it as newer. Build into the temp folder,
  prepend the PATH, and the gate judges the tree with the code that wrote it (0312).
- **A reason a doctor older than the tree does not know is a typo to it.** `odd/rules.toml` now names
  `gate:verification_missing`, and an installed 0.47.0 reports it as a problem. Build the tree's binary to
  read the table it wrote; CI reads it after the release.
- **The proof names the index tree** (`git write-tree`). The order that closes a feature: the change
  commit, then `plan --done` on top of it, then stage the records and measure, then commit the records.
  The proof measures the closing tree, which is the tree the records commit will have.
- **A document commit after a close leaves the local `check` red**, on purpose. CI regenerates its own
  proof from a clean checkout.
- **The undo row answers four different things**: `noticed` with the symbol the compiler named, `the copy
  did not build` for a failure that names nothing the change added, `not owed: the change adds no
  production behavior`, and `missed` for a suite that stayed green.
- **Text with backticks is not edited with a double-quoted shell string.** PowerShell reads `` `t `` and
  `` `b `` as escapes. Edit with the file tool or single quotes, and read the result back.
- **PowerShell's `Set-Content -NoNewline` on an array joins the lines.** It flattened a Go file into one
  line and the build said "expected 'package', found EOF". Write whole files with the file tool.
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
# the record the close owes: one command, the prompt on stdin, the answer on stdout
odachi verify --exec '<command that runs odachi-verify with {prompt}>'
odachi verify        # the prompt it would ask, and nothing runs
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
# the probe stores each prompt beside its answer and spends one extra call on the access canary
odachi review probe --runs 2 --exec 'C:\Users\USER\AppData\Local\Temp\odachi-rev.bat {model} {prompt} C:\Users\USER\AppData\Local\Temp\review.json'
```

```sh
# the release: a branch that renames Unreleased, a PR, then the tag on its merge
git switch -c release/vX.Y.Z && git commit -am "release vX.Y.Z" && gh pr create --base main
git tag -a vX.Y.Z <merge commit> -m "vX.Y.Z: <the sentence the notes carry>" && git push origin vX.Y.Z
irm https://github.com/biggs-100/odachi/releases/download/vX.Y.Z/install.ps1 | iex
odachi --version
```
