# a-role-follows
Base: 7c6087a078a560dae5664bd18bfa6fac9fd8b3b5

## Requirements (verbatim, from the user)
R1 · "odachi role <slug> writes .agents/skills/<slug>/SKILL.md in the project, whatever the scope: in a project that just adopted odachi, that single skill makes derivesMinimal answer this project carries the craft, so the next plain init installs the whole curriculum there and rewrites the block to the project's copy ... Role should respect the scope the project has, which is a change of its own"

## Done when
- in a project whose store says skills = home, odachi role <slug> writes the skill under the home's .agents/skills and leaves the project without .agents
- in a project that adopted odachi with no stored choice, the same happens: role does not create the copy that flips the next init, and a plain init afterwards still writes the minimal footprint
- in a project whose craft lives in the project, by choice or by default, role writes .agents/skills/<slug>/SKILL.md as it always did
- odachi role, the listing, reads the same destination and says up to date
- the line the human reads names the folder the skill went to, ./.agents/skills or ~/.agents/skills
- a store, a home or a copy odachi cannot read is refused by name, never read as an absence

## Steps
- [x] T1 · the skills scope is decided by one function, and init asks it   R1 · tree:715906238526ace54ba674041c16bfb3a0438c93
- [x] T2 · a role's skill lands where the scope says, and the listing reads the same place   R1 · tree:22a149a07d18e3ced43c61cd21c85b73d3ab8916
- [x] T3 · the help, the README and the changelog say what a role follows   R1 · tree:86aaa1c0681cee95a431af64c3067d62bd37ca3b

## History
L1 · "odachi role <slug> writes .agents/skills/<slug>/SKILL.md in the project, whatever the scope: in a project that just adopted odachi, that single skill makes derivesMinimal answer this project carries the craft, so the next plain init installs the whole curriculum there and rewrites the block to the project's copy ... Role should respect the scope the project has, which is a change of its own"
L2 · OFICIO · "a role's skill follows the scope the project has, and the scope is decided by one function nobody reimplements" (porque a role writing into .agents/skills whatever the scope handed the adoption inference a copy odachi wrote itself, and the next plain init installed the whole curriculum in a project that keeps it in the machine)
L3 · OFICIO · "the review's WARNING is verified false and pinned by a test, not argued in prose" (porque init and role both stat .agents/skills, and a repository that keeps notes in .agents/ is read the same way by both; the test makes the claim impossible to re-raise without a red)
