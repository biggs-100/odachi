# the-proof-says
Base: bfcdb9f616a4a89ebba17d1e6ea14ea92ae3da7c

## Requirements (verbatim, from the user)
R1 · "the proof says which old files it cannot read, and keeps the line it can"

## Done when
- a new proof whose output starts with the terminator keeps that line, pinned by a test
- the parser says which old files are ambiguous and why, with the date the terminator arrived
- the ledger row moves to accepted, with the gating effect named

## Steps
- [x] T1 · the terminator line in the output survives the round trip, and the ambiguity is written where it is decided   R1 · 1b5f03c

## History
L1 · "the proof says which old files it cannot read, and keeps the line it can"
