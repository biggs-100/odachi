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
- [x] T1 · the store carries the skills scope, and its absence still means the project   R1 · tree:bc74c2e17c6e80b8c9c8ab6197fdd4af4d2c8eb8
- [x] T2 · a plain init honours the stored scope and names the choice it used   R1 · tree:0d8f4d7f4e5d0fdfadc7d4341e591aa6aac246ff
- [x] T3 · one test per acceptance line, and this repo takes the change   R1 · tree:2f55ef2990f30e2eaccb258dc1ca2cf6caca226a

## History
L1 · "the store remembers the skills scope: a plain odachi init honours the choice the project recorded, --skills changes it and says so in the output, and a project that never chose keeps the answer it always had"
L2 · OFICIO · "one schema counter serves the three strict files, and the store's key moved it to 2" (porque a per-file counter would let a file the binary cannot fully read pass a check that only looks at its own number; the cost -a verdict declares 2 without changing shape- is written where the counter lives)
L3 · OFICIO · "the Skills port takes the scope as an argument instead of a field" (porque only init can read the store, so a caller that fixed the scope in advance could not honour the choice the project recorded)
