# Next session

Written after `v0.38.11` shipped. Nothing is owed.

## Where the project is

- `v0.38.11` is published: annotated tag on `2dc4123` (`Merge pull request #523`), sixteen assets, and the
  binary installed on this machine reports `odachi v0.38.11`. `main` is that merge commit.
- The patch it names came in with #522: `odachi start --light` takes a requirement that starts with a dash,
  and a step may start with a dash because the rule now refuses the markdown bullet (0270).
- 268 features, 648 steps, 0 open questions; `docs/review-debt.md` holds no open row.
- The closing tree carries `proof/0270-el-requisito-con-guion.txt` and one verdict
  (`.git/odachi/reviews/1508cb8….json`, reliability, base `b6975a7`, no findings).
- `## Unreleased` does not exist: the next feature writes it again, above `## v0.38.11`.

## What shipped

- **0270, a requirement may start with a dash.** Opening a feature about `--runs 0` failed twice, both times
  for the wrong reason: without Go's separator the message was the flag package's `invalid flag`, and with
  it, `a step does not start with a dash` came from a rule written to stop a pasted `- [ ] T1 …`. The rule
  now refuses the bullet and says "bullet"; `start --light` names the form a line that starts with a dash
  needs. Two reds were observed in the domain and one in the CLI.

## The questions that wait

- **None open.** The next feature writes the new `## Unreleased`.
- **The brief is per feature, not per project.** `brief.md` is replaced by every `start --force` and holds
  the current requirement; the project's doctrine calls it the vision. That tension is still the oldest
  open question after the fence, and it needs the human's word before code (0270 was the papercut, not it).
- **One approval is one draw.** The lens approved the closing tree with no findings, and earlier draws of
  similar trees produced real blockers. A verdict is a sample, not proof that nothing is left (0265).

## The limits to remember

- **The environment is part of the red.** Four installer tests failed identically on `main` and passed in
  CI because a child Windows PowerShell inherits `PSModulePath`; measure with a probe before blaming the
  change (0269).
- **The store can block a verdict.** A causal BLOCKER or CRITICAL row stays potent while its blob is
  unchanged, and every new verdict inherits it. `docs/review-debt.md` is the list a human reads.
- **One tree, one verdict**, and a wider base covers a later feature, so review with the oldest feature's
  base when two close together (0268).
- **A docs-only change owes no new proof**: `check` says "the proof belongs to the code it ran on, and only
  documents moved since".
- **This file goes stale in one release.** Three releases today, three rewrites; write it when the facts
  move, not only when the session ends.
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
# a requirement that starts with a flag: the separator is part of the call
odachi start --light -- "--runs 0 se rechaza"
```

```sh
# the release, after the merge it names: heading only, an annotated tag, the binary rebuilt
go build -trimpath -ldflags "-s -w -X main.version=vX.Y.Z" -o %LOCALAPPDATA%\odachi\bin\odachi.exe ./cmd/odachi
```
