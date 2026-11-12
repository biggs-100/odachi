# Next session

Written when the session closed, after `v0.40.1` and 0279 shipped. Nothing is owed.

## Where the project is

- `v0.40.1` is published: annotated tag on `467e446` (`Merge pull request #546`), sixteen assets
  built and attested, and the binary on this machine reports `odachi v0.40.1`.
- `main` is `467e446`. 277 features, 675 steps, 0 open questions.
- The closing tree of 0279 carries `proof/0279-el-piso-corre.txt` and one verdict
  (`.git/odachi/reviews/de8cad6b….json`), a correction of `1c0469c9` that measured the fence
  wording the rebase brought and carried the rest.
- **This project now has a floor**: `odd/review.toml` declares `golangci-lint run ./...` at
  `CRITICAL`, `.golangci.yml` pins the five linters, and this clone approved that exact list. Every
  review runs it before any lens.
- `## Unreleased` does not exist: the next feature writes it again, above `## v0.40.1`.

## What shipped

- **0279, the floor runs for real.** Pinning the ruleset in `.golangci.yml` measured 54 findings
  where the installed default reported 46 -three `SA5011` in tests, four quickfix suggestions and an
  unused field were hidden-, and all 54 are paid: 38 errors now say what they do with them, four dead
  things are deleted, a split's dead assignment is gone.
- **The fence was narrowed and it cost two PRs.** The gate reads the policy from the base ref, so a
  change cannot touch what the base protects; and a loosened fence needs a brief change in the same
  change. The new rule landed first (#547, carrying 0279's opening), then the code (#546).
- **The changelog gained the `## v0.40.0` section it shipped without.** A release renames
  `## Unreleased` to `## vX.Y.Z - date`; v0.40.0 was published with 0278's entry still in the
  worklog.

## The decision that waits

**Part 8's second item: a standing blocker visible in the gate's own line.** A close that carries
`changes_requested` should say so where `check` speaks and need a recorded human waiver to proceed,
instead of a "ready" with a footnote in a document. The research ranks it next, it touches the gate's
voice rather than its rules, and the store already knows how to hold a named waiver
(`review waive --because`).

The other candidates, in the order the research ranks them: repeat the probe across models (the
fixture is still the one model's), and heterogeneous judges for full-tier trees. Separately,
`govulncheck` is still not installed: installing it is the decision, and the check table is ready for
it the day it exists.

**One wart, named and not fixed:** a verdict's `reviewer` field concatenates its merges and now also
every correction's command, so one field reads as a paragraph of names. The identity the gate
compares is `models`; deduping the field is a small change nobody has asked for.

## The limits to remember

- **A fence change is two problems, not one.** The base ref is the policy: touching a protected path
  needs the rule already there, and loosening a rule needs `brief.md` to move in the same change.
  Order: the rule and the brief first, the code after.
- **A rebase moves the proof, not only the commits.** A step's tree is content, so it survives; the
  proof and the verdict name content too, so a document that changes wording after the close costs a
  new proof and, at most, a `--since` correction of the delta.
- **A correction of a document-only delta needs `--because`**: the evidence asks for no lens, and the
  scope record is the only place a later reader learns who asked.
- **`golangci-lint run ./...` is ~1 minute on a warm cache** and is now part of every review here. A
  teammate without the binary gets a `CRITICAL` row saying the command could not run, which is the
  honest answer and not a pass.
- **The prompt of 0279 is about 145 KB** against `max_prompt_kb = 120`, so it splits in two parts and
  a split costs one call per part per draw.
- **PowerShell's `Set-Content`/`Add-Content` write CRLF and the repo is LF**: after editing Go code,
  `gofmt -w` is not optional, and `TestEveryGoFileIsGofmtClean` is what says so.
- **The local hook can be skipped with `--no-verify`**, and the one place it is right is a change
  that is a single document: CI runs the whole suite on it. Never for code.
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
# what the floor is, and whether this clone agreed to run it
odachi review checks
odachi review checks approve
golangci-lint run ./...     # what the declared check runs
```

```sh
# the release, after the merge it names: rename Unreleased, then the tag, then the binary
# (in CHANGELOG.md: "## Unreleased" becomes "## vX.Y.Z - <date>")
git tag -a vX.Y.Z <merge commit> -m "vX.Y.Z: <the sentence the notes carry>" && git push origin vX.Y.Z
go build -trimpath -ldflags "-s -w -X main.version=vX.Y.Z" -o %LOCALAPPDATA%\odachi\bin\odachi.exe ./cmd/odachi
```
