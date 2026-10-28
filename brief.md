# the-home-only

## What must exist
- the home-only choice survives a plain init, and the store is what remembers it

## Done when
- a project that recorded no-project-agents answers no when asked to write them, and a plain odachi init writes no .opencode/agents
- the choice lives in the clone's store next to the client selection, and a human reads it there
- the flag changes the stored choice in one run, and the opposite flag puts the mirrors back
- doctor reads the stored choice and does not call a mirrorless project pending
- a fake port drives both answers in a test, observed red first

## Not included (first version)
- no change to where the store lives
- no new verb and no new flag
- init never deletes files it did not write

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: how a decision survives a second run: the difference between a flag and a stored choice.
