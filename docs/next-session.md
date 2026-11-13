# Next session

Written after 0289, the ratchet reaching `review` and closing over the three packages
that talk to the agent.

## Where the project is

- `v0.41.0` is published: annotated tag on `68a2ae2` (`Merge pull request #568`), sixteen
  assets built and attested, and the binary on this machine reports `odachi v0.41.0`.
- `main` is past 0289's merge (PR #575). 287 features, 700 steps, 0 open questions.
- **The ratchet covers three packages**: `internal/app/check` (4 sentences, 5 unreadable),
  `internal/app/start` (2, 3) and `internal/app/review` (52, 5 across two files). Each
  freezes its debt in a `refusals.baseline`, and the mechanism -walk, both directions, the
  counts- lives in `internal/refusal`.
- **The scope of a ratchet is the constructor it reads**: `blocked(...)`. `refused(...)`
  stays outside -7 calls in `review`, 4 in `check`-, named in 0289's decision line and not
  fixed, because widening the walker moves every baseline under it.
- **`main` is protected:** every change enters through a pull request, the release rename
  included. A fence change is still two problems -the policy lands before the code-.
- **Every review runs two tools before any lens**: `lint` and `govulncheck ./...`, both at
  `CRITICAL`, approved in this clone; the second needs the network.
- The repository asks for **Go 1.26.11** (`go.mod`); the verdict's schema is **7**, its `at`
  is the clone's clock and `declared_at` is what the reviewer claimed.
- `odachi doctor` is clean: 15 checks, no warnings.

## What shipped

- **0289, the ratchet reaches review** (no release: pinning a rule is not a capability).
  52 sentences and 2 count rows, generated from the walk so no row is a transcription; both
  directions observed by hand before the baseline existed; the verdict approved by the
  reliability lens at two draws. `check` and `start` did not move a byte.

## The decision that waits

**Heterogeneous lenses is deferred by the human until they say so** -no money for more
models-, so it is not next.

**No slice of the ratchet waits.** The three packages are covered, and the next debt is a
different shape: issue #572, the gate certifying a green command rather than a test that can
fail. It needs its brief before code, and the fence protects `internal/domain/gate/**`, so it
is two pull requests and a vision decision first.

**The third model waits for a call this size.** The free tier refuses a probe-sized prompt
after 45 to 90 seconds while a small one answers in two.

**Named and not fixed:** a refutation still writes its own time (`refute.go` calls
`instant`); the gate's continuation test covers seven of its ten reasons and the fence
protects the package; `go.mod` asks for a patch of Go.

## The limits to remember

- **A ratchet has two directions and both must fail**, including the counts: a shrink
  reported through a log is a shrink nobody sees in CI.
- **The scope of a ratchet is the constructor it reads.** Widening the walker past
  `blocked(...)` moves every baseline under it, so a scope change is its own feature with
  all of them in the same diff.
- **A mechanism with a second consumer gets a home, and the extraction includes the prose.**
  The walk moved before the count comparison did, and the review caught the half that stayed
  duplicated.
- **The fence protects paths, not intentions.** `internal/domain/gate/**` is closed to a
  test-only fix, and loosening it is two PRs. Check the fence before planning that step.
- **`plan --done` runs even when the commit before it failed**, recording the previous HEAD's
  tree. `odachi undo --artifact <taskdoc>`, then re-run after a commit that succeeded.
- **A record line after the proof voids the proof**, and once the document is closed the hook
  refuses the commit: write the records, measure, commit.
- **`proof/` is ignored by git.** The proof is evidence in the working tree and CI regenerates
  it; `git add proof/<feature>.txt` fails by design, and the closing commit passes because the
  tree, not the file, is what the proof names.
- **The classifier reads `go/token` as touching secrets**, so a test that parses Go asks for
  all four lenses.
- **The actor does not write its own receipt**; the tool stamps what it proves.
- **A field written under the old schema number is what the number exists to announce.**
- **A floor is declared after the debt is paid**, and the debt can be the toolchain itself.
- **A commit's message is safer rewritten with `git commit-tree <tree> -p <parent> -F msg`**
  than with checkout+amend; the reflog recovers the trees.
- **A release renames `## Unreleased` before the tag, and the rename is a PR**
  (`release/vX.Y.Z`), with the tag on that merge.
- **PowerShell's `Set-Content`/`Add-Content` write CRLF and the repo is LF**: `gofmt -w` after
  Go code, and check the bytes after a `-replace` that renames a heading.
- **One heavy gate per machine**: `go test ./...` takes about 700 s here.

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
