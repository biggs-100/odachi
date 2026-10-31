# a-review-whose
Base: 02736c2061ee7d351f9d0f4fab3f3a2b15b71ce1

## Requirements (verbatim, from the user)
R1 · "a review whose prompt does not fit the project's limit is split into parts that do, and one verdict answers the tree"

## Done when
- a candidate whose prompt exceeds max_prompt_kb is sent as parts under the limit, and the store holds one verdict for the tree
- every changed path is answered by exactly one part, and the verdict's coverage is complete
- a single file whose diff alone exceeds the limit is refused, naming the file
- a candidate under the limit runs as before: one reviewer, one prompt, nothing extra written
- the prompt says which part it is, and the changelog and the ADR say why

## Steps
- [x] T1 · a prompt over the limit is split into parts that fit, each with its own paths, diff and part number   R1 · tree:c2438081ac37d3a6ab6f1c7248683615045a488a
- [x] T2 · the review runs one reviewer per part and the store merges them, refusing only a single file that cannot fit   R1 · tree:2cb69d8a4ca90a42497001b3fc636e127b117d34
- [x] T3 · the changelog and the ADR say why a big review is split, and the impact map names the split   R1 · tree:2c3601586ed9235e8dfbdb50daeba185d5818582

## History
L1 · "a review whose prompt does not fit the project's limit is split into parts that do, and one verdict answers the tree"
