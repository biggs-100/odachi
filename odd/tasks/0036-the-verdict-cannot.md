# the-verdict-cannot

## Requirements (verbatim, from the user)
R1 · "the verdict cannot be from the model that wrote the tree: the proof and the verdict record who produced them, and a self-review is refused with the command that fixes it"

## Done when
- a proof run by an agent session records the model it ran under, and an old proof without that line still parses
- a verdict records the model that judged, and an old verdict without it still parses
- check --require-review refuses a verdict whose model is the one in the proof, and names review --model
- the README says this is disclosure, not identity: a model name can be lied about

## Steps
- [x] T1 · a proof records the model its run happened under, and an older proof still parses   R1 · 3531319
- [x] T2 · a verdict records the model that judged, and check refuses a self-review naming the fix   R1 · d5462bf
- [ ] T3 · the README says what this proves and what it does not, and a test pins it   R1

## History
L1 · "the verdict cannot be from the model that wrote the tree: the proof and the verdict record who produced them, and a self-review is refused with the command that fixes it"
L2 · OFICIO · "The identity is a field on both files, not text parsed out of the reviewer command" (porque a command is a string a human edits; comparing two declared identities is something a gate can do without guessing)
