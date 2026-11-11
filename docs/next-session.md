# Next session

Written after `v0.38.13` shipped. Nothing is owed.

## Where the project is

- `v0.38.13` is published: annotated tag on `6bc8864` (`Merge pull request #530`), sixteen assets, and the
  binary installed on this machine reports `odachi v0.38.13`. `main` is that merge commit.
- Two patches went in today for it: `review --since` resolves the prefix a person copies to the stored
  verdict (0272), and the public face says what the repository does -`--help` names every verb the
  dispatcher answers, and the README lost the counts that rot (0273).
- 271 features, 654 steps, 0 open questions; `docs/review-debt.md` holds no open row.
- The closing tree carries `proof/0273-la-cara-dice-la-verdad.txt` and one verdict
  (`.git/odachi/reviews/68fecbb0….json`, reliability, base `94645f0`, no findings).
- `## Unreleased` does not exist: the next feature writes it again, above `## v0.38.13`.

## What shipped

- **0272, the tree a human copies.** `domain.Name` accepted any hash-like string, so `--since bbf385ca`
  looked for `reviews/bbf385ca.json` and answered "there is no verdict for the tree --since names" about a
  verdict the clone held -it cost a full review instead of the cheap delta. The correction resolves the
  prefix against the store listing `--all` prints, refuses an ambiguous one naming its candidates, and says
  which tree it completed.
- **0273, the public face.** `design` and `review probe` were answered and absent from the usage block; the
  README claimed eighteen surfaces (twenty-three), 44 packages (62), no provenance (`gh attestation verify`
  exits 0), and no automated test for `install.ps1` (the Windows suite runs it over `file://`). The counts
  are gone instead of updated, and a test reads the dispatcher's own cases so a verb cannot hide again.

## The questions that wait

- **None open.** The next feature writes the new `## Unreleased`.
- **One approval is one draw.** Every verdict today was a single sample; two of them approved a tree and the
  next tree's review found real defects in what the first had passed.

## The limits to remember

- **A test that reads source can be fooled.** The usage test matched `odachi skill` inside `odachi skills`
  and stopped its scan at any brace: the review of the tree found both. A reader needs a token boundary and
  the block's own closing brace, and a guard that fails when the scan does not reach the last case.
- **`## Unreleased` appears twice in `CHANGELOG.md`**: once in the heading and once in the intro prose that
  explains it. `strings.Replace` on the first match rewrites the sentence, not the heading; rename the line
  whose content is exactly `## Unreleased`.
- **A run queued with zero jobs is GitHub's problem**, not a busy queue: `gh run view <id> --json jobs` says
  `total=0`, and closing and reopening the pull request triggers a fresh run that starts.
- **A rebase of a closing branch needs a fresh proof and a fresh verdict**, and the conflicted commit cannot
  run the hook: resolve, run `check --run`, review the delta, then push with `--force-with-lease`.
- **The store's rows are keyed by bytes** (0265), **the console is not UTF-8** (0271), **one tree, one
  verdict** and a wider base covers a later feature (0268), **a docs-only change owes no new proof**.
- **One heavy gate per machine**: `go test ./...` takes about 700 s and `check --run` writes the proof.

## The commands that work

```sh
odachi check --run "go test ./..." --timeout 1800
```

```sh
# a verdict for the tree, with the answer kept beside it
odachi review --base <feature base> --lens reliability --model opencode-go/deepseek-v4.1-flash \
  --exec 'C:\Users\USER\AppData\Local\Temp\odachi-rev.bat {model} {prompt} C:\Users\USER\AppData\Local\Temp\review.json'
```

```sh
# a correction takes the prefix the listing shows
odachi review --since bbf385ca --lens reliability --exec '...'
```

```sh
# the release, after the merge it names: heading only, an annotated tag, the binary rebuilt
go build -trimpath -ldflags "-s -w -X main.version=vX.Y.Z" -o %LOCALAPPDATA%\odachi\bin\odachi.exe ./cmd/odachi
```
