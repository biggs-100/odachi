# Next session

Written after 0290, the gate measuring whether a test can fail, and after the release
that carries it.

## Where the project is

- `v0.42.0` is published: annotated tag on `e4b9bf9` (`Merge pull request #580`), sixteen
  assets built and attested, and the binary on this machine reports `odachi v0.42.0`.
- `main` is past 0290's merge (PR #577), the policy that licensed it (PR #578), and the
  handover (PR #579). 288 features, 706 steps, 0 open questions.
- **The ratchet covers three packages**: `internal/app/check` (4 sentences, 5 unreadable),
  `internal/app/start` (2, 3) and `internal/app/review` (52, 5 across two files). The
  mechanism -walk, both directions, the counts- lives in `internal/refusal`.
- **The gate runs one command of its own**: the undo row of ADR 0052. It exports the tree
  with `git archive`, reverts the production `.go` files in the copy, runs the same command
  that just passed, and blocks only a suite that stayed green. Everything else it still
  reads: lint and `govulncheck` run in the review and travel as rows.
- **`main` is protected**, and a fence change is two pull requests: the policy first. CI
  reads the fence from `main`, so a rule retired inside a branch does not stop that branch
  from being judged by it -0290 learned this the hard way and split itself in two-.
- Every review runs two tools before any lens: `lint` and `govulncheck ./...`, both at
  `CRITICAL`, approved in this clone; the second needs the network.
- The repository asks for **Go 1.26.11** (`go.mod`); the verdict's schema is **7**, its `at`
  is the clone's clock and `declared_at` is what the reviewer claimed.
- `odachi doctor` is clean at v0.42.0: 15 checks, no problems.

## What shipped

- **0290, `check` measures whether a test can fail** (released in v0.42.0, with 0288 and
  0289). The row is evidence, not an action: `proof/<feature>.undone.txt` starts with
  `# undone <tree>` and every later `check` reads it back, so the block does not last only as
  long as the invocation that had `--run`. A close with no row owes it (`undo_missing`); a
  copy that does not build is a note; "there is nothing to undo" is a stored verdict; a
  refactor is waived by the human with a reason (`start --waive-undo --because "<why>"`), and
  the close names it.
- **Its first review asked for changes**, and it was right: the row lived in the run and the
  evidence was never parsed. The correction moved the block onto the evidence and added the
  missing-row reason. The second verdict approved the delta.

## The decision that waits

**The next debt has no name yet.** The ratchet covers its three packages, the gate runs the
one command ADR 0052 allows, and #572 shipped as 0290. What is named and waiting is smaller:
`review`'s `refused(...)` helper stays outside the ratchet, and the gate's continuation test
covers seven of its ten reasons.

**Heterogeneous lenses is still deferred by the human** -no money for more models-, so it is
not next.

**The third model waits for a call this size.** The free tier refuses a probe-sized prompt
after 45 to 90 seconds while a small one answers in two.

**Named and not fixed:** a refutation still writes its own time (`refute.go` calls
`instant`); the gate's continuation test covers seven of its ten reasons; `go.mod` asks for a
patch of Go.

## The limits to remember

- **The fence is read from the base, not from the branch.** A policy lands first, in its own
  pull request; the code it licenses follows. A branch that retires a rule while touching the
  path that rule protects is refused by CI with the old policy.
- **The undo row cannot judge a change that adds a symbol its tests call.** Reverting the
  production file leaves a copy whose tests do not compile, and that is a note, never a
  block. It is the documented limit of ADR 0052, and 0290 itself closes under it.
- **A refactor is the human's declared waiver**, never the tool's guess: a suite that stays
  green because nothing changed is indistinguishable from one that notices nothing.
- **The row is evidence.** It is read on every close; a file for another tree is not this
  tree's answer, and a close without it owes it the way it owes the proof.
- **A stub proof command is blocked by the row.** A fixture that closes with `echo ok` stays
  green when reverted, so it waives the row by name -the three tests that did it are the
  record-.
- **A ratchet has two directions and both must fail**, including the counts: a shrink
  reported through a log is a shrink nobody sees in CI.
- **A mechanism with a second consumer gets a home, and the extraction includes the prose.**
- **`plan --done` runs even when the commit before it failed**, recording the previous HEAD's
  tree. `odachi undo --artifact <taskdoc>`, then re-run after a commit that succeeded.
- **A record line after the proof voids the proof**, and once the document is closed the hook
  refuses the commit: write the records, measure, commit.
- **`proof/` is ignored by git.** The proof is evidence in the working tree and CI regenerates
  it; the closing commit passes because the tree, not the file, is what the proof names.
- **The classifier reads `go/token` as touching secrets**, so a test that parses Go asks for
  all four lenses.
- **The actor does not write its own receipt**; the tool stamps what it proves.
- **A field written under the old schema number is what the number exists to announce.**
- **A floor is declared after the debt is paid**, and the debt can be the toolchain itself.
- **A commit's message is safer rewritten with `git commit-tree <tree> -p <parent> -F msg`**
  than with checkout+amend; the reflog recovers the trees.
- **A rebase keeps the tree hashes when the content does not move**, so a proof and a verdict
  survive it: the commit ids change and the evidence does not.
- **A release renames `## Unreleased` before the tag, and the rename is a PR**
  (`release/vX.Y.Z`), with the tag on that merge.
- **PowerShell's `Set-Content`/`Add-Content` write CRLF and the repo is LF**: `gofmt -w` after
  Go code, and check the bytes after a `-replace` that renames a heading.
- **One heavy gate per machine**: `go test ./...` takes about 700 s here, and `check --run`
  runs it once -the undo row runs in the copy only when the binary in use has 0290-.

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
odachi review --since <tree prefix> --runs 2 --because "<why>"
```

```sh
# when a commit was refused and a step recorded the wrong tree
odachi undo --artifact odd/tasks/<the taskdoc>.md
```

```sh
# the release: a branch that renames Unreleased, a PR, then the tag on its merge
git switch -c release/vX.Y.Z && git commit -am "release vX.Y.Z" && gh pr create --base main
git tag -a vX.Y.Z <merge commit> -m "vX.Y.Z: <the sentence the notes carry>" && git push origin vX.Y.Z
go build -trimpath -ldflags "-s -w -X main.version=vX.Y.Z" -o %LOCALAPPDATA%\odachi\bin\odachi.exe ./cmd/odachi
```
