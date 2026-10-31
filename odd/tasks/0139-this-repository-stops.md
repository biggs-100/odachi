# this-repository-stops
Base: 773b12b4fa0532a4af2a6980e559c0d2bd78bc87

## Requirements (verbatim, from the user)
R1 · "this repository stops carrying its own copy of the craft, and the doctor refuses a block that names the folder the store did not choose"

## Done when
- the project's copy and .opencode/ leave the tree and AGENTS.md names ~/.agents/skills/
- odachi doctor refuses a block that names the folder the store did not choose, and names the command that fixes it
- a project that chose the project's copy keeps it, and the doctor accepts its block

## Steps
- [x] T1 · the doctor refuses a block whose skills line names the folder the store did not choose   R1 · tree:2548658a83c08c65d97cf3ddc5e1eaf639f4c809
- [x] T2 · the doctor asks the same scope the init asks, and both variants stay pinned   R1
- [ ] T3 · this repo drops .agents/skills and .opencode/, regenerates the block and says why   R1

## History
L1 · "this repository stops carrying its own copy of the craft, and the doctor refuses a block that names the folder the store did not choose"
L2 · OFICIO · "T1 and T2 land in one commit: the red test names the behaviour the fix adds" (porque a test that expects the doctor to refuse the wrong folder is red until the check exists, and a commit with red tests is never allowed; the same tree closes both steps)
