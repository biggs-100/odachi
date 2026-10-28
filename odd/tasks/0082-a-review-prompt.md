# a-review-prompt
Base: 66bd37a0b4919a688ce1404dbc52db53e0f153b9

## Requirements (verbatim, from the user)
R1 · "a review prompt and its answer have an owner, so neither lands in the tree nor in a scroll"

## Done when
- odachi review writes the prompt into the clone's store beside the verdict and names the path, so an agent never improvises one
- odachi review --exec reads the answer from the {verdict} file the command was given, and stdout becomes diagnosis
- without {verdict} the answer is still read from stdout, so no existing reviewer breaks
- a test drives both capture paths and the prompt path, observed red first

## Steps
- [x] T1 · the printed review writes its prompt into the clone's store and names the path   R1 · tree:889c6b78fb562000d55417d13a1b0fb5754f745e
- [x] T2 · exec reads the answer from the {verdict} file the command was given, and stdout is diagnosis   R1 · tree:d1462551e86c3cf765c5104a783197b1b978b604
- [x] T3 · the help and the changelog say where the prompt and the answer live   R1
- [x] T4 · the corrections the review asked for: a quoted answer path, a separate folder, and no inheritance   R1

## History
L1 · "a review prompt and its answer have an owner, so neither lands in the tree nor in a scroll"
