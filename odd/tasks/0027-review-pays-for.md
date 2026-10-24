# review-pays-for

## Requirements (verbatim, from the user)
R1 · "review pays for itself: it runs when the tree asks, and it stops after one correction"

## Done when
- review --show names the lenses this tree asks for and why, without writing a verdict
- a documents-only tree says none is due, and the verb still exits blocked when there is no verdict
- the drive skill and the README state the gate and the one-correction cap
- the prompt, the verdict format and the store are untouched, and the suite stays green

## Steps
- [x] T1 · T1 - review --show names the lenses this tree asks for and why, without a verdict   R1 · 9bbc10a
- [x] T2 · T2 - the drive skill: the gate and one correction, one re-review   R1 · 5d91929
- [x] T3 · T3 - README: when a review pays, and when it does not   R1 · b5951b6

## History
L1 · "review pays for itself: it runs when the tree asks, and it stops after one correction"
L2 · OFICIO · "The reason names the path or the added lines, never a tier word" (porque a label a human cannot check does not help decide whether the review pays)
L3 · OFICIO · "A passive tree says none is due and still exits blocked" (porque --show answers whether a verdict exists, not whether one is required; scripts keep their meaning)
L4 · OFICIO · "One review at the end over the whole feature" (porque an intermediate tree is paid for and then invalidated by the next commit)
L5 · OFICIO · "The missing line says required, not "this tree asks for"" (porque --lens can add lenses the tree did not ask for, and the sentence must stay true)
