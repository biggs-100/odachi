# Next session

Written when the session closed, after `v0.40.0` and 0278 shipped. Nothing is owed.

## Where the project is

- `v0.40.0` is published: annotated tag on `a23ea84` (`Merge pull request #543`), sixteen assets
  built and attested by the release workflow, and the binary installed on this machine reports
  `odachi v0.40.0`.
- `main` is `a23ea84`. 276 features, 670 steps, 0 open questions.
- The closing tree of 0278 carries `proof/0278-la-herramienta-corre.txt` and one verdict
  (`.git/odachi/reviews/0bba7325….json`, two draws per correction, fourteen draws of reliability
  over the feature), approved with no open row.
- `## Unreleased` does not exist: the next feature writes it again, above `## v0.40.0`.

## What shipped

- **0278, the tool runs before the model.** `odd/review.toml` declares deterministic tools -name,
  command and the severity a failure is worth- and the review runs each one on the candidate before
  any lens reads it. A tool that exits non-zero, hangs past `check_timeout_s` or names a command
  nobody installed is filed as a finding at the declared severity, with its own words as the
  evidence; the prompt carries it so no lens pays to find it again. A tool's row has no path and no
  line: `--show` prints `check:lint`, and a blocking one is dismissed with `--dismiss check:lint
  --because "<why>"`, the one blocking row a person may retire without a refuter.
- **The declaration travels, the agreement does not.** `odachi review checks approve` writes the
  exact list's fingerprint in the clone's store, so a cloned `review.toml` runs nothing until the
  person who cloned it says so, and a command that changes asks again. The fingerprint is
  length-prefixed, because the review of this tree showed a separator inside a command could make
  one check hash like two.
- **The review found two defects in its own security control** -the forgeable fingerprint and a
  two-line command written as invalid TOML- and both are fixed with tests. That is the second tree
  in a row where a full read found what a correction had approved.
- The verdict and `odd/review.toml` are schema 6, and `docs/releases.md` is why this was a minor.

## The decision that waits

**What the review does next: run this against a real ruleset, and then the standing blocker.** The
mechanism exists and this project declares nothing, which is the first thing to fix and the last
thing to do it: `golangci-lint run ./...` reports 46 findings here today -38 of them `errcheck`-, so
declaring it means choosing a ruleset (`.golangci.yml`) and fixing or excluding what it names.
`govulncheck` is not installed on this machine, and installing it is a decision, not a step. After
that, Part 8's second item is still untouched: a close that carries `changes_requested` should say
so in the gate's own line and need a recorded human waiver, instead of "ready" with a footnote.

**One wart, named and not fixed:** a verdict's `reviewer` field concatenates its merges, and it now
also concatenates the reviewer commands of every correction, so one field reads as a paragraph of
six names. The identity the gate compares is `models`; deduping the field is a small change nobody
has asked for.

## The limits to remember

- **A finding about the record costs a full cycle, and so does one about the code.** 0278 took three
  proofs and three reviews: two real defects in the control it built, and one done-when the step had
  rewritten to match the release it wanted.
- **A correction is not a second reading.** The delta of a document-only change asks for no lens, and
  `--lens` beside it needs `--because`: the scope record is the only place a later reader learns who
  asked.
- **The prompt of this feature is 148 KB** against `max_prompt_kb = 120`, so it splits in two parts,
  and a split costs one call per part per draw. Raising the limit means a new tree and a new proof.
- **A check is judged by its exit status.** `gofmt -l .` lists files and exits zero, so it is a check
  that always passes; `golangci-lint` is the one that fails.
- **PowerShell's `Add-Content` writes CRLF and the repo is LF**: after appending Go code, `gofmt -w`
  is not optional, and `TestEveryGoFileIsGofmtClean` is what says so.
- **The console is not UTF-8** (0271), one tree one verdict (0268), a docs-only change owes no new
  proof *after* the close.
- **One heavy gate per machine**: `go test ./...` takes about 700 s here and three to five minutes on
  CI; `./cmd/...` alone is about 330 s.

## The commands that work

```sh
odachi check --run "go test ./..." --timeout 1800
```

```sh
# the review of a closing tree: four lenses, two draws each, the candidate split in two parts
odachi review --runs 2 --base <feature base> --lens reliability --model opencode-go/deepseek-v4.1-flash \
  --exec 'C:\Users\USER\AppData\Local\Temp\odachi-rev.bat {model} {prompt} C:\Users\USER\AppData\Local\Temp\review.json'
```

```sh
# a correction measures only the delta and carries the predecessor's rows; a document-only
# delta asks for no lens, so --lens needs --because
odachi review --since <tree prefix> --runs 2 --lens reliability --because "<why>" --exec '...'
```

```sh
# the checks a project declares, and whether this clone agreed to run them
odachi review checks            # what odd/review.toml declares
odachi review checks approve    # this clone agrees, by fingerprint
```

```sh
# the release, after the merge it names: the tag, then the binary
git tag -a vX.Y.Z <merge commit> -m "vX.Y.Z: <the sentence the notes carry>" && git push origin vX.Y.Z
go build -trimpath -ldflags "-s -w -X main.version=vX.Y.Z" -o %LOCALAPPDATA%\odachi\bin\odachi.exe ./cmd/odachi
```
