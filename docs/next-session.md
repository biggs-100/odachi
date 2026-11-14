# Next session

Written after 0291 to 0295, the release that carried the first three, and the study of gentle-shell that
produced the last two.

## Where the project is

- `v0.43.0` is published: annotated tag on `6b2f120`, sixteen assets signed and attested, and the
  binary on this machine was installed from that release -not built by hand- and reports `odachi
  0.43.0`. Everything after it rides the next release.
- `main` is past 0295's merge (PR #588), 293 features, 719 steps, 0 open questions.
- **`schema.Current` is 8**: the verdict, the refutation, `agents.toml` and `odd/review.toml`. A
  binary at `v0.42.0` or older refuses a verdict or refutation written by this one and says
  "upgrade".
- **The ratchet covers three packages through declared builders**, and now all three declare every
  builder they own: `check` and `review` read `blocked` at its call sites and `refused` in its body,
  `start` reads `blocked`.
- **The gate's reasons are declared once**: one table judges the state, the words and the
  continuation, and a test reads the `Reason…` constants from `gate.go` with `go/ast`.
- **The text a session pays for is pinned by bytes**: two block voices, four lens prompts and
  `SKILLS.md`, in `internal/archtest/bytes.baseline`, failing in both directions.
- The fence keeps its five active rules; none protects `internal/domain/gate` since 0290.
- `odachi doctor` is clean once this file names 0295: 15 checks.

## What shipped

- **0291, minor.** A refutation's `at` is the instant the clone sealed it, the refuter's declared hour
  lives in `declared_at`, and `--show` prints one line per witness. The refutation declares its
  schema, which moved the counter to 8. The first review named three real rows and all three were
  fixed.
- **0292, patch.** `refusal.Walk` takes the builders a package declares; `review`'s refusals entered
  the mechanism, and the dead end it could then see pays: an unreadable verdict names the file and
  `odachi review --attach <file>`.
- **0293, patch.** The gate's reasons are one table; the seven-state hand list is gone; a reason in
  the source with no row fails the suite.
- **0294, no release.** `check` declares its `refused` helper too: the last asymmetry between the two
  packages closed.
- **0295, no release.** `internal/domain/budget` pins the bytes of what an agent reads, in both
  directions, with the numbers measured by the test and never typed.

## The next feature, already designed: 0296, corrections that cannot repeat in silence

Requirement R1, in the human's words: "una correccion no se repite sin una razon: el veredicto
recuerda a cual corrige, y la tercera vuelta se niega salvo que el humano diga por que".

**What is missing today.** A stored verdict does not remember which verdict it corrects: `--since` is
an input to the run, not a field of the file. Without that link a third correction is
indistinguishable from the first, and each one costs a paid review.

**The design, measured before writing it:**

1. `Verdict` gains `corrects` (the predecessor's tree, optional), `Seal` keeps it, `--since` writes
   it. `schema.Current` moves to 9: a field under the old number is the lie the number exists to
   announce (0286's rule, paid again in 0291).
2. With the link, the chain can be walked from the store: a correction whose predecessor is itself a
   correction would be the third round, and it is refused unless `--because` names the reason -the
   same shape as the undo row's waiver-. `review` already has `--because`, and it is worth checking
   whether reusing it muddies its dismissal meaning or deserves its own flag.
3. It is a **minor** release, not a patch: the shape of a stored record changes.

**The limit to write in the task doc, not hide:** the bound that can be checked mechanically is the
*chain depth*. gentle-shell's finer rule -a second correction only when the same blocker still
fails- needs findings compared across verdicts, and the ids are positional inside each file. A bound
that holds beats a close approximation that only looks like one.

## The rest of the plan from the gentle-shell study

- **0297, mutation as a mode of the undo row.** Today, proving that a guard bites is a habit the agent
  has to remember; the tool could own it: mutate the source of truth, assert the exact message, and
  restore. We did it by hand for `gate.go` (0293) and for `check`'s sentence (0294).
- **0298, the measurement document.** A fixed shape -criterion written before the run, verdict table,
  exact commands, excluded runs with their reason, and a section on what the data does not show- plus
  the honest vocabulary: bytes are not tokens, and instruction accounting is not a cost measurement.
  Without it, 0296 and 0297 have nowhere to prove they helped.

## The limits to remember

- **The proof names the index tree** (`git write-tree`): a close writes its records, stages them,
  measures, and only then commits. The commit hook refuses any commit after the task doc closes until
  a proof exists for that tree.
- **A document commit after a close leaves the local `check` red**, on purpose: the tree moved and the
  proof did not. CI regenerates its own proof from a clean checkout.
- **The undo row answers three different things**: `the copy did not build` (a change that adds a
  symbol its tests call), `not owed: the change adds no production behavior`, and a stored verdict
  for a suite that stayed green. 0293, 0294 and 0295 gave one of each in the same session.
- **A guard that passes the first time proves nothing.** Three of this session's guards were verified
  by mutation: a reason constant added to `gate.go`, the command removed from `check`'s sentence, and
  a line appended to `SKILLS.md`. Two of them passed by luck before the fix.
- **A pin is exact, not a ceiling**, and it fails in both directions: a text that shrinks leaves the
  number above it, and the next raise hides in that gap.
- **Measure the debt before opening the feature.** The gate's note said ten reasons and two pull
  requests; the tree had eleven and needed one, because 0290 had retired the rule that cost the
  second. The refusal debt said eighteen calls and had six.
- **A release is a tag on the merge of the release pull request**, whose only change is renaming
  `## Unreleased`. The tag publishes; the workflow builds, signs and attests. Install from the
  release; a local `go build` is not published bytes. A version is never moved or reused.
- **PowerShell's `Add-Content` writes CRLF and a trailing newline**: `internal/archtest` runs `gofmt`
  over the whole repo and fails a file written that way in the working copy -the index may already be
  right-. `gofmt -w` after writing Go through a here-string.
- **The fence is read from the base, not from the branch**: a policy change lands first, in its own
  pull request.
- **The gate runs one command of its own**: the undo row. Lint and `govulncheck` run in the review and
  travel as rows, both at `CRITICAL`.
- **`odachi review --since <tree>` needs `--exec`, takes no `--base`, and carries the predecessor's
  rows.**
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
```
