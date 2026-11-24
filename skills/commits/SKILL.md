---
name: commits
description: Trigger: when deciding what goes into a commit, how to split the work, or what the message and PR description say.
---

## When

When a change is ready to commit, a diff cannot be reviewed in one sitting, or history no longer reads in order.

## Rules

- A commit is one task-doc step and one revertible thing; a message needing "and" names two commits.
- Every commit says what a revert would take with it: files or behavior that cannot be removed alone; otherwise it cannot revert alone.
- The title says what changed **for the person using the project**, imperatively; how and why belong in the body.
- Tests and documentation travel in the same commit, never a later "tests" or "docs" commit.
- A refactor and a behavior change do not share a commit: a break must identify which one caused it.
- Before committing: `odachi review --show`, a real `odachi review --exec "<command>"`, any required
  `odachi verify --exec "<command>"`, then `odachi check`; `odachi check --run "<command>"` writes the closing proof.
- With review mode on, a non-passive close blocks without a valid verdict, and a named
  risk blocks without independent verification. Mode off leaves both optional;
  `odachi check --no-review` bypasses one run (CI has no clone store and makes no model calls).
- Close the body with `Risk: none - <written reason>` or
  `Risk: data|security|contract|concurrency|delivery|no-test`. On non-passive code, missing, unknown or bare `none` is invalid.
- A PR's commits read **in order** and in twenty minutes; split it when they do not or step on each other.

## Risk by blast radius

The `Risk:` line decides the ceremony. None of the six skips the fence or proof or buys a waiver;
a waiver is the human's `start --waive-review` for a change touching none.

| Risk | What it covers | Lens | Verifier |
| --- | --- | --- | --- |
| data | schema, migration, stored records, backup | reliability | yes |
| security | permissions, secrets, personal data, input that decides whose data is read | risk | yes |
| contract | a shape someone else consumes: CLI, JSON, file format | reliability | yes |
| concurrency | shared state, locks, retries, background work | resilience | yes |
| delivery | the path a person installs or upgrades through | resilience | yes |
| no-test | the change cannot produce a test | reliability | yes |

The lens uses `odachi review --lens <lens>` and adds to the floor, never replaces it.
The verifier is `odachi-verify`, read-only: the whole task doc, a verdict per requirement with probes,
never against your summary. `odachi verify --exec "<command>"` writes its record;
with review mode on, a named risk cannot close without it.
Review approval never authorizes push, pull request, merge or release.

## Output

The commit and message: user-facing change, why, `Risk:` line, and any split that lets each commit stand alone.

## Never

- Commit with red tests "so the work is not lost".
- Mix formatting, reordering or renames with logic.
- A message that does not say what changed: "fix", "wip", "changes", "update".
- Commit somebody else's work that was already in the tree.
- One giant commit "because it is all together".
