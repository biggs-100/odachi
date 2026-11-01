# a-second-reading-or-a-second-opinion
Base: 3b4baeb35e722a21c0fd06a15ae782606b7e1967

## Requirements (verbatim, from the user)
R1 · "the gate names when the verdict came from the model that wrote the proof instead of refusing it by default, refuses only when the run asks for another model, and a human who owns one model can record that decision on an open feature, with its reason"

## Done when
- a verdict by the proof's model leaves the tree ready and says what it is, and the same tree under --require-review is refused naming the model -both observed in a test
- the comparison ignores the variant, so provider/name#high and provider/name are the same model, and the message names the effort each one ran with
- odachi review waive --because "<why>" writes Review: waived into the open feature's task doc, keeps its history, and check reports the waiver instead of a missing verdict
- an ADR says why the rule names instead of refusing, and what a second reading with the same model does and does not buy

## Steps
- [ ] T1 · the comparison ignores the variant, so the same model at another effort is still the same model   R1
- [ ] T2 · a verdict from the builder's model is named by default and refused only when another model is asked for   R1
- [ ] T3 · a human can waive an open feature's review with its reason, and check says so   R1
- [ ] T4 · the ADR says what a second reading buys and what it does not   R1

## History
L1 · "the gate names when the verdict came from the model that wrote the proof instead of refusing it by default, refuses only when the run asks for another model, and a human who owns one model can record that decision on an open feature, with its reason"
