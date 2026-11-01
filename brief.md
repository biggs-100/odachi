# a-second-reading-or-a-second-opinion

## What must exist
- the gate names when the verdict came from the model that wrote the proof instead of refusing it by default, refuses only when the run asks for another model, and a human who owns one model can record that decision on an open feature, with its reason

## Done when
- a verdict by the proof's model leaves the tree ready and says what it is, and the same tree under --require-review is refused naming the model -both observed in a test
- the comparison ignores the variant, so provider/name#high and provider/name are the same model, and the message names the effort each one ran with
- odachi review waive --because "<why>" writes Review: waived into the open feature's task doc, keeps its history, and check reports the waiver instead of a missing verdict
- an ADR says why the rule names instead of refusing, and what a second reading with the same model does and does not buy

## Not included (first version)
- no change to the quorum of 0160 or to the severity vocabulary: cited falls to one refutation, argued to two, three witnesses at most
- no network and no provider query: the tool cannot discover which models exist, and a gate that calls out to answer is not a gate
- the waiver does not skip the proof or the fence, and it is not a per-run flag: it is a decision on the document, taken by the human
- the four lens prompts are not edited

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: why a fresh session with the same model is a second reading and not a second opinion.
