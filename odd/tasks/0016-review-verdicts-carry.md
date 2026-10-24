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
- [x] T3 · deterministic floor picks the lens class and the cited lines before any LLM   R1 · ac4b744
- [x] T4 · dismissals: reason plus fingerprint, never re-raised in later candidates   R1 · c4011a3
- [x] T5 · full-candidate diffs use merge-base instead of the two-dot diff   R1 · 33f80c3
- [x] T6 · docs: CHANGELOG, impact map and dogfood of the new verdict flow   R1 · 6412a80

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
L10 · OFICIO · "A refuted finding travels as a dismissed row with its reason and fingerprint, and a re-raise with the same fingerprint stays dismissed" (porque The ADR wants a review cumulative inside a feature: a judged defect must not be re-litigated by the next candidate, and the fingerprint makes that identity deterministic rather than editorial)
L11 · OFICIO · "The fingerprint covers path, line, lens and normalized evidence, not the blob" (porque The blob already gates potency: while the bytes are unchanged the line cannot drift, and once they change the judgment is over and a re-raise is legitimate)
L12 · OFICIO · "A reason is only valid on a dismissed row and the status vocabulary is closed to open|dismissed" (porque A dismissal without evidence is silence, and a reason on an open finding is a note pretending to be a judgment)
L13 · OFICIO · "A dismissed row still answers coverage and never blocks or downgrades" (porque The dismissal is the current answer for its (path, blob, lens) row; its severity is history, not a live claim)
L14 · OFICIO · "Diff and ChangedPaths compare against merge-base(HEAD, base), not the base tip" (porque Two dots turn every commit the base gained after the fork into a deletion, so a branch appears to delete code it never touched (the promptfoo bug); the fork point is the candidate's parent)
L15 · OFICIO · "An empty base or the empty tree stays the first day: no merge-base, everything is new" (porque There is no fork to find before the first commit, and the existing behavior must keep holding for a project with no history)
L16 · OFICIO · "AddedDeps still reads go.mod at the base tip, not at the merge-base" (porque The fence asks which dependency this candidate would introduce into the base, and a module the base removed but the branch still uses does arrive in the merge; the diff is the authored change, the dependency is a merge consequence)
L17 · OFICIO · "T6 documents the whole feature in one Unreleased block and updates the impact map in the same commit, and dogfoods the coverage gate on the T5 commit with a real attached verdict" (porque Docs that land apart from the code they describe drift, and a docs-only worktree is a passive candidate: the real gate has to be exercised on a code candidate (--base 33f80c3^) or the dogfood proves nothing)
