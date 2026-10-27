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
- [ ] T1 · the clone store: one resolver reads the git common dir, and reviews, backups and the lock move to .git/odachi   R1
- [ ] T2 · the rules move: fence, profile, review and roles live under odd, and agents.toml lives in the clone store   R1
- [ ] T3 · the migration: store move takes the ignored state byte-preserving, repeats, and doctor names what is left   R1
- [ ] T4 · the map follows: docs, the AGENTS block and the drive skill say where the store lives, and this repo moves its own   R1

## History
L1 · "the store leaves the project, and the clone's state lives in .git/odachi"
L2 · OFICIO · "the ignored state moves first and the tracked files second" (porque each step leaves the repo green, and the store can be proven without touching what the human commits)
L3 · OFICIO · "the migration is its own verb, odachi store move, not a flag on init" (porque init is idempotent setup and moving evidence looks destructive; a verb that only migrates is provable on its own)
