# the-home-only
Base: 4d8bc9128b8c040c9c9f810c683ab6785bffd0eb

## Requirements (verbatim, from the user)
R1 · "the home-only choice survives a plain init, and the store is what remembers it"

## Done when
- a project that recorded no-project-agents answers no when asked to write them, and a plain odachi init writes no .opencode/agents
- the choice lives in the clone's store next to the client selection, and a human reads it there
- the flag changes the stored choice in one run, and the opposite flag puts the mirrors back
- doctor reads the stored choice and does not call a mirrorless project pending
- a fake port drives both answers in a test, observed red first

## Steps
- [ ] T1 · the store remembers the project-agents choice, and a plain init honors it while --agents writes the mirrors back   R1

## History
L1 · "the home-only choice survives a plain init, and the store is what remembers it"
