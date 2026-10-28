# the-verdict-says
Base: 64f16214895d8892b592c94250a079281a8e4f05

## Requirements (verbatim, from the user)
R1 · "the verdict says who judged it, and the gate refuses one that names no model when the proof names one"

## Done when
- check refuses a verdict whose required lens names no model while the proof names one, and names the lens and the tree
- the refusal offers the two honest ways out: re-review naming the model that judged, or the human's waiver recorded in the task document
- a verdict that names a model different from the proof's still passes, and one that names the proof's model is still refused
- the prompt asks the answer to name the model that judged it
- a test drives each answer, observed red first

## Steps
- [x] T1 · the review prompt asks the answer to name the model that judged it   R1
- [x] T2 · check refuses a verdict whose required lens names no model while the proof names one, and offers the waiver   R1
- [x] T3 · the start help names the waiver lane it already writes   R1

## History
L1 · "the verdict says who judged it, and the gate refuses one that names no model when the proof names one"
L2 · OFICIO · "a gate refuses a silence, and the way out is a decision on the record, never an omission" (porque the identity gate compared two declarations and a verdict that named none passed by saying less; the two roads out are a named reviewer or the human's waiver, which the help now lists)
L3 · OFICIO · "the refusal names the verdict's file, not only the lens and the model" (porque the review of this feature read the Done-when and found the first version naming the lens and the model but not the tree, so a project with two trees could not tell which review it read)
