# Next session

Written while closing 0266 and 0267. One thing waits, and it is the human's: the release.

## Where the project is

- `v0.38.8` is the published and installed binary; `## Unreleased` in `CHANGELOG.md` holds two patches
  (0266 the fence merge, 0267 zero draws) that no tag names yet.
- The work is on `feat/0266-la-cerca-se-fusiona`, five commits, no pull request yet; `main` does not have it.
- 265 features, 643 steps, 0 open questions; `docs/review-debt.md` holds no open row.
- The closing tree carries `proof/0267-un-runs-0.txt` and one verdict,
  `.git/odachi/reviews/4e20476….json` (reliability, base `dac1e1b`, no findings).

## What this session did

- **0266, the fence keeps what it protects** (`## Unreleased`, patch): `start` reads `odd/fence.toml` before
  writing it, keeps every rule as it was, appends only the non-goals the run brings, reports how many it
  kept, and blocks on a fence it cannot parse instead of overwriting it. The rules are read and written by
  one adapter, `internal/adapters/fencetoml`, so a file a run writes is a file the gate can read. Reading
  the code found one defect of its own -the fence was merged after the brief and the profile were already
  replaced, so a blocked run left the human's files changed- and it is paid, with a test that pins it.
- **0267, zero draws is refused** (patch): the row 0265's review left on `probe.go` -`--runs 0` silently
  measured one draw- is paid with the test observed red on the zero.

## The questions that wait

- **The release.** `## Unreleased` names both patches; whether they ship as `v0.38.9`, and when, is the
  human's call. The chain is in `docs/releases.md`: rename the heading only, PR, merge, annotated tag.
- **One approval is one draw.** The final review of this tree approved it with no findings, and an earlier
  draw of the same code produced a causal blocker that turned out to be real. The lens is not evidence
  that nothing is left; it is a sample (0265).

## The limits to remember

- **The store can block a verdict.** A causal BLOCKER or CRITICAL row stays potent while its blob is
  unchanged, and every new verdict inherits it: `review` then refuses to store an approval that contradicts
  it. `docs/review-debt.md` is the list a human reads; when a row is missing there, the store still holds it.
- **One tree, one verdict.** A second `review --base <sha>` for a tree that already has one is refused
  ("this tree already has a verdict"): the verdict is content-addressed by tree, not by base.
- **`start --force` rewrote the fence once more** while opening 0266, before the fix was in the tree; the
  run restored it by hand (recorded as a craft decision in the task doc). With 0266 in, a run keeps it.
- **One heavy gate per machine**: `go test ./...` takes about 700 s and `check --run` writes the proof; two
  ran this session, one per closing tree.
- **A light feature's requirement cannot start with a dash**: `start --light -- "un --runs 0 …"` is the
  shape that works; `"--runs 0 …"` is read as a flag.

## The commands that work

```sh
odachi check --run "go test ./..." --timeout 1800
```

```sh
# a verdict for the tree, with the answer kept beside it
odachi review --base dac1e1b --lens reliability --model opencode-go/deepseek-v4.1-flash \
  --exec 'C:\Users\USER\AppData\Local\Temp\odachi-rev.bat {model} {prompt} C:\Users\USER\AppData\Local\Temp\review.json'
```

```sh
# the measurement: eight prompts, N draws each; zero is refused now
odachi review probe --runs 3
```
