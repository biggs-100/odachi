# the-guard-that

## What must exist
- the guard that decides whether an answer may merge into an existing verdict is a predicate the package pins, so asking the wrong question fails a test

## Done when
- a run that landed the lens merges its rows into that lens
- a run that landed nothing is refused, whatever shape the candidate has
- making the guard ask the candidate's shape again fails a test

## Not included (first version)
- no change to the proof format, and the stored proof is never rewritten
- no change to the bundled skills

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: por que un predicado pequeno se puede pinnear y un booleano cableado no.
