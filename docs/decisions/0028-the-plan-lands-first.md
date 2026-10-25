# 0028 - The plan lands first

Status: accepted · Date: 2026-10 · Session: after v0.11.9

## Context

`odachi start` writes three artifacts that belong together: `brief.md` (the
vision and the non-goals), `.odachi/fence.toml` (the policy derived from those
non-goals) and the task doc (the plan). The agent then builds on a branch, and
the first commits carry the code.

That gap cost real time. Three times a `git reset --hard origin/main` discarded
the plan because it was never committed: the brief and the fence of the feature
were in the worktree and nowhere else. Once it was worse: the brief of feature
0031 rode inside a fix commit while the fence stayed uncommitted, so `main`
ended with the new brief and the previous feature's fence. The close pull
request that brought the fence was then blocked by the rule of feature 0031
itself -a fence may only be loosened with new non-goals- and needed a one-time
admin merge to land. The gate was right; the workflow had split a pair that
belongs together.

## Decision

- **The plan is committed before the work starts.** `brief.md`,
  `.odachi/fence.toml` and the task doc go in one commit, before the branch.
- **`start` says it.** Its last line is now "commit the plan (brief, fence and
  task doc) before you build", so the instruction appears at the moment the
  files are written, not in a manual read an hour earlier.
- **The drive skill makes it step 2 of the feature cycle.** The order is: start,
  commit the plan, plan the steps, work, prove, commit each step, close.
- **The task doc travels with the plan.** It is the plan, not a log written
  after the fact; committing it early also makes the steps visible to a reviewer
  before the code exists.

## Alternatives rejected

- **Auto-commit from `start`.** The tool would decide what lands and on which
  branch; it does not know the contributor's conventions, and a tool that commits
  is a tool that can commit the wrong thing. The agent commits; the tool writes.
- **A gate that checks whether the plan is committed.** `check` reads the
  worktree, not the branch or the reflog: it cannot tell an uncommitted plan from
  a committed one, and a rule it cannot measure is a rule that will lie.
- **Leave it to memory.** The incidents above are the evidence: three resets and
  one split pair in two days.

## Consequences

- One more commit per feature, the cheapest one in the history: it is the plan,
  and it can be reverted alone.
- The brief and the fence can no longer drift apart between features, which is
  what feature 0031's rule needs to stay honest.
- A human who edits the brief or the fence later commits that edit with the
  change that motivated it; the plan commit is the first version, not the only
  one.
- The close commit still carries the task doc's final state (steps and history);
  the plan commit carries its first, which is what the work was built from.

## Sources

- Feature 0031: the split pair and the admin merge.
- The three `reset --hard` losses of an uncommitted brief and fence.
