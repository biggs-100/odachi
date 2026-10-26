# the-questions-arrive
Base: 6aeb87793d824f56e301c9947f2055d835bc53d0

## Requirements (verbatim, from the user)
R1 · "the questions arrive in the report with their id and their words, not only their count"

## Done when
- progress --json publishes each open question with its id and its words
- the human report names them from the same state, and a feature with none prints what it printed before
- the contract test pins the names id and text, and an empty list instead of null

## Steps
- [ ] T1 · the state carries each open question with its id and its words   R1
- [ ] T2 · the human report names them from the same state, and says nothing when there are none   R1

## History
L1 · "the questions arrive in the report with their id and their words, not only their count"
L2 · OFICIO · "the state carries each question as id and text, without the answer command" (porque the command is a rendering of the id; a second copy inside another program would drift from ask)
L3 · OFICIO · "the human report names the questions between the headline and the steps" (porque a question is what stops the work, and check already puts the words before the remedy; the count alone sends the human to another file)
