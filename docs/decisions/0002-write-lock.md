# 0002 · The write lock

Status: accepted · Date: 2026-10 · Session: T17

## Context

Every artifact is written atomically, so no file can be half-written. What atomic
writes do not protect is the *set*: two runs that interleave can leave the
`brief.md` of one beside the `.odachi/fence.toml` of the other. Every file is
valid, and the project is incoherent — the one failure mode atomic writes cannot
see. `start` writes four files and `undo` writes an artifact plus the backup of
what it replaced.

`plan`, `ask` and `level --to` write a single file each, atomically. A lock there
would protect nothing and would only add a way to be annoyed by your own tool.

## Decision

- One lock file, `.odachi/lock`, created with `O_CREATE|O_EXCL` and written by the
  run that owns it: one line with the verb, the pid and the instant. Exclusive
  create is the whole mutual exclusion — there is no daemon, no socket, no state
  beyond one file a human can delete.
- Only `start` and `undo` take it, before their first read that decides anything,
  and release it with `defer` so a run that stops on its own rules releases it too.
- **Ten minutes is the staleness boundary.** Younger than that, the other run is
  writing and the answer is "wait and try again": exit 1, naming the verb, the pid,
  the age and the file. odachi never deletes a live lock. Older than that, the run
  that wrote it is gone; the lock is taken over (remove, then exclusive create, so
  two runs cannot both take the same old lock) and the takeover is announced in one
  line.
- **A lock is never ignored.** Content that cannot be read does not unblock
  anything: the modification time still decides, and a fresh unreadable lock blocks
  with a message that names the file. Ignoring it "because it looks broken" is how
  the incoherent set gets written.
- The port lives in `internal/app/lock` and the domain takes no part: there is no
  business rule here, only exclusion. The adapter (in `cmd/odachi`) owns `O_EXCL`,
  the modification time and the removal; the messages and the staleness constant
  stay in the app package, where every other Spanish line lives.
- Release removes the file **only if it still holds what this run wrote**. A run
  that outlives ten minutes can have its lock taken over, and deleting the new
  owner's lock would break the rule above.

## Alternatives rejected

- **Lock the individual file writes.** They are already atomic; the failure is
  between files, not inside one.
- **A lock in every verb.** `plan`, `ask` and `level --to` write one file. A lock
  that protects nothing is friction, and friction is what makes people work around
  a tool.
- **A timeout that waits for the lock.** Waiting turns a two-second mistake into a
  hang with no output, and the human cannot tell "working" from "stuck". Refusing
  with the pid and the age is information; waiting is not.
- **A pid liveness check.** Portable process inspection does not exist: a pid can
  be reused, and the check is a different answer on every platform. The age of the
  file is the same fact everywhere.
- **Deleting a lock that looks abandoned.** "Looks abandoned" is exactly the guess
  that would let two runs write at once when it is wrong.
