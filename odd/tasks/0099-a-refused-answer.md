# a-refused-answer
Base: e79b7ea274b4218f8a4789ba1983d689542fc0bc

## Requirements (verbatim, from the user)
R1 · "a refused answer writes nothing: the scope record is written only when the verdict it explains is written, the way gentle-ai creates no authority for a refusal"

## Done when
- an answer refused for its declaration or its coverage leaves no scope record, and the store keeps only the refused answer
- the record still lands when the verdict lands, with the same fields it had
- a test drives the refusal, and it fails without the move

## Steps
- [x] T1 · a refused answer writes no scope record, because the record follows the verdict   R1 · tree:ef387c31a99d06d5fffa095bd506fa4b83460726
- [ ] T2 · the changelog names the defect and what the fix changes   R1

## History
L1 · "a refused answer writes nothing: the scope record is written only when the verdict it explains is written, the way gentle-ai creates no authority for a refusal"
L2 · OFICIO · "the human delegated the design to the gentle-ai comparison, and the comparison decides the shape: gentle-ai writes nothing for a refusal and freezes the escalation with the authority, so odachi writes no scope record without the verdict it explains" (porque in odachi the verdict is the authority; a record beside a verdict that does not exist is the one thing gentle-ai never allows, and a refusal that mutates is a refusal that lies)
