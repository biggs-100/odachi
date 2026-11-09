# the-first-commit
Base: 1f963e60c32eab2fef84e142b123a5377baa1f3c

## Requirements (verbatim, from the user)
R1 · "a change that names itself in its commit can be committed: the hook that sees the message judges it, and the proof is owed after the commit exists"

## Done when
- `odachi init` installs a `commit-msg` hook written by odachi, idempotent, that leaves a foreign one alone and names it, the same rules the pre-commit hook follows; `odachi doctor` names a missing or foreign one
- a branch whose commit message carries `Odachi-Change: <line>` passes `git commit`: the gate reads the line from the message file, checks the fence against the worktree, the candidate the gate always reads, and defers the proof, which is owed by the tree that carries the commit
- the lane's refusals keep working from the message file: an empty line, two different lines, and a trailer on the default branch; the fork point is checked once the commit exists, in the commit lane
- a repository with no remote keeps the lane closed, the same rule the commit lane already follows
- a message with no trailer keeps today's answer, and the pre-commit hook keeps refusing a commit the gate has not cleared
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- Unreleased names the change: a contract change, minor

## Steps
- [x] T1 · init writes the gate at commit-msg, retires odachi's pre-commit, and doctor names the gate hook   R1 · route:inline · tree:8c82d488dbaae4466204c12540e68cafe3204318
- [x] T2 · the gate reads the trailer from the message file, defers the proof before the commit, and keeps the refusals   R1 · route:inline · tree:8c82d488dbaae4466204c12540e68cafe3204318
- [x] T3 · a real first commit on a branch passes and the proof is owed after it   R1 · route:inline · tree:8c82d488dbaae4466204c12540e68cafe3204318
- [x] T4 · Unreleased names the change as a contract change, minor   R1 · route:inline · tree:8c82d488dbaae4466204c12540e68cafe3204318

## History
L1 · "a change that names itself in its commit can be committed: the hook that sees the message judges it, and the proof is owed after the commit exists"
L2 · OFICIO · "the brief's promises name what the gate measures, not what it should" (porque the verifier found three sentences that promised more than the code: the fence reads the worktree candidate, the fork point is only reachable after the commit exists, and a repository with no remote closes the lane)
