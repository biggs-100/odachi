# installing-odachi-is

## Requirements (verbatim, from the user)
R1 · "installing odachi is enough: the agent sets up the project without the human typing commands"

## Done when
- a new odachi setup leaves the bootstrap skill in the home, the door for the clients it finds, and OpenCode's default agent, writing nothing in the current folder
- the installer runs setup after copying the binary, so the first conversation finds the agent ready
- the door and the bootstrap skill tell the agent to run init and start itself, asking permission before writing in a repository that is not the human's
- a machine already set up is left as it is: setup is idempotent
- the full suite stays green and the proof runs once per tree

## Steps
- [x] T1 · odachi setup writes the global half and nothing in the current folder   R1 · f406eca
- [x] T2 · the installer runs setup, and the door tells the agent to run the commands   R1 · 6adf484
- [x] T3 · docs: the install is the setup, and ADR 0021 records the reversal   R1 · 063d7fb

## History
L1 · "installing odachi is enough: the agent sets up the project without the human typing commands"
L2 · OFICIO · "odachi setup is a verb, not a flag on init" (porque the project half and the machine half are different jobs, and the installer needs one that writes nothing in the folder it runs in)
L3 · OFICIO · "Setup detects the clients in the home; --agents is the override" (porque an installer does not create another tool's config folder)
L4 · OFICIO · "The installer runs setup: running the installer is the consent" (porque that is the moment the human says put odachi on this machine; a plain init still leaves the home alone)
L5 · OFICIO · "The door and the bootstrap skill say the agent runs init and start" (porque the knowledge and the action belong together; asking the human to type was the defect)
L6 · OFICIO · "The permission rule stays: ask before init in a repository that is not the human's" (porque writing is the line, not typing; the human answers in words)
