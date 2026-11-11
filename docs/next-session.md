# Next session

Written after `v0.38.10` shipped. Nothing is owed.

## Where the project is

- `v0.38.10` is published: annotated tag on `358b78c` (`Merge pull request #520`), sixteen assets, and the
  binary installed on this machine reports `odachi v0.38.10`. `main` is that merge commit.
- Two patches name it: `odachi check` no longer calls a verdict judged from an earlier base "another diff"
  (0268), and `scripts/install.ps1` extracts the release without `Expand-Archive` (0269).
- 267 features, 646 steps, 0 open questions; `docs/review-debt.md` holds no open row.
- The closing tree carries `proof/0269-install-ps1-no.txt` and one verdict
  (`.git/odachi/reviews/3bfb0f4….json`, reliability and resilience, base `0d9293c`, no findings).
- `## Unreleased` does not exist: the next feature writes it again, above `## v0.38.10`.

## What shipped

- **0268, a base that came before is not another diff.** `check` compared the verdict's base with the
  feature's by equality, so a review that judged a wider diff read as "judged something else": the run said
  `ready` and, in the same output, `the feature is closed and its tree has no verdict`, naming a
  `review --base` the store answers with "this tree already has a verdict". The gate now asks git whether
  that base came before -earlier or equal covers, later starts inside the feature and is still refused- and
  the store's search asks the same question, so a close that only moved documents keeps its verdict. The
  port gained `Ancestor` for it.
- **0269, the installer stops trusting a cmdlet.** Four Windows-only installer tests were red on a machine
  whose Windows PowerShell 5.1 inherits PowerShell 7's `PSModulePath`: the modern `Archive` module's
  `-Force` calls `ExtractToDirectory` with the three-argument overload .NET Framework lacks, and the
  installer died before verifying anything. `install.ps1` opens the zip with `ZipFile` now.

## The questions that wait

- **None open.** The next feature writes the new `## Unreleased`.
- **One approval is one draw**, and one fix is one sample: the lens approved this tree with no findings,
  after an earlier draw of a similar tree produced a causal blocker that was real (0265).

## The limits to remember

- **The environment is part of the red.** Four installer tests failed identically on `main` and passed in
  CI: a child Windows PowerShell inherits `PSModulePath`, so which `Expand-Archive` it loads depends on who
  started it. A red that does not reproduce in CI is a machine fact until proved otherwise; measure it with
  a probe before believing the change broke it.
- **The store can block a verdict.** A causal BLOCKER or CRITICAL row stays potent while its blob is
  unchanged and every new verdict inherits it. `docs/review-debt.md` is what a human reads; when a row is
  missing there, the store still holds it.
- **One tree, one verdict.** Review with the oldest feature's base when two features close on one branch;
  with 0268 in, a wider verdict covers the later feature, and the exact-base complaint is gone.
- **A docs-only change owes no new proof**, and `check` says so: "the proof belongs to the code it ran on,
  and only documents moved since".
- **This file goes stale in one release.** It has now been rewritten twice in a day for that reason; write
  it when the facts move, not only when the session ends.
- **Two features in one pull request is a last resort** and it has one justification: the second is what
  makes the first provable here. Write that in the body, or split.
- **One heavy gate per machine**: `go test ./...` takes about 700 s and `check --run` writes the proof.

## The commands that work

```sh
odachi check --run "go test ./..." --timeout 1800
```

```sh
# a verdict for the tree, with the answer kept beside it
odachi review --base <oldest feature base> --lens reliability,resilience \
  --because "<why the extra lens>" --model opencode-go/deepseek-v4.1-flash \
  --exec 'C:\Users\USER\AppData\Local\Temp\odachi-rev.bat {model} {prompt} C:\Users\USER\AppData\Local\Temp\review.json'
```

```sh
# the release, after the merge it names: heading only, an annotated tag, the binary rebuilt
go build -trimpath -ldflags "-s -w -X main.version=vX.Y.Z" -o %LOCALAPPDATA%\odachi\bin\odachi.exe ./cmd/odachi
```
