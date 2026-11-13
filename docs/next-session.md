# Next session

Written after `v0.40.5`, 0284 and 0285, the day the probe's third model could not be measured.

## Where the project is

- `v0.40.5` is published: annotated tag on `ca815ff` (`Merge pull request #562`), sixteen
  assets built and attested, and the binary on this machine reports `odachi v0.40.5`.
- `main` is past 0285's merge (PR #564). 283 features, 690 steps, 0 open questions.
- **`main` is protected:** every change enters through a pull request, the release rename
  included. A fence change is still two problems -the policy lands before the code that
  touches it-.
- **Every review runs two tools before any lens**: `lint` (`golangci-lint run ./...`) and
  `govulncheck ./...`, both at `CRITICAL`, approved in this clone. The second needs the
  network; a machine that cannot reach the vulnerability database fails the row instead of
  passing it, and the way past is `--dismiss check:govulncheck --because "<why>"`.
- The repository asks for **Go 1.26.11** (`go.mod`): the release that clears the standard
  library vulnerabilities the code reaches.
- A verdict's `reviewer` field names each judge **once**, in the order it appeared.
- `odachi doctor` is clean: 15 checks, no warnings.

## What shipped

- **0284, a verdict names each of its judges once** (`v0.40.5`). The `reviewer` field
  appended every merge whole, so the same judge was written again by each draw and each
  correction: the 0283 verdict named six authors where there were two. `joinReviewers` keeps
  the record's shape and spelling and writes a name once.
- **0285, the third model could not be measured** (no release: a measurement is not a
  capability). `docs/review-research.md` Part 5d writes the attempt:
  `mimo-v2.6-flash` answered five of six calls, two to eight minutes each, and stalled on the
  sixth at the probe's 600 s ceiling. Its five answers were valid -the SQL blocker and the
  IDOR on the seeded risk lens, nothing on the clean one- but a draw with a missing arm is
  not a score. The free models (`opencode/glm-5.3-flash-free`,
  `opencode/minimax-m3.1-flash-free`) answer `Free usage limit reached` on this machine, and
  the paid flash siblings were declined. The finding it did leave:
  **availability is part of a judge's fitness.**

## The decision that waits

**The third model waits for a free quota or a cheap model that answers at this size.** The
question is exactly where Part 5c left it: whether "suggests on clean code" is one vendor's
habit or a property of the lens prompt. When a model answers, the run is
`odachi review probe --runs 5 --model <provider/model> --exec '<bat> {model} {prompt} <out>'`
and the reading goes to Part 5d.

After that, in the research's order: heterogeneous judges for full-tier trees, which the
format already supports as one model per lens in `odd/review.toml`.

**One wart is named and not fixed:** `go.mod`'s `go` directive is a patch version, so a
contributor's toolchain is the repository's decision and not their own. That is the price of
a standard library fix.

## The limits to remember

- **The probe needs nine answers per draw** (four lenses on the seeded arm, four on the clean
  one) and one stall voids the draw while leaving the paid answers in the store. A candidate
  judge is measured on availability first: two to eight minutes per call and one stall in six
  disqualifies it before its judgment is read.
- **The free tier of this machine is spent** (`Free usage limit reached`), and the reviewer
  `.bat` hides the provider's exit code -it redirects to a file and `type`s it-, so the probe
  reports "the verdict is empty" instead of "the reviewer failed". Read the stored answer
  before blaming the parser.
- **`odachi plan --done --commit <sha>` records the tree of that commit**, not the worktree's,
  so the artifacts `odachi start` wrote need their own commit; committing them last is honest.
- **A record line written after the proof voids it.** Once the document is closed the hook
  demands a proof for every commit: write every record, then measure, then commit.
- **A floor is declared after the debt is paid**, and the debt can be the toolchain itself.
  `govulncheck` names a "fixed in" per advisory; the answer is the newest patch of the minor.
- **A CRITICAL row that needs the network is a contract with whoever clones.**
- **A commit's message is safer rewritten with `git commit-tree <tree> -p <parent> -F msg`**
  than with checkout+amend; the reflog recovers the trees.
- **A rebase does not move a proof or a verdict when the content does not change.**
- **A release renames `## Unreleased` before the tag, and the rename is a PR**
  (`release/vX.Y.Z`), with the tag on that merge.
- **The probe's answers live in `.git/odachi/probe/<fixture-tree>-<arm>-<lens>-<n>.json`**,
  the name does not carry the model -the `model` field inside does-, and a replay deck re-reads
  a measured run without paying for it again.
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
# the floor's own evidence is in the prompt, not in the exit line
odachi review --base <feature base> | Select-String "govulncheck: passed"
```

```sh
# the probe, when a model answers at this size
odachi review probe --runs 5 --model opencode-go/<model> --exec '<bat> {model} {prompt} <out.json>'
```

```sh
# the release: a branch that renames Unreleased, a PR, then the tag on its merge
git switch -c release/vX.Y.Z && git commit -am "release vX.Y.Z" && gh pr create --base main
git tag -a vX.Y.Z <merge commit> -m "vX.Y.Z: <the sentence the notes carry>" && git push origin vX.Y.Z
go build -trimpath -ldflags "-s -w -X main.version=vX.Y.Z" -o %LOCALAPPDATA%\odachi\bin\odachi.exe ./cmd/odachi
```
