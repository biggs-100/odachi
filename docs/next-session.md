# Next session

Written after 0288, the ratchet reaching `start` with its walker in a package of its own.

## Where the project is

- `v0.41.0` is published: annotated tag on `68a2ae2` (`Merge pull request #568`), sixteen
  assets built and attested, and the binary on this machine reports `odachi v0.41.0`.
- `main` is past 0288's merge (PR #573). 286 features, 699 steps, 0 open questions.
- **The ratchet covers two packages**: `internal/app/check` (4 sentences, 5 unreadable) and
  `internal/app/start` (2 sentences, 3 unreadable) freeze their debt in a
  `refusals.baseline` each, and the mechanism lives in `internal/refusal` (walk, both
  directions, the counts). **`internal/app/review` still owes its slice**: 114 refusals, 30
  with a way out.
- **`main` is protected:** every change enters through a pull request, the release rename
  included. A fence change is still two problems -the policy lands before the code-.
- **Every review runs two tools before any lens**: `lint` and `govulncheck ./...`, both at
  `CRITICAL`, approved in this clone; the second needs the network.
- The repository asks for **Go 1.26.11** (`go.mod`); the verdict's schema is **7**, its `at`
  is the clone's clock and `declared_at` is what the reviewer claimed.
- `odachi doctor` is clean: 15 checks, no warnings.

## What shipped

- **0288, the ratchet reaches start** (no release: pinning a rule is not a capability).
  `internal/refusal` holds the walk, `Grew`, `Shrunk` and `CountIssues`; `check` kept its
  baseline byte for byte and lost 202 lines of duplicated walker; `start` froze two
  sentences and three unreadable calls. Its own review found the count comparison had been
  left in both tests and it was paid in a correction over the delta.

## The decision that waits

**Heterogeneous lenses is deferred by the human until they say so** -no money for more
models-, so it is not next. The cheap model prices say a review with two vendors costs the
same calls and cents, and that argument is recorded; the decision is not.

**The next slice of the ratchet is `internal/app/review`**: 114 refusals, 30 with a way out.
It is the biggest surface and the one that talks to the agent most often.

**The third model waits for a call this size.** The free tier refuses a probe-sized prompt
after 45 to 90 seconds while a small one answers in two.

**Named and not fixed:** a refutation still writes its own time (`refute.go` calls
`instant`); the gate's continuation test covers seven of its ten reasons and the fence
protects the package; and `go.mod` asks for a patch of Go.

## The limits to remember

- **A ratchet has two directions and both must fail**, including the counts: a shrink
  reported through a log is a shrink nobody sees in CI.
- **A mechanism with a second consumer gets a home, and the extraction includes the prose.**
  The walk moved before the count comparison did, and the review caught the half that stayed
  duplicated.
- **The fence protects paths, not intentions.** `internal/domain/gate/**` is closed to a
  test-only fix, and loosening it is two PRs. Check the fence before planning that step.
- **`plan --done` runs even when the commit before it failed**, recording the previous HEAD's
  tree. `odachi undo --artifact <taskdoc>`, then re-run after a commit that succeeded.
- **A record line after the proof voids the proof**, and once the document is closed the hook
  refuses the commit: write the records, measure, commit.
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
