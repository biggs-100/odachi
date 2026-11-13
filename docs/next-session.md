# Next session

Written after `v0.41.0` and 0286, the feature that gave the record its clock.

## Where the project is

- `v0.41.0` is published: annotated tag on `68a2ae2` (`Merge pull request #568`), sixteen
  assets built and attested, and the binary on this machine reports `odachi v0.41.0`.
- `main` is past 0286's merge (PR #567). 284 features, 693 steps, 0 open questions.
- **The verdict's schema is 7.** `at` is the instant the clone sealed the file; the time the
  reviewer wrote is `declared_at`, kept only when it parses and disagrees. A binary older
  than v0.41.0 refuses a schema-7 file and names the upgrade.
- **`main` is protected:** every change enters through a pull request, the release rename
  included. A fence change is still two problems -the policy lands before the code-.
- **Every review runs two tools before any lens**: `lint` (`golangci-lint run ./...`) and
  `govulncheck ./...`, both at `CRITICAL`, approved in this clone. The second needs the
  network; a machine that cannot reach the database fails the row instead of passing it, and
  the way past is `--dismiss check:govulncheck --because "<why>"`.
- The repository asks for **Go 1.26.11** (`go.mod`).
- A verdict's `reviewer` field names each judge once, in the order it appeared.
- `odachi doctor` is clean: 15 checks, no warnings.

## What shipped

- **0286, the clock stamps the verdict** (`v0.41.0`, minor). `Seal` kept the reviewer's own
  `at`, so every verdict in the sample claimed a round hour. The tool now stamps the instant
  it wrote the file and keeps the answer's time as `declared_at` when it differs; `--show`
  and the line a run ends with print both. A re-seal (a completion or a correction) keeps the
  claim the first part made. The feature proved itself on its own review: that verdict
  declared `at: 2026-11-12T21:30:12Z`, the exact instant written in the tests the same diff
  added -the model copied a timestamp out of the text it was reviewing.

## The decision that waits

**The research's queue is the work, and nothing is owed.** In its order: the third model for
the probe with `--runs 5` -blocked while the free tier refuses a call the size of a probe
prompt and the paid flash is declined- and then heterogeneous judges for full-tier trees,
which the format already supports as one model per lens in `odd/review.toml`.

**Two things are named and not fixed.** A refutation also declares its time with the writer's
pen (`internal/domain/review/refute.go` still calls `instant`), and the same fix applies; the
requirement was about a verdict, and the other surface is its own change. And `go.mod`'s `go`
directive is a patch version: a contributor's toolchain is the repository's decision.

## The limits to remember

- **The actor does not write its own receipt.** The tool stamps what it can prove -the tree,
  the schema, the instant- and keeps what the actor claims beside it, labelled as a claim.
  The same rule applies to any field a reviewer writes about itself.
- **A field written under the old schema number is what the number exists to announce.** Bump
  the schema in the commit that adds the field, not in the one that documents it.
- **The installed binary is the one that runs in reviews.** A feature's own verdict is sealed
  by the released binary, not by the tree under review, so its `schema` is the released one.
- **`odachi plan --done --commit <sha>` records the tree of that commit**, not the worktree's.
- **A record line written after the proof voids it.** Write every record, then measure, then
  commit.
- **A floor is declared after the debt is paid**, and the debt can be the toolchain itself.
- **A CRITICAL row that needs the network is a contract with whoever clones.**
- **A commit's message is safer rewritten with `git commit-tree <tree> -p <parent> -F msg`**
  than with checkout+amend; the reflog recovers the trees.
- **A rebase does not move a proof or a verdict when the content does not change.**
- **A release renames `## Unreleased` before the tag, and the rename is a PR**
  (`release/vX.Y.Z`), with the tag on that merge.
- **The probe needs nine answers per draw** and one stall voids it; the free tier of this
  machine refuses a call the size of a probe prompt after 45 to 90 seconds, while a small
  prompt answers in two.
- **The probe's answers live in `.git/odachi/probe/<fixture-tree>-<arm>-<lens>-<n>.json`**,
  the name does not carry the model -the `model` field inside does.
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
# a correction measures only the delta, carries the predecessor's rows, and takes no --base
odachi review --since <tree prefix> --runs 2 --because "<why>"
```

```sh
# what the record says about itself: who judged, when the tool sealed it, and the claim
odachi review --show
```

```sh
# the release: a branch that renames Unreleased, a PR, then the tag on its merge
git switch -c release/vX.Y.Z && git commit -am "release vX.Y.Z" && gh pr create --base main
git tag -a vX.Y.Z <merge commit> -m "vX.Y.Z: <the sentence the notes carry>" && git push origin vX.Y.Z
go build -trimpath -ldflags "-s -w -X main.version=vX.Y.Z" -o %LOCALAPPDATA%\odachi\bin\odachi.exe ./cmd/odachi
```
