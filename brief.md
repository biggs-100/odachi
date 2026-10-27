# the-store-leaves

## What must exist
- the store leaves the project, and the clone's state lives in .git/odachi

## Done when
- reviews, backups and the lock live under .git/odachi and the worktree carries none of them
- .odachi no longer exists: fence, profile, review and roles live under odd, and agents.toml lives in the clone store
- reading never creates the store; only start, review and undo --collect do, under a lock
- one resolver reads the git common dir, so linked worktrees share the store and there is no second path
- the migration moves the existing ignored state byte-preserving, repeats without change, and doctor names what is left behind

## Not included (first version)
- proof/ does not move: the evidence of a close stays where the human reads it
- no env override for the clone store
- no retention or size cap for backups and reviews
- no sharing of state between clones of the same repository
- no change to how init writes the project agents

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: how a tool finds its own state without a key.
