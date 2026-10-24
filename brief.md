# review-verdicts-carry

## What must exist
- review verdicts carry per-finding content receipts (path, blob, lens) so correcting a candidate re-reviews only what changed, never weakening the gate

## Done when
- a corrected candidate yields a verdict that inherits potent findings from unchanged blobs and re-runs only the changed files' path-class lenses
- a deterministic floor selects the lens class and the cited lines before any LLM runs
- verdicts written before blob keys grant no coverage and the gate says so instead of reusing them

## Not included (first version)
- signing or trust ladders
- multi-model consensus
- review-at-merge phase
- changes to lens selection rules or the human switch

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: no rush.
- Wants to learn: how to give an LLM verdict content identity while keeping the tree-hash receipt honest.
