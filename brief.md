# the-first-commit

## What must exist
- a change that names itself in its commit can be committed: the hook that sees the message judges it, and the proof is owed after the commit exists

## Done when
- `odachi init` installs the gate at `commit-msg`, the hook that sees the message, idempotent, that leaves a foreign `commit-msg` alone and names it, and retires odachi's own `pre-commit` hook while leaving one a project wrote untouched; `odachi doctor` names a missing or foreign gate hook
- a branch whose commit message carries `Odachi-Change: <line>` passes `git commit`: the gate reads the line from the message file, checks the fence against the worktree, the candidate the gate always reads, and defers the proof, which is owed by the tree that carries the commit
- the lane's refusals keep working from the message file: an empty line, two different lines, and a trailer on the default branch; the fork point is checked once the commit exists, in the commit lane
- a repository with no remote keeps the lane closed, the same rule the commit lane already follows
- a message with no trailer keeps today's answer: the gate refuses a commit whose tree has no proof, exactly as the `pre-commit` hook did
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- Unreleased names the change: a contract change, minor

## Not included (first version)
- no change to the bundled skill texts: the 19 keep their words — back when the lane needs a sentence they do not already say; 0245 needs none
- no change to the review or the classifier — back when a verdict has to read something the diff cannot show
- no change to odachi bench and no new verb — back when a bench journey crosses a commit, or a human asks for a verb
- no change to the stored proofs: never touched — never; they are the evidence
- no change to the design canvas, its token or the installers — back when an install path changes; the hook 0245 adds is written by `init`, not by the installers
- no change to the release workflow — back when a release must run a command the workflow does not run today

## Decisions taken
- It is used from the computer, by a person committing a one-line change.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: whether a second hook pays for its own failure mode.
- The line is read where git keeps it: `pre-commit` runs before the message file exists (probed), and `commit-msg` receives its path.
