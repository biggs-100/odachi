# Next session

Written after `v0.38.12` shipped. Nothing is owed.

## Where the project is

- `v0.38.12` is published: annotated tag on `c0f8c55` (`Merge pull request #526`), sixteen assets, and the
  binary installed on this machine reports `odachi v0.38.12`. `main` is that merge commit.
- The patch it names came in with #525: `odachi start` appends a block to `brief.md` instead of writing it
  from scratch, and the fence is built from every non-goal the brief carries (0271).
- 269 features, 650 steps, 0 open questions; `docs/review-debt.md` holds no open row.
- The closing tree carries `proof/0271-el-brief-acumula.txt` and one verdict
  (`.git/odachi/reviews/b82fded7….json`, reliability, base `c3c7004`).
- `## Unreleased` does not exist: the next feature writes it again, above `## v0.38.12`.

## What shipped

- **0271, the brief keeps every feature it has seen.** The fence has been append-only since 0266, so a rule
  outlived the feature that wrote it while the sentence behind it did not. A run now appends a block named
  after the feature, the block repeats the canonical headers `brief.Parse` already reads, and the four lists
  are the union: outcomes, done-when, non-goals and decisions. The same requirement is not written twice,
  `--force` still writes the file when nothing is new because `undo` reads the backup, and the fence is
  built from the union of the blocks.
- **Two findings from its own review, both real and both paid.** The brief went through a console that
  decoded UTF-8 as CP850 during the hand merge that opened the feature (`â€¦`, `Ã³`), and the tree's review
  found it; the third outcome said the fence was built from the union and it was not, and the tree's review
  found that too. Both corrections are in the pull request with the reds that pin them.

## The questions that wait

- **None open.** The brief question is answered: `brief.md` is the project's vision and it accumulates, one
  block per feature, with the same headers the parser reads. A light feature (`start --light`) still writes
  no brief: it is one line, and its line lives in its task document.
- **One approval is one draw.** The lens approved the closing tree, and the same lens had found two real
  defects in the previous tree. A verdict is a sample, not a proof that nothing is left (0265).

## The limits to remember

- **A review row is keyed by bytes, not by truth.** The WARNING about `odd/fence.toml` stayed potent after
  its fix landed, because the row names the fence file and the fix was in the code. `docs/review-debt.md`
  records it as paid; the store retires it when those bytes change.
- **The console is not UTF-8.** PowerShell on this machine decodes UTF-8 as CP850 on the way through, so a
  file rebuilt through `Get-Content`/`Set-Content` comes back with `â€¦`. Rebuild artifacts with Python or
  with the tool; `brief.md` is the artifact whose requirements must stay verbatim.
- **The environment is part of the red** (0269), **the store can block a verdict** (0265), **one tree, one
  verdict** and a wider base covers a later feature (0268), **a docs-only change owes no new proof**.
- **This file goes stale in one release.** Four releases today, four rewrites; write it when the facts move.
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
# the release, after the merge it names: heading only, an annotated tag, the binary rebuilt
go build -trimpath -ldflags "-s -w -X main.version=vX.Y.Z" -o %LOCALAPPDATA%\odachi\bin\odachi.exe ./cmd/odachi
```
