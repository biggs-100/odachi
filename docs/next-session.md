# Next session

Written after `v0.38.9` shipped. Nothing is owed.

## Where the project is

- `v0.38.9` is published: annotated tag on `63a5657` (`Merge pull request #517`), and the binary installed on
  this machine reports `odachi v0.38.9`. `main` is that merge commit.
- The two patches it names went in with #516: `odachi start` keeps the fence it finds instead of drawing it
  again (0266), and `odachi review probe --runs 0` is refused instead of being measured as one draw (0267).
- 265 features, 643 steps, 0 open questions; `docs/review-debt.md` holds no open row.
- `## Unreleased` no longer exists: the next feature writes it again, above the newest version.

## What shipped

- **0266, the fence keeps what it protects.** `start` reads `odd/fence.toml` before writing it, keeps every
  rule as it was, appends only the non-goals the run brings, reports how many it kept, and blocks on a fence
  it cannot parse instead of overwriting it. `Decode` and `Render` live together in
  `internal/adapters/fencetoml`, so the writer and the gate share one grammar. Proved with the installed
  binary in a scratch project: three runs, a rule added by hand between them, `kept 3 rules it already had
  and added 1`, and the hand-written rule still there.
- **0267, zero draws is refused.** The row 0265's review left on `probe.go` is paid; the range test iterates
  `{-1, 0, MaxRuns + 1}` and was observed red on the zero.

## The questions that wait

- **None open.** The release question this file carried is answered: `v0.38.9` exists and the chain in
  `docs/releases.md` was followed (rename the heading only, PR, merge, annotated tag).
- **One approval is one draw.** The final review of the closing tree approved it with no findings, and an
  earlier draw of the same code produced a causal blocker that turned out to be real. The lens is a sample,
  not a proof that nothing is left (0265).

## The limits to remember

- **The store can block a verdict.** A causal BLOCKER or CRITICAL row stays potent while its blob is
  unchanged, and every new verdict inherits it. `docs/review-debt.md` is the list a human reads; when a row
  is missing there, the store still holds it -that is how 0266 found 0265's row.
- **One tree, one verdict.** A second `odachi review --base <sha>` for a tree that already has a verdict is
  refused: the verdict is content-addressed by tree, not by base. When two features close on one branch,
  review with the older feature's base.
- **A docs-only change owes no new proof.** `check` says "the proof belongs to the code it ran on, and only
  documents moved since", which is why closing documents commit without another 700-second run.
- **A document that describes the past goes stale in one release.** This file named a PR that was already
  merged and a release that had already shipped; the same happened at the close of 0265 with "Nothing is
  owed". Rewrite it when the facts move, not only when the session ends.
- **A light feature's requirement cannot start with a dash**: `odachi start --light -- "un --runs 0 …"` is
  the shape that works.
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
# the release, after the merge it names: heading only, an annotated tag, and the binary rebuilt
go build -trimpath -ldflags "-s -w -X main.version=vX.Y.Z" -o %LOCALAPPDATA%\odachi\bin\odachi.exe ./cmd/odachi
```
