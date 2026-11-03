# the-store-prunes
Base: 0f43ae2dbbf5a01bec504ee374b96eeaa4f3cd13

## Requirements (verbatim, from the user)
R1 · "the store prunes what it can regenerate and names what it keeps: a stored verdict disposes of the prompts and scaffolding its keep_prompts stopped keeping, a ready close names the store size and what clean --apply would free, and one ADR says which evidence odachi never bounds"

## Done when
- with keep_prompts = 0 a stored verdict leaves no prompt of its tree in the store, and the verdict is stored
- with keep_prompts = 1 the newest tree keeps its prompt when its verdict is stored, and an older tree does not
- a review that gets no verdict removes nothing, because there is nothing stored
- a ready close prints one line with the store size and what clean --apply would free, and the gate answer and exit code do not change
- an ADR records that proof/ and odd/tasks are never bounded by odachi, and clean keeps its owner lines
- the suite pins every one of them and go test ./... is green

## Steps
- [x] T1 · the prompt policy has one answer: review owns which trees keep theirs, and clean asks it   R1 · route:inline · tree:30c49763a771f2bb06fb0b2440e7ceaecc12ce9f
- [x] T2 · a stored verdict removes the prompts beyond keep_prompts and its answer file, and a failed run removes nothing   R1 · route:inline · tree:0bed2b4065d25a8a9a174b584262387b8c22e56f
- [x] T3 · a ready close names the store growth and what clean --apply would free, and it never deletes   R1 · route:inline · tree:e443b4817f912efbe1964fb1f30fbc6c069331f9
- [x] T4 · one ADR records the cleanup policy: disposable at write, named at close, and proof/ and odd/tasks never bounded   R1 · route:inline · tree:813e0a987ebd658ba8a80002450df477948f0386

## History
L1 · "the store prunes what it can regenerate and names what it keeps: a stored verdict disposes of the prompts and scaffolding its keep_prompts stopped keeping, a ready close names the store size and what clean --apply would free, and one ADR says which evidence odachi never bounds"
