# pack-installs-the
Base: 15c0c46c28cafc453c7f1592ae839a12c043a8e0

## Requirements (verbatim, from the user)
R1 · "pack installs the skills of the project too, so a skill created with skill new reaches the folder the agent reads"

## Done when
- odachi pack installs the bundled skills and the ones in ./skills, and a project skill wins over a bundled one with the same name, naming the replacement
- a skill created with odachi skill new reaches .agents/skills after pack
- a project without skills/ installs exactly the bundled set, as today, and --dry-run writes nothing
- the tests pin the merge, the clash and the empty project, and the impact map records it

## Steps
- [x] T1 · pack merges the project skills with the bundled ones, and the tests pin the clash and the empty project   R1 · tree:961830616c0889d5c2c632af5531c4fe8353fbdc

## History
L1 · "pack installs the skills of the project too, so a skill created with skill new reaches the folder the agent reads"
L2 · OFICIO · "a project skill wins over a bundled one with the same name, and the replacement is named" (porque the bundled skill is a starter and the project copy is the project decision; silently preferring either one hides which bytes the agent reads)
