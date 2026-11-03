# the-diff-speaks
Base: 5a5662cc9c0576ea921d81afb0170ccf01820b09

## Requirements (verbatim, from the user)
R1 · "el primer commit lleva `Odachi-Change: <una línea>`; check y review la leen del rango desde el merge-base"

## Done when
- a branch whose commits carry `Odachi-Change: <line>` is judged against that line, with the base taken from the merge-base with the project's default branch, and no task document is written
- the gate demands the proof for that change, and the review when the diff asks for one; a proof that does not cover the line is named as it is for a document
- `odachi review` builds the same prompt for a document-less change, with the trailer line where a light document's line goes, and the prompt says the diff is the record
- a change with no trailer and no open document keeps today's answer: the newest document is the feature, and a project that never writes a trailer sees nothing new
- a trailer that is empty, appears twice with different lines, sits on the default branch, or cannot find its merge-base is refused by name
- tests pin the read, the base, the refusals and the unchanged answer without a trailer

## Steps
- [x] T1 · the commits since the merge-base name the change: a trailer is read, and its refusals are named   R1 · route:inline · tree:333871c7054bc8f6834b5e9644404385adcf5b7a
- [ ] T2 · the gate judges a document-less change against its trailer, with the proof it already owes   R1 · route:inline
- [ ] T3 · the review's vision for a document-less change is the trailer line, and the prompt says the diff is the record   R1 · route:inline
- [ ] T4 · the drive skill and the changelog name the lane, so an agent knows it exists   R1 · route:inline

## History
L1 · "el primer commit lleva `Odachi-Change: <una línea>`; check y review la leen del rango desde el merge-base"
