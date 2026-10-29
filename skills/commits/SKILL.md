---
name: commits
description: Trigger: when deciding what goes into a commit, how to split the work, or what the message and PR description say.
---

## When

When the change is ready and has to be committed, when a diff cannot be reviewed
in one sitting, or when the history no longer reads in order.

## Rules

- A commit is one task-doc step and one thing that can be reverted alone. If the
  message needs "and", they are two commits.
- The message says what changed **for the person using the project**, in one
  line, in the imperative. The how and why go in the body, not in the title.
- The tests and documentation of the change travel in the same commit. There is
  no later commit called "tests" or "docs".
- A refactor and a behavior change do not share a commit: if something breaks,
  you do not know which one it was.
- Before committing, the gate: `odachi check`. The proof is owed by the feature's
  close, not by every step: run `odachi check --run "<command>"` on the tree that
  closes it, and let CI judge the pull request that carries it.
- Close the body with `Risk: none` or
  `Risk: data|security|contract|concurrency|delivery|no-test`. A named risk asks
  for `odachi review` before the commit.
- A PR is a set of commits that read **in order**. If it does not read in twenty
  minutes, split it; if the commits step on each other, the order is wrong.

## Risk by blast radius

The `Risk:` line decides the ceremony. None of the six skips the fence or the
proof, and none allows a waiver: that is for a change touching none, recorded by
the human with `start --waive-review`.

| Risk | What it covers | Lens | Verifier |
| --- | --- | --- | --- |
| data | schema, migration, stored records, backup | reliability | yes |
| security | permissions, secrets, personal data, input that decides whose data is read | risk | yes |
| contract | a shape someone else consumes: CLI, JSON, file format | reliability | yes |
| concurrency | shared state, locks, retries, background work | resilience | yes |
| delivery | the path a person installs or upgrades through | resilience | yes |
| no-test | the change cannot produce a test | reliability | yes |

The lens is asked with `odachi review --lens <lens>` and is added to the floor's,
never instead. The verifier is `odachi-verify`, read-only: the whole task doc, a
verdict per requirement with probes, never against your summary.

## Never

- Commit with red tests "so the work is not lost".
- Mix formatting, reordering or renames with logic.
- A message that does not say what changed: "fix", "wip", "changes", "update".
- Commit somebody else's work that was already in the tree.
- One giant commit "because it is all together".
