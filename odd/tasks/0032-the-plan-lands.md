# the-plan-lands

## Requirements (verbatim, from the user)
R1 · "the plan lands before the work: brief, fence and task doc in one commit"

## Done when
- odachi start ends by telling the agent to commit the plan before building
- the drive skill's cycle starts with the plan commit and says why the pair must not split
- a test pins the sentence and the suite stays green

## Steps
- [x] T1 · T1 - the start message says the plan lands first, pinned by a test   R1 · bc3acb7
- [x] T2 · T2 - the drive skill: the plan commit before the branch, and why   R1 · b86b978
- [x] T3 · T3 - ADR 0028: the split pair and the lost plan   R1 · d0e31b0

## History
L1 · "the plan lands before the work: brief, fence and task doc in one commit"
L2 · OFICIO · "The plan is committed before the branch" (porque a plan that does not land can be lost by a reset, and a split pair blocks the next change)
