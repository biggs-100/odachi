# review-verdicts-carry

## Requirements (verbatim, from the user)
R1 · "review verdicts carry per-finding content receipts (path, blob, lens) so correcting a candidate re-reviews only what changed, never weakening the gate"

## Done when
- a corrected candidate yields a verdict that inherits potent findings from unchanged blobs and re-runs only the changed files' path-class lenses
- a deterministic floor selects the lens class and the cited lines before any LLM runs
- verdicts written before blob keys grant no coverage and the gate says so instead of reusing them

## Steps
- [x] T1 · verdict findings keyed by path, blob, lens, status; the tool derives the blob from git   R1 · 7400207
- [x] T2 · potency and coverage in the gate: inherit from matching blobs; old verdicts grant no coverage   R1 · 9596b28
- [x] T3 · deterministic floor picks the lens class and the cited lines before any LLM   R1 · 21687cd
- [ ] T4 · dismissals: reason plus fingerprint, never re-raised in later candidates   R1
- [ ] T5 · full-candidate diffs use merge-base instead of the two-dot diff   R1
- [ ] T6 · docs: CHANGELOG, impact map and dogfood of the new verdict flow   R1

## History
L1 · "review verdicts carry per-finding content receipts (path, blob, lens) so correcting a candidate re-reviews only what changed, never weakening the gate"
L2 · OFICIO · "Deleted paths owe no coverage: the gate drops rows whose blob measures empty" (porque A receipt about absence is a fact about nothing; the ADR demands an answer for every changed blob, and a deleted path has no blob)
L3 · OFICIO · "A pass without a blob validates: it is the reviewer's fresh answer awaiting the tool's stamp, inert until stamped" (porque The stamp is derived like the finding blob; refusing blob-less passes would make every fresh answer invalid)
L4 · OFICIO · "Inheritance is survival-biased: only standing findings travel; findings refuted with evidence die with their verdict" (porque Refutations are index-based (F1) and judge one verdict's findings; grandchildren inheritance would launder killed findings through new trees)
L5 · OFICIO · "Coverage demand is per file (Need), not per tree tier: one file may ask reliability while another asks all four" (porque ADR 0016 makes assignment per file even though selection stays path-based; a tree-wide demand would over-review docs the verdict already covers)
L6 · OFICIO · "The floor reports the path class; it does not change lens selection, which the fence keeps in ADR 0010" (porque The brief fence forbids changes to the lens selection rules; the deterministic class is the one the lenses already followed, reported with its signals)
L7 · OFICIO · "Earned blocking downgrades only causal BLOCKER/CRITICAL findings that cannot cite the added lines" (porque A WARNING never blocked and a non-causal finding is a follow-up by definition; downgrading them removes information without strengthening the gate)
L8 · OFICIO · "A floor with no readable rows judges nothing" (porque Silence is not evidence of an uncited finding: a binary diff or a fixture must not silently demote a real blocker)
L9 · OFICIO · "The audit set is the diff's added lines; context lines belong to nobody's citation" (porque The candidate's writing is what it wrote; the hunk's context is the past, and FloorContext=3 already covers a signature and its caller)
