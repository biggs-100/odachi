# a-proof-is
Base: 83b56fde34cf9187e9ca5ade21b538bd5bcd8c63

## Requirements (verbatim, from the user)
R1 · "a proof is owed by the feature that closes, not by every step that builds it"

## Done when
- while the feature is open, check says it is ready for a step and names the proof as owed at the close
- when the feature closes, check demands the proof for that tree and refuses a missing or stale one
- a proof taken with --run mid-feature is still written and reported, and a stale one is a note
- the hook needs no change, and a test drives each answer

## Steps
- [x] T1 · an open feature does not owe the proof, and the run says why   R1
- [x] T2 · the closing tree still owes it, and a stale proof is named   R1

## History
L1 · "a proof is owed by the feature that closes, not by every step that builds it"
