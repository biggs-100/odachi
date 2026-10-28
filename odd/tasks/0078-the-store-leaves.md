# the-store-leaves
Base: 70d3bf54ab8d7dbb7665f9f235c8606f4fcd2909

## Requirements (verbatim, from the user)
R1 · "the store leaves the project, and the clone's state lives in .git/odachi"

## Done when
- reviews, backups and the lock live under .git/odachi and the worktree carries none of them
- .odachi no longer exists: fence, profile, review and roles live under odd, and agents.toml lives in the clone store
- reading never creates the store; only start, review and undo --collect do, under a lock
- one resolver reads the git common dir, so linked worktrees share the store and there is no second path
- the migration moves the existing ignored state byte-preserving, repeats without change, and doctor names what is left behind

## Steps
- [x] T1 · the resolver and the lock: gitx names the common dir, store names the root, and the lock moves into it   R1 · tree:a22f874b6127150ee5fa0c65388175aa2571a4e3
- [x] T2 · the backups leave: every verb that saves a version writes into the store, and undo --collect adopts the old folder   R1 · tree:9b73aac950b21b37f0a2a766029431819bb57c3d
- [x] T3 · the verdicts leave: review reads and writes in the store, and the ignore machinery retires from init, doctor and tier   R1 · tree:82137011baa034d5e60c3efc8ce8d4d8479b8fc2
- [x] T4 · the rules move: fence, profile, review and roles live under odd, and agents.toml lives in the clone store   R1 · tree:31cde8be49c0561877bb89099296ebb40bc41f69
- [ ] T5 · the leftovers: doctor names what an old project still carries, and init prints the moves git has to make   R1
- [ ] T6 · the map follows: docs, the AGENTS block and the drive skill say where the store lives, and this repo moves its own   R1

## History
L1 · "the store leaves the project, and the clone's state lives in .git/odachi"
L2 · OFICIO · "the ignored state moves first and the tracked files second" (porque each step leaves the repo green, and the store can be proven without touching what the human commits)
L3 · OFICIO · "the migration is its own verb, odachi store move, not a flag on init" (porque init is idempotent setup and moving evidence looks destructive; a verb that only migrates is provable on its own)
L4 · OFICIO · "the first step is cut in six: eight verbs write backups through one adapter, and the verdict store drags the ignore machinery" (porque a step is sized by the outcomes it changes, not by the files it touches; three stores in one commit would hide the review change inside a path change)
L5 · OFICIO · "the store is injected in the files adapter, and every verb gets it from one disk() helper" (porque eight verbs write backups through one port, and a path each verb computes on its own is eight chances to drift)
L6 · OFICIO · "discovery reads the store, the old project store and the side of the artifact" (porque losing a version during a migration is worse than reading one extra directory)
L7 · OFICIO · "the verdict's message names the absolute path in check and the project-relative one in review" (porque check may run from any directory and the human has to be able to delete the file; review's --attach takes a path from the project root)
L8 · OFICIO · "the rules live under odd/, next to the plan, and the store keeps the client choice" (porque the fence and the profile are the team's decisions and travel with the repository; which clients receive the agent is the person's, so it does not)
L9 · OFICIO · "store move adopts the evidence and prints the git mv lines; it never moves a tracked file" (porque a tracked file belongs to git and to the human's commit, and a migration that rewrites history behind their back is not a migration)
L10 · OFICIO · "a leftover .odachi/ fails doctor only when the rules inside it are the ones not being read" (porque an old copy next to the new one is dead weight, but a fence nobody enforces is worse than no fence: the human believes it is there)
