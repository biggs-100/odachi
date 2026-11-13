# Next session

Written after 0287, the ratchet over the refusals this package can read.

## Where the project is

- `v0.41.0` is published: annotated tag on `68a2ae2` (`Merge pull request #568`), sixteen
  assets built and attested, and the binary on this machine reports `odachi v0.41.0`.
- `main` is past 0287's merge (PR #570). 285 features, 696 steps, 0 open questions.
- **`internal/app/check` has a ratchet**: `refusals.baseline` freezes four sentences that
  name no way out and five calls the walk cannot read. It fails on a new one and on a
  frozen one that disappears, so the number follows the debt in the same commit.
  `internal/app/start` (33 refusals, 20 with a way out) and `internal/app/review` (114, 30)
  still owe their slice.
- **`main` is protected:** every change enters through a pull request, the release rename
  included. A fence change is still two problems -the policy lands before the code-.
- **Every review runs two tools before any lens**: `lint` and `govulncheck ./...`, both at
  `CRITICAL`, approved in this clone. The second needs the network; the way past a
  disconnected machine is `--dismiss check:govulncheck --because "<why>"`.
- The repository asks for **Go 1.26.11** (`go.mod`); the verdict's schema is **7**, and its
  `at` is the clone's clock while `declared_at` is what the reviewer claimed.
- `odachi doctor` is clean: 15 checks, no warnings.

## What shipped

- **0287, a refusal that names no way out cannot land** (no release: pinning a rule is not a
  capability). A test walks `internal/app/check` with `go/parser`, keeps the sentence of every
  `blocked(...)` it can read statically and counts the calls whose message is a value. Both
  are frozen in `refusals.baseline`, and both directions fail: a new refusal breaks the
  suite, and a frozen one that disappears breaks it until the file follows. Its own review
  found the second half was only a `t.Logf` -invisible in CI- and the fix landed in a
  correction over the delta.

## The decision that waits

**Heterogeneous lenses is a decision, not a build.** `odd/review.toml` already names one
model per lens and all four are `deepseek-v4.1-flash`: one opinion in four disguises. Giving
one or two lenses another vendor costs the same calls (four lenses, one call each per draw),
and the cheap ones sit at or below today's price. What nobody has measured is whether two
vendors disagree usefully; that measurement is the probe with two models, blocked below.

**The third model waits for a call this size.** The free tier refuses a probe-sized prompt
after 45 to 90 seconds while a small one answers in two, and the paid flash was declined.
`docs/review-research.md` Part 5d keeps the attempt and the rule it found: availability is
part of a judge's fitness.

**Named and not fixed:** a refutation still writes its own time (`refute.go` calls
`instant`); the gate's continuation test covers seven of its ten reasons and the fence
protects the package; and `go.mod` asks for a patch of Go.

## The limits to remember

- **A ratchet has two directions and both must fail.** A shrink reported through a log is a
  shrink nobody sees: `go test ./...` without `-v` discards it, so the baseline drifts above
  the debt forever.
- **The fence protects paths, not intentions.** `internal/domain/gate/**` is closed to a
  test-only fix, and loosening it is two PRs. Check the fence before planning a step that
  edits it, not after the commit is refused.
- **`plan --done` runs even when the commit before it failed.** Chaining `git commit; odachi
  plan --done` records the previous HEAD's tree. `odachi undo --artifact <taskdoc>` restores
  the record, and the tree is written again after a commit that succeeded.
- **The classifier reads a test that imports `go/token` as touching secrets.** It asks for all
  four lenses; the review is due for any code change anyway.
- **The actor does not write its own receipt.** The tool stamps what it can prove and keeps
  the claim beside it.
- **A field written under the old schema number is what the number exists to announce.**
- **A floor is declared after the debt is paid**, and the debt can be the toolchain itself.
- **A CRITICAL row that needs the network is a contract with whoever clones.**
- **A commit's message is safer rewritten with `git commit-tree <tree> -p <parent> -F msg`**
  than with checkout+amend; the reflog recovers the trees.
- **A rebase does not move a proof or a verdict when the content does not change.**
- **A release renames `## Unreleased` before the tag, and the rename is a PR**
  (`release/vX.Y.Z`), with the tag on that merge.
- **The probe needs nine answers per draw** and one stall voids it.
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
