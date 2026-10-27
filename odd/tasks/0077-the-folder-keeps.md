# the-folder-keeps
Base: 1e89b88693088aabc55d5b636faaf0a7cb24e28b

## Requirements (verbatim, from the user)
R1 · "the folder keeps only the rules and the plan, and the agent files of a project can live in the home"

## Done when
- init --global --no-project-agents writes no agent file in the project, and the flag without --global is refused naming the fix
- start and level do not recreate the project agents when they are absent
- doctor accepts the home as the source, and an empty project agents folder is the clean mode
- the tests pin the flag, the refusal, the no-recreate and the doctor note, and the docs record it

## Steps
- [ ] T1 · init gains the flag: no project agents, refused without --global, and start and level do not recreate them   R1
- [ ] T2 · doctor accepts the home as the source, and the docs record the clean mode   R1

## History
L1 · "the folder keeps only the rules and the plan, and the agent files of a project can live in the home"
L2 · OFICIO · "the flag is --no-project-agents and it requires --global" (porque skipping the project agents without a door leaves no agent at all, and a flag that can leave you with nothing must name what it needs)
