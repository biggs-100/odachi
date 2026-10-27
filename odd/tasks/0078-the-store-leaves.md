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
- [ ] T2 · the backups leave: every verb that saves a version writes into the store, and undo --collect adopts the old folder   R1
- [ ] T3 · the verdicts leave: review reads and writes in the store, and the ignore machinery retires from init, doctor and tier   R1
- [ ] T4 · the rules move: fence, profile, review and roles live under odd, and agents.toml lives in the clone store   R1
- [ ] T5 · the leftovers: doctor names what an old project still carries, and init prints the moves git has to make   R1
- [ ] T6 · the map follows: docs, the AGENTS block and the drive skill say where the store lives, and this repo moves its own   R1

## History
L1 · "the store leaves the project, and the clone's state lives in .git/odachi"
L2 · OFICIO · "the ignored state moves first and the tracked files second" (porque each step leaves the repo green, and the store can be proven without touching what the human commits)
L3 · OFICIO · "the migration is its own verb, odachi store move, not a flag on init" (porque init is idempotent setup and moving evidence looks destructive; a verb that only migrates is provable on its own)
