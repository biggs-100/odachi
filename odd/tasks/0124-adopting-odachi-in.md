# adopting-odachi-in
Base: abbf3033d1f9219693a048a1858d42a8b7a1f3f0

## Requirements (verbatim, from the user)
R1 · "adopting odachi in a repository that already lives proposes the minimal footprint: the project keeps the rules and the hook, the craft stays in the machine, and the copies come back only when the human asks for them"

## Done when
- a plain odachi init in a repository with a commit and no stored choice writes AGENTS.md, .gitignore and the hook, no .agents/skills and no project agent, and says where the craft went
- a fresh repository keeps today's install: the craft in the project, the native agent in the project
- a project that already carries .agents/skills keeps it: the inference never moves a project that already has the craft
- --skills project and --agents put the copies back, are stored, and a later plain init honours them

## Steps
- [x] T1 · a repository that already lives gets the minimal footprint, inferred and not stored   R1 · tree:562beb2bf1ebfd6efc54e73e2d9be0fbec4ca792
- [ ] T2 · the docs and the help say what adoption writes, and what brings the copies back   R1

## History
L1 · "adopting odachi in a repository that already lives proposes the minimal footprint: the project keeps the rules and the hook, the craft stays in the machine, and the copies come back only when the human asks for them"
