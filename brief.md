# installing-odachi-is

## What must exist
- installing odachi is enough: the agent sets up the project without the human typing commands

## Done when
- a new odachi setup leaves the bootstrap skill in the home, the door for the clients it finds, and OpenCode's default agent, writing nothing in the current folder
- the installer runs setup after copying the binary, so the first conversation finds the agent ready
- the door and the bootstrap skill tell the agent to run init and start itself, asking permission before writing in a repository that is not the human's
- a machine already set up is left as it is: setup is idempotent
- the full suite stays green and the proof runs once per tree

## Not included (first version)
- the project half of init: AGENTS.md, hook, .gitignore and project skills
- installing a door for a client that is not on the machine
- an interactive installer
- changing the interview or the gates
- a new dependency
- the release: a separate decision

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: no rush.
- Wants to learn: how an agent takes over the setup without owning the human's folder.
