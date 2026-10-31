# 0043 - The review never authorises delivery

Status: accepted — Date: 2026-10 — Session: after v0.11.84

## Context

ADR 0042 recorded why the review gate lives in the clone: the verdict is evidence
about one tree, it is named after that tree's content, and evidence about a tree
cannot live inside that tree. Committing it changes the tree it names, so the gate
asks again for the new one.

That ADR left the promise standing: `mode = on` made `check` refuse a close without a
verdict, and the hook refused the commit. CI, which cannot see the store, ran
`check --no-review`. The promise was therefore larger than its mechanism: a
contributor could skip the review with `git commit --no-verify` and nothing anywhere
would know, while the README read like a gate.

The comparison with gentle-ai made the choice visible: that tool declares that review
never authorises delivery -its named gates are informational, and commit, push, PR and
release follow ordinary repository policy- and so it has no gap to close. Its review
authority is local exactly like this one; it simply does not claim delivery depends on
it.

## Decision

- **The review does not authorise delivery.** With the project's switch on, a tree
  that owes a verdict and does not have one is **ready**, and the run **names** it:
  the reminder appears in the report, and the hook -which runs `check --quiet`- prints
  it on stderr and lets the commit through.
- **`--require-review` is the only thing that refuses**, and it refuses for that run.
  A project that wants to be stopped asks to be stopped.
- **What the review found still stops a ship.** A causal `BLOCKER` in the tree's
  verdict blocks until it is refuted, with or without the flag: that is the review
  saying no, not the review being absent. The line is between evidence that exists and
  evidence that does not.
- **CI stops passing `--no-review`.** With the switch alone the run is ready and says
  what it could not verify, so every pull request carries the sentence instead of a
  flag that hid it.
- **The docs say it**: the README's review section, the `review mode` help, the switch's
  own comment, and this ADR. ADR 0042 keeps the analysis of locality and loses its
  claim of a gate.

## What this does not do

- **It does not make the review verifiable in CI.** That needs a trusted job that runs
  the review with its own model and key, or a receipt signed outside the tree. Neither
  exists, and the README names them as the way to get there.
- **It does not weaken the fence or the proof.** Those are verified from the base ref,
  because CI rebuilds the gate and runs the suite itself: they never depended on a
  local store.
- **It does not touch the four hard rules.** The review was never one of them.

## Consequences

- A contributor can close a feature without a review, and nothing refuses it. That is
  the trade, chosen on purpose: the alternative was a promise the mechanism could not
  keep.
- The moment a second person or a second machine contributes, the gap becomes real
  again; the two ways out are written down, and neither is a flag.
- `mode = on` now means "ask, and say so", not "require"; the flag keeps its old
  meaning, so a project that wants the old behaviour changes one call, not the file.
