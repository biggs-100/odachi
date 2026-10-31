# 0042 - A correction is the delta, and the gate lives where the evidence does

Status: accepted — Date: 2026-10 — Session: after v0.11.80

## Context

Two findings from comparing the review with gentle-ai, both measured on this
repository.

**A fixed finding costs the whole candidate.** A verdict is named by the tree it
judged, so a fix makes a new tree and the next review sends the whole diff again.
0139 needed five rounds: the first prompt was 64 KB and each correction re-sent it
for a handful of changed lines. gentle-ai answers this with one bounded correction
and a targeted validation of just the fix; it costs a state machine -lineage,
revision, tokens, burn- that this project deliberately does not have.

**The review gate is local, and the documentation did not say how local.** A first
proposal was to make CI check the review by reading it from the proof. It cannot
work, and the reason is structural rather than an oversight: `proof/` is in
`.gitignore` and holds no tracked files, so CI never sees the local proof; and even
if it did, evidence about a tree cannot live inside that tree -committing the
verdict changes the tree the verdict names, and the gate asks again-. gentle-ai does
not solve this either: it declares that review never authorises delivery.

## Decision

- **`review --since <tree>` reviews the delta.** The prompt is built from the paths
  and the diff between the tree that verdict judged and the worktree, through a
  `Correction` read on the repository port: one throwaway index, no merge-base,
  because the tree is a fact and not a branch.
- **The prompt says it is a correction** -the header names the tree, and a
  paragraph says the diff is only what moved and the rest is already answered- so a
  reviewer handed a delta does not report the rest of the candidate as missing.
- **The stored verdict carries the predecessor's rows and its base.** A correction is
  sealed as the union of the verdict it completes and its own answer, so every path
  of the feature stays answered and the gate's coverage keeps its meaning. A
  correction with nothing to judge is refused, naming the read that already holds the
  answer.
- **`check` names the review debt a close leaves**, reading the open rows of
  `docs/review-debt.md`: the file a person reads before choosing the next feature is
  named at the moment the feature closes, instead of trusting memory.
- **The gate stays in the clone, and the docs say so.** The README and the review
  section name the store as the place the gate lives. Making CI enforce the review
  needs evidence that travels *and* proves it was not written by the candidate: a
  trusted job that runs `odachi review` itself, or a signed receipt outside the tree.
  Neither is this feature, and neither is claimed.

## What this does not do

- **No state machine.** A correction is one more tree in the same store; there is no
  lineage, no revision, no acknowledgement, no burn. The predecessor is found by the
  tree the human names, and the honest failure is a refusal that says so.
- **No automatic correction.** The tool does not notice that a fix follows a
  blocker: the human runs `--since` with the tree `review --show` names.
- **No change to the proof format**, and the fence keeps protecting it.
- **The review is still not enforced by CI.** A contributor can skip the hook, and a
  forged record would be a lie visible in the diff rather than prevented.

## Consequences

- A fix after a blocker costs the delta plus one prompt header, and the verdict it
  produces answers exactly what the previous one answered plus what moved.
- Two readings of one feature now exist: the whole candidate (`--base`) and the
  correction (`--since`). Both store under the tree they judge, so the store stays
  one file per tree.
- The structural finding is written down: evidence about a tree cannot live inside
  that tree. That is why gentle-ai separates review from delivery, and why this
  project's gate runs where the store is.
