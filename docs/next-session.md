# Next session

Written after 0291, 0292 and 0293, and after the release that carries the three of them.

## Where the project is

- `v0.43.0` is published: annotated tag on `6b2f120` (`Merge pull request #585`), sixteen
  assets built, signed and attested, and the binary on this machine was installed from that
  release -not built by hand- and reports `odachi 0.43.0`.
- `main` is past 0293's merge (PR #584), 291 features, 716 steps, 0 open questions.
- **`schema.Current` is 8 for all four strict decoders**: the verdict, the refutation,
  `agents.toml` and `odd/review.toml`. A binary at `v0.42.0` or older refuses a verdict or a
  refutation written by this one and says "upgrade": that is the cost `schema.go` declares it
  pays at every move, and it is why the release is a minor.
- **The ratchet covers three packages through declared builders**: `internal/app/check`,
  `internal/app/start` and `internal/app/review`. `refusal.AtTheCall(name)` reads a
  constructor's first argument at its call sites; `refusal.InTheBody(name)` reads the sentence a
  helper hardcodes, once. A package declares its own, so `check` and `start` keep their counts
  while `review` reads `refused(...)` too.
- **The gate's reasons are declared once**: one table in `internal/domain/gate/gate_test.go`
  judges the state, the exact words and the continuation, and a test reads the `Reason…`
  constants from `gate.go` with `go/ast` and fails when one has no row -either direction-.
- The fence keeps its five active rules; none protects `internal/domain/gate` since 0290.
- `odachi doctor` is clean at v0.43.0: 15 checks, no problems.

## What shipped

- **0291, minor.** A refutation's `at` is the instant this clone sealed the file, the refuter's
  declared hour lives beside it in `declared_at` when it disagrees, and `--show` prints one line
  per witness. The refutation declares its schema, which is what moved the counter to 8. The
  first review asked for changes and named three real rows: a witness line that called an
  unattributable hour a seal for every record already in a store, a second seal that read the
  tool's own stamp as the refuter's claim -and depended on how far apart the two seals fell-,
  and `instant` left without a caller.
- **0292, patch.** `internal/refusal.Walk` takes the builders a package declares, so `review`'s
  refusals are inside the mechanism, and the one dead end it could then see pays: an unreadable
  verdict names the file and `odachi review --attach <file>`. The baseline row was added and
  deleted in the same branch, which is the ratchet working in both directions.
- **0293, patch.** The gate's continuation is judged from the one table; the seven-state hand
  list and its test are gone; a reason in the source with no row fails the suite.

## The decision that waits

**The next debt has no name yet.** What is named and waiting:

- `go.mod` asks for a patch of Go. A one-line diff and a full measurement.
- The classifier reads `go/token` as touching secrets, so any test that parses Go asks for all
  four lenses. 0293 paid it; the note stays because the next one will too.
- `check`'s `refused(...)` is not declared in its ratchet. It already names the command, so it
  owes nothing today; the reason to declare it is that nothing watches it.
- Heterogeneous lenses stay deferred by the human -no money for more models-.

## The limits to remember

- **The proof names the index tree** (`git write-tree`), not `HEAD`: a close writes its records,
  stages them, measures, and only then commits, or the proof names a tree the commit never had.
  The commit hook refuses any commit after the task doc closes until a proof exists for it.
- **A document commit after a close makes the local `check` red** (`changed since the proof`) and
  that is expected: the tree moved and the proof did not. CI regenerates its own proof from a
  clean checkout, so the release pull request passes on the workflow, not on this clone.
- **The undo row answers three different things**: `the copy did not build` (a change that adds a
  symbol its tests call -the ADR 0052 limit-), `not owed: the change adds no production
  behavior` (tests or documents only), and a stored verdict for a suite that stayed green.
- **A guard that passes the first time proves nothing.** 0293's `go/ast` test was verified by
  mutation: a reason constant was added to `gate.go` on purpose, the test named it, the constant
  was removed. 0291's second-seal test passed by luck before the fix -the clock decided it-.
- **A release is a tag, and the tag is on the merge of the release pull request**, whose only
  change is renaming `## Unreleased`. The tag is what publishes: the workflow builds, signs and
  attests. Install from the release -the installer verifies -the signature and the sha256 with
  `cosign`, which must be on the PATH-; a `go build` in this folder is not published bytes.
- **A version is never moved, edited or reused**: a mistake is fixed by the next one.
- **PowerShell's `Add-Content` writes CRLF and a trailing newline**: `internal/archtest` runs
  `gofmt` over the whole repo and will fail a file written that way, in the working copy -the
  index may already be right-. `gofmt -w` after writing Go through a here-string.
- **The fence is read from the base, not from the branch**: a policy change lands first, in its
  own pull request.
- **The gate runs one command of its own**: the undo row. Lint and `govulncheck` run in the
  review and travel as rows, both at `CRITICAL`.
- **`odachi review --since <tree>` needs `--exec` and takes no `--base`**: a correction measures
  the delta, carries the predecessor's rows, and refuses without a reviewer to ask.
- **One heavy gate per machine**: `go test ./...` takes about 700 s here, and `check --run` runs
  it twice when the undo row can revert the production files.

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
