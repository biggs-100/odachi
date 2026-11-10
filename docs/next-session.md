# Next session

Written while closing 0263. Nothing is owed.

## Where the project is

- `v0.38.6` is published (tag on `342117c`); the session shipped `v0.38.5` (0262) and `v0.38.6` (0263),
  and `main` carries both task docs.
- The installed binary is `0.38.6`; `doctor` is all clear with one warning, this file, which this run pays.
- 261 features, 636 steps, 0 open questions, and `docs/review-debt.md` is empty.
- `v0.38.6` renamed `Unreleased`: the next feature writes under it again.

## What this session did

- **0262, the two trees meet** (`v0.38.5`): `review` and `check` name one tree for one worktree. The
  issue blamed an ignored folder; the cause was `read-tree <base>`, which keeps the index entry of a path
  a later commit untracked, so `add --all` staged it again although git ignored it. `stageWorktree` reads
  the index from HEAD for `TreeHash`, `Candidate` and `Diff`. RED observed first; consequence recorded:
  a correction now reports such a path as a deletion.
- **The research** (`docs/review-research.md`): six sources with sample sizes (Harvard/Jellyfish, Sonar,
  DORA, arXiv 2603.15911, METR, Morgan Stanley), eight failure classes, the coverage matrix, and our own
  numbers -15 of 16 kept verdicts approved, which is the field's signature of a rubber stamp and was the
  argument for measuring.
- **0263, the lenses are measured** (`v0.38.6`): `odachi review probe` asks the four shipping prompts
  about a fixture with five planted defects and a clean arm, and scores found, low, missed and noise
  against a written key. First run: **5 of 5** planted defects found, and **10 findings on the clean
  arm**, every one of them real. Both arms are asked; the answers live in `.git/odachi/probe/`, never in
  the review store.
- **The review of 0263 found five defects and all were paid**: the probe never measured the clean arm
  (the half of the requirement), the fixture told the reviewer it was an exam, `causal: false` was
  counted as both miss and noise, two runs of one tree overwrote each other, and the summary printed
  twice. Two correction rounds on the delta, then approved with no findings.
- **The fence**: `start --force` rewrote `odd/fence.toml` from the new brief and dropped the rule that
  protects `proof/**`. The archtest caught it, the rule is back, and the decision is recorded. A fence
  outlives the feature that wrote it.

## The questions that wait

- **The clean arm is not clean.** Ten findings on code with nothing planted, all real: the handler
  answers every database error with 404, it never checks the remote status, the response write error is
  discarded, `limit` does not clamp negatives, and the handler and the cache have no tests. Until that
  arm is code a person would approve, the number that reads as noise is the fixture's sloppiness, not
  the lens's precision. This is the next feature on the probe, and it is the fixture's job.
- **A verdict cannot stop a close.** `check` says ready while a `changes_requested` verdict stands on the
  tree; today the only thing preventing a merge over a blocker is the agent's judgment. Reopen when a
  merge over a standing blocker has to be refused by the tool rather than by a person.

## The limits to remember

- **The probe's key is written by the same person who wrote the fixture.** Five defects in a 120-line
  fixture is one tree and one model, not a population; recall is measured, precision is not yet.
- **Two earlier probe runs are discarded, not data.** They came from an instrument that never asked the
  clean arm and a fixture that leaked it was an exam. The numbers 3/4 and 4/5 in `docs/review-research.md`
  history are the reason both fixes exist; do not read them as a result.
- **A review pass is one draw.** Not proven by this session's instrument, but the skill probes found the
  same shape (2, 4 and 4 edges across runs) and the review has no repetition lane yet.
- **The gate's own tools are the cheapest coverage.** No `govulncheck`, no linter with a ruleset, no
  secret scan; JetBrains puts 20-25% of AI hallucinations inside what static analysis catches. That is a
  floor row, not a lens.
- **`start --force` regenerates the fence from the brief it is given.** A rule that is doctrine and not a
  non-goal of this feature will be dropped silently; the archtest is what catches it.
- **A probe answer is not a verdict.** The store under `.git/odachi/probe/` is separate on purpose:
  writing a fixture's answer into `reviews/` would make `check` believe a tree was judged.
- **`review --since <tree>` is the cheap correction lane.** Three rounds this session, each on a delta of
  a few lines, and each merge waited for the delta to be approved.
- **The release chain**: branch `release/vX.Y.Z`, rename `## Unreleased` to the version and nothing else,
  PR, merge, then an annotated tag on the merge commit. Renaming every `## Unreleased` in the file also
  rewrites the prose that explains the section; the first attempt this session did exactly that.
- One heavy gate per machine; `go test ./...` takes about 700 s and `check --run` writes the proof.

## The commands that work

```sh
odachi check --run "go test ./..." --timeout 1800
```

```sh
# the measurement: prints eight prompts (two arms, four lenses) and runs nobody
odachi review probe
```

```sh
# the measurement with a reviewer; answers land in .git/odachi/probe/
odachi review probe --model opencode-go/deepseek-v4.1-flash \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

```sh
# no --model: each lens uses the model odd/review.toml names
odachi review --base <feature-base> --lens reliability \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```
