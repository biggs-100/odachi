# Next session

Written after `v0.40.4`, 0283 and the release that published it.

## Where the project is

- `v0.40.4` is published: annotated tag on `96cb342` (`Merge pull request #559`), sixteen
  assets built and attested, and the binary on this machine reports `odachi v0.40.4`.
- `main` is past 0283's merge (PR #558). 281 features, 687 steps, 0 open questions.
- **`main` is protected:** every change enters through a pull request, the release rename
  included. A fence change is still two problems -the policy lands before the code that
  touches it- and #553 was that first PR for 0282.
- **Every review runs two tools before any lens**: `lint` (`golangci-lint run ./...`) and
  `govulncheck ./...`, both at `CRITICAL`, approved in this clone. The second one needs the
  network; a machine that cannot reach the vulnerability database fails the row instead of
  passing it, and the way past is `--dismiss check:govulncheck --because "<why>"`.
- The repository asks for **Go 1.26.11** (`go.mod`): the release that clears the standard
  library vulnerabilities the code reaches. A machine with an older toolchain downloads it.
- `odachi doctor` is clean: 15 checks, no warnings.

## What shipped

- **0283, the floor reads the dependencies too** (`v0.40.4`). `govulncheck ./...` is a check
  at `CRITICAL`, so every review runs a linter and a vulnerability scan before any lens
  reads the candidate. The debt came first: the tool named 25 vulnerabilities the code
  reached, all in the standard library, and `go get go@1.26.11` was the fix -`1.26.3`
  cleared six and left nineteen-. The survey also reports one vulnerability in imported
  packages and two in required modules that the code never calls; the row measures what the
  code reaches, which is what the exit status says.
- **The new toolchain refused a latent test defect.** `go vet` on 1.26.11 rejected a `%q`
  in `reviewtoml_test.go` whose branch only runs when `Decode` succeeds, so it had never
  printed. Paid in its own commit, before the row was declared.
- **A trailer change corrected the research document** (#557, no document): Part 8 said
  "Repeating the probe across models is still open" and Part 5c is that repetition. The
  note now says what was kept and what waits.

## The decision that waits

**Nothing is owed, and the research's queue is the work.** In its order: a third model from
another family for the probe with `--runs 5`, which says whether "suggests on clean code" is
one vendor's habit or a property of the lens prompt; then heterogeneous judges for full-tier
trees, which the format already supports as one model per lens in `odd/review.toml`.

**Two warts are named and not fixed.** A verdict's `reviewer` field concatenates its merges
and every correction's command, so one field reads as a paragraph of names -visible in the
0283 verdict, which names six. And `go.mod`'s `go` directive is a patch version, so a
contributor's toolchain is the repository's decision and not their own; that is the price of
a standard library fix, not an accident.

## The limits to remember

- **A floor is declared after the debt is paid**, and the debt can be the toolchain itself.
  `govulncheck` names a "fixed in" per advisory; the answer is the newest patch of the
  minor, not the first one listed.
- **A CRITICAL row that needs the network is a contract with whoever clones.** It fails
  loudly instead of passing quietly, which is the point, and it can stop a close on a plane.
- **A new toolchain changes the compiler's opinion.** `go vet` gets stricter: a `%q` with a
  struct survived several releases and stopped the suite the day the directive moved.
- **The commit hook refuses a commit when a proof already exists for an older tree.** A
  correction runs `check --run` first and commits after; the tree it writes is the tree the
  commit will carry.
- **A commit's message is easier to rewrite with `commit-tree` than with checkout+amend.**
  A failed `git switch` leaves `HEAD` where it was, and `git commit --amend` then rewrites the
  wrong commit. Pin the SHAs, build the new commits with `git commit-tree <tree> -p <parent>
  -F <message>`, and move the branch last.
- **A rebase does not move a proof or a verdict when the content does not change.** A tree id
  is content: five commits were rebased onto the merge of their own opening and the proof and
  verdict still named `0e44c60`.
- **A release renames `## Unreleased` before the tag, and the rename is a PR**
  (`release/vX.Y.Z`), with the tag on that merge. A tag with no section publishes no notes.
- **The probe's answers live in `.git/odachi/probe/<fixture-tree>-<arm>-<lens>-<n>.json`**,
  the name does not carry the model -the `model` field inside does-, and a replay deck re-reads
  a measured run without paying for it again.
- **PowerShell's `Set-Content`/`Add-Content` write CRLF and the repo is LF**: `gofmt -w` after
  Go code, and check the bytes after a `-replace`.
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
