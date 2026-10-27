# pack-installs-the

## What must exist
- pack installs the skills of the project too, so a skill created with skill new reaches the folder the agent reads

## Done when
- odachi pack installs the bundled skills and the ones in ./skills, and a project skill wins over a bundled one with the same name, naming the replacement
- a skill created with odachi skill new reaches .agents/skills after pack
- a project without skills/ installs exactly the bundled set, as today, and --dry-run writes nothing
- the tests pin the merge, the clash and the empty project, and the impact map records it

## Not included (first version)
- skill new --scope home
- a registry or a check for the home
- changing init
- installing automatically when skill new runs

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: how a source and its copies stay one thing.
