# Next session

Written after 0291 to 0296, the two releases that carried them, and the study of gentle-shell that
produced 0295 and the plan behind 0297.

## Where the project is

- `v0.44.0` is published: annotated tag on `1ae4306` (the merge of release PR #591), sixteen assets
  signed and attested, and the binary on this machine was installed from that release -not built by
  hand- and reports `odachi 0.44.0`.
- `main` is at that merge, 294 features, 722 steps, 0 open questions.
- **`schema.Current` is 9**: the verdict carries `corrects`, the refutation carries `declared_at`, and
  `agents.toml` and `odd/review.toml` share the same counter. A binary at `v0.43.0` or older refuses a
  verdict written by this one and says "upgrade": that is what the minor bought, and it is the promise
  a reader with a store needs.
- **The ratchet covers three packages through declared builders**, and all three declare every builder
  they own: `check` and `review` read `blocked` at its call sites and `refused` in its body, `start`
  reads `blocked`.
- **The gate's reasons are declared once**: one table judges the state, the words and the
  continuation, and a test reads the `Reason…` constants from `gate.go` with `go/ast`.
- **The text a session pays for is pinned by bytes**: two block voices, four lens prompts and
  `SKILLS.md`, in `internal/archtest/bytes.baseline`, failing in both directions.
- **A correction remembers its predecessor**: `corrects` names the direct predecessor's tree, the store
  can be walked, and a third correction needs the human's reason.
- The fence keeps its five active rules; none protects `internal/domain/gate` since 0290.
- `odachi doctor` is clean at v0.44.0: 15 checks, no problems.

## What shipped

- **0291, minor (v0.43.0).** A refutation's `at` is the instant the clone sealed it, the refuter's
  declared hour lives in `declared_at`, and `--show` prints one line per witness. The refutation
  declares its schema, which moved the counter to 8. The first review named three real rows and all
  three were fixed.
- **0292, patch (v0.43.0).** `refusal.Walk` takes the builders a package declares; `review`'s refusals
  entered the mechanism, and the dead end it could then see pays: an unreadable verdict names the file
  and `odachi review --attach <file>`.
- **0293, patch (v0.43.0).** The gate's reasons are one table; the seven-state hand list is gone; a
  reason in the source with no row fails the suite.
- **0294, no entry.** `check` declares its `refused` helper too: the last asymmetry between the two
  packages closed. Nothing a person installs changed.
- **0295, no entry.** `internal/domain/budget` pins the bytes of what an agent reads, in both
  directions, with the numbers measured by the test and never typed.
- **0296, minor (v0.44.0).** A verdict remembers the tree it corrects, the chain can be walked, and a
  third correction is refused with the flag that continues it. The walk stops at a link the store no
  longer holds: a pruned store forgets old trees, and counting past a hole would invent rounds.

## The next feature: 0297, mutation as a mode of the undo row

Today, proving that a guard bites is a habit the agent has to remember. Three times this session it was
done by hand: a reason constant added to `gate.go`, the command taken out of `check`'s refusal sentence,
and a line appended to `SKILLS.md`. Each time the evidence lived in a commit message, which is a record
nobody can re-run.

The shape worth designing: `odachi check --mutate <file>` (or a flag on `--run`) reverts the *source of
truth* rather than the production file, runs the suite, asserts the exact failure message, and restores
the file. The undo row already knows how to revert and restore inside a copy; this is the same machinery
pointed at a smaller target.

Two questions to settle before the code: what counts as "the source of truth" for a mutation (a rule
text, a constant, a table row), and whether the mutation is declared in the task doc or typed on the
run. gentle-shell's version -`Triangulation A` mutates the canonical fixture, `Triangulation B` the
mirror, and both must fail with the exact message- is the reference.

## Then: 0298, the measurement document

A fixed shape for any measurement this project takes: the pass/fail criterion written **before** the
run, a verdict table, the exact commands, the excluded runs with their reason, and a section on what the
data does not show. The vocabulary comes from gentle-shell's reports: *"bytes are not tokens"* and
*"instruction accounting is not a cost measurement"*.

Without it, 0296 and 0297 have nowhere to prove they helped, and the next "diet" is a story.

Neither 0297 nor 0298 needs a release on its own: 0297 adds a flag (a capability, so a patch when it is
usable), and 0298 is documents. Both ride the release after them.

## The limits to remember

- **The proof names the index tree** (`git write-tree`): a close writes its records, stages them,
  measures, and only then commits. The commit hook refuses any commit after the task doc closes until a
  proof exists for that tree.
- **A document commit after a close leaves the local `check` red**, on purpose: the tree moved and the
  proof did not. CI regenerates its own proof from a clean checkout.
- **The undo row answers three different things**: `the copy did not build` (a change that adds a symbol
  its tests call), `not owed: the change adds no production behavior` (tests or documents only), and a
  stored verdict for a suite that stayed green. Every feature this session gave one of the three.
- **A guard that passes the first time proves nothing.** Every new guard this session was verified by
  mutation, and two of them passed by luck before their fix.
- **A pin is exact, not a ceiling**, and it fails in both directions: a text that shrinks leaves the
  number above it, and the next raise hides in that gap.
- **A walk that counts stops at a hole.** A pruned store forgets trees, and counting past a missing link
  would refuse on a fact nobody has.
- **`## Unreleased` has to be re-created after every release**: the release renames it, and the worklog
  is not a permanent heading.
- **Measure the debt before opening the feature.** The gate's note said ten reasons and two pull
  requests; the tree had eleven and needed one. The refusal debt said eighteen calls and had six.
- **A release is a tag on the merge of the release pull request**, whose only change is renaming
  `## Unreleased`. The tag publishes; the workflow builds, signs and attests. Install from the release;
  a local `go build` is not published bytes. **Minor is decided by the stored shape, not the diff size.**
- **PowerShell's `Add-Content` writes CRLF and a trailing newline**: `internal/archtest` runs `gofmt`
  over the whole repo and fails a file written that way in the working copy -the index may already be
  right-. `gofmt -w` after writing Go through a here-string.
- **The fence is read from the base, not from the branch**: a policy change lands first, in its own
  pull request.
- **The gate runs one command of its own**: the undo row. Lint and `govulncheck` run in the review and
  travel as rows, both at `CRITICAL`.
- **`odachi review --since <tree>` needs `--exec`, takes no `--base`, carries the predecessor's rows,
  and now refuses a third round without `--because`.**
- **One heavy gate per machine**: `go test ./...` takes about 700 s, and `check --run` runs it twice
  when the undo row can revert the production files.

## The commands that work

```sh
odachi check --run "go test ./..." --timeout 1800
```

```sh
# the review of a closing tree: the lenses the evidence asks for, two draws each
odachi review --runs 2 --base <feature base> --exec 'C:\Users\USER\AppData\Local\Temp\odachi-rev.bat {model} {prompt} C:\Users\USER\AppData\Local\Temp\review.json'
```

```sh
# a correction measures only the delta, carries the predecessor's rows, takes no --base
odachi review --since <tree prefix> --runs 2 --because "<why>" --exec '<the same command>'
```

```sh
# the release: a branch that renames Unreleased, a PR, then the tag on its merge
git switch -c release/vX.Y.Z && git commit -am "release vX.Y.Z" && gh pr create --base main
git tag -a vX.Y.Z <merge commit> -m "vX.Y.Z: <the sentence the notes carry>" && git push origin vX.Y.Z
irm https://github.com/biggs-100/odachi/releases/download/vX.Y.Z/install.ps1 | iex
odachi --version
```
