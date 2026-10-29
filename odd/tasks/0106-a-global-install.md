# a-global-install
Base: 2bbda63c3417ff6d94f923b91f4c39955c63f6e4

## Requirements (verbatim, from the user)
R1 · "a global install leaves the project alone: odachi init --global writes the door in the home and no agent files in the project, and a plain init is the run that writes a project's own"

## Done when
- odachi init --global in a project writes the home door and no .opencode/agents, and the store says project_agents = none
- odachi init without --global still writes the project's agent files and the store says write
- the flag that asked for what is now the default is refused, and the help names the two runs
- a test drives each one, and one of them fails without the change

## Steps
- [x] T1 · init --global leaves the project alone, and a plain init is the run that writes it   R1
- [x] T2 · the help and the changelog say which run writes what   R1 · tree:8da8af892eac13b2bccb4d9047d53719bce0dadd

## History
L1 · "a global install leaves the project alone: odachi init --global writes the door in the home and no agent files in the project, and a plain init is the run that writes a project's own"
