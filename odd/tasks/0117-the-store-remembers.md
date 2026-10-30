# the-store-remembers
Base: 7e496ea17e4554fe4a561380f6a89287386ef153

## Requirements (verbatim, from the user)
R1 · "the store remembers the skills scope: a plain odachi init honours the choice the project recorded, --skills changes it and says so in the output, and a project that never chose keeps the answer it always had"

## Done when
- after odachi init --skills home, a plain odachi init writes no .agents/skills in the project
- after odachi init --skills project, a plain odachi init writes .agents/skills again
- a project that never chose keeps today's answer: a plain init writes the project copy
- the store names the choice in its output and in the file, and an older binary that opens it says upgrade instead of unknown key

## Steps
- [ ] T1 · the store carries the skills scope, and its absence still means the project   R1
- [ ] T2 · a plain init honours the stored scope and names the choice it used   R1
- [ ] T3 · one test per acceptance line, and this repo takes the change   R1

## History
L1 · "the store remembers the skills scope: a plain odachi init honours the choice the project recorded, --skills changes it and says so in the output, and a project that never chose keeps the answer it always had"
