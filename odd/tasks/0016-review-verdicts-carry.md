# review-verdicts-carry

## Requirements (verbatim, from the user)
R1 · "review verdicts carry per-finding content receipts (path, blob, lens) so correcting a candidate re-reviews only what changed, never weakening the gate"

## Done when
- a corrected candidate yields a verdict that inherits potent findings from unchanged blobs and re-runs only the changed files' path-class lenses
- a deterministic floor selects the lens class and the cited lines before any LLM runs
- verdicts written before blob keys grant no coverage and the gate says so instead of reusing them

## Steps
- [ ] T1 · verdict findings keyed by path, blob, lens, status; the tool derives the blob from git   R1
- [ ] T2 · potency and coverage in the gate: inherit from matching blobs; old verdicts grant no coverage   R1
- [ ] T3 · deterministic floor picks the lens class and the cited lines before any LLM   R1
- [ ] T4 · dismissals: reason plus fingerprint, never re-raised in later candidates   R1
- [ ] T5 · full-candidate diffs use merge-base instead of the two-dot diff   R1
- [ ] T6 · docs: CHANGELOG, impact map and dogfood of the new verdict flow   R1

## History
L1 · "review verdicts carry per-finding content receipts (path, blob, lens) so correcting a candidate re-reviews only what changed, never weakening the gate"
