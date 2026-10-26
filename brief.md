# the-proof-says

## What must exist
- the proof says which old files it cannot read, and keeps the line it can

## Done when
- a new proof whose output starts with the terminator keeps that line, pinned by a test
- the parser says which old files are ambiguous and why, with the date the terminator arrived
- the ledger row moves to accepted, with the gating effect named

## Not included (first version)
- a change of the proof format
- reading old proofs differently
- the review store
- the proofs on other machines

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: how to tell an accepted ambiguity from an unfixed bug.
