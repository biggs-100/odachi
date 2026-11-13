# Next session

Written after `v0.40.5`, 0284 and the release that published it.

## Where the project is

- `v0.40.5` is published: annotated tag on `ca815ff` (`Merge pull request #562`), sixteen
  assets built and attested, and the binary on this machine reports `odachi v0.40.5`.
- `main` is past 0284's merge (PR #561). 282 features, 689 steps, 0 open questions.
- **`main` is protected:** every change enters through a pull request, the release rename
  included. A fence change is still two problems -the policy lands before the code that
  touches it-.
- **Every review runs two tools before any lens**: `lint` (`golangci-lint run ./...`) and
  `govulncheck ./...`, both at `CRITICAL`, approved in this clone. The second needs the
  network; a machine that cannot reach the vulnerability database fails the row instead of
  passing it, and the way past is `--dismiss check:govulncheck --because "<why>"`.
- The repository asks for **Go 1.26.11** (`go.mod`): the release that clears the standard
  library vulnerabilities the code reaches.
- A verdict's `reviewer` field names each judge **once**, in the order it appeared: the
  0284 verdict reads `opencode/odachi-review + odachi-reviewer`.
- `odachi doctor` is clean: 15 checks, no warnings.

## What shipped

- **0284, a verdict names each of its judges once** (`v0.40.5`). The `reviewer` field
  appended every merge whole, so the same judge was written again by each draw and each
  correction: the 0283 verdict named six authors where there were two. `joinReviewers` keeps
  the record's shape and spelling -one sealed string, ` + ` between names- and writes a name
  once. It is not a list: a schema change would cost a minor release and force every older
  binary to say "update" for a field that only had to stop repeating itself. The trade-off is
  named in the code: a command containing ` + ` would be split like a name, and the merge puts
  both halves back.

## The decision that waits

**The research's queue is the work**, and nothing is owed. In its order: a third model from
another family for the probe with `--runs 5`, which says whether "suggests on clean code" is
one vendor's habit or a property of the lens prompt; then heterogeneous judges for full-tier
trees, which the format already supports as one model per lens in `odd/review.toml`.

**One wart is named and not fixed:** `go.mod`'s `go` directive is a patch version, so a
contributor's toolchain is the repository's decision and not their own. That is the price of
a standard library fix, not an accident.

## The limits to remember

- **`odachi plan --done --commit <sha>` records the tree of that commit**, not the worktree's.
  So the artifacts `odachi start` wrote -the brief block, the fence templates- are in no
  step's tree until they get their own commit, and committing them last is honest: the closing
  tree is the worktree's, and moving the opening to the front would rewrite trees the task
  document already names.
- **A record line written after the proof voids the proof.** `ask --craft` and `plan --done`
  edit the task document, which moves the tree; the commit hook then refuses with "changed
  since the proof". Write every record, then measure, then commit.
- **A floor is declared after the debt is paid**, and the debt can be the toolchain itself.
  `govulncheck` names a "fixed in" per advisory; the answer is the newest patch of the minor.
- **A CRITICAL row that needs the network is a contract with whoever clones.** It fails loudly
  instead of passing quietly, which is the point, and it can stop a close on a plane.
- **A new toolchain changes the compiler's opinion.** `go vet` gets stricter.
- **A commit's message is safer rewritten with `git commit-tree <tree> -p <parent> -F msg`**
  than with checkout+amend: a failed `git switch` leaves `HEAD` where it was and the amend
  rewrites the wrong commit. The reflog recovers the trees.
- **A rebase does not move a proof or a verdict when the content does not change.** A tree id
  is content.
- **A release renames `## Unreleased` before the tag, and the rename is a PR**
  (`release/vX.Y.Z`), with the tag on that merge. A tag with no section publishes no notes.
- **The probe's answers live in `.git/odachi/probe/<fixture-tree>-<arm>-<lens>-<n>.json`**,
  the name does not carry the model -the `model` field inside does-, and a replay deck re-reads
  a measured run without paying for it again.
- **PowerShell's `Set-Content`/`Add-Content` write CRLF and the repo is LF**: `gofmt -w` after
  Go code, and check the bytes after a `-replace` that renames a heading.
- **The local hook can be skipped with `--no-verify`** for a single-document change; CI runs
  the whole suite with `--fence-from`. Never for code.
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
# the floor's own evidence is in the prompt, not in the exit line
odachi review --base <feature base> | Select-String "govulncheck: passed"
```

```sh
# re-read a measured run without paying for it
odachi review probe --exec 'replay.bat {prompt}'
```

```sh
# the release: a branch that renames Unreleased, a PR, then the tag on its merge
git switch -c release/vX.Y.Z && git commit -am "release vX.Y.Z" && gh pr create --base main
git tag -a vX.Y.Z <merge commit> -m "vX.Y.Z: <the sentence the notes carry>" && git push origin vX.Y.Z
go build -trimpath -ldflags "-s -w -X main.version=vX.Y.Z" -o %LOCALAPPDATA%\odachi\bin\odachi.exe ./cmd/odachi
```
