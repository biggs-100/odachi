# a-global-install

## What must exist
- a global install leaves the project alone: odachi init --global writes the door in the home and no agent files in the project, and a plain init is the run that writes a project's own

## Done when
- odachi init --global in a project writes the home door and no .opencode/agents, and the store says project_agents = none
- odachi init without --global still writes the project's agent files and the store says write
- the flag that asked for what is now the default is refused, and the help names the two runs
- a test drives each one, and one of them fails without the change

## Not included (first version)
- no change to the verdict shape
- no change to the proof format, and the stored proof is never rewritten

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: how a default and the flag that asked for it stop living in the same place.
