# Next session: the release, and a sweep that cannot clean itself

Written at the close of 0187. The feature is committed on `feat/0187-the-store-prunes`, not pushed;
`main` is clean. The installed binary is still 0.14.5, so the new rule is proven by the suite and
not yet by the tool on this machine.

## What shipped

- **0187, unreleased**: a stored verdict disposes of the prompts its `keep_prompts` stopped
  keeping, by the same domain answer `odachi clean` reports; the tree that closes a feature names
  the store's size and what `clean --apply` would free, without deleting; and ADR 0049 records the
  split between disposable and evidence, with `proof/` and `odd/tasks` never bounded.

## The lessons this session paid for

**The policy was asked in two places.** Which prompts survive was a question only `odachi clean`
could answer; the write that stores a verdict had to answer it too. It now lives once, in the
domain that owns the store's names: the trees in order, the files that are questions, and the ones
a retention forgets. The test pins keep 0, keep 1, and the edge that a tree with no prompt still
spends a place.

**A measurement that cannot read is not an empty store.** The close says `cannot free anything: N
paths could not be read`, and the sweep keeps every prompt and names the path it could not read.
The strict side is deliberate: a leftover costs a file, a guess costs a paid answer.

**The installed binary judged this feature.** `odachi review` ran with 0.14.5, so the prompt of
0187 is still in the store while `keep_prompts = 0`: the defect the feature removes, seen live. The
suite pins the write; the release is what makes the tool do it here.

## What was measured, so nobody re-measures it

- `go test ./...` green on the closing tree; the proof is `proof/0187-the-store-prunes.txt`.
- The store at the close: 34.9 MB in 1352 files, 34.0 MB recoverable, 31.1 MB of it scratch.
- Eight paths under `.git/odachi/scratch/lane-smoke` and `probe-upstream` are deeper than the
  walk's six levels.
- The review: approved by `mimo-v2.6-flash` (reliability), one suggestion open and recorded in
  `docs/review-debt.md`.

## Queued

1. **Release v0.14.6 and install it.** The changelog's `Unreleased` is the release body; the
   installed 0.14.5 is the only reason the new prune has not run here.
2. **The dead end the close found.** `odachi clean --apply` refuses to remove anything while a
   cleanable class holds an unreadable path, and the nested repositories under `.git/odachi/scratch`
   (0182's probes) are exactly that: eight paths deeper than the walk. The switch cannot remove the
   folder that blocks it. A decision is needed: a deeper walk with a bound, or naming the unreadable
   path and letting the rest of the class go (which weakens 0125), or making the nested-repo shape a
   class of its own. It is a policy decision, not a bug fix.
3. **Before any `clean --apply`**: `light-base-probe`, `probe-upstream` and the 0182 build under
   scratch are what the last session kept; move anything still wanted to `docs/`.

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

## How to find the next real work

`docs/review-debt.md` has one open row (the two untested failure branches of the sweep). The clean
dead end above is the map today.
