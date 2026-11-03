# the-diff-speaks

## What must exist
- el primer commit lleva `Odachi-Change: <una línea>`; check y review la leen del rango desde el merge-base

## Done when
- a branch whose commits carry `Odachi-Change: <line>` is judged against that line, with the base taken from the merge-base with the project's default branch, and no task document is written
- the gate demands the proof for that change, and the review when the diff asks for one; a proof that does not cover the line is named as it is for a document
- `odachi review` builds the same prompt for a document-less change, with the trailer line where a light document's line goes, and the prompt says the diff is the record
- a change with no trailer and no open document keeps today's answer: the newest document is the feature, and a project that never writes a trailer sees nothing new
- a trailer that is empty, appears twice with different lines, sits on the default branch, or cannot find its merge-base is refused by name
- tests pin the read, the base, the refusals and the unchanged answer without a trailer

## Not included (first version)
- no new verb: the trailer is read by check and review, and nothing writes it
- the light lane stays: a one-line change may still write a one-step document when a person wants the record
- a stored proof is never rewritten, whatever the writer learns

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: why a change that can be read from its diff does not need a second text.
