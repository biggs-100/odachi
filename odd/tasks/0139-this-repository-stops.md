# this-repository-stops
Base: 773b12b4fa0532a4af2a6980e559c0d2bd78bc87

## Requirements (verbatim, from the user)
R1 · "this repository stops carrying its own copy of the craft, and the doctor refuses a block that names the folder the store did not choose"

## Done when
- the project's copy and .opencode/ leave the tree and AGENTS.md names ~/.agents/skills/
- odachi doctor refuses a block that names the folder the store did not choose, and names the command that fixes it
- a project that chose the project's copy keeps it, and the doctor accepts its block

## Steps
- [x] T1 · the doctor refuses a block whose skills line names the folder the store did not choose   R1 · tree:714fe71c8b0960575b8ebc59a17c2cae5da597e3
- [x] T2 · the doctor asks the same scope the init asks, and both variants stay pinned   R1 · tree:714fe71c8b0960575b8ebc59a17c2cae5da597e3
- [x] T3 · this repo drops .agents/skills and .opencode/, regenerates the block and says why   R1

## History
L1 · "this repository stops carrying its own copy of the craft, and the doctor refuses a block that names the folder the store did not choose"
L2 · OFICIO · "T1 and T2 land in one commit: the red test names the behaviour the fix adds" (porque a test that expects the doctor to refuse the wrong folder is red until the check exists, and a commit with red tests is never allowed; the same tree closes both steps)
L3 · OFICIO · "the first fix commit was made on main by mistake, so it was remade on feat/0139 and the ledger's tree rewritten" (porque the plan and the fix share one commit if the branch is not created before the work; the commit was unreferenced and never pushed, and the ledger must name a tree that exists)
