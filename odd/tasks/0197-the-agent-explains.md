# the-agent-explains
Base: bf6451f60e8aca0b3c49174ab62cfed48e3107ce

## Requirements (verbatim, from the user)
R1 · "the agent explains as it works: every decision and every question teaches the human the craft"

## Done when
- the novato voice explains the why of every decision in a short paragraph of 2 to 4 lines, and names and defines the concept the first time it appears
- the senior voice leads with the decision and still explains its why and its cost in 2 or 3 lines, defining any concept the human has not seen; the two voices stay distinguishable
- a question the agent asks says what the answer changes and what each option costs, never a bare question
- odachi start prints a why of 2 to 3 lines per interview question, saying what the answer decides
- the anti-sermon guard stays: no repetition, no glossary, never two explanations in a row
- the AGENTS.md block and the native agent files carry the same bytes, pinned by the existing test
- go test ./... is green

## Steps
- [ ] T1 · the novato and senior voices explain in a short paragraph and define each concept the first time   R1 · route:inline
- [ ] T2 · the interview asks each question with a why of 2 to 3 lines that says what the answer decides   R1 · route:inline

## History
L1 · "the agent explains as it works: every decision and every question teaches the human the craft"
