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

## History
L1 · "a review prompt and its answer have an owner, so neither lands in the tree nor in a scroll"
