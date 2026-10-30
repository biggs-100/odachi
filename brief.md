# the-store-remembers

## What must exist
- the store remembers the skills scope: a plain odachi init honours the choice the project recorded, --skills changes it and says so in the output, and a project that never chose keeps the answer it always had

## Done when
- after odachi init --skills home, a plain odachi init writes no .agents/skills in the project
- after odachi init --skills project, a plain odachi init writes .agents/skills again
- a project that never chose keeps today's answer: a plain init writes the project copy
- the store names the choice in its output and in the file, and an older binary that opens it says upgrade instead of unknown key

## Not included (first version)
- no change to the proof format, and the stored proof is never rewritten
- no change to odachi pack, and no new verb: the scope is a key in the store and a flag init already has
- no change to which clients receive the native agent

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: como un archivo del clon guarda una decision y un flag la cambia sin que el siguiente init la borre.
