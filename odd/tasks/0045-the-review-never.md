# the-review-never
Base: 261ee6266f9bab916a85002cc90709f6f62df326

## Requirements (verbatim, from the user)
R1 · "the review never seals a verdict that covers less than the tree, and names the rows it is missing"

## Done when
- an answer that misses a row writes no verdict and names the missing file and lens
- a --lens run still stores: its gaps are legitimate and the store inherits them
- an answer that covers every row writes the verdict exactly as today
- tests pin the refusal, the stored partial run and the silence

## Steps
- [x] T1 · a full answer with a gap writes nothing and names the missing path and lens   R1 · 027eabe
- [x] T2 · a --lens partial run still stores, and an answer that covers everything is untouched   R1 · b3c5057

## History
L1 · "the review never seals a verdict that covers less than the tree, and names the rows it is missing"
