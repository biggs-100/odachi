# Next session: a sweep that cannot clean itself

Written at the close of 0187. `main` is clean, v0.14.6 is published and installed, and `doctor` is
14/14.

## What shipped

- **0187, v0.14.6**: a stored verdict disposes of the prompts its `keep_prompts` stopped keeping,
  by the same domain answer `odachi clean` reports; the tree that closes a feature names the store's
  size and what `clean --apply` would free, without deleting; and ADR 0049 records the split between
  disposable and evidence, with `proof/` and `odd/tasks` never bounded.

## The lessons this session paid for

**The policy was asked in two places.** Which prompts survive was a question only `odachi clean`
could answer; the write that stores a verdict had to answer it too. It now lives once, in the
domain that owns the store's names: the trees in order, the files that are questions, and the ones
a retention forgets. The test pins keep 0, keep 1, and the edge that a tree with no prompt still
spends a place.

**A measurement that cannot read is not an empty store.** The close says `cannot free anything: N
paths could not be read`, and the sweep keeps every prompt and names the path it could not read.
The strict side is deliberate: a leftover costs a file, a guess costs a paid answer.

**An old binary does not obey the rule it is about to gain.** 0.14.5 judged and stored 0187, so the
feature's own prompt is still in the store under `keep_prompts = 0`; the installed 0.14.6 prunes the
next review it stores. Until then, that prompt is the defect the feature removed, kept by the binary
that could not see it.

## What was measured, so nobody re-measures it

- `go test ./...` green on the closing tree; the proof is `proof/0187-the-store-prunes.txt`.
- The store at the close: 34.9 MB in 1352 files, 34.0 MB recoverable, 31.1 MB of it scratch.
- Eight paths under `.git/odachi/scratch/lane-smoke` and `probe-upstream` are deeper than the
  walk's six levels.
- The review: approved by `mimo-v2.6-flash` (reliability), one suggestion open and recorded in
  `docs/review-debt.md`.
- The documented Windows install (`irm .../install.ps1 | iex`) prints a spurious
  `?#Requires` error: the script's UTF-8 BOM reaches the pipe as a character. The install completes;
  the noise is the defect.

## Queued

1. **The dead end the close found.** `odachi clean --apply` refuses to remove anything while a
   cleanable class holds an unreadable path, and the nested repositories under `.git/odachi/scratch`
   are exactly that: eight paths deeper than the walk. The switch cannot remove the folder that
   blocks it. A decision is needed: a deeper walk with a bound, or naming the unreadable path and
   letting the rest of the class go (which weakens 0125), or making the nested-repo shape a class of
   its own. It is a policy decision, not a bug fix.
2. **Before any `clean --apply`**: `light-base-probe`, `probe-upstream` and the 0182 build under
   scratch are what the last session kept; move anything still wanted to `docs/`.
3. **The installer's first line.** `install.ps1` opens with a `#Requires` the documented
   `irm | iex` cannot read because of the BOM. A project whose documented install starts with an
   error teaches the reader to ignore errors; strip it or read the script another way.

## What is already true, so it is not re-decided

- A stored verdict obeys `keep_prompts` at write time; a run that gets no verdict prunes nothing; a
  store the run cannot read prunes nothing and says so.
- The close names the growth and never deletes: the switch stays `clean --apply`, and the line
  changes no answer or exit code.
- ADR 0049: `proof/` and `odd/tasks` are evidence odachi never bounds; a new class earns a lifetime
  by saying who reads it and what is lost when it goes.
- The retention numbers are the project's (`odd/review.toml`: `keep_verdicts = 10`,
  `keep_prompts = 0`), and the fence protects that file from a feature that only obeys it.
- `main` is protected: four checks, everything through a pull request.
- **v0.14.6 is published, signed and installed**; the home carries the agents and the skills.

## How to find the next real work

`docs/review-debt.md` has one open row (the two untested failure branches of the sweep). The clean
dead end above is the map today.
